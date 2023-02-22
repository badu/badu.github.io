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

func NewFire() *Fire {
	result := Fire{}
	newCh := make(chan Payload)
	result.at.Store(&newCh)
	return &result
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

func RifleManPointer(ctx context.Context, f *Fire) {
	for {
		select {
		case payload := <-f.Get():
			fmt.Println("conscript received command : ", payload, payload.Command)
		case <-ctx.Done():
			return
		}
	}
}

func TankPointer(ctx context.Context, f *Fire) {
	for {
		select {
		case payload := <-f.Get():
			fmt.Println("tank received command : ", payload, payload.Command)
		case <-ctx.Done():
			return
		}
	}
}

func ArtilleryPointer(ctx context.Context, f *Fire) {
	for {
		select {
		case payload := <-f.Get():
			fmt.Println("artillery received command : ", payload, payload.Command)
		case <-ctx.Done():
			return
		}
	}
}

func TestBroadcastPointer(t *testing.T) {
	fire := NewFire()
	ctx, cancel := context.WithCancel(context.Background())

	go RifleManPointer(ctx, fire)
	go TankPointer(ctx, fire)
	go ArtilleryPointer(ctx, fire)

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
