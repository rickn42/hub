package hub_test

import (
	"testing"

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

func BenchmarkHub_2Connector(b *testing.B) {

	c1 := hub.NewBufferedConnector(1)
	c2 := hub.NewBufferedConnector(1)

	hub := hub.NewHub()
	hub.PlugIn(c1)
	hub.PlugIn(c2)

	in := c1.InC()
	out := c2.OutC()
	for i := 0; i < b.N; i++ {
		in <- i
		<-out
	}
}

func BenchmarkHub_10Connector(b *testing.B) {

	cs := []hub.Connector{
		hub.NewBufferedConnector(1),
		hub.NewBufferedConnector(1),
		hub.NewBufferedConnector(1),
		hub.NewBufferedConnector(1),
		hub.NewBufferedConnector(1),
		hub.NewBufferedConnector(1),
		hub.NewBufferedConnector(1),
		hub.NewBufferedConnector(1),
		hub.NewBufferedConnector(1),
		hub.NewBufferedConnector(1),
	}

	hub := hub.NewHub()
	for _, c := range cs {
		hub.PlugIn(c)
	}

	in := cs[0].InC()
	for i := 0; i < b.N; i++ {
		in <- i
		for i := 1; i < 10; i++ {
			<-cs[i].OutC()
		}
	}
}
