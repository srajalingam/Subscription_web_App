package models

import (
	"context"
	"database/sql"
	"strings"
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
	IsRecurring    bool   `json:"is_recurring"`
	PlanID         string `json:"plan_id"`
	CreatedAt      string `json:"-"`
	UpdatedAt      string `json:"-"`
}

type Order struct {
	ID            int         `json:"id"`
	WidgetID      int         `json:"widget_id"`
	TransactionID int         `json:"transaction_id"`
	CustomerId    int         `json:"customer_id"`
	StatusID      int         `json:"status_id"`
	Quantity      int         `json:"quantity"`
	Amount        int         `json:"amount"`
	CreatedAt     string      `json:"-"`
	UpdatedAt     string      `json:"-"`
	Widget        Widget      `json:"widget"`
	Transaction   Transaction `json:"transaction"`
	Customer      Customer    `json:"customer"`
}

type Status struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"-"`
	UpdatedAt string `json:"-"`
}

type Transaction struct {
	ID                  int    `json:"id"`
	Amount              int    `json:"amount"`
	Currency            string `json:"currency"`
	LastFour            string `json:"last_four"`
	ExpiryMonth         int    `json:"expiry_month"`
	ExpiryYear          int    `json:"expiry_year"`
	PaymentIntent       string `json:"payment_intent"`
	PaymentMethod       string `json:"payment_method"`
	BankReturnCode      string `json:"bank_return_code"`
	TransactionStatusID int    `json:"transaction_status_id"`
	CreatedAt           string `json:"-"`
	UpdatedAt           string `json:"-"`
}

type User struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	CreatedAt string `json:"-"`
	UpdatedAt string `json:"-"`
}

type Customer struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	CreatedAt string `json:"-"`
	UpdatedAt string `json:"-"`
}

func (m *DBModel) GetWidgets(id int) (Widget, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var widgets Widget
	rows := m.DB.QueryRowContext(ctx, `
		SELECT 
			id, name , description, inventory_level, price, coalesce(image,''),is_recurring, plan_id, created_at, updated_at
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
		&widgets.IsRecurring,
		&widgets.PlanID,
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
				(amount, currency, last_four, bank_return_code,expiry_month,expiry_year, payment_intent, payment_method, transaction_status_id, created_at, updated_at)
				values(?,?,?,?,?,?,?,?,?,?,?)
			`
	result, err := m.DB.ExecContext(ctx, stmt,
		txn.Amount,
		txn.Currency,
		txn.LastFour,
		txn.BankReturnCode,
		txn.ExpiryMonth,
		txn.ExpiryYear,
		txn.PaymentIntent,
		txn.PaymentMethod,
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
				(widget_id, transaction_id, status_id, quantity,customer_id, amount, created_at, updated_at)
				values(?,?,?,?,?,?,?,?)
			`
	result, err := m.DB.ExecContext(ctx, stmt,
		order.WidgetID,
		order.TransactionID,
		order.StatusID,
		order.Quantity,
		order.CustomerId,
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

func (m *DBModel) InsertCustomer(customer Customer) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `insert into customers
				(first_name, last_name, email, created_at, updated_at)
				values(?,?,?,?,?)
			`
	result, err := m.DB.ExecContext(ctx, stmt,
		customer.FirstName,
		customer.LastName,
		customer.Email,
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

// GetUserByEmail returns a user by email address
func (m *DBModel) GetUserByEmail(email string) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	email = strings.ToLower(email)
	var user User

	row := m.DB.QueryRowContext(ctx, `
		SELECT
			id, first_name, last_name, email, password, created_at, updated_at
		FROM
			users	
		WHERE email = ?`, email)
	err := row.Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return user, err
	}
	return user, nil
}

// UpdateUserPassword
func (m *DBModel) UpdateUserPassword(u User, hash string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `update users set password = ? where id = ?`
	_, err := m.DB.ExecContext(ctx, stmt, hash, u.ID)
	if err != nil {
		return err
	}
	return nil
}

