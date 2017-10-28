package hub

import (
	"sync"
)

type port struct {
	Connector
	signalChan chan signal
	filters    []Filter
}

type signal struct {
	done  chan struct{}
	from  Connector
	value interface{}
}

type hub struct {
	sig     chan signal
	destroy chan struct{}
	mu      sync.RWMutex
	ports   map[Connector]port
}

func NewHub() Hub {
	h := &hub{
		sig:     make(chan signal),
		destroy: make(chan struct{}),
		ports:   make(map[Connector]port),
	}

	go func() {
		for signal := range h.sig {
			select {
			case <-h.destroy:
				continue
			default:
			}

			h.propagate(signal)
		}
	}()

	return h
}

func (h *hub) PlugIn(c Connector, filters ...Filter) {
	h.mu.Lock()

	port := port{
		Connector:  c,
		signalChan: make(chan signal),
		filters:    filters,
	}

	h.ports[c] = port

	h.mu.Unlock()

	go func() {
		in := c.InC()
		for {
			select {
			case <-h.destroy:
				h.PlugOut(c)
				return
			default:
			}

			value := <-in
			done := make(chan struct{})
			h.sig <- signal{
				from:  c,
				value: value,
				done:  done,
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
	close(port.signalChan)

	if len(h.ports) == 0 {
		select {
		case <-h.destroy:
			close(h.sig)
		default:
		}
	}
}

func (h *hub) propagate(sig signal) {
	h.mu.RLock()

	for c, port := range h.ports {
		if c == sig.from {
			continue
		}

		value := sig.value
		ok := true
		for _, filter := range port.filters {
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

	close(sig.done)

	h.mu.RUnlock()
}

func (h *hub) Destory() {
	h.mu.Lock()
	defer h.mu.Unlock()

	close(h.destroy)
	if len(h.ports) == 0 {
		close(h.sig)
	}
}
