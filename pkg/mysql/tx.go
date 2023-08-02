package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
)

func Tx(ctx context.Context, db *sqlx.DB, wrapper func(tx *sqlx.Tx) error) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}

	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.WithError(err).Errorf("tx rollback")
		}
	}()

	if err := wrapper(tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		log.WithError(err).Errorf("tx commit")
		return err
	}

	return nil
}
