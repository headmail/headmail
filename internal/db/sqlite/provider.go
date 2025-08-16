// Copyright 2025 JC-Lab
// SPDX-License-Identifier: AGPL-3.0-or-later

package sqlite

import (
	"github.com/headmail/headmail/pkg/config"
	"github.com/headmail/headmail/pkg/db"
	"github.com/headmail/headmail/pkg/repository"
)

// provider implements the db.Provider interface for SQLite.
type provider struct{}

func init() {
	db.RegisterDefaultProvider("sqlite", &provider{})
}

// New creates a new SQLite database connection.
func (p *provider) New(cfg config.DatabaseConfig) (repository.DB, error) {
	return New(cfg)
}
