# Hub

Hub get channel connectors. And connectors got signal each other.

### Basic

`Connector` is interface which has 2 channel.

`InC` returns an input channel which data get into the hub.

`OutC` returns an output channel which data get out from the hub.

`TryAndPass` is signed if pass or not when output channel not ready.

```go
type Connector interface {
    InC() chan interface{}
    OutC() chan interface{}
    TryAndPass() bool
}
```

### Plug-in

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

### Plug-out 

```go
h := hub.NewHub()
h.PlugIn(someConnector)
h.PlugOut(someConnector)

// Re-plugin is still working.
h.PlugIn(someConnector)
```

### Filter 

```go
type Filter = func(old interface{}) (new interface{}, ok bool)
```

```go
evenFilter = func(v interface{}) (interface{}, bool) {
    if v.(int) % 2 == 0 {
        return v, true
    } 
    return nil, false 
}

h := hub.NewHub()
h.PlugIn(someConnector)
h.PlugIn(anotherConnector, evenFilter) 

someConnector.InC() <- 1
someConnector.InC() <- 2
someConnector.InC() <- 3

// only got 2
<-anotherCennector.OutC()
```

### Try and Pass 

The value passing in hub is working synchronizely.

So if other connector's 'OutC()` channel not in use and no more buffer, 

then value passing is blocked.

```go
// Channel buffer size is 0
someConnector := hub.NewBufferedConnector(0))
zeroBufferConnector := hub.NewBufferedConnector(0)

h := hub.NewHub()
h.PlugIn(someConnector)
h.PlugIn(zeroBufferConnector)

// It is blocked! anotherConnector's output channel is not ready!
someConnector.InC() <- 1
```

Use `TryAndPass()` set true.

It will just pass channel when channel is not in use.

```go
// This wrapper just has TryAndPass() always true method.
anotherConnector := hub.WrapConnectorWithTryAndPass(zeroBufferConnector)

h := hub.NewHub()
h.PlugIn(someConnector)
h.PlugIn(anotherConnector)

// Channel is not blocked. 
someConnector.InC() <- 1
someConnector.InC() <- 2
someConnector.InC() <- 3

// Waiting a moment for all value passing.
time.Sleep(time.Millisecond)

// Now set zeroBufferConnector output channel ready.
go func() {
	// It will returns 100
	<-zeroBufferConnector.OutC() 
}()
runtime.Gosched()

// It will not be blocked.
someConnector.InC() <- 100
```

