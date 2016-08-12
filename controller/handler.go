package controller

import (
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
)

// TXHandler is handler for working with transaction.
// This is wrapper function for commit and rollback.
func TXHandler(c *gin.Context, db *sql.DB, f func(*sql.Tx) error) {
	tx, err := db.Begin()
	if err != nil {
		c.JSON(500, gin.H{"err": "start transaction failed"})
		return
	}
	defer func() {
		if err := recover(); err != nil {
			tx.Rollback()
			log.Print("rollback operation.")
			return
		}
	}()
	if err := f(tx); err != nil {
		c.JSON(500, gin.H{"err": "operation failed"})
		return
	}
}
