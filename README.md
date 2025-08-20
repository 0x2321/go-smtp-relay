# go-smtp-relay

Lightweight local SMTP relay that accepts mail on a local port and forwards it to a configured upstream SMTP server.

It is intended as a simple replacement for sendmail/postfix, to provide a local SMTP endpoint for applications, containers, and scripts.

## Features
- Listens on a local address/port and relays to an upstream SMTP server
- Optional SMTP authentication to upstream (username/password)
- Optional sender overwrite (envelope/from)

## Installation
```bash
# Quick install
curl -fsSL https://raw.githubusercontent.com/0x2321/go-smtp-relay/main/scripts/install.sh | bash
```
```bash
# Or download, review, then run
curl -fsSL https://raw.githubusercontent.com/0x2321/go-smtp-relay/main/scripts/install.sh -o install.sh
less install.sh
bash install.sh
```

## Configuration
Configuration can be provided via:
1) Config file (default path: /etc/smtp-relay.yaml)
2) Environment variables (prefix: SMTP_RELAY_)
3) CLI flags

### Config file example
```yaml
# /etc/smtp-relay.yaml
listen:
  address: 127.0.0.1
  port: 25

upstream:
  host: smtp.example.com
  port: 587
  user: myuser@example.com
  password: "change-me"

# Optional overwrite of the sender envelope (omit or leave empty to disable)
overwrite:
  sender: noreply@example.com
```

## systemd service
A sample unit file is provided at smtp-relay.service. It uses systemd notify readiness.

Install as a service:
```bash
# Build and install binary
sudo install -m 0755 smtp-relay /usr/local/bin/smtp-relay

# Install service unit
sudo install -m 0644 smtp-relay.service /etc/systemd/system/smtp-relay.service

# Reload and enable
sudo systemctl daemon-reload
sudo systemctl enable --now smtp-relay

# Check status and logs
systemctl status smtp-relay
journalctl -u smtp-relay -f
```

## How it works (internals)
- The server listens via github.com/mhale/smtpd and handles incoming messages.
- Messages are parsed using net/mail; headers are largely preserved and the message body is streamed through.
- A new Message-ID is generated (and set) for the outbound message.
- If overwrite-sender is set, the sender address is replaced before relaying.
- Upstream delivery is performed using github.com/wneessen/go-mail, with optional SMTP auth.
- The service notifies systemd when ready/stopping via github.com/coreos/go-systemd/daemon.

## Security notes
- Protect your config file; it may contain credentials. Restrict permissions to root where appropriate.
- Prefer environment variables or a secret manager for sensitive values in production.
- Only bind to 0.0.0.0 if you understand the exposure. By default the listener binds to 127.0.0.1.

## Troubleshooting
- "Please specify the upstream SMTP host and port": ensure --upstream-host and --upstream-port (or env/config) are set.
- Connection/auth errors: validate host, port, credentials, and that your upstream accepts relaying from your source IP.

