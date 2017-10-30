package hub

import (
	"sync"
)

type port struct {
	Connector
	filters    []Filter
	once       sync.Once
	plugOut    chan struct{}
	pluggedOut chan struct{}
}

func newPort(c Connector, filters ...Filter) *port {
	return &port{
		Connector:  c,
		filters:    filters,
		plugOut:    make(chan struct{}),
		pluggedOut: make(chan struct{}),
	}
}

func (p *port) donePlugOut() {
	c := p.plugOut
	p.once.Do(func() {
		close(c)
		return
	})
}
