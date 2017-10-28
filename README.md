# Hub

Hub get channel connectors. And connectors got signal each other.

### Basic

`Connector` has 2 channel. 

`InC()` return input channel which data get into it.

`OutC()` return output which data get out. 


```go
type Connector interface {
	InC() chan interface{}
	OutC() chan interface{}
}
```

### How to use 

```go
h := hub.NewHub()
h.PlugIn(someConnector)
h.PlugIn(anotherConnector)
h.PlugIn(andAnotherConnector)

// Input value propagate all other connectors.
someConnector.InC() <- 1
someConnector.InC() <- 2
someConnector.InC() <- 3

// Hub guarantees value order correct.
<- anotherConnector.OutC() // 1
<- anotherConnector.OutC() // 2
<- anotherConnector.OutC() // 3

<- andAnotherConnector.OutC() // 1
<- andAnotherConnector.OutC() // 2
<- andAnotherConnector.OutC() // 3
```


### Filter 

```go
h := hub.NewHub()
h.PlugIn(someConnector)
h.PlugIn(anotherConnector, func(o interface{}) (n interface{}, ok bool) {
	if o.(int) % 3 == 0 {
		return o, true
	} 
	return nil, false 
})

someConnector.InC() <- 2
someConnector.InC() <- 3
someConnector.InC() <- 4

// only got 3 
<-anotherCennector.OutC()
```

