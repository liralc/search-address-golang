package cep

import (
	"context"
	"errors"
	"testing"
	"time"
)

// fakeProvider responde após um atraso, respeitando o cancelamento do contexto.
type fakeProvider struct {
	name  string
	delay time.Duration
	addr  Address
	err   error
}

func (f fakeProvider) Name() string { return f.name }

func (f fakeProvider) Fetch(ctx context.Context, _ string) (Address, error) {
	select {
	case <-time.After(f.delay):
		return f.addr, f.err
	case <-ctx.Done():
		return Address{}, ctx.Err()
	}
}

func TestFetchFastest(t *testing.T) {
	fast := Address{CEP: "01001-000", City: "rápida"}
	slow := Address{CEP: "01001-000", City: "lenta"}

	tests := []struct {
		name       string
		timeout    time.Duration
		providers  []Provider
		wantSource string
		wantCity   string
		wantErr    error
	}{
		{
			name:    "primeira resposta vence",
			timeout: time.Second,
			providers: []Provider{
				fakeProvider{name: "A", delay: 200 * time.Millisecond, addr: slow},
				fakeProvider{name: "B", delay: 10 * time.Millisecond, addr: fast},
			},
			wantSource: "B",
			wantCity:   "rápida",
		},
		{
			name:    "erro rápido não vence resposta válida mais lenta",
			timeout: time.Second,
			providers: []Provider{
				fakeProvider{name: "A", delay: 5 * time.Millisecond, err: errors.New("boom")},
				fakeProvider{name: "B", delay: 50 * time.Millisecond, addr: slow},
			},
			wantSource: "B",
			wantCity:   "lenta",
		},
		{
			name:    "timeout quando ninguém responde a tempo",
			timeout: 50 * time.Millisecond,
			providers: []Provider{
				fakeProvider{name: "A", delay: time.Second, addr: slow},
				fakeProvider{name: "B", delay: time.Second, addr: slow},
			},
			wantErr: ErrTimeout,
		},
		{
			name:    "todos falham",
			timeout: time.Second,
			providers: []Provider{
				fakeProvider{name: "A", delay: time.Millisecond, err: ErrNotFound},
				fakeProvider{name: "B", delay: time.Millisecond, err: ErrNotFound},
			},
			wantErr: ErrNotFound,
		},
		{
			name:    "sem provedores",
			timeout: time.Second,
			wantErr: ErrNoProviders,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			got, err := FetchFastest(ctx, "01001000", tt.providers...)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("erro = %v, esperado %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("erro inesperado: %v", err)
			}
			if got.Source != tt.wantSource || got.Address.City != tt.wantCity {
				t.Errorf("resultado = %+v, esperado source=%s city=%s", got, tt.wantSource, tt.wantCity)
			}
		})
	}
}

func TestFetchFastestRespectsTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := FetchFastest(ctx, "01001000", fakeProvider{name: "A", delay: 5 * time.Second})
	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("erro = %v, esperado ErrTimeout", err)
	}
	if elapsed := time.Since(start); elapsed > 500*time.Millisecond {
		t.Errorf("FetchFastest demorou %v; deveria retornar logo após o timeout", elapsed)
	}
}
