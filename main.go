package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/IHorvalds/mailcode-v2/pkg/persistence"
	"github.com/IHorvalds/mailcode-v2/pkg/service"

	"github.com/IHorvalds/mailcode-v2/pkg/configs"
)

func startService(cfg service.Config) {
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, os.Interrupt)

	db, err := persistence.NewRepository(cfg.DbPath)
	if err != nil {
		log.Panicf("Failed to connect to database at %s: %s", cfg.DbPath, err)
	}

	srv := service.NewServer(fmt.Sprintf("127.0.0.1:%d", cfg.Port))
	configs.Register(srv.Mux, db)
	c := srv.ListenAndServe()
	select {
	case <-stopCh:
		log.Println("Shutting down the server...")
	case err := <-c:
		log.Printf("Server stopped unexpectedly: %s", err)
	}
}

func main() {
	// TODO: Read this from a file in $XDG_CONFIG_HOME or $HOME
	cfg := service.Config{
		Port: 8080,
	}
	startService(cfg)
}
