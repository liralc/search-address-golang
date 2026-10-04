# Desafio Multithreading — CEP mais rápido

CLI em Go que consulta um CEP **simultaneamente** na [BrasilAPI](https://brasilapi.com.br) e na [ViaCEP](https://viacep.com.br), exibe a resposta da API que responder primeiro e descarta a outra. Se nenhuma responder em **1 segundo**, o programa exibe um erro de timeout.

## Requisitos

- Go 1.22 ou superior

## Como executar

```bash
go run ./cmd/cep 01001000
# o CEP também pode ser informado com máscara
go run ./cmd/cep 01001-000
```

Exemplo de saída:

```
API mais rápida: BrasilAPI
  CEP:        01001000
  Logradouro: Praça da Sé
  Bairro:     Sé
  Cidade:     São Paulo
  UF:         SP
```

O timeout padrão é de 1 segundo. Para alterar (útil para ver o erro de timeout):

```bash
go run ./cmd/cep -timeout 1ms 01001000
# erro: timeout: nenhuma API respondeu a tempo
```

Para gerar um binário:

```bash
make build      # ou: go build -o bin/cep ./cmd/cep
./bin/cep 01001000
```

Em caso de erro (timeout, CEP inválido ou inexistente), a mensagem vai para o `stderr` e o código de saída é `1`.

## Como testar

Os testes não acessam a internet: os provedores são testados com `httptest` e a lógica de corrida com provedores falsos.

```bash
make test       # ou: go test -race -count=1 ./...
make cover      # cobertura
```

## Estrutura

```
cmd/cep/            ponto de entrada do CLI (flags, timeout, saída)
internal/cep/
  cep.go            modelo Address, interface Provider, validação e cliente HTTP
  brasilapi.go      provedor BrasilAPI
  viacep.go         provedor ViaCEP
  fastest.go        FetchFastest: a corrida entre os provedores
```

## Como funciona

`FetchFastest` dispara uma goroutine por provedor. Todas enviam o resultado para um channel e um `select` espera pelo primeiro sucesso ou pelo fim do prazo (`ctx.Done()`).

- **Timeout:** o prazo de 1 segundo vem de um `context.WithTimeout`. O contexto é repassado às requisições HTTP (`http.NewRequestWithContext`), então as conexões são canceladas de verdade ao final do prazo, e não apenas ignoradas.
- **Descarte da mais lenta:** quando há um vencedor, o `defer cancel()` cancela as requisições restantes.
- **Sem vazamento de goroutines:** o channel tem buffer do tamanho do número de provedores, então as goroutines perdedoras conseguem enviar o resultado e terminar, mesmo que ninguém mais o leia.
- **Falhas:** um erro rápido (ex.: HTTP 500) não vence a corrida. O sistema aguarda a outra API dentro do prazo e só retorna erro se todas falharem.
- **CEP inexistente:** a BrasilAPI responde `404` e a ViaCEP responde `200` com `{"erro": true}`. Os dois casos viram o mesmo `ErrNotFound`.
- **Extensível:** para adicionar uma nova API, basta implementar a interface `Provider` (`Name()` e `Fetch()`).