// GetAllOrders returns all the orders in the database
func (m *DBModel) GetAllOrders() ([]Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var orders []Order

	query := `
	    select 
			o.id,o.widget_id,o.transaction_id,o.customer_id,o.status_id,o.quantity,o.amount,o.created_at,o.updated_at,w.id,w.name,
			t.id,t.amount,t.currency,t.last_four,t.expiry_month,t.expiry_year,t.payment_intent,t.bank_return_code,
			c.id,c.first_name,c.last_name,c.email
			
			from 
				orders o
				left join widgets w on (o.widget_id = w.id)
				left join transactions t on (o.transaction_id=t.id)
				left join customers c on(o.customer_id=c.id)
			where
				w.is_recurring=0
		order by
			o.created_at desc
	`
	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var o Order
		err := rows.Scan(
			&o.ID,
			&o.WidgetID,
			&o.TransactionID,
			&o.CustomerId,
			&o.StatusID,
			&o.Quantity,
			&o.Amount,
			&o.CreatedAt,
			&o.UpdatedAt,
			&o.Widget.ID,
			&o.Widget.Name,
			&o.Transaction.ID,
			&o.Transaction.Amount,
			&o.Transaction.Currency,
			&o.Transaction.LastFour,
			&o.Transaction.ExpiryMonth,
			&o.Transaction.ExpiryYear,
			&o.Transaction.PaymentIntent,
			&o.Transaction.BankReturnCode,
			&o.Customer.ID,
			&o.Customer.FirstName,
			&o.Customer.LastName,
			&o.Customer.Email,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}

// GetOrderByID returns an order by ID
func (m *DBModel) GetOrderByID(id int) (Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var o Order
	query := `
	    select 
			o.id,o.widget_id,o.transaction_id,o.customer_id,o.status_id,o.quantity,o.amount,o.created_at,o.updated_at,w.id,w.name,
			t.id,t.amount,t.currency,t.last_four,t.expiry_month,t.expiry_year,t.payment_intent,t.bank_return_code,
			c.id,c.first_name,c.last_name,c.email
			
			from 
				orders o
				left join widgets w on (o.widget_id = w.id)
				left join transactions t on (o.transaction_id=t.id)
				left join customers c on(o.customer_id=c.id)
			where
				o.id=?
	`
	row := m.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&o.ID,
		&o.WidgetID,
		&o.TransactionID,
		&o.CustomerId,
		&o.StatusID,
		&o.Quantity,
		&o.Amount,
		&o.CreatedAt,
		&o.UpdatedAt,
		&o.Widget.ID,
		&o.Widget.Name,
		&o.Transaction.ID,
		&o.Transaction.Amount,
		&o.Transaction.Currency,
		&o.Transaction.LastFour,
		&o.Transaction.ExpiryMonth,
		&o.Transaction.ExpiryYear,
		&o.Transaction.PaymentIntent,
		&o.Transaction.BankReturnCode,
		&o.Customer.ID,
		&o.Customer.FirstName,
		&o.Customer.LastName,
		&o.Customer.Email,
	)
	if err != nil {
		return o, err
	}
	return o, nil
}

// GetAllSubscriptions returns all the subscriptions in the database
func (m *DBModel) GetAllSubscriptions() ([]Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var orders []Order
	query := `
	    select 
			o.id,o.widget_id,o.transaction_id,o.customer_id,o.status_id,o.quantity,o.amount,o.created_at,o.updated_at,w.id,w.name,
			t.id,t.amount,t.currency,t.last_four,t.expiry_month,t.expiry_year,t.payment_intent,t.bank_return_code,
			c.id,c.first_name,c.last_name,c.email

			from 
				orders o
				left join widgets w on (o.widget_id = w.id)
				left join transactions t on (o.transaction_id=t.id)
				left join customers c on(o.customer_id=c.id)
			where
				w.is_recurring=1
		order by
			o.created_at desc
	`
	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var o Order
		err := rows.Scan(
			&o.ID,
			&o.WidgetID,
			&o.TransactionID,
			&o.CustomerId,
			&o.StatusID,
			&o.Quantity,
			&o.Amount,
			&o.CreatedAt,
			&o.UpdatedAt,
			&o.Widget.ID,
			&o.Widget.Name,
			&o.Transaction.ID,
			&o.Transaction.Amount,
			&o.Transaction.Currency,
			&o.Transaction.LastFour,
			&o.Transaction.ExpiryMonth,
			&o.Transaction.ExpiryYear,
			&o.Transaction.PaymentIntent,
			&o.Transaction.BankReturnCode,
			&o.Customer.ID,
			&o.Customer.FirstName,
			&o.Customer.LastName,
			&o.Customer.Email,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}

// GetSubscriptionByID returns a subscription by ID
func (m *DBModel) GetSubscriptionByID(id int) (Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var o Order
	query := `
	    select 
			o.id,o.widget_id,o.transaction_id,o.customer_id,o.status_id,o.quantity,o.amount,o.created_at,o.updated_at,w.id,w.name,
			t.id,t.amount,t.currency,t.last_four,t.expiry_month,t.expiry_year,t.payment_intent,t.bank_return_code,
			c.id,c.first_name,c.last_name,c.email
			from 
				orders o
				left join widgets w on (o.widget_id = w.id)
				left join transactions t on (o.transaction_id=t.id)
				left join customers c on(o.customer_id=c.id)
			where
				o.id=?
				and w.is_recurring=1
	`
	row := m.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&o.ID,
		&o.WidgetID,
		&o.TransactionID,
		&o.CustomerId,
		&o.StatusID,
		&o.Quantity,
		&o.Amount,
		&o.CreatedAt,
		&o.UpdatedAt,
		&o.Widget.ID,
		&o.Widget.Name,
		&o.Transaction.ID,
		&o.Transaction.Amount,
		&o.Transaction.Currency,
		&o.Transaction.LastFour,
		&o.Transaction.ExpiryMonth,
		&o.Transaction.ExpiryYear,
		&o.Transaction.PaymentIntent,
		&o.Transaction.BankReturnCode,
		&o.Customer.ID,
		&o.Customer.FirstName,
		&o.Customer.LastName,
		&o.Customer.Email,
	)
	if err != nil {
		return o, err
	}
	return o, nil
}

// update order status
func (m *DBModel) UpdateOrderStatus(id int, statusID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	stmt := `update orders set status_id = ? where id = ?`
	_, err := m.DB.ExecContext(ctx, stmt, statusID, id)
	if err != nil {
		return err
	}
	return nil
}
