# cbp

Circuit-Breaker Proxy.

Simple TCP proxy with built-in circuit breaker functionality.

In large systems, it becomes important to be able to isolate faults
quickly to prevent wider scale damage. One common technique for this
is the "circuit breaker" pattern that is discussed in Michael
Nygaard's book "Release It!". A component that may fail is accessed
through a "circuit breaker", which starts "closed", allowing traffic
to flow through. If a certain threshold of errors is crossed, the
circuit breaker "opens" and blocks all traffic to the component,
giving it time to recover. Eventually (either after a fixed time, or
with exponential backoff), the circuit breaker "closes" again, letting
traffic flow again.

`cbp` implements this pattern as a simple TCP proxy. This is cruder
than a typical circuit breaker implementation within an application
(which can do very fine-grained error detection), but allows you to
insert a circuit breaker between components that you may not want to
make source code level changes to (or can't).

Eg, part of your system may make API calls to
`http://api.example.com/`. If that service is failing, instead of
hammering it with requests, it's better to back off a bit, show the
user a message (or otherwise handle it on the client end), and let it
recover. So you'd run

    $ cbp -l localhost:8000 -r api.example.com:80 -t .05

And make API requests to `http://localhost:8000/` instead. If more
than 5% of those requests fail, the circuit breaker pops open and goes
into exponential backoff mode. Traffic is blocked for 1 second, turned
back on for 1 second, if there are more failures, it blocks traffic
for 2 seconds, 4 seconds, 8 seconds, etc.

## Flags

`-l` - [required] local address:port to listen on.

`-r` - [required] remote address:port to proxy to. if it can't resolve
the remote address, it will fail immediately.

`-t` - failure threshold. Defaults to 0.5

`-ms` - minimum number of samples. Defaults to 5. Fewer samples than
        this in the total window and it won't trip.

## LICENSE

BSD
