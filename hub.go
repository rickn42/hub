package hub

import (
	"errors"
	"sync"
)

// signal is a inner message structure in hub.
type signal struct {
	from Connector
	// message passing done
	done    chan struct{}
	message interface{}
}

// hub is implementation of Hub interface.
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

	startHubGoroutine(h)

	return h
}

func startHubGoroutine(h *hub) {
	// It will return when sig-channel closed.
	go func() {
		for signal := range h.sig {
			select {
			case <-h.destroy:
				// Suck all signal out until sig channel is closed.
				continue
			default:
			}

			h.broadcast(signal)
		}
	}()
}

func (h *hub) PlugIn(c Connector, filters ...Filter) error {
	select {
	case <-h.destroy:
		return errors.New("the hub was destroyed")
	default:
	}

	p := newPort(c, filters)
	h.mu.Lock()
	h.ports[c] = p
	h.mu.Unlock()

	startConnectorGoroutine(h, p)
	return nil
}

func startConnectorGoroutine(h *hub, p *port) {
	// It will return when plugging out.
	go func() {
		defer func() {
			close(p.pluggedOut)
		}()

		c := p.Connector
		in := c.InC()

		if in == nil {
			<-p.plugOut // nil channel. Just wait for plugging out.
			return
		}

		for {
			select {
			case <-p.plugOut: // Just wait for plugging out.
				return

			case msg, ok := <-in:
				if !ok {
					// closed channel. Just wait plugging out.
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
	}()
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
		return
	}
	delete(h.ports, c)

	p.notifyPlugOut()
	return p.pluggedOut
}

// broadcast send signal to all connector.
func (h *hub) broadcast(sig signal) {
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
	h.notifyDestroy()
	var ds []chan struct{}
	for c := range h.ports {
		ds = append(ds, h.plugOut(c))
	}
	h.mu.Unlock()

	for _, done := range ds {
		<-done
	}

	h.doDestroy()
}

// notifyDestroy close destroy-channel for notifying hub goroutine to vanish all signals.
func (h *hub) notifyDestroy() {
	c := h.destroy
	h.once.Do(func() {
		close(c)
	})
}

// doDestroy close sig-channel so terminate hub goroutine.
func (h *hub) doDestroy() {
	c := h.sig
	h.onceD.Do(func() {
		close(c)
	})
}
