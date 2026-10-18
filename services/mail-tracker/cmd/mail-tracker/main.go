// Command mail-tracker runs the mail-tracker service.
package main

import (
	"log/slog"
	"os"

	"github.com/pathroute/suite/libs/go/aiclient"
	"github.com/pathroute/suite/libs/go/authz"
	"github.com/pathroute/suite/libs/go/platform"
	mailtracker "github.com/pathroute/suite/services/mail-tracker"
)

func main() {
	cfg, err := platform.LoadConfig("mail-tracker", nil)
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}
	log := platform.NewLogger(os.Stdout, cfg)
	srv := mailtracker.NewServer(authz.SignerFromEnv(), aiclient.Retry{Next: aiclient.Fake{}, Attempts: 3})
	if err := platform.Run(cfg, log, srv.Handler()); err != nil {
		log.Error("serve", "err", err)
		os.Exit(1)
	}
}
