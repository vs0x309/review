package main

import (
	"context"
	"exchanges/pkg/exchange/bybit"
	"exchanges/pkg/exchange/gateio"
	"exchanges/pkg/exchange/okx"
	"exchanges/pkg/server"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	logFile string
	addr    string
)

func init() {
	flag.StringVar(&logFile, "logFile", "", "Path to log file")
	flag.StringVar(&addr, "addr", ":8080", "server addres")
	flag.Parse()
}

func main() {
	if len(logFile) > 0 {
		logFile, err := os.Open(logFile)
		if err != nil {
			log.Fatal(err)
		}

		defer func() {
			_ = logFile.Close()
		}()

		log.SetOutput(logFile)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill, syscall.SIGTERM)
	defer stop()

	srv := server.NewServer()

	srv.SetExchange(gateio.NewAPI())
	srv.SetExchange(bybit.NewAPI())
	srv.SetExchange(okx.NewAPI())

	log.Print("Application start")

	if err := srv.Run(ctx, addr); err != nil {
		log.Fatalf("Application stop: %v", err)
	}

	log.Print("Application stop")
}
