package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// DBModel is the type for database connection values
type DBModel struct {
	DB *sql.DB
}

// Models is the wrapper for all models
type Models struct {
	DB DBModel
}

// NewModels returns a model type with database connection pool
func NewModels(db *sql.DB) Models {
	return Models{
		DB: DBModel{DB: db},
	}
}

// Widget is the type for all widgets
type Widget struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	InventoryLevel int       `json:"inventory_level"`
	Price          int       `json:"price"`
	Image          string    `json:"image"`
	IsRecurring    bool      `json:"is_recurring"`
	PlanID         string    `json:"plan_id"`
	CreatedAt      time.Time `json:"-"`
	UpdatedAt      time.Time `json:"-"`
}

// Order is the type for all orders
type Order struct {
	ID            int         `json:"id"`
	WidgetID      int         `json:"widget_id"`
	TransactionID int         `json:"transaction_id"`
	CustomerID    int         `json:"customer_id"`
	StatusID      int         `json:"status_id"`
	Quantity      int         `json:"quantity"`
	Amount        int         `json:"amount"`
	CreatedAt     time.Time   `json:"-"`
	UpdatedAt     time.Time   `json:"-"`
	Widget        Widget      `json:"widget"`
	Transaction   Transaction `json:"transaction"`
	Customer      Customer    `json:"customer"`
}

// Status is the type for order statuses
type Status struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// TransactionStatus is the type for transaction statuses
type TransactionStatus struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// Transaction is the type for transactions
type Transaction struct {
	ID                  int       `json:"id"`
	Amount              int       `json:"amount"`
	Currency            string    `json:"currency"`
	LastFour            string    `json:"last_four"`
	ExpiryMonth         int       `json:"expiry_month"`
	ExpiryYear          int       `json:"expiry_year"`
	BankReturnCode      string    `json:"bank_return_code"`
	TransactionStatusID int       `json:"transaction_status_id"`
	PaymentInent        string    `json:"payment_intent"`
	PaymentMethod       string    `json:"payment_method"`
	CreatedAt           time.Time `json:"-"`
	UpdatedAt           time.Time `json:"-"`
}

// User is the type for users
type User struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// Customer is the type for customers
type Customer struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

func (m *DBModel) GetWidget(id int) (Widget, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var widget Widget

	row := m.DB.QueryRowContext(ctx,
		`SELECT id, 
				name,
				description,
				inventory_level,
				price,
				coalesce(image, ''),
				is_recurring,
				plan_id,
				created_at,
				updated_at
		 FROM	widgets 
		 WHERE 	id = ?`, id)

	err := row.Scan(
		&widget.ID,
		&widget.Name,
		&widget.Description,
		&widget.InventoryLevel,
		&widget.Price,
		&widget.Image,
		&widget.IsRecurring,
		&widget.PlanID,
		&widget.CreatedAt,
		&widget.UpdatedAt)

	if err != nil {
		return widget, err
	}

	return widget, nil
}

