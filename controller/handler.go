package controller

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/pkg/errors"
)

// TXHandler is handler for working with transaction.
// This is wrapper function for commit and rollback.
func TXHandler(db *sql.DB, f func(*sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return errors.Wrap(err, "start transaction failed")
	}
	defer func() {
		if err := recover(); err != nil {
			tx.Rollback()
			log.Print("rollback operation.")
			return
		}
	}()
	if err := f(tx); err != nil {
		return errors.Wrap(err, "transaction: operation failed")
	}
	return nil
}

func Error(w http.ResponseWriter, err error, code int) {
	http.Error(w, fmt.Sprintf("%s", err), code)
}
