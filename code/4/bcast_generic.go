package main_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Payload struct {
	Command string
}

type BChan[T any] struct {
	mu    sync.Mutex
	at    atomic.Pointer[chan T]
	calls int
}

func NewBChan[T any]() BChan[T] {
	result := BChan[T]{}
	newCh := make(chan T)
	result.at.Store(&newCh)
	return result
}

func (f *BChan[T]) Sub() <-chan T {
	p := f.at.Load()
	if p == nil {
		fmt.Println("FATAL ERROR : channel is not present in atomic pointer")
		return nil
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	f.calls++ // this needs to be updated, if in the `select` we have a different `exit` ( like context.Done() )

	return *p

}

func (f *BChan[T]) Send(payload T) {
	ch := f.at.Load()
	if ch == nil {
		fmt.Println("FATAL ERROR : channel is not present in atomic pointer")
		return
	}

	for i := f.calls; i > 0; i-- { // for all the folks expecting our event, we're replicating what was just said
		*ch <- payload
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = 0
}

// a must have, in order to not have a blocking channel
func (f *BChan[T]) Cancel() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls--
}

func RifleMan(ctx context.Context, f ISimpleBroadcaster[Payload]) {
	for {
		select {
		case payload := <-f.Sub():
			fmt.Println("conscript received command : ", payload, payload.Command)
		case <-ctx.Done():
			f.Cancel() // yes, we have to decrease the count for the rest of the world to function
			return
		}
	}
}

func Tank(ctx context.Context, f ISimpleBroadcaster[Payload]) {
	for {
		select {
		case payload := <-f.Sub():
			fmt.Println("tank received command : ", payload, payload.Command)
		case <-ctx.Done():
			f.Cancel() // yes, we have to decrease the count for the rest of the world to function
			return
		}
	}
}

func Artillery(ctx context.Context, f ISimpleBroadcaster[Payload]) {
	for {
		select {
		case payload := <-f.Sub():
			fmt.Println("artillery received command : ", payload, payload.Command)
		case <-ctx.Done():
			f.Cancel() // yes, we have to decrease the count for the rest of the world to function
			return
		}
	}
}

type ISimpleBroadcaster[T any] interface {
	Sub() <-chan T
	Cancel() // so this, MUST be called in the event our context is Done, we've died, or have been thrown under a bus
}

func TestBroadcastingGenericWithRules() {
	fire := NewBChan[Payload]()
	ctx, cancel := context.WithCancel(context.Background())

	go RifleMan(ctx, &fire)
	go Tank(ctx, &fire)
	go Artillery(ctx, &fire)

	<-time.After(1 * time.Second)

	fire.Send(Payload{Command: "fire at will!"})

	<-time.After(2 * time.Second)

	fire.Send(Payload{Command: "fire again"})

	cancel()

	<-time.After(2 * time.Second) // drain all messages if necessary
}
