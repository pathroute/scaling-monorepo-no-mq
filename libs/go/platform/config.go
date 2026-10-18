// Package platform holds what every service shares: config, logging and the HTTP server.
package platform

import (
	"fmt"
	"os"
	"strconv"
)

// Config is the runtime configuration common to all services, read from the environment.
type Config struct {
	Service string
	Addr    string
	Debug   bool
}

// LoadConfig reads PORT and DEBUG for the named service. getenv is os.Getenv outside tests.
func LoadConfig(service string, getenv func(string) string) (Config, error) {
	if getenv == nil {
		getenv = os.Getenv
	}
	c := Config{Service: service, Addr: ":8080"}
	if p := getenv("PORT"); p != "" {
		n, err := strconv.Atoi(p)
		if err != nil || n <= 0 || n > 65535 {
			return Config{}, fmt.Errorf("platform: invalid PORT %q", p)
		}
		c.Addr = ":" + p
	}
	if d := getenv("DEBUG"); d != "" {
		v, err := strconv.ParseBool(d)
		if err != nil {
			return Config{}, fmt.Errorf("platform: invalid DEBUG %q", d)
		}
		c.Debug = v
	}
	return c, nil
}
