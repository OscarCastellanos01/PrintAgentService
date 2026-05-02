package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/kardianos/service"

	"print-agent/internal/config"
	"print-agent/internal/handlers"
	"print-agent/internal/middleware"
)

type Program struct {
	server *http.Server
}

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	serviceConfig := &service.Config{
		Name: config.ServiceName,
		DisplayName: config.ServiceDisplayName,
		Description: config.ServiceDescription,
	}

	program := &Program{}

	s, err := service.New(program, serviceConfig)
	if err != nil {
		log.Fatal(err)
	}

	if len(os.Args) > 1 {
		command := os.Args[1]
		if err := service.Control(s, command); err != nil {
			log.Fatalf("Error ejecutando comando %s: %v", command, err)
		}
		fmt.Printf("Comando %s ejecutado correctamente\n", command)
		return
	}

	if err := s.Run(); err != nil {
		log.Fatal(err)
	}
}

func (p *Program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *Program) run() {
	mux := http.NewServeMux()
	mux.HandleFunc("/status", handlers.HandleStatus)
	mux.HandleFunc("/printers", handlers.HandlePrinters)
	mux.HandleFunc("/print/pdf", handlers.HandlePrintPDF)
	mux.HandleFunc("/print/zpl/raw", handlers.HandlePrintZPLRaw)

	p.server = &http.Server{
		Addr: config.Addr(),
		Handler: middleware.CORS(mux),
		ReadTimeout: config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout: config.IdleTimeout,
	}

	slog.Info("Print Agent iniciado", "addr", p.server.Addr)

	if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("error iniciando servidor", "err", err)
	}
}

func (p *Program) Stop(s service.Service) error {
	slog.Info("Deteniendo Print Agent...")

	if p.server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.ShutdownTimeout)
	defer cancel()

	return p.server.Shutdown(ctx)
}
