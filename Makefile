CEP ?= 01001000

.PHONY: run build test cover vet fmt

run: ## Executa o CLI (ex.: make run CEP=01001-000)
	go run ./cmd/cep $(CEP)

build: ## Gera o binário em bin/cep
	go build -o bin/cep ./cmd/cep

test: ## Roda os testes com detector de race condition
	go test -race -count=1 ./...

cover: ## Roda os testes com relatório de cobertura
	go test -cover ./...

vet:
	go vet ./...

fmt:
	gofmt -w .
