package main_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type Payload struct {
	Command string
}

type Fire struct {
	mu sync.Mutex // protects field(s) below
	at atomic.Pointer[chan Payload]
}

func NewFire() Fire {
	result := Fire{}
	newCh := make(chan Payload)
	result.at.Store(&newCh)
	return result
}

func (f *Fire) Get() <-chan Payload {
	p := f.at.Load()
	if p == nil {
		fmt.Println("FATAL ERROR : channel is not present in atomic pointer")
		return nil
	}

	return *p

}

func (f *Fire) Send(payload Payload) {
	ch := f.at.Load()
	if ch == nil {
		fmt.Println("FATAL ERROR : channel is not present in atomic pointer")
		return
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	newCh := make(chan Payload)
	f.at.Store(&newCh)

	*ch <- payload // write the payload, so at least one is receiving

	close(*ch) // this causes all "listeners" to receive a payload (for the majority this is empty)
}

func RifleMan(ctx context.Context, f IFire) {
	for {
		select {
		case payload := <-f.Get():
			fmt.Println("conscript received command : ", payload, payload.Command)
		case <-ctx.Done():
			// fmt.Println("conscript closing up radio : context it's done")
			return
		}
	}
}

func Tank(ctx context.Context, f IFire) {
	for {
		select {
		case payload := <-f.Get():
			fmt.Println("tank received command : ", payload, payload.Command)
		case <-ctx.Done():
			// fmt.Println("tank closing up radio : context it's done")
			return
		}
	}
}

func Artillery(ctx context.Context, f IFire) {
	for {
		select {
		case payload := <-f.Get():
			fmt.Println("artillery received command : ", payload, payload.Command)
		case <-ctx.Done():
			// fmt.Println("artillery closing up radio : context it's done")
			return
		}
	}
}

type IFire interface { // use interface, it's more elegant
	Get() <-chan Payload
}

func TestBroadcastCloseChannel(t *testing.T) {
	fire := NewFire()
	ctx, cancel := context.WithCancel(context.Background())

	go RifleMan(ctx, &fire)
	go Tank(ctx, &fire)
	go Artillery(ctx, &fire)

	t.Log("army is listening. waiting 1 second...")

	<-time.After(1 * time.Second)

	t.Log("sending command")
	fire.Send(Payload{Command: "fire at will!"})

	<-time.After(2 * time.Second)

	t.Log("sending command again")
	fire.Send(Payload{Command: "fire again"})

	cancel()

	<-time.After(2 * time.Second) // drain all messages if necessary
	t.Log("Done-done!")

}
