# Hub

Hub get channel connectors. And connectors got signal each other.

### Basic

`Connector` has 2 channel. 

`InC()` return input channel which data get into it.

`OutC()` return output which data get out. 


```go
type Connector interface {
	// InC returns an input channel which data get into the hub.
    InC() chan interface{}
    // OutC returns an output channel which data get out from the hub.
    OutC() chan interface{}
    // TryAndPass is signed if pass or not when output channel not ready.
    TryAndPass() bool
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

### try and pass 

The value passing in hub is working synchronizely.

So if other connector did not use `OutC()` channel, the whole hub is blocked until ready.

Use `TryAndPass() = true`.

```go
// channel buffer size is 0
someConnector := hub.NewBufferedConnector(0))
zeroBufferConnector := hub.NewBufferedConnector(0)

h := hub.NewHub()
h.PlugIn(someConnector)
h.PlugIn(zeroBufferConnector)

// It is blocked! anotherConnector's output channel is not ready!
someConnector.InC() <- 1
```

```go
h := hub.NewHub()
h.PlugIn(someConnector)
// This wrapper just has TryAndPass()=true method.  
h.PlugIn(hub.WrapConnectorWithTryAndPass(zeroBufferConnector))

// Input channel is working. The zeroBufferConnector is just passed.
someConnector.InC() <- 1
someConnector.InC() <- 2
someConnector.InC() <- 3

// waiting a moment for all value passing.
time.Sleep(time.Millisecond)

// Now get zeroBufferConnector output channel ready.
go func() {
	<-zeroBufferConnector.OutC() // It will returns 100
}()
runtime.Gosched()

someConnector.InC() <- 100
```

