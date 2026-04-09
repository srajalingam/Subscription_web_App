package models

import "database/sql"

// DBModel is the wrapper for sql.DB connection pool
type DBModel struct {
	DB *sql.DB
}

// Model is the wrapper for all the models
type Model struct {
	DB DBModel
}

// NewModels returns a new instance of Model with the given sql.DB connection pool
func NewModels(db *sql.DB) Model {
	return Model{
		DB: DBModel{DB: db},
	}
}

//widget is the type for the widget table in the database
type Widget struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	InventortLevel int    `json:"inventory_level"`
	Price          int    `json:"price"`
	CreatedAt      string `json:"-"`
	UpdatedAt      string `json:"-"`
}
