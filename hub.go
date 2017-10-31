package hub

import (
	"sync"
)

type signal struct {
	from Connector
	// message passing done
	done    chan struct{}
	message interface{}
}

type hub struct {
	// notify hub destroy signal
	once    sync.Once
	destroy chan struct{}

	// destroy hub
	onceD sync.Once
	sig   chan signal

	mu    sync.RWMutex
	ports map[Connector]*port
}

func NewHub() Hub {
	h := &hub{
		sig:     make(chan signal),
		destroy: make(chan struct{}),
		ports:   make(map[Connector]*port),
	}

	go func() {
		for signal := range h.sig {
			select {
			case <-h.destroy:
				// Continue loop and suck all messages.
				continue
			default:
			}

			h.propagate(signal)
		}
	}()

	return h
}

func (h *hub) PlugIn(c Connector, filters ...Filter) {
	select {
	case <-h.destroy:
		panic("The hub was destoryed")
	default:
	}

	p := newPort(c, filters...)

	h.mu.Lock()
	h.ports[c] = p
	h.mu.Unlock()

	go func(h *hub, p *port) {
		defer func() {
			close(p.pluggedOut)
		}()

		c := p.Connector

		in := c.InC()
		if in == nil {
			<-p.plugOut
			return
		}

		for {
			select {
			case <-p.plugOut: // Check if the connector plugged out.
				return

			case msg := <-in:
				if msg == nil {
					// If nil msg received, consider as input channel closed.
					<-p.plugOut
					return
				}

				done := make(chan struct{})
				h.sig <- signal{
					from:    c,
					message: msg,
					done:    done,
				}
				<-done
			}
		}
	}(h, p)
}
func (h *hub) PlugOut(c Connector) {
	h.mu.Lock()
	done := h.plugOut(c)
	h.mu.Unlock()

	<-done
}

func (h *hub) plugOut(c Connector) (done chan struct{}) {

	p, ok := h.ports[c]
	if !ok {
		h.mu.Unlock()
		return
	}
	delete(h.ports, c)

	p.donePlugOut()

	return p.pluggedOut
}

func (h *hub) propagate(sig signal) {
	h.mu.RLock()

	from := sig.from

	for c, port := range h.ports {
		if c == from {
			continue
		}

		if port.OutC() == nil {
			continue
		}

		msg := sig.message
		ok := true

		for _, filter := range port.filters {
			msg, ok = filter(msg)
			if !ok {
				break
			}
		}
		if !ok {
			continue
		}

		if port.TryAndPass() {
			select {
			case port.OutC() <- msg:
			default:
			}
		} else {
			port.OutC() <- msg
		}
	}

	close(sig.done)

	h.mu.RUnlock()
}

func (h *hub) Destory() {
	h.mu.Lock()

	h.notiDestory()

	var ds []chan struct{}
	for c := range h.ports {
		ds = append(ds, h.plugOut(c))
	}

	h.mu.Unlock()

	for _, done := range ds {
		<-done
	}

	h.doDestory()
}

func (h *hub) notiDestory() {
	c := h.destroy
	h.once.Do(func() {
		close(c)
	})
}

func (h *hub) doDestory() {
	c := h.sig
	h.onceD.Do(func() {
		close(c)
	})
}
