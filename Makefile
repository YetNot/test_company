.PHONY: run
run:
	go run main.go

.PHONY: build
build:
	go build -o bot main.go