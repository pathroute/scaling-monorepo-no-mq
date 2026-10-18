// Command issues runs the issues service.
package main

import (
	"log/slog"
	"os"

	"github.com/pathroute/suite/libs/go/aiclient"
	"github.com/pathroute/suite/libs/go/authz"
	"github.com/pathroute/suite/libs/go/platform"
	"github.com/pathroute/suite/services/issues"
)

func main() {
	cfg, err := platform.LoadConfig("issues", nil)
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}
	log := platform.NewLogger(os.Stdout, cfg)
	srv := issues.NewServer(authz.SignerFromEnv(), aiclient.Retry{Next: aiclient.Fake{}, Attempts: 3})
	if err := platform.Run(cfg, log, srv.Handler()); err != nil {
		log.Error("serve", "err", err)
		os.Exit(1)
	}
}
