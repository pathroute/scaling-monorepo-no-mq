// Command notifications runs the notifications service.
package main

import (
	"log/slog"
	"os"

	"github.com/pathroute/suite/libs/go/platform"
	"github.com/pathroute/suite/services/notifications"
	"github.com/pathroute/suite/services/notifications/api"
)

func main() {
	cfg, err := platform.LoadConfig("notifications", nil)
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}
	log := platform.NewLogger(os.Stdout, cfg)
	srv := notifications.NewServer(map[string]notifications.Sender{
		api.ChannelEmail:   &notifications.LogSender{},
		api.ChannelPush:    &notifications.LogSender{},
		api.ChannelWebhook: &notifications.LogSender{},
	})
	if err := platform.Run(cfg, log, srv.Handler()); err != nil {
		log.Error("serve", "err", err)
		os.Exit(1)
	}
}
