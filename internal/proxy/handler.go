package proxy

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/mail"
	"slices"

	client "github.com/wneessen/go-mail"
)

var skippedHeaders = []string{
	string(client.HeaderFrom),
	string(client.HeaderTo),
}

func mailHandler(c *client.Client, overwriteFrom string) func(net.Addr, string, []string, []byte) error {
	return func(origin net.Addr, from string, to []string, data []byte) error {
		// Create new outgoing message and generate Message-ID early
		out := client.NewMsg()
		out.SetMessageID()
		id := out.GetMessageID()

		log.Printf("[msg %s] Incoming message: origin=%s envelope_from=%q rcpt_count=%d bytes=%d", id, origin.String(), from, len(to), len(data))

		// Read message
		in, err := mail.ReadMessage(bytes.NewReader(data))
		if err != nil {
			log.Printf("[msg %s] Failed to read message: %v", id, err)
			return fmt.Errorf("failed to read message: %s", err)
		}

		subject := in.Header.Get(string(client.HeaderSubject))
		contentTypeHeader := in.Header.Get(string(client.HeaderContentType))
		if contentTypeHeader == "" {
			contentTypeHeader = "text/plain; charset=us-ascii"
		}
		log.Printf("[msg %s] Parsed headers: Subject=%q Content-Type=%q", id, subject, contentTypeHeader)

		// Copy headers
		for header, values := range in.Header {
			if slices.Contains(skippedHeaders, header) {
				continue
			}
			out.SetGenHeader(client.Header(header), values...)
		}

		out.SetMessageIDWithValue(id)

		// Copy body
		out.SetBodyWriter(client.ContentType(contentTypeHeader), func(w io.Writer) (int64, error) {
			return io.Copy(w, in.Body)
		})

		// Overwrite sender address
		if overwriteFrom != "" {
			from = overwriteFrom
		}

		// Set FROM address
		if err = out.From(from); err != nil {
			log.Printf("[msg %s] Failed to set sender address: %v", id, err)
			return fmt.Errorf("failed to set sender address: %s", err)
		}

		// Set TO addresses
		if err = out.To(to...); err != nil {
			log.Printf("[msg %s] Failed to set receiver addresses: %v", id, err)
			return fmt.Errorf("failed to set receiver addresses: %s", err)
		}

		// Send message
		if err = c.DialAndSend(out); err != nil {
			log.Printf("[msg %s] Upstream send failed: %v", id, err)
			return err
		}

		log.Printf("[msg %s] Successfully relayed", id)
		return nil
	}
}
