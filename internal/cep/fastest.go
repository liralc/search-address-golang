package cep

import (
	"context"
	"errors"
	"fmt"
)

// ErrTimeout indica que nenhum provedor respondeu dentro do prazo.
var ErrTimeout = errors.New("timeout: nenhuma API respondeu a tempo")

// Result é o endereço vencedor junto com o nome do provedor que o entregou.
type Result struct {
	Address Address
	Source  string
}

type outcome struct {
	result Result
	err    error
}

// FetchFastest consulta todos os provedores em paralelo e retorna a primeira
// resposta bem-sucedida. As demais requisições são canceladas e descartadas.
//
// O prazo é controlado pelo ctx (use context.WithTimeout). Se o prazo expirar,
// retorna ErrTimeout; se todos falharem antes disso, retorna os erros combinados.
func FetchFastest(ctx context.Context, cep string, providers ...Provider) (Result, error) {
	if len(providers) == 0 {
		return Result{}, ErrNoProviders
	}

	// Cancela as requisições perdedoras assim que houver um vencedor.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Buffer do tamanho do número de provedores: as goroutines perdedoras
	// conseguem enviar sem bloquear, evitando vazamento de goroutines.
	results := make(chan outcome, len(providers))
	for _, p := range providers {
		go func() {
			addr, err := p.Fetch(ctx, cep)
			if err != nil {
				err = fmt.Errorf("%s: %w", p.Name(), err)
			}
			results <- outcome{result: Result{Address: addr, Source: p.Name()}, err: err}
		}()
	}

	var errs []error
	for range providers {
		select {
		case out := <-results:
			if out.err == nil {
				return out.result, nil
			}
			errs = append(errs, out.err)
		case <-ctx.Done():
			return Result{}, contextError(ctx)
		}
	}
	// Os provedores podem ter falhado justamente por causa do prazo.
	if ctx.Err() != nil {
		return Result{}, contextError(ctx)
	}
	return Result{}, errors.Join(errs...)
}

func contextError(ctx context.Context) error {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return ErrTimeout
	}
	return ctx.Err()
}
