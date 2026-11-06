package mail

import (
	"context"
	"fmt"
	"time"

	mail "github.com/xhit/go-simple-mail/v2"
)

type SMTPClient struct {
	server    *mail.SMTPServer
	fromEmail string
}

func NewSMTPClient(host string, port int, username, password, encryption string, fromEmail string) (*SMTPClient, error) {
	server := mail.NewSMTPClient()
	server.Host = host
	server.Port = port
	server.Username = username
	server.Password = password
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	switch encryption {
	case "SSL":
		server.Encryption = mail.EncryptionSSL
	case "TLS", "STARTTLS":
		server.Encryption = mail.EncryptionSTARTTLS
	default:
		server.Encryption = mail.EncryptionNone
	}

	return &SMTPClient{
		server:    server,
		fromEmail: fromEmail,
	}, nil
}

func (s *SMTPClient) SendHTML(ctx context.Context, to string, subject string, htmlBody string) error {
	client, err := s.server.Connect()
	if err != nil {
		return fmt.Errorf("failed establishing active smtp socket session: %w", err)
	}
	defer client.Close()

	email := mail.NewMSG()
	email.SetFrom(s.fromEmail).
		AddTo(to).
		SetSubject(subject).
		SetBody(mail.TextHTML, htmlBody)

	if email.Error != nil {
		return fmt.Errorf("malformed mail envelope parameters: %w", email.Error)
	}

	errChan := make(chan error, 1)
	go func() {
		errChan <- email.Send(client)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errChan:
		if err != nil {
			return fmt.Errorf("smtp transmission error: %w", err)
		}
		return nil
	}
}
