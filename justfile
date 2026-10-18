e2e: vet test build smoke

build:
	go build -o gh-next-pr

smoke:
	./gh-next-pr cli/cli

test:
	go test ./...

vet:
	go vet ./...
