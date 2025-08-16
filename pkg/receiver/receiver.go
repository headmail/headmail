// Copyright 2025 JC-Lab
// SPDX-License-Identifier: AGPL-3.0-or-later

package receiver

import "context"

type Event struct {
	DeliveryID        string
	MessageID         string
	Subject           string
	BouncedRecipients []string
	Reason            string
}

// Receiver defines an interface for inbound mail receivers (bounce processors, webhooks, etc).
type Receiver interface {
	// Start starts background processing (returns when started or on error).
	Start(ctx context.Context) (chan *Event, error)
	// Stop requests graceful stop and waits for completion.
	Stop(ctx context.Context) error
}
