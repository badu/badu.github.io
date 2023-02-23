package main_test

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func Rifleman(ctx context.Context, c chan string) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Rifleman: I'm done")
			return
		case msg := <-c:
			fmt.Println("Rifleman:", msg)
		}
	}
}
func Tank(ctx context.Context, c chan string) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Tank: I'm done")
			return
		case msg := <-c:
			fmt.Println("Tank:", msg)
		}
	}
}
func Artilery(ctx context.Context, c chan string) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Artillery: I'm done")
			return
		case msg := <-c:
			fmt.Println("Artillery:", msg)
		}
	}
}

/*
*
Including the AI expects the following output:
Rifleman: Fire!
Tank: Fire!
Artillery: Fire!
Rifleman: I'm done
Tank: I'm done
Artillery: I'm done
*/
func TestBrokenBroadcast(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan string)

	go Rifleman(ctx, c)
	go Tank(ctx, c)
	go Artilery(ctx, c)

	c <- "Fire!"

	cancel()

	time.Sleep(time.Second) // drain messages

}
