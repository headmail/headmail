package mailer

import (
	"context"

	"github.com/headmail/headmail/pkg/domain"
)

// Mailer is an abstraction for sending emails.
type Mailer interface {
	// Send delivers the given delivery. Implementations are responsible for
	// constructing message headers/body and performing the send.
	Send(ctx context.Context, d *domain.Delivery) error
}
