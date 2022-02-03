package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/kelseyhightower/envconfig"
	"github.com/kofoworola/gojson/logging"
)

type Config struct {
	Port        string `default:"8080"`
	Environment string `envconfig:"env" default:"development"`
	LogPath     string `default:"/var/log/"`
}

func main() {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatal(err)
	}

	logger, err := logging.New(cfg.Environment, cfg.LogPath)
	if err != nil {
		panic(err)
	}

	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-ch
		cancel()
	}()

	handler, err := NewHandler(logger)
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/", handler)

	go func() {
		log.Printf("listening on :%s\n", cfg.Port)
		if err := http.ListenAndServe(":"+cfg.Port, nil); err != nil {
			log.Fatal(err)
		}

	}()
	<-ctx.Done()
	if err := logger.Close(); err != nil {
		fmt.Printf("[WARN] error closing log file: %v", err)
	}
}
