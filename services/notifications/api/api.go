// Package api defines the notifications service's wire types.
package api

import "time"

// Channels a notification can go out on.
const (
	ChannelEmail   = "email"
	ChannelPush    = "push"
	ChannelWebhook = "webhook"
)

// Notify is the body of POST /notify.
type Notify struct {
	User     string   `json:"user"`
	Channels []string `json:"channels"`
	Message  string   `json:"message"`
}

// Delivery is one notification sent on one channel.
type Delivery struct {
	ID      string    `json:"id"`
	User    string    `json:"user"`
	Channel string    `json:"channel"`
	Message string    `json:"message"`
	OK      bool      `json:"ok"`
	Error   string    `json:"error,omitempty"`
	SentAt  time.Time `json:"sent_at"`
}
