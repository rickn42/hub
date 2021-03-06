package hub_test

import (
	"runtime"
	"testing"
	"time"

	"github.com/rickn42/hub"
)

// TestHub_ValueOrder test if Hub ensure the receive-order.
func TestHub_ValueOrder(t *testing.T) {

	c1 := hub.NewBufferedConnector(100)
	c2 := hub.NewBufferedConnector(100)
	c3 := hub.NewBufferedConnector(100)

	h := hub.NewHub()
	h.PlugIn(c1)
	h.PlugIn(c2)
	h.PlugIn(c3)

	// The interface value get into c1 and then get to c2, c3 through the hub.
	expect := [4]interface{}{1, 2, 3, 4}
	for _, v := range expect {
		c1.InC() <- v
	}

	got := [4]interface{}{<-c2.OutC(), <-c2.OutC(), <-c2.OutC(), <-c2.OutC()}
	if got != expect {
		t.Error("Signal receive failed, got=", got)
	}
	got = [4]interface{}{<-c3.OutC(), <-c3.OutC(), <-c3.OutC(), <-c3.OutC()}
	if got != expect {
		t.Error("Signal receive failed, got=", got)
	}

	// The interface value get into c2 and then get to c1, c3 through the hub.
	expect = [4]interface{}{"hello", "world", "awesome", "golang"}
	for _, v := range expect {
		c2.InC() <- v
	}

	got = [4]interface{}{<-c1.OutC(), <-c1.OutC(), <-c1.OutC(), <-c1.OutC()}
	if got != expect {
		t.Error("Signal receive failed, got=", got)
	}
	got = [4]interface{}{<-c3.OutC(), <-c3.OutC(), <-c3.OutC(), <-c3.OutC()}
	if got != expect {
		t.Error("Signal receive failed, got=", got)
	}
}

func TestHub_Filter(t *testing.T) {
	testIn := make(chan interface{}, 100)
	testOut := make(chan interface{}, 100)
	nothingOut := make(chan interface{}, 100)

	c1 := hub.NewConnectorWithChannels(testIn, nil)
	c2 := hub.NewConnectorWithChannels(nil, testOut)
	c3 := hub.NewConnectorWithChannels(nil, nothingOut)

	evenFilter := func(old interface{}) (new interface{}, ok bool) {
		if old.(int)%3 == 0 {
			return old, true
		}
		return nil, false
	}

	h := hub.NewHub()
	h.PlugIn(c1)
	h.PlugIn(c2, evenFilter)
	h.PlugIn(c3, hub.FilterNothing)

	for i := 1; i < 10; i++ {
		testIn <- i
	}

	expect := [3]interface{}{3, 6, 9}
	got := [3]interface{}{<-testOut, <-testOut, <-testOut}
	if got != expect {
		t.Error("Filter not working! got=", got)
	}

	select {
	case v := <-nothingOut:
		t.Error("FilterNothing not working.", v)
	default:
	}
}

func TestHub_TryAndPass(t *testing.T) {
	c1 := hub.NewBufferedConnector(1)
	c2 := hub.WrapConnectorWithTryAndPass(hub.NewBufferedConnector(0))

	h := hub.NewHub()
	h.PlugIn(c1)
	h.PlugIn(c2)

	// This is not blocked. c2 connector is just passed by the hub.
	c1.InC() <- 1
	c1.InC() <- 2

	// waiting a moment for all value passing.
	time.Sleep(time.Millisecond)

	// Now set c2 connector output channel ready.
	res := make(chan interface{}, 1)
	go func() {
		res <- <-c2.OutC()
	}()
	runtime.Gosched()

	c1.InC() <- 3
	if <-res != 3 {
		t.Error("TryAndPass not working")
	}
}

func BenchmarkHub_10Connector(b *testing.B) {

	const cnt = 10
	cs := [cnt]hub.Connector{}

	for i := range cs {
		cs[i] = hub.NewBufferedConnector(100)
	}

	h := hub.NewHub()
	for _, c := range cs {
		h.PlugIn(c)
	}

	in := cs[0].InC()
	for i := 0; i < b.N; i++ {
		in <- i
		for i := 1; i < cnt; i++ {
			<-cs[i].OutC()
		}
	}
}