// InsertTransaction inserts a new row into the Transactions table
// and returns the primary key for the newly created row and an
// error (if any)
func (m *DBModel) InsertTransaction(txn Transaction) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
		INSERT INTO transactions (
			amount,
			currency,
			last_four,
			bank_return_code,
			transaction_status_id,
			expiry_month,
			expiry_year,
			payment_intent,
			payment_method,
			created_at,
			updated_at
		) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := m.DB.ExecContext(ctx, stmt,
		txn.Amount,
		txn.Currency,
		txn.LastFour,
		txn.BankReturnCode,
		txn.TransactionStatusID,
		txn.ExpiryMonth,
		txn.ExpiryYear,
		txn.PaymentInent,
		txn.PaymentMethod,
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

// InsertOrder inserts a new row into the Transactions table
// and returns the primary key for the newly created row and an
// error (if any)
func (m *DBModel) InsertOrder(order Order) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
		INSERT INTO orders (
			widget_id,
			transaction_id,
			customer_id,
			status_id,
			quantity,
			amount,
			created_at,
			updated_at
		) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := m.DB.ExecContext(ctx, stmt,
		order.WidgetID,
		order.TransactionID,
		order.CustomerID,
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

// InsertCustomer inserts a new row into the Customers table
// and returns the primary key for the newly created row and an
// error (if any)
func (m *DBModel) InsertCustomer(cust Customer) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
		INSERT INTO customers (
			first_name,
			last_name,
			email,
			created_at,
			updated_at
		) 
		VALUES (?, ?, ?, ?, ?)`

	result, err := m.DB.ExecContext(ctx, stmt,
		cust.FirstName,
		cust.LastName,
		cust.Email,
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

// GetUserByEmail retrieves User by email address
func (m *DBModel) GetUserByEmail(email string) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	email = strings.ToLower(email)
	var u User
	sql := `SELECT	id,
					first_name,
					last_name,
					email,
					password,
					created_at,
					updated_at
			FROM	users
			WHERE	email = ?`

	row := m.DB.QueryRowContext(ctx, sql, email)

	err := row.Scan(
		&u.ID,
		&u.FirstName,
		&u.LastName,
		&u.Email,
		&u.Password,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		return u, err
	}

	return u, nil
}

func (m *DBModel) Authenticate(email, password string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int
	var hashedPassword string

	sql := "SELECT id, password FROM users WHERE email = ?"

	row := m.DB.QueryRowContext(ctx, sql, email)
	err := row.Scan(&id, &hashedPassword)
	if err != nil {
		return id, err
	}

	logSnippet := "[internal][models][Authenticate] =>"

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {

		if err == bcrypt.ErrMismatchedHashAndPassword {
			fmt.Printf("%s Invalid password for %s", logSnippet, email)
			return 0, errors.New("incorrect password")
		} else {
			fmt.Println(err)
			return 0, err
		}
	}

	fmt.Printf("%s Valid password found for %s (userID: %d)", logSnippet, email, id)
	return id, nil
}

func (m *DBModel) UpdatePasswordForUser(u User, hash string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `UPDATE users SET password = ? WHERE id = ?`
	_, err := m.DB.ExecContext(ctx, stmt, hash, u.ID)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (m *DBModel) GetAllOrders(recurring bool) ([]*Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var orders []*Order

	query := `
		select
			o.id, o.widget_id, o.transaction_id, o.customer_id, 
			o.status_id, o.quantity, o.amount, o.created_at,
			o.updated_at, w.id, w.name, t.id, t.amount, t.currency,
			t.last_four, t.expiry_month, t.expiry_year, t.payment_intent,
			t.bank_return_code, c.id, c.first_name, c.last_name, c.email
			
		from
			orders o
			left join widgets w on (o.widget_id = w.id)
			left join transactions t on (o.transaction_id = t.id)
			left join customers c on (o.customer_id = c.id)
		where
			w.is_recurring = ?
		order by
			o.created_at desc`

	var isRecurring int8 = 0
	if recurring {
		isRecurring = 1
	}

	rows, err := m.DB.QueryContext(ctx, query, isRecurring)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var o Order
		err = rows.Scan(
			&o.ID,
			&o.WidgetID,
			&o.TransactionID,
			&o.CustomerID,
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
			&o.Transaction.PaymentInent,
			&o.Transaction.BankReturnCode,
			&o.Customer.ID,
			&o.Customer.FirstName,
			&o.Customer.LastName,
			&o.Customer.Email,
		)
		if err != nil {
			return nil, err
		}

		orders = append(orders, &o)
	}

	return orders, nil
}

func (m *DBModel) GetOrderByID(id int) (Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var o Order

	query := `
		select
			o.id, o.widget_id, o.transaction_id, o.customer_id, 
			o.status_id, o.quantity, o.amount, o.created_at,
			o.updated_at, w.id, w.name, t.id, t.amount, t.currency,
			t.last_four, t.expiry_month, t.expiry_year, t.payment_intent,
			t.bank_return_code, c.id, c.first_name, c.last_name, c.email
			
		from
			orders o
			left join widgets w on (o.widget_id = w.id)
			left join transactions t on (o.transaction_id = t.id)
			left join customers c on (o.customer_id = c.id)
		where
			o.id = ?`

	row := m.DB.QueryRowContext(ctx, query, id)

	err := row.Scan(
		&o.ID,
		&o.WidgetID,
		&o.TransactionID,
		&o.CustomerID,
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
		&o.Transaction.PaymentInent,
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

func (m *DBModel) UpdateOrderStatus(id, statusID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := "UPDATE orders SET status_id = ? WHERE id = ?"

	_, err := m.DB.ExecContext(ctx, stmt, statusID, id)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (m *DBModel) GetAllOrdersPaginated(recurring bool, pageSize, page int) ([]*Order, int, int, error) {
	logSnippet := "\n[internal][models][GetAllOrdersPaginated] =>"
	fmt.Printf("%s (pageSize): %d", logSnippet, pageSize)
	fmt.Printf("%s (page)....: %d", logSnippet, page)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	offset := (page - 1) * pageSize
	fmt.Printf("%s (offset)..: %d", logSnippet, offset)

	var orders []*Order

	query := `
		select
			o.id, o.widget_id, o.transaction_id, o.customer_id, 
			o.status_id, o.quantity, o.amount, o.created_at,
			o.updated_at, w.id, w.name, t.id, t.amount, t.currency,
			t.last_four, t.expiry_month, t.expiry_year, t.payment_intent,
			t.bank_return_code, c.id, c.first_name, c.last_name, c.email
			
		from
			orders o
			left join widgets w on (o.widget_id = w.id)
			left join transactions t on (o.transaction_id = t.id)
			left join customers c on (o.customer_id = c.id)
		where
			w.is_recurring = ?
		order by
			o.created_at desc
		limit ?
		offset ?`

	var isRecurring int8 = 0
	if recurring {
		isRecurring = 1
	}

	rows, err := m.DB.QueryContext(ctx, query, isRecurring, pageSize, offset)
	if err != nil {
		fmt.Println(err)
		return nil, 0, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var o Order
		err = rows.Scan(
			&o.ID,
			&o.WidgetID,
			&o.TransactionID,
			&o.CustomerID,
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
			&o.Transaction.PaymentInent,
			&o.Transaction.BankReturnCode,
			&o.Customer.ID,
			&o.Customer.FirstName,
			&o.Customer.LastName,
			&o.Customer.Email,
		)
		if err != nil {
			fmt.Println(err)
			return nil, 0, 0, err
		}

		orders = append(orders, &o)
	}

	query = `
		SELECT 
			COUNT(o.id)
		FROM 
			orders o
		LEFT JOIN 
			widgets w ON (o.widget_id = w.id) 
		WHERE 
			w.is_recurring = ?`

	var totalRecords int
	countRow := m.DB.QueryRowContext(ctx, query, isRecurring)
	err = countRow.Scan(&totalRecords)
	if err != nil {
		fmt.Println(err)
		return nil, 0, 0, err
	}

	lastPage := totalRecords / pageSize

	fmt.Printf("%s (len(orders)).: %d", logSnippet, len(orders))
	fmt.Printf("%s (totalRecords): %d", logSnippet, totalRecords)
	fmt.Printf("%s (lastPage)....: %d", logSnippet, lastPage)

	return orders, lastPage, totalRecords, nil
}

func (m *DBModel) GetAllUsers() ([]*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var users []*User

	query :=
		`select
			id, last_name, first_name, email, created_at, updated_at
		from
			users
		order by
			last_name, first_name`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	for rows.Next() {
		var u User
		err = rows.Scan(
			&u.ID,
			&u.LastName,
			&u.FirstName,
			&u.Email,
			&u.CreatedAt,
			&u.UpdatedAt,
		)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		users = append(users, &u)
	}
	return users, nil
}

func (m *DBModel) GetOneUser(id int) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query :=
		`select
			id, last_name, first_name, email, created_at, updated_at
		from
			users
		where
			id = ?`

	row := m.DB.QueryRowContext(ctx, query, id)

	var u User
	err := row.Scan(
		&u.ID,
		&u.LastName,
		&u.FirstName,
		&u.Email,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		fmt.Println(err)
		return u, err
	}

	return u, nil
}

func (m *DBModel) EditUser(u User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
		UPDATE users SET
			first_name = ?,
			last_name  = ?,
			email      = ?,
			updated_at = ?
		WHERE
			id = ?`

	_, err := m.DB.ExecContext(ctx, stmt, u.FirstName, u.LastName, u.Email, u.UpdatedAt, u.ID)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (m *DBModel) InsertUser(u User, hash string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
		INSERT INTO users
		(
			first_name,
			last_name,
			email,
			password,
			created_at,
			updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?)`

	_, err := m.DB.ExecContext(
		ctx,
		stmt,
		u.FirstName,
		u.LastName,
		u.Email,
		hash,
		time.Now(),
		time.Now(),
	)

	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (m *DBModel) DeleteUser(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `DELETE FROM users WHERE id = ?`

	_, err := m.DB.ExecContext(ctx, stmt, id)
	if err != nil {
		fmt.Println(err)
		return err
	}

	stmt = "DELETE FROM tokens WHERE user_id = ?"
	_, err = m.DB.ExecContext(ctx, stmt, id)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
