// Command admin-api runs the admin-api service.
package main

import (
	"log/slog"
	"os"

	"github.com/pathroute/suite/libs/go/authz"
	"github.com/pathroute/suite/libs/go/platform"
	adminapi "github.com/pathroute/suite/services/admin-api"
)

func main() {
	cfg, err := platform.LoadConfig("admin-api", nil)
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}
	log := platform.NewLogger(os.Stdout, cfg)
	srv := adminapi.NewServer(authz.SignerFromEnv())
	if err := platform.Run(cfg, log, srv.Handler()); err != nil {
		log.Error("serve", "err", err)
		os.Exit(1)
	}
}
