package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/sethvargo/go-envconfig"

	"github.com/k911mipt/geolocation/loader"
	"github.com/k911mipt/geolocation/store"
)

type appConfig struct {
	CsvFilePath        string `env:"CSV_FILE_PATH"`
	BatchSize          int64  `env:"BATCH_SIZE"`
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
	err := runImport(ctx)
	stop()
	if err != nil {
		log.Fatalln(err)
	}
}

func runImport(ctx context.Context) error {
	var cfg appConfig
	if err := envconfig.Process(ctx, &cfg); err != nil {
		return err
	}

	st, err := store.NewStore(ctx, cfg.PgConnectionString)
	if err != nil {
		return err
	}

	file, err := os.Open(cfg.CsvFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	loader := loader.NewFromCSV(st)
	stats, err := loader.Run(ctx, file, cfg.CsvFilePath, int(cfg.BatchSize))
	if err != nil {
		return err
	}
	return stats.Print(os.Stdout)
}
