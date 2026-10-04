// Comando cep busca um endereço consultando BrasilAPI e ViaCEP em paralelo
// e exibe a resposta da API mais rápida.
//
// Uso:
//
//	go run ./cmd/cep [-timeout 1s] <cep>
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/seu-usuario/desafio-multithreading/internal/cep"
)

const defaultTimeout = time.Second

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := run(ctx, os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		stop()
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, out io.Writer) error {
	fs := flag.NewFlagSet("cep", flag.ContinueOnError)
	timeout := fs.Duration("timeout", defaultTimeout, "tempo limite para a resposta das APIs")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Uso: %s [-timeout 1s] <cep>\n", fs.Name())
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		fs.Usage()
		return errors.New("informe exatamente um CEP")
	}

	code, err := cep.Normalize(fs.Arg(0))
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, *timeout)
	defer cancel()

	client := &http.Client{}
	res, err := cep.FetchFastest(ctx, code,
		cep.NewBrasilAPI(client),
		cep.NewViaCEP(client),
	)
	if err != nil {
		return err
	}

	printResult(out, res)
	return nil
}

func printResult(w io.Writer, r cep.Result) {
	a := r.Address
	fmt.Fprintf(w, "API mais rápida: %s\n", r.Source)
	fmt.Fprintf(w, "  CEP:        %s\n", a.CEP)
	fmt.Fprintf(w, "  Logradouro: %s\n", a.Street)
	fmt.Fprintf(w, "  Bairro:     %s\n", a.Neighborhood)
	fmt.Fprintf(w, "  Cidade:     %s\n", a.City)
	fmt.Fprintf(w, "  UF:         %s\n", a.State)
}
