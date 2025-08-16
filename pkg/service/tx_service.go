// Copyright 2025 JC-Lab
// SPDX-License-Identifier: AGPL-3.0-or-later

package service

import (
	"github.com/headmail/headmail/pkg/repository"
)

// TxService provides business logic for transactional email sending.
type TxService struct {
	deliveryRepo repository.DeliveryRepository
	// May need other dependencies like a mailer service
}

// NewTxService creates a new TxService.
func NewTxService(deliveryRepo repository.DeliveryRepository) *TxService {
	return &TxService{deliveryRepo: deliveryRepo}
}

// TODO: Implement transactional email service methods
