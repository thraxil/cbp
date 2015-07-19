package main

import (
	"flag"
	"io"
	"log"
	"net"

	"github.com/rubyist/circuitbreaker"
)

var localAddr *string = flag.String("l", "localhost:9999", "local address")
var remoteAddr *string = flag.String("r", "localhost:80", "remote address")

func Proxy(cliConn *net.TCPConn, rAddr *net.TCPAddr) error {
	srvConn, err := net.DialTCP("tcp", nil, rAddr)
	if err != nil {
		cliConn.Close()
		return err
	}
	log.Println("ok")
	defer srvConn.Close()

	// channels to wait on the close event for each connection
	serverClosed := make(chan struct{}, 1)
	clientClosed := make(chan struct{}, 1)

	go broker(srvConn, cliConn, clientClosed)
	go broker(cliConn, srvConn, serverClosed)

	var waitFor chan struct{}
	select {
	case <-clientClosed:
		// the client closed first
		srvConn.SetLinger(0)
		srvConn.CloseRead()
		waitFor = serverClosed
	case <-serverClosed:
		cliConn.CloseRead()
		waitFor = clientClosed
	}

	<-waitFor
	return nil
}

func broker(dst, src net.Conn, srcClosed chan struct{}) {
	_, err := io.Copy(dst, src)

	if err != nil {
		log.Printf("Copy error: %s", err)
	}
	if err := src.Close(); err != nil {
		log.Printf("Close error: %s", err)
	}
	srcClosed <- struct{}{}
}

func handleConn(in <-chan *net.TCPConn, out chan<- *net.TCPConn, rAddr *net.TCPAddr, cb *circuit.Breaker) {
	for conn := range in {
		cb.Call(func() error {
			return Proxy(conn, rAddr)
		}, 0)
	}
}

func closeConn(in <-chan *net.TCPConn) {
	for conn := range in {
		conn.Close()
	}
}

func main() {
	flag.Parse()

	log.Printf("%v -> %v\n", *localAddr, *remoteAddr)

	addr, err := net.ResolveTCPAddr("tcp", *localAddr)
	if err != nil {
		panic(err)
	}
	rAddr, err := net.ResolveTCPAddr("tcp", *remoteAddr)
	if err != nil {
		panic(err)
	}

	cb := circuit.NewRateBreaker(0.95, 5)
	events := cb.Subscribe()

	go func() {
		for {
			e := <-events
			switch e {
			case circuit.BreakerTripped:
				log.Println("breaker tripped")
			case circuit.BreakerReset:
				log.Println("breaker reset")
			case circuit.BreakerFail:
				log.Println("breaker fail")
			case circuit.BreakerReady:
				log.Println("breaker ready")
			}
		}
	}()

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}

	pending, complete := make(chan *net.TCPConn), make(chan *net.TCPConn)

	for i := 0; i < 5; i++ {
		go handleConn(pending, complete, rAddr, cb)
	}
	go closeConn(complete)

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			panic(err)
		}
		pending <- conn
	}
}
