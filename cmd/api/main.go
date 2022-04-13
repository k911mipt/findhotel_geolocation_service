package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/sethvargo/go-envconfig"

	"github.com/k911mipt/geolocation/server"
	"github.com/k911mipt/geolocation/store"
)

type appConfig struct {
	ApiPort            string `env:"API_PORT"`
	PgConnectionString string `env:"PG_CONNECTION_STRING"`
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer func() {
		stop()
		if r := recover(); r != nil {
			log.Fatalf("application panic %s %s", "panic", r)
		}
	}()
	err := runServer(ctx)
	stop()
	if err != nil {
		log.Fatalln(err)
	}
}

func runServer(ctx context.Context) error {
	var cfg appConfig
	if err := envconfig.Process(ctx, &cfg); err != nil {
		return err
	}

	st, err := store.NewStore(ctx, cfg.PgConnectionString)
	if err != nil {
		return err
	}

	server := server.New(st)
	return server.Run(ctx, cfg.ApiPort)
}
