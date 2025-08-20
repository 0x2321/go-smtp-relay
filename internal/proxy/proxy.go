package proxy

import (
	"fmt"
	"log"

	"github.com/coreos/go-systemd/daemon"
	"github.com/mhale/smtpd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	client "github.com/wneessen/go-mail"
)

func Run(cmd *cobra.Command, args []string) {
	listenAddr := viper.GetString("listen.address")
	listenPort := viper.GetInt("listen.port")
	upstreamHost := viper.GetString("upstream.host")
	upstreamPort := viper.GetInt("upstream.port")
	upstreamUser := viper.GetString("upstream.user")
	upstreamPass := viper.GetString("upstream.password")
	overwriteSender := viper.GetString("overwrite.sender")

	if upstreamHost == "" || upstreamPort == 0 {
		log.Fatalf("Please specify the upstream SMTP host and port (flags: --upstream-host and --upstream-port, env: SMTP_RELAY_UPSTREAM_HOST/PORT, or config file)")
	}

	// Announce effective configuration (without exposing secrets)
	log.Printf("smtp-relay starting: listen=%s:%d upstream=%s:%d auth=%t", listenAddr, listenPort, upstreamHost, upstreamPort, upstreamUser != "" || upstreamPass != "")
	if upstreamUser != "" {
		log.Printf("Using SMTP authentication with username %q", upstreamUser)
	} else {
		log.Printf("No SMTP authentication configured (no username provided)")
	}

	if overwriteSender != "" {
		log.Printf("Overwriting sender to %s", overwriteSender)
	}

	// Create smtp client
	log.Printf("Initializing upstream SMTP client for %s:%d", upstreamHost, upstreamPort)
	c, err := client.NewClient(upstreamHost,
		client.WithPort(upstreamPort),
		client.WithSMTPAuth(client.SMTPAuthAutoDiscover),
		client.WithUsername(upstreamUser),
		client.WithPassword(upstreamPass),
	)
	if err != nil {
		log.Fatalf("failed to create client: %s", err)
	}
	defer c.Close()

	// Start server
	_, _ = daemon.SdNotify(false, daemon.SdNotifyReady)
	addr := fmt.Sprintf("%s:%d", listenAddr, listenPort)
	log.Printf("Starting SMTP server on %s", addr)

	if err = smtpd.ListenAndServe(
		addr,
		mailHandler(c, overwriteSender),
		"smtp-relay",
		"",
	); err != nil {
		_, _ = daemon.SdNotify(false, daemon.SdNotifyStopping)
		log.Fatalf("failed to run server: %s", err)
	}
}
