package hub

// Hub is collection of Connectors. It mediates channel values.
type Hub interface {
	PlugIn(c Connector, filters ...Filter) error
	PlugOut(c Connector)
	Destory()
}

// Connector is a pair of channels.
type Connector interface {
	// InC returns an input channel which data get into the hub.
	InC() chan interface{}
	// OutC returns an output channel which data get out from the hub.
	OutC() chan interface{}
	// TryAndPass is signed if pass or not when output channel not ready.
	TryAndPass() bool
}

// Filter is a alias function that change value or vanish value
// If ok is false, that value is vanished.
type Filter = func(old interface{}) (new interface{}, ok bool)

var FilterNothing = func(interface{}) (_ interface{}, ok bool) {
	return
}
