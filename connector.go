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

// TryAndPass default is false. So default message passing type is sync.
func (c *connector) TryAndPass() (ok bool) {
	return false
}

func NewConnectorWithChannels(in chan interface{}, out chan interface{}) *connector {
	return &connector{in: in, out: out}
}

func NewBufferedConnector(bufSize int) *connector {
	return &connector{
		in:  make(chan interface{}, bufSize),
		out: make(chan interface{}, bufSize),
	}
}

// tryAndPassWrapper wrap Connector with TryAndPass() always true method.
type tryAndPassWrapper struct {
	Connector
}

func (c *tryAndPassWrapper) TryAndPass() (ok bool) {
	return true
}

func WrapConnectorWithTryAndPass(c Connector) Connector {
	return &tryAndPassWrapper{
		Connector: c,
	}
}
