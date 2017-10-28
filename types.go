package hub

type Hub interface {
	PlugIn(c Connector)
	PlugOut(c Connector)
	Destory()
}

type Connector interface {
	InC() chan interface{}
	OutC() chan interface{}
}
