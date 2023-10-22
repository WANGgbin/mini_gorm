package gorm

import (
	"context"
)

type Session struct {
	DryRun            bool
	AllowGlobalUpdate bool
	AllowGlobalDelete bool
	Ctx               context.Context
}

func (db *DB) Session(config *Session) (tx *DB) {
	tx = db.newSession()
	if config.DryRun {
		tx.cfg.DryRun = true
	}

	if config.AllowGlobalUpdate {
		tx.cfg.AllowGlobalUpdate = true
	}

	if config.AllowGlobalDelete {
		tx.cfg.AllowGlobalDelete = true
	}

	if config.Ctx != nil {
		tx.stmt.ctx = config.Ctx
	}

	return tx
}

func (db *DB) WithContext(ctx context.Context) (tx *DB) {
	return db.Session(&Session{Ctx: ctx})
}
