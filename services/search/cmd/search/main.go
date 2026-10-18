// Command search runs the search service.
package main

import (
	"log/slog"
	"os"

	"github.com/pathroute/suite/libs/go/authz"
	"github.com/pathroute/suite/libs/go/platform"
	"github.com/pathroute/suite/services/search"
)

func main() {
	cfg, err := platform.LoadConfig("search", nil)
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}
	log := platform.NewLogger(os.Stdout, cfg)
	srv := search.NewServer(authz.SignerFromEnv())
	if err := platform.Run(cfg, log, srv.Handler()); err != nil {
		log.Error("serve", "err", err)
		os.Exit(1)
	}
}
