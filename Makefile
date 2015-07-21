cbp: *.go
	go build .

run: cbp
	./cbp -l localhost:9000 -r localhost:443

test: *.go
	go test .

install_deps:
	go get github.com/cenkalti/backoff
	go get github.com/rubyist/circuitbreaker
	go get github.com/peterbourgon/g2s
