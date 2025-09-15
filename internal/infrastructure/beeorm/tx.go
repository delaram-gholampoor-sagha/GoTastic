package beeinfra

import (
	"context"

	beeorm "git.ice.global/packages/beeorm/v4"
)

type BeeTx struct{ db *beeorm.DB }

func newBeeTx(db *beeorm.DB) *BeeTx { return &BeeTx{db: db} }

// repository.Tx impl:
func (t *BeeTx) Commit(ctx context.Context) error   { t.db.Commit(); return nil }
func (t *BeeTx) Rollback(ctx context.Context) error { t.db.Rollback(); return nil }
