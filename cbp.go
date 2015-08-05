package main

import (
	"flag"
	"io"
	"log"
	"net"
	"time"

	"github.com/peterbourgon/g2s"
	"github.com/rubyist/circuitbreaker"
)

var localAddr = flag.String("l", "localhost:9999", "local address")
var remoteAddr = flag.String("r", "localhost:80", "remote address")
var threshold = flag.Float64("t", 0.5, "error threshold for tripping")
var minSamples = flag.Int64("ms", 5, "minimum samples")
var windowTime = flag.Int64("window-time", 10000, "window time (ms)")
var windowBuckets = flag.Int64("window-buckets", 10, "window Buckets")
var verbose = flag.Bool("v", false, "verbose")

var statsdHost = flag.String("statsd", "", "statsd host. eg: localhost8125")
var metricBase = flag.String("metric-base", "", "statsd metric base")
var metricName = flag.String("metric-name", "", "statsd metric name")

func proxy(cliConn *net.TCPConn, rAddr *net.TCPAddr) error {
	srvConn, err := net.DialTCP("tcp", nil, rAddr)
	if err != nil {
		cliConn.Close()
		return err
	}
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
		if *verbose {
			log.Printf("Copy error: %s", err)
		}
	}
	if err := src.Close(); err != nil {
		if *verbose {
			log.Printf("Close error: %s", err)
		}
	}
	srcClosed <- struct{}{}
}

func handleConn(in <-chan *net.TCPConn, out chan<- *net.TCPConn, rAddr *net.TCPAddr, cb *circuit.Breaker) {
	for conn := range in {
		cb.Call(func() error {
			return proxy(conn, rAddr)
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
	if *verbose {
		log.Printf("%v -> %v\n", *localAddr, *remoteAddr)
	}

	addr, err := net.ResolveTCPAddr("tcp", *localAddr)
	if err != nil {
		log.Fatal("cannot resolve local address: ", err)
	}
	rAddr, err := net.ResolveTCPAddr("tcp", *remoteAddr)
	if err != nil {
		log.Fatal("cannot resolve remote address: ", err)
	}
	options := circuit.Options{
		ShouldTrip:    circuit.RateTripFunc(*threshold, *minSamples),
		WindowTime:    time.Duration(*windowTime) * time.Millisecond,
		WindowBuckets: int(*windowBuckets),
	}
	cb := circuit.NewBreakerWithOptions(&options)
	events := cb.Subscribe()

	if *statsdHost != "" && *metricBase != "" && *metricName != "" {
		log.Println("logging to statsd")
		s, err := g2s.Dial("udp", *statsdHost)
		if err != nil {
			log.Fatal(err)
		}
		panel := circuit.NewPanel()
		panel.StatsPrefixf = *metricBase + ".%s"
		panel.Statter = s
		panel.Add(*metricName, cb)
	}
	if *verbose {
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
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatal("cannot bind to local port: ", err)
	}

	pending, complete := make(chan *net.TCPConn), make(chan *net.TCPConn)

	for i := 0; i < 5; i++ {
		go handleConn(pending, complete, rAddr, cb)
	}
	go closeConn(complete)

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Fatal("error starting listener: ", err)
		}
		pending <- conn
	}
}
