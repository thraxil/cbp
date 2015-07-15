cbp: *.go
	go build .

run: cbp
	./cbp -l localhost:9000 -r localhost:443

install_deps:
	go get github.com/cenkalti/backoff
