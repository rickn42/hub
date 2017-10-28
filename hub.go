package hub

import (
	"sync"
)

type Port struct {
	Connector
	SignalChan chan Signal
	Filters    []Filter
}

type Signal struct {
	Done  chan struct{}
	From  Connector
	Value interface{}
}

type hub struct {
	sig       chan Signal
	destoryed chan struct{}
	mu        sync.RWMutex
	ports     map[Connector]Port
}

func (h *hub) PlugIn(c Connector, filters ...Filter) {
	h.mu.Lock()

	port := Port{
		Connector:  c,
		SignalChan: make(chan Signal),
		Filters:    filters,
	}

	h.ports[c] = port

	h.mu.Unlock()

	go func() {
		in := c.InC()
		for {
			select {
			case value := <-in:
				done := make(chan struct{})
				h.sig <- Signal{
					From:  c,
					Value: value,
					Done:  done,
				}
				<-done
			case <-h.destoryed:
				return
			}
		}
	}()
}

func (h *hub) PlugOut(c Connector) {
	h.mu.Lock()
	defer h.mu.Unlock()

	port, ok := h.ports[c]
	if !ok {
		return
	}

	delete(h.ports, c)

	close(port.SignalChan)
}

func NewHub() Hub {
	h := &hub{
		sig:       make(chan Signal),
		destoryed: make(chan struct{}),
		ports:     make(map[Connector]Port),
	}

	go func() {
		for signal := range h.sig {
			h.propagate(signal)
		}
	}()

	return h
}

func (h *hub) Destory() {
	close(h.sig)
	close(h.destoryed)
}

func (h *hub) propagate(sig Signal) {
	h.mu.RLock()

	for c, port := range h.ports {
		if c == sig.From {
			continue
		}

		value := sig.Value
		ok := true
		for _, filter := range port.Filters {
			value, ok = filter(value)
			if !ok {
				break
			}
		}
		if !ok {
			continue
		}

		if port.TryAndPass() {
			select {
			case port.OutC() <- value:
			default:
			}
		} else {
			port.OutC() <- value
		}
	}

	close(sig.Done)

	h.mu.RUnlock()
}
