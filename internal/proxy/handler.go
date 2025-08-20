package proxy

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/mail"

	client "github.com/wneessen/go-mail"
)

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

		subject := in.Header.Get("Subject")
		contentTypeHeader := in.Header.Get("Content-Type")
		log.Printf("[msg %s] Parsed headers: Subject=%q Content-Type=%q", id, subject, contentTypeHeader)

		// Copy headers
		for header, values := range in.Header {
			out.SetGenHeader(client.Header(header), values...)
		}

		out.SetMessageIDWithValue(id)

		// Copy body
		out.SetBodyWriter(client.ContentType(contentTypeHeader), func(w io.Writer) (int64, error) {
			return io.Copy(w, in.Body)
		})

		// Overwrite from address
		if overwriteFrom != "" {
			if err = out.From(overwriteFrom); err != nil {
				log.Printf("[msg %s] Failed to overwrite sender to %q: %v", id, overwriteFrom, err)
				return fmt.Errorf("failed to overwrite address: %s", err)
			}
		}

		if err = c.DialAndSend(out); err != nil {
			log.Printf("[msg %s] Upstream send failed: %v", id, err)
			return err
		}

		log.Printf("[msg %s] Successfully relayed", id)
		return nil
	}
}
