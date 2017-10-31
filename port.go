package hub

import (
	"sync"
)

// port is a wrapper of Connector with stuff for hub using.
type port struct {
	Connector
	filters    []Filter
	once       sync.Once
	plugOut    chan struct{}
	pluggedOut chan struct{}
}

func newPort(c Connector, filters []Filter) *port {
	return &port{
		Connector:  c,
		filters:    filters,
		plugOut:    make(chan struct{}),
		pluggedOut: make(chan struct{}),
	}
}

// notifyPlugOut close plugOut channel.
// So notify to Connector goroutine.
func (p *port) notifyPlugOut() {
	c := p.plugOut
	p.once.Do(func() {
		close(c)
	})
}
