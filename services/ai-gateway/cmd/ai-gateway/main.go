// Command ai-gateway runs the ai-gateway service.
package main

import (
	"log/slog"
	"os"

	"github.com/pathroute/suite/libs/go/aiclient"
	"github.com/pathroute/suite/libs/go/authz"
	"github.com/pathroute/suite/libs/go/platform"
	aigateway "github.com/pathroute/suite/services/ai-gateway"
)

func main() {
	cfg, err := platform.LoadConfig("ai-gateway", nil)
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}
	log := platform.NewLogger(os.Stdout, cfg)
	srv := aigateway.NewServer(authz.SignerFromEnv(), map[string]aigateway.Route{
		"small": {Provider: "fake", Client: aiclient.Retry{Next: aiclient.Fake{}, Attempts: 3}},
	}, aigateway.NewQuota(1_000_000))
	if err := platform.Run(cfg, log, srv.Handler()); err != nil {
		log.Error("serve", "err", err)
		os.Exit(1)
	}
}
