// Package cep consulta endereços a partir de um CEP em múltiplos provedores
// concorrentemente, retornando a resposta do mais rápido.
package cep

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"unicode"
)

// Erros de domínio expostos pelo pacote.
var (
	ErrInvalidCEP  = errors.New("CEP inválido: deve conter 8 dígitos")
	ErrNotFound    = errors.New("CEP não encontrado")
	ErrNoProviders = errors.New("nenhum provedor configurado")
)

// Address é a representação normalizada de um endereço, independente do provedor.
type Address struct {
	CEP          string `json:"cep"`
	Street       string `json:"street"`
	Neighborhood string `json:"neighborhood"`
	City         string `json:"city"`
	State        string `json:"state"`
}

// Provider é uma fonte de consulta de CEP.
type Provider interface {
	Name() string
	Fetch(ctx context.Context, cep string) (Address, error)
}

// Normalize remove caracteres não numéricos (ex.: "01001-000") e valida o tamanho.
func Normalize(raw string) (string, error) {
	cep := strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) {
			return r
		}
		return -1
	}, raw)
	if len(cep) != 8 {
		return "", ErrInvalidCEP
	}
	return cep, nil
}

// getJSON executa um GET respeitando o contexto e decodifica o corpo em v.
func getJSON(ctx context.Context, client *http.Client, url string, v any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("criando requisição: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("executando requisição: %w", err)
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusNotFound:
		return ErrNotFound
	case resp.StatusCode != http.StatusOK:
		return fmt.Errorf("status inesperado: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf("decodificando resposta: %w", err)
	}
	return nil
}
