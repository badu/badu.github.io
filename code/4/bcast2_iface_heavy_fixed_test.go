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
	mu   sync.Mutex // protects field(s) below
	at   atomic.Pointer[chan Payload]
	gets int // keeps the number of Get() calls created by "listeners". Set() will use this to multiply the "message" to those listeners
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

	f.mu.Lock()
	defer f.mu.Unlock()

	f.gets++

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

	for i := f.gets; i > 0; i-- {
		*ch <- payload
	}

	f.gets = 0
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

func heavyOperation(ctx context.Context, payload Payload) {
	fmt.Println("tank will execute command after 3 seconds ", payload.Command)
	for {
		select {
		case <-time.After(3 * time.Second):
			fmt.Println("tank heavy ops done!")
		case <-ctx.Done():
			// fmt.Println("tank closing up radio : context it's done")
			return
		}
	}
}

func Tank(ctx context.Context, f IFire) {
	for {
		select {
		case payload := <-f.Get():
			go heavyOperation(payload)
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

/*
*
So, almost everytime we're using this "technique" we should be sure that we run a goroutine with the heavy lifting inside the listener (see 'heavyOperation' function above).
Also note that 'heavyOperation' is now listening for the context Done(), so it can abandon if context finishes first
*/
func TestBroadcastHeavyOperation(t *testing.T) {
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

	<-time.After(4 * time.Second) // drain all messages if necessary - note that we've increased this value, so tank can finish too
	t.Log("Done-done!")

}
