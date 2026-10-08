package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hendelsantos/Goping/internal/monitor"
	"github.com/hendelsantos/Goping/internal/web"
)

func main() {
	address := flag.String("addr", defaultAddress(), "endereço HTTP do servidor")
	dataPath := flag.String("data", "data/goping.json", "arquivo de persistência JSON")
	staticDirectory := flag.String("web", ".", "diretório dos arquivos do frontend")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	store, err := monitor.NewStore(*dataPath)
	if err != nil {
		logger.Error("não foi possível abrir o armazenamento", "error", err)
		os.Exit(1)
	}

	server := &http.Server{
		Addr:              *address,
		Handler:           web.NewServer(store, logger, *staticDirectory),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	shutdownSignal, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("GoPing iniciado", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("servidor HTTP encerrado com erro", "error", err)
			os.Exit(1)
		}
	}()

	<-shutdownSignal.Done()
	logger.Info("encerrando GoPing")

	shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownContext); err != nil {
		logger.Error("encerramento forçado", "error", err)
		os.Exit(1)
	}
}

func defaultAddress() string {
	if port := os.Getenv("PORT"); port != "" {
		return ":" + port
	}
	return ":8080"
}