func TestHub_PlugOut_Connector(t *testing.T) {
	h := hub.NewHub()

	startCnt := runtime.NumGoroutine()
	connector := hub.NewConnectorWithChannels(nil, nil)
	h.PlugIn(connector)
	h.PlugOut(connector)

	time.Sleep(time.Millisecond)
	if curCnt := runtime.NumGoroutine(); curCnt != startCnt {
		t.Error("PlugOut error.", "startCnt=", startCnt, ", curCnt=", curCnt)
	}

	startCnt = runtime.NumGoroutine()
	in := make(chan interface{})
	connector = hub.NewConnectorWithChannels(in, nil)
	h.PlugIn(connector)
	h.PlugOut(connector)
	close(in)

	time.Sleep(time.Millisecond)
	if curCnt := runtime.NumGoroutine(); curCnt != startCnt {
		t.Error("PlugOut error.", "startCnt=", startCnt, ", curCnt=", curCnt)
	}

	startCnt = runtime.NumGoroutine()
	out := make(chan interface{})
	connector = hub.NewConnectorWithChannels(nil, out)
	h.PlugIn(connector)
	h.PlugOut(connector)
	close(out)

	time.Sleep(time.Millisecond)
	if curCnt := runtime.NumGoroutine(); curCnt != startCnt {
		t.Error("PlugOut error.", "startCnt=", startCnt, ", curCnt=", curCnt)
	}

	startCnt = runtime.NumGoroutine()
	in = make(chan interface{})
	out = make(chan interface{})
	connector = hub.NewConnectorWithChannels(in, out)
	h.PlugIn(connector)
	h.PlugOut(connector)
	close(in)
	close(out)

	time.Sleep(time.Millisecond)
	if curCnt := runtime.NumGoroutine(); curCnt != startCnt {
		t.Error("PlugOut error.", "startCnt=", startCnt, ", curCnt=", curCnt)
	}
}

func TestHub_PlugOut_Re_PlugIn(t *testing.T) {

	h := hub.NewHub()
	c := hub.NewBufferedConnector(10)
	c2 := hub.NewBufferedConnector(10)

	h.PlugIn(c)
	h.PlugIn(c2)

	h.PlugOut(c)
	h.PlugIn(c)

	n := 100
	in := c.InC()
	out := c2.OutC()

	in <- n
	select {
	case v := <-out:
		if v != n {
			t.Error("Re PlugIn failed.", v)
		}
	case <-time.NewTimer(time.Millisecond).C:
		t.Error("Re PlugIn not working.")
	}
}

func TestHub_Destory_EmptyHub(t *testing.T) {
	startCnt := runtime.NumGoroutine()

	h := hub.NewHub()
	time.Sleep(time.Millisecond)
	h.Destory()
	time.Sleep(time.Millisecond)

	if curCnt := runtime.NumGoroutine(); curCnt != startCnt {
		t.Error("Destory error.", "startCnt=", startCnt, ", curCnt=", curCnt)
	}
}

func TestHub_Destroy(t *testing.T) {
	startCnt := runtime.NumGoroutine()

	c1 := make(chan interface{})
	c2 := make(chan interface{})
	c3 := make(chan interface{})
	c4 := make(chan interface{})
	c5 := make(chan interface{})
	c6 := make(chan interface{})

	closed := make(chan interface{})
	close(closed)

	h := hub.NewHub()

	// nil channel
	h.PlugIn(hub.NewConnectorWithChannels(nil, nil))
	h.PlugIn(hub.NewConnectorWithChannels(c1, nil))
	h.PlugIn(hub.NewConnectorWithChannels(nil, c2))
	h.PlugIn(hub.NewConnectorWithChannels(c3, c4))

	// closed channel
	h.PlugIn(hub.NewConnectorWithChannels(closed, nil))
	h.PlugIn(hub.NewConnectorWithChannels(nil, closed))
	h.PlugIn(hub.NewConnectorWithChannels(closed, closed))

	// mix channel
	h.PlugIn(hub.NewConnectorWithChannels(c5, closed))
	h.PlugIn(hub.NewConnectorWithChannels(closed, c6))

	h.Destory()
	close(c1)
	close(c2)
	close(c3)
	close(c4)
	close(c5)
	close(c6)

	// for some goroutine termination delayed.
	time.Sleep(time.Millisecond)

	if curCnt := runtime.NumGoroutine(); curCnt != startCnt {
		t.Error("Destory error.", "startCnt=", startCnt, ", curCnt=", curCnt)
	}
}
