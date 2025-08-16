// Copyright 2025 JC-Lab
// SPDX-License-Identifier: AGPL-3.0-or-later

package mailer

import (
	"context"
	"strings"

	"github.com/headmail/headmail/pkg/domain"
)

// Mailer is an abstraction for sending emails.
type Mailer interface {
	// Send delivers the given delivery. Implementations are responsible for
	// constructing message headers/body and performing the send.
	Send(ctx context.Context, d *domain.Delivery) error
}

const HeadmailDeliveryHeaderName = "X-Headmail-Delivery"

var HeadmailDeliveryHeaderNameAsLower = strings.ToLower(HeadmailDeliveryHeaderName)
