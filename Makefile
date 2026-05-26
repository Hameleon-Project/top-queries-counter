.PHONY: build test bench loadtest run up down publish

build:
	go build -o bin/server ./cmd/server
	go build -o bin/publisher ./cmd/publisher

test:
	go test ./... -count=1

bench:
	go test ./internal/store -bench=. -benchmem -count=3

loadtest:
	chmod +x scripts/loadtest.sh
	./scripts/loadtest.sh

run:
	go run ./cmd/server

up:
	docker compose up --build -d

down:
	docker compose down

publish:
	go run ./cmd/publisher -query "iphone 15" -n 5
