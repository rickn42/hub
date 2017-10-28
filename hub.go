package hub

import (
	"sync"
)

type Port struct {
	Connector
	SignalChan chan Signal
}

type Signal struct {
	Done  chan struct{}
	From  Connector
	Value interface{}
}

type hub struct {
	sig   chan Signal
	mu    sync.RWMutex
	ports map[Connector]Port
}

func (h *hub) PlugIn(c Connector) {
	h.mu.Lock()

	port := Port{
		Connector:  c,
		SignalChan: make(chan Signal),
	}

	h.ports[c] = port

	h.mu.Unlock()

	go func() {
		in := c.InC()
		for value := range in {
			done := make(chan struct{})
			h.sig <- Signal{
				From:  c,
				Value: value,
				Done:  done,
			}
			<-done
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
		sig:   make(chan Signal),
		ports: make(map[Connector]Port),
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
}

func (h *hub) propagate(sig Signal) {
	h.mu.RLock()

	for c, port := range h.ports {
		if c == sig.From {
			continue
		}

		port.OutC() <- sig.Value
	}

	close(sig.Done)

	h.mu.RUnlock()
}
