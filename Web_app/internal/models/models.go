package models

import (
	"context"
	"database/sql"
	"time"
)

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

// widget is the type for the widget table in the database
type Widget struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	InventortLevel int    `json:"inventory_level"`
	Price          int    `json:"price"`
	Image          string `json:"image"`
	CreatedAt      string `json:"-"`
	UpdatedAt      string `json:"-"`
}

type Order struct {
	ID            int    `json:"id"`
	WidgetID      int    `json:"widget_id"`
	TransactionID int    `json:"transaction_id"`
	CustomerId    int    `json:"customer_id"`
	StatusID      int    `json:"status_id"`
	Quantity      int    `json:"quantity"`
	Amount        int    `json:"amount"`
	CreatedAt     string `json:"-"`
	UpdatedAt     string `json:"-"`
}

type Status struct {
	ID        int    `json:"id"`
	Name      int    `json:"name"`
	CreatedAt string `json:"-"`
	UpdatedAt string `json:"-"`
}

type Transaction struct {
	ID                  int    `json:"id"`
	Amount              int    `json:"amount"`
	Currency            string `json:"currency"`
	LastFour            string `json:"last_four"`
	BankReturnCode      string `json:"bank_return_code"`
	TransactionStatusID int    `json:"transaction_status_id"`
	CreatedAt           string `json:"-"`
	UpdatedAt           string `json:"-"`
}

type User struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type Customer struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

func (m *DBModel) GetWidgets(id int) (Widget, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var widgets Widget
	rows := m.DB.QueryRowContext(ctx, `
		SELECT 
			id, name , description, inventory_level, price, coalesce(image,''),	created_at, updated_at
		FROM 
			widgets 
		WHERE id = ?`, id)
	err := rows.Scan(
		&widgets.ID,
		&widgets.Name,
		&widgets.Description,
		&widgets.InventortLevel,
		&widgets.Price,
		&widgets.Image,
		&widgets.CreatedAt,
		&widgets.UpdatedAt,
	)
	if err != nil {
		return widgets, err
	}
	return widgets, nil
}

func (m *DBModel) InsertTransaction(txn Transaction) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `insert into transactions
				(amount, currency, last_four, bank_return_code, transaction_status_id, created_at, updated_at)
				values(?,?,?,?,?,?)
			`
	result, err := m.DB.ExecContext(ctx, stmt,
		txn.Amount,
		txn.Currency,
		txn.LastFour,
		txn.BankReturnCode,
		txn.TransactionStatusID,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (m *DBModel) InsertOrder(order Order) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `insert into orders
				(widget_id, transaction_id, status_id, quantity, amount, created_at, updated_at)
				values(?,?,?,?,?,?)
			`
	result, err := m.DB.ExecContext(ctx, stmt,
		order.WidgetID,
		order.TransactionID,
		order.StatusID,
		order.Quantity,
		order.Amount,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}
