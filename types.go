package hub

type Hub interface {
	PlugIn(c Connector, filters ...Filter)
	PlugOut(c Connector)
	Destory()
}

type Connector interface {
	InC() chan interface{}
	OutC() chan interface{}
}

type Filter = func(old interface{}) (new interface{}, ok bool)

var FilterNothing = func(interface{}) (_ interface{}, ok bool) {
	return
}
