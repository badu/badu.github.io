---
title: My Thoughts On Net/Http Package - Week 2
tags: ["Golang", "Net", "Http", "Analysis"]
date: 2018-02-24
description : A deep dive into it net/http package.
---
## TL;DR

This series is about my questions and thoughts regarding net/http package. The process of learning is based on mistakes, therefor I'm inviting you to learn aside me.

You are allowed to judge the code. You are not allowed to judge the people.

[Part 1](/post/net-http-week1/)

## ListenAndServe

As you might well know, using http package is easy :
```go
package main

import (
    "io"
	"net/http"
	"log"
)

func main() {
    http.HandleFunc("/hello", func (w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, "hello, world!\n")
	})
	log.Fatal(http.ListenAndServe(":12345", nil))
}
```

#### Chain of Responsibility

The design pattern on which the Golang authors has decided to use is called Chain of Responsibility and it looks like [this](https://github.com/badu/go_design_pattern/blob/master/chain_of_responsibility/chain_of_responsibility.go).

Because it can be simplified using closure functions there was no need to use the "next" property.

#### Inside ListenAndServe

Calling ListenAndServe() will create a new pointer to a [Server](https://github.com/golang/go/blob/release-branch.go1.9/src/net/http/server.go#L2343) and call it's receiver method [ListenAndServe](https://github.com/golang/go/blob/release-branch.go1.9/src/net/http/server.go#L2627). In order to listen on a port we need to use net.Listen("tcp", address) which will return an interface : [net.Listener](https://github.com/golang/go/blob/release-branch.go1.9/src/net/net.go#L361) having interface signature:
```go
    Accept() (Conn, error)
    Close() error
    Addr() Addr
```
As the comment above the interface says, multiple goroutines may invoke methods on a Listener simultaneously.

Of course, the above net.Listener is [net.TCPListener](https://github.com/golang/go/blob/release-branch.go1.9/src/net/tcpsock.go#L224) implementation, since we've mentioned "tcp" as a parameter of our call.

Because we want to handle our own [Accept()](https://github.com/golang/go/blob/release-branch.go1.9/src/net/http/server.go#L3119) this net.TCPListener implementation is type asserted to [tcpKeepAliveListener](https://github.com/golang/go/blob/release-branch.go1.9/src/net/http/server.go#L3115) which is actually embedding a pointer to net.TCPListener, thus allowing us to "override" the Accept method. Once we've prepared this, the receiver method [Serve()](https://github.com/golang/go/blob/release-branch.go1.9/src/net/http/server.go#L2678) is being called, having the above listener as parameter.

#### Serving

A naive approach to serving on our own would look like:

```go
    // ask net to create a tcp listener and return us the interface
    lsn, err := net.Listen("tcp", ":8080")
    if err != nil {
        // handle error
    }
    // ensure that we're releasing the listener
    defer lsn.Close()
    for {
        type accepted struct {
            conn net.Conn
            err  error
        }
        // create a channel to
        c := make(chan accepted, 1)
        go func() {
            conn, err := lsn.Accept() // accept incoming connections
            c <- accepted{conn, err} // send the struct to the channel
        }()
        select {
            case a := <-c: // receive from the channel
                // if the error of the struct is not nil
                if a.err != nil {
                    // handle error and continue, for the next struct to get here
                    continue
                }
                // no error has occurred, we handle the connection
                go handleConnection(a.conn)
            case e := <-ev: // let's say we have a ev channel which transports shutdown requests
                // handle shutdown event
                return
        }
    }
```

In [Serve()](https://github.com/golang/go/blob/release-branch.go1.9/src/net/http/server.go#L2678) method, despite the fact that is seems extra complicated, basic idea is the same. After accepting an incoming connection a [conn](https://github.com/golang/go/blob/release-branch.go1.9/src/net/http/server.go#L229) struct is being created and the accepted connection (which is a net.Conn interface) is being passed to it. Also, the reference to the Server is being passed, because later is used to access timeout values (read, write, idle), but probably the most noticeable thing is [this](https://github.com/golang/go/blob/release-branch.go1.9/src/net/http/server.go#L1801) - read the comment above.

Worth noticing that inside the serve() function of the conn struct is the only place where server [recovers from panic](https://github.com/golang/go/blob/release-branch.go1.9/src/net/http/server.go#L1693). The effective reading of the tcp connection happens on functions of another struct, called [connReader](https://github.com/golang/go/blob/release-branch.go1.9/src/net/http/server.go#L624) - which is an io.Reader wrapper.

One should know that buffer readers and writers are kept in a sync.Pool.

#### A word about tests

For some reason [testHookServerServe](https://github.com/golang/go/blob/release-branch.go1.9/src/net/http/server.go#L2680) - which is a function declared by the tests, was left to go in production. It's not a big deal, because it's used only in one test [TestServeTLS](https://github.com/golang/go/blob/release-branch.go1.9/src/net/http/serve_test.go#L1360). However, there are many test ["hooks"](https://github.com/golang/go/blob/release-branch.go1.9/src/net/http/transport.go#L1877) left around inside the production code.

I've decided to replace them with the following technique:
```go
    type(
        ServerEventType int 
        
        srvEvDispatcher struct {
            lsns map[ServerEventType][]srvEvListner
            mu   sync.RWMutex
        }
    
        srvEvListner struct {
            ch chan ServerEventType
        }
        // a helper struct which embeds a waitgroup
        ServerEventHandler struct {
            sync.WaitGroup
            ch          chan ServerEventType // channel for receiving events
            handler     func()               // function which gets called if event is met
            eventType   ServerEventType      // which kind of event we're listening to
            willRemount bool                 // internal, so we can continuosly listen
        }
    )
    
    const(
        killListeners               ServerEventType = 0
        ServerServe                 ServerEventType = 1
        EnterRoundTripEvent         ServerEventType = 2
        RoundTripRetriedEvent       ServerEventType = 3
        PrePendingDialEvent         ServerEventType = 4
        PostPendingDialEvent        ServerEventType = 5
        WaitResLoopEvent            ServerEventType = 6
        ReadLoopBeforeNextReadEvent ServerEventType = 7
    )
        
    func (r *srvEvDispatcher) dispatch(event ServerEventType) {
    	if len(r.lsns[event]) == 0{ 
    	     return 
        }
        r.mu.Lock()
        defer r.mu.Unlock()
        // for each listener of that event type
        for i := 0; i < len(r.lsns[event]); i++ {
            lisn := r.lsns[event][i] 
            select {
            case lisn.ch <- event: // we're writting into the channel
            default:
            }
        }
    }
    // "mounting" the effective listener
    func (r *srvEvDispatcher) on(event ServerEventType) chan ServerEventType {
        r.mu.Lock()
        defer r.mu.Unlock()
        ch := make(chan ServerEventType, 1)
        r.lsns[event] = append(r.lsns[event], srvEvListner{ch: ch})
        return ch
    }
    // helper method that will receive an event via a channel, then mount itself to listen for more
    func (h ServerEventHandler) Next() {
        h.Add(1)
        go func() {
            defer h.Done()
            func() {
                switch <-h.ch {
                case h.eventType:
                    h.handler()
                case killListeners:
                    // on kill, we will not do "next" execution
                    h.willRemount = false
                }
            }()
        }()
        h.Wait()
        if h.willRemount {
            // next execution
            go h.Next()
        }
    }
    // usage "defer eventListener.Kill()". Will use a custom type that tells the above helper to stop mounting itself
    func (h ServerEventHandler) Kill() {
        h.ch <- killListeners
    }
    // called from tests, to listen for server events
    func ListenTestEvent(eventType ServerEventType, f func()) ServerEventHandler {
        wg := ServerEventHandler{ch: testEventsEmitter.on(eventType), handler: f, eventType: eventType, willRemount: true}
        // first execution
        go wg.Next()
        return wg
    }
```

You can find the code [here](https://github.com/badu/http).

To be continued.