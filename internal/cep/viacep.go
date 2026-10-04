package cep

import (
	"context"
	"net/http"
)

const viaCEPDefaultURL = "http://viacep.com.br/ws/"

// ViaCEP consulta https://viacep.com.br.
type ViaCEP struct {
	Client  *http.Client
	BaseURL string // permite apontar para um servidor de teste
}

// NewViaCEP cria o provedor com a URL oficial.
func NewViaCEP(client *http.Client) *ViaCEP {
	return &ViaCEP{Client: client, BaseURL: viaCEPDefaultURL}
}

func (*ViaCEP) Name() string { return "ViaCEP" }

type viaCEPResponse struct {
	CEP        string `json:"cep"`
	Logradouro string `json:"logradouro"`
	Bairro     string `json:"bairro"`
	Localidade string `json:"localidade"`
	UF         string `json:"uf"`
	// ViaCEP responde 200 com {"erro": true} (bool ou string) para CEP inexistente.
	Erro any `json:"erro"`
}

func (v *ViaCEP) Fetch(ctx context.Context, cep string) (Address, error) {
	var r viaCEPResponse
	if err := getJSON(ctx, v.Client, v.BaseURL+cep+"/json/", &r); err != nil {
		return Address{}, err
	}
	if r.Erro != nil {
		return Address{}, ErrNotFound
	}
	return Address{
		CEP:          r.CEP,
		Street:       r.Logradouro,
		Neighborhood: r.Bairro,
		City:         r.Localidade,
		State:        r.UF,
	}, nil
}
