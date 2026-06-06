.PHONY: run build test tidy clean docker fmt vet seed-reset

BINARY := bin/lim-server

run:            ## Run the API server (seeded demo data)
	go run ./cmd/server

build:          ## Build the server binary into bin/
	@mkdir -p bin
	go build -trimpath -ldflags="-s -w" -o $(BINARY) ./cmd/server

test:           ## Run the test suite
	go test ./...

vet:            ## Static analysis
	go vet ./...

fmt:            ## Format the code
	gofmt -w .

tidy:           ## Tidy go.mod/go.sum
	go mod tidy

docker:         ## Build the Docker image
	docker build -t lim-backend:latest .

seed-reset:     ## Delete the local data file to force a fresh seed on next run
	rm -f lim-data.json lim-data.json.tmp

clean:          ## Remove build artefacts
	rm -rf bin lim-data.json lim-data.json.tmp

help:           ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'
