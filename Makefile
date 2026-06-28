test:
	go test ./... -cover

vet:
	go vet ./...

fmt:
	gofmt -w .

ci:
	go vet ./...
	go test ./... -cover
