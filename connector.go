package hub

type connector struct {
	in, out chan interface{}
}

func (c *connector) InC() chan interface{} {
	return c.in
}

func (c *connector) OutC() chan interface{} {
	return c.out
}

func NewChannelWrapperConnector(in, out chan interface{}) *connector {
	return &connector{in, out}
}

func NewBufferedConnector(bufSize int) *connector {
	if bufSize < 1 {
		bufSize = 1
	}
	return &connector{
		make(chan interface{}, bufSize),
		make(chan interface{}, bufSize),
	}
}
