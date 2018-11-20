package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

type ChanInt chan int

type PrimeFinder struct {
	intSource         func() int
	main              ChanInt
	multiplexedStream ChanInt
	runners           []ChanInt
	fanWG             sync.WaitGroup
}

func (pf PrimeFinder) stream() ChanInt {
	valueStream := make(ChanInt)
	go func() {
		defer close(valueStream)
		for {
			select {
			case <-pf.main:
				return
			case valueStream <- pf.intSource():
			}
		}
	}()
	return valueStream
}

func (pf PrimeFinder) take(capacity int) ChanInt {
	takeStream := make(ChanInt)
	go func() {
		defer close(takeStream)
		for i := 0; i < capacity; i++ {
			select {
			case <-pf.main:
				return
			case takeStream <- <-pf.fanIn():
			}
		}
	}()
	return takeStream
}

func (pf PrimeFinder) checkPrime(intStream ChanInt) ChanInt {
	primeStream := make(ChanInt)
	go func() {
		defer close(primeStream)
		for integer := range intStream {
			integer -= 1
			prime := true
			for divisor := integer - 1; divisor > 1; divisor-- {
				if integer%divisor == 0 {
					prime = false
					break
				}
			}

			if prime {
				select {
				case <-pf.main:
					return
				case primeStream <- integer:
				}
			}
		}
	}()
	return primeStream
}

func (pf PrimeFinder) multiplex(workerChannel ChanInt) {
	defer pf.fanWG.Done()
	for i := range workerChannel {
		select {
		case <-pf.main:
			return
		case pf.multiplexedStream <- i:
		}
	}
}

func (pf PrimeFinder) fanIn() ChanInt {
	// Select from all the channels
	for _, ch := range pf.runners {
		go pf.multiplex(ch)
	}

	// Wait for all the reads to complete
	go func() {
		pf.fanWG.Wait()
		close(pf.multiplexedStream)
		close(pf.main)
	}()

	return pf.multiplexedStream
}

func newPrimeFinder(sourceFn func() int, capacity int) PrimeFinder {
	// prepare channels and slice of runners
	result := PrimeFinder{
		main:              make(ChanInt),
		runners:           make([]ChanInt, capacity),
		multiplexedStream: make(ChanInt),
		intSource:         sourceFn,
	}
	// prepare stream
	randIntStream := result.stream()

	// prepare runners
	for i := 0; i < capacity; i++ {
		result.runners[i] = result.checkPrime(randIntStream)
	}
	// add runners len to waitgroup
	result.fanWG.Add(capacity)
	// return processor
	return result
}

func main() {
	start := time.Now()

	source := func() int { return rand.Intn(50000000) }

	numFinders := runtime.NumCPU()
	fmt.Printf("Spinning up %d prime finders.\n", numFinders)

	processor := newPrimeFinder(source, numFinders)

	fmt.Println("Primes:")
	for prime := range processor.take(10) {
		fmt.Printf("\t%d\n", prime)
	}

	fmt.Printf("Search took: %v", time.Since(start))
}
