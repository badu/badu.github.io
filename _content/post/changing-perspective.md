---
title: Changing Perspective
tags: ["Go", "Channels", "Grouping Methods"]
date: 2018-11-20
description : Changing Perspective Might Help You Understand
---

## Changing perspective

Using abstractions is more about `what` your code can do. Encapsulation is about `how` we achieve that functionality.

#### Foreword

"I was introduced to complex concepts almost immediately... with examples I'd never use in real life. It confused me, I didn't understand it, and eventually quit trying to learn Go because I thought I'd never get it" - quote from a beginner captured by [Mat Ryer](https://matryer.com/)

The present writing is about concurrency and use of channels in Go.

#### Original Code

I'm assuming that you know what pipelining is. One of the interesting properties of pipelines is the ability they give you to operate on the stream of data using a combination of separate, often reorderable stages. It allows you to reuse stages of the pipeline multiple times.

There is a pattern called `fan-out fan-in` which reuses a single stage of a pipeline on multiple goroutines in an attempt to parallelize pulls from an upstream stage - which, in the end, results in improving performance of the pipeline.

Let's look at a piece of code, that is generating a stream of random numbers which are converted into an integer stream and then it's passed to a `heavy duty doing` function. In this case, is about finding prime numbers.

The code is taken from the [github](https://github.com/kat-co/concurrency-in-go-src) code of the book [Concurrency In Go](https://katherine.cox-buday.com/concurrency-in-go/) by [Katherine Cox-Buday](https://katherine.cox-buday.com/blog/).

```go
package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

func main() {
	repeatFn := func(
		done <-chan interface{},
		fn func() interface{},
	) <-chan interface{} {
		valueStream := make(chan interface{})
		go func() {
			defer close(valueStream)
			for {
				select {
				case <-done:
					return
				case valueStream <- fn():
				}
			}
		}()
		return valueStream
	}
	take := func(
		done <-chan interface{},
		valueStream <-chan interface{},
		num int,
	) <-chan interface{} {
		takeStream := make(chan interface{})
		go func() {
			defer close(takeStream)
			for i := 0; i < num; i++ {
				select {
				case <-done:
					return
				case takeStream <- <-valueStream:
				}
			}
		}()
		return takeStream
	}
	toInt := func(done <-chan interface{}, valueStream <-chan interface{}) <-chan int {
		intStream := make(chan int)
		go func() {
			defer close(intStream)
			for v := range valueStream {
				select {
				case <-done:
					return
				case intStream <- v.(int):
				}
			}
		}()
		return intStream
	}
	primeFinder := func(done <-chan interface{}, intStream <-chan int) <-chan interface{} {
		primeStream := make(chan interface{})
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
					case <-done:
						return
					case primeStream <- integer:
					}
				}
			}
		}()
		return primeStream
	}
	fanIn := func(
		done <-chan interface{},
		channels ...<-chan interface{},
	) <-chan interface{} { // <1>
		var wg sync.WaitGroup // <2>
		multiplexedStream := make(chan interface{})

		multiplex := func(c <-chan interface{}) { // <3>
			defer wg.Done()
			for i := range c {
				select {
				case <-done:
					return
				case multiplexedStream <- i:
				}
			}
		}

		// Select from all the channels
		wg.Add(len(channels)) // <4>
		for _, c := range channels {
			go multiplex(c)
		}

		// Wait for all the reads to complete
		go func() { // <5>
			wg.Wait()
			close(multiplexedStream)
		}()

		return multiplexedStream
	}

	done := make(chan interface{})
	defer close(done)

	start := time.Now()

	rand := func() interface{} { return rand.Intn(50000000) }

	randIntStream := toInt(done, repeatFn(done, rand))

	numFinders := runtime.NumCPU()
	fmt.Printf("Using %d Cores.\n", numFinders)
	finders := make([]<-chan interface{}, numFinders)
	fmt.Println("Found Primes:")
	for i := 0; i < numFinders; i++ {
		finders[i] = primeFinder(done, randIntStream)
	}
	num := 0
	for prime := range take(done, fanIn(done, finders...), 10) {
		num++
		fmt.Printf("\t%d - %d\n", num, prime)
	}

	fmt.Printf("Whole work took: %v", time.Since(start))
}
```

Can you read it ? Of course, you can - it's just Go!

Now, let's consider that you'll have to explain it to a beginner. First thing that you can say about the code is that it is very hard to follow despite the fact that functions and variables are named, so you can work your way towards understanding what it does.

In summary, you could say "we now have a number() equal to the number of CPUs) of goroutines pulling from the random number generator and attempting to determine whether the number is prime" and let the poor beginner read (or break) the code until it gets it.

Let's change the code, by using a technique called grouping of methods. This method has both advantages and disadvantages. Main advantage is that for an "OOP brain", it shows the blue prints of the separation of concerns in that code.

First thing on my list of rearranging the above code is inventory of the model parts:

1. we have a main channel, passed around to `toInt`, `repeatFn`, `primeFinder`, `take`, `fanIn` functions allowing them to know when the mission is accomplished.
2. we have a bunch of channels called `finders` in the above code, that allows multiplexing of the work
3. we have a channel for fanning in (`multiplexedStream`) the work
4. we have a wait group (`wg`) that is being used for fanning in the work

Having that in mind, let's do the following declaration:

```go
type ChanInt chan int

type PrimeFinder struct {
	intSource         func() int
	main              ChanInt
	multiplexedStream ChanInt
	runners           []ChanInt
	fanWG             sync.WaitGroup
}
```

Now, it's only a matter of creating methods from the above code. In the end, we're getting the completely rearranged code as following:

#### Modified Code

```go
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
```

#### Instead of Conclusion

Can you read it? Do you find it better, easier to understand, follow and modify?

One can argue that I gave up on the syntactic sugar that shows you the direction of the channel, thus making it more confusing. I can say that it is not true, because the actual operations happen inside the `select` statements - where the developer should look to follow the code. Also, by grouping the functions as methods, it's hiding those kind of details for the developer - you only need to replace the worker function, resting assured that the workflow won't change.

Be warned that this method has disadvantages as well - from the top of my head, I can mention the compiler inlining of the functions, which might affect your speed.

However, I prefer readability and the advantage of being able to explain in simple terms what's going on : what's the encapsulation and what's the abstraction.  

Code before and after can be found [here](https://github.com/badu/badu.github.io/blob/master/code/3).