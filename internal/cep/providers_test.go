package cep

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newServer(t *testing.T, wantPath string, status int, body string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != wantPath {
			t.Errorf("path = %q, esperado %q", r.URL.Path, wantPath)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestBrasilAPIFetch(t *testing.T) {
	srv := newServer(t, "/api/cep/v1/01001000", http.StatusOK,
		`{"cep":"01001000","state":"SP","city":"São Paulo","neighborhood":"Sé","street":"Praça da Sé"}`)

	p := &BrasilAPI{Client: srv.Client(), BaseURL: srv.URL + "/api/cep/v1/"}
	got, err := p.Fetch(context.Background(), "01001000")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	want := Address{CEP: "01001000", Street: "Praça da Sé", Neighborhood: "Sé", City: "São Paulo", State: "SP"}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestBrasilAPINotFound(t *testing.T) {
	srv := newServer(t, "/api/cep/v1/99999999", http.StatusNotFound, `{"message":"not found"}`)

	p := &BrasilAPI{Client: srv.Client(), BaseURL: srv.URL + "/api/cep/v1/"}
	if _, err := p.Fetch(context.Background(), "99999999"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("erro = %v, esperado ErrNotFound", err)
	}
}

func TestViaCEPFetch(t *testing.T) {
	srv := newServer(t, "/ws/01001000/json/", http.StatusOK,
		`{"cep":"01001-000","logradouro":"Praça da Sé","bairro":"Sé","localidade":"São Paulo","uf":"SP"}`)

	p := &ViaCEP{Client: srv.Client(), BaseURL: srv.URL + "/ws/"}
	got, err := p.Fetch(context.Background(), "01001000")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	want := Address{CEP: "01001-000", Street: "Praça da Sé", Neighborhood: "Sé", City: "São Paulo", State: "SP"}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestViaCEPNotFound(t *testing.T) {
	for _, body := range []string{`{"erro": true}`, `{"erro": "true"}`} {
		srv := newServer(t, "/ws/99999999/json/", http.StatusOK, body)

		p := &ViaCEP{Client: srv.Client(), BaseURL: srv.URL + "/ws/"}
		if _, err := p.Fetch(context.Background(), "99999999"); !errors.Is(err, ErrNotFound) {
			t.Errorf("body %s: erro = %v, esperado ErrNotFound", body, err)
		}
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{in: "01001000", want: "01001000"},
		{in: "01001-000", want: "01001000"},
		{in: " 01.001-000 ", want: "01001000"},
		{in: "123", wantErr: true},
		{in: "abcdefgh", wantErr: true},
		{in: "", wantErr: true},
	}
	for _, tt := range tests {
		got, err := Normalize(tt.in)
		if tt.wantErr {
			if !errors.Is(err, ErrInvalidCEP) {
				t.Errorf("Normalize(%q) erro = %v, esperado ErrInvalidCEP", tt.in, err)
			}
			continue
		}
		if err != nil || got != tt.want {
			t.Errorf("Normalize(%q) = %q, %v; esperado %q", tt.in, got, err, tt.want)
		}
	}
}
