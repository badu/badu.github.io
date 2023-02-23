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
	newCh := make(chan Payload, 1)
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
		case <-ctx.Done():
			// fmt.Println("conscript closing up radio : context it's done")
			return
		case payload := <-f.Get():
			fmt.Println("conscript received command : ", payload, payload.Command)
		}
	}
}

func Tank(ctx context.Context, f IFire) {
	for {
		select {
		case <-ctx.Done():
			// fmt.Println("tank closing up radio : context it's done")
			return
		case payload := <-f.Get():
			fmt.Println("tank received command : ", payload, payload.Command)
		}
	}
}

func Artillery(ctx context.Context, f IFire) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("artillery closing up radio : context it's done")
			return
		case payload := <-f.Get():
			fmt.Println("artillery received command : ", payload, payload.Command)
			// add return here, if you're going to receive this message only once (you might as well give up on listening context)
		}
	}
}

type IFire interface { // use interface, it's more elegant
	Get() <-chan Payload
}

// Now, lets simulate that Artillery will participate in receiving only one event.
// We provide a different context and cancel that context after the first payload.
// After performing the experiment, I've noticed that if `case payload := <-f.Get():` is written first
// a panic will occur.
// Attempted to fix it by moving the `<-ctx.Done()` as first condition didn't worked (panics).
// Attempted to fix it by using a buffered channel with a size 1, works but one 'actor' receives a duplicate message.
// Last attempt : added a return immediately after the `case payload := <-f.Get():`, so the goroutine would exit.

// In conclusion :
// 1. do not use this pattern unless all 'listeners' are added and cancelled in the same time.
// 2. do not do heavy load operations inside 'listeners'. Use another goroutine to take over a copy of the payload they have received.
// 3. use interface (`IFire` here), in order to avoid pointers (the elegant solution).
// 4. when you close a channel, all 'listeners' of that channel will receive an 'empty' payload. This can be used in the same manner as 'context.Done()', by checking if the payload received is empty (channel was closed).
// 5. do not use more channels in the select. It will increment 'gets'.
// 6. last, but not least, it seems the AI has no idea how to solve this problem. To me, that's funny.

func TestBroadcastLastExperiment(t *testing.T) {
	fire := NewFire()
	ctx, cancel := context.WithCancel(context.Background())
	ctx2, cancel2 := context.WithCancel(context.Background())

	go RifleMan(ctx, &fire)
	go Tank(ctx, &fire)
	go Artillery(ctx2, &fire)

	t.Log("army is listening. waiting 1 second...")

	<-time.After(1 * time.Second)

	t.Log("sending command")
	fire.Send(Payload{Command: "fire at will!"})
	cancel2()

	<-time.After(2 * time.Second)

	t.Log("sending command again")
	fire.Send(Payload{Command: "fire again"})

	cancel()

	<-time.After(2 * time.Second) // drain all messages if necessary
	t.Log("Done-done!")

}
