package driver

import (
	"database/sql"
	"fmt"
)

func OpenDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		fmt.Println("Error pinging database:", err)
		return nil, err
	}
	return db, nil
}
