package cep

import (
	"context"
	"net/http"
)

const brasilAPIDefaultURL = "https://brasilapi.com.br/api/cep/v1/"

// BrasilAPI consulta https://brasilapi.com.br.
type BrasilAPI struct {
	Client  *http.Client
	BaseURL string // permite apontar para um servidor de teste
}

// NewBrasilAPI cria o provedor com a URL oficial.
func NewBrasilAPI(client *http.Client) *BrasilAPI {
	return &BrasilAPI{Client: client, BaseURL: brasilAPIDefaultURL}
}

func (*BrasilAPI) Name() string { return "BrasilAPI" }

type brasilAPIResponse struct {
	CEP          string `json:"cep"`
	State        string `json:"state"`
	City         string `json:"city"`
	Neighborhood string `json:"neighborhood"`
	Street       string `json:"street"`
}

func (b *BrasilAPI) Fetch(ctx context.Context, cep string) (Address, error) {
	var r brasilAPIResponse
	if err := getJSON(ctx, b.Client, b.BaseURL+cep, &r); err != nil {
		return Address{}, err
	}
	return Address{
		CEP:          r.CEP,
		Street:       r.Street,
		Neighborhood: r.Neighborhood,
		City:         r.City,
		State:        r.State,
	}, nil
}
