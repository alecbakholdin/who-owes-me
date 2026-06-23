package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"who-owes-me/internal/envutil"
)

var DB *sql.DB

func InitDB() {
	var err error
	dbPath := envutil.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data.db"
	}

	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	createTables()
}

func createTables() {
	usersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		oidc_sub TEXT UNIQUE NOT NULL,
		aid_class TEXT NOT NULL,
		actual_payee_id TEXT NOT NULL
	);`

	expenseSplitsTable := `
	CREATE TABLE IF NOT EXISTS expense_splits (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		actual_transaction_id TEXT NOT NULL,
		user_id INTEGER NOT NULL,
		amount_owed INTEGER NOT NULL,
		auto_created INTEGER NOT NULL DEFAULT 0,
		FOREIGN KEY (user_id) REFERENCES users (id),
		UNIQUE(actual_transaction_id, user_id)
	);`

	_, err := DB.Exec(usersTable)
	if err != nil {
		log.Fatalf("Error creating users table: %v", err)
	}

	_, err = DB.Exec(expenseSplitsTable)
	if err != nil {
		log.Fatalf("Error creating expense_splits table: %v", err)
	}

	// Apply migrations
	DB.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_expense_splits_tx_user ON expense_splits(actual_transaction_id, user_id)")
	DB.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_users_payee ON users(actual_payee_id) WHERE actual_payee_id != ''")
	DB.Exec("ALTER TABLE expense_splits ADD COLUMN auto_created INTEGER NOT NULL DEFAULT 0")
	DB.Exec("ALTER TABLE expense_splits ADD COLUMN expense_date TEXT NOT NULL DEFAULT ''")
	DB.Exec("ALTER TABLE expense_splits ADD COLUMN expense_note TEXT NOT NULL DEFAULT ''")

	fmt.Println("Database initialized successfully.")
}

// User represents a user in the system
type User struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	OIDCSub       string `json:"oidc_sub"`
	AidClass      string `json:"aid_class"`
	ActualPayeeID string `json:"actual_payee_id"`
}

// ExpenseSplit represents how an Actual Budget transaction is split
type ExpenseSplit struct {
	ID                  int    `json:"id"`
	ActualTransactionID string `json:"actual_transaction_id"`
	UserID              int    `json:"user_id"`
	AmountOwed          int    `json:"amount_owed"` // in cents
	AutoCreated         bool   `json:"auto_created"`
	ExpenseDate         string `json:"expense_date"`
	ExpenseNote         string `json:"expense_note"`
}

// --- User Queries ---

func CreateUser(name, oidcSub, aidClass, actualPayeeID string) error {
	_, err := DB.Exec(`
		INSERT INTO users (name, oidc_sub, aid_class, actual_payee_id) 
		VALUES (?, ?, ?, ?)
	`, name, oidcSub, aidClass, actualPayeeID)
	return err
}

func UpdateUser(id int, name, oidcSub, aidClass, actualPayeeID string) error {
	_, err := DB.Exec(`
		UPDATE users 
		SET name = ?, oidc_sub = ?, aid_class = ?, actual_payee_id = ?
		WHERE id = ?
	`, name, oidcSub, aidClass, actualPayeeID, id)
	return err
}

func GetUserBySub(sub string) (*User, error) {
	row := DB.QueryRow("SELECT id, name, oidc_sub, aid_class, actual_payee_id FROM users WHERE oidc_sub = ?", sub)
	var u User
	err := row.Scan(&u.ID, &u.Name, &u.OIDCSub, &u.AidClass, &u.ActualPayeeID)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func GetUserByID(id int) (*User, error) {
	row := DB.QueryRow("SELECT id, name, oidc_sub, aid_class, actual_payee_id FROM users WHERE id = ?", id)
	var u User
	err := row.Scan(&u.ID, &u.Name, &u.OIDCSub, &u.AidClass, &u.ActualPayeeID)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func GetAllUsers() ([]User, error) {
	rows, err := DB.Query("SELECT id, name, oidc_sub, aid_class, actual_payee_id FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.OIDCSub, &u.AidClass, &u.ActualPayeeID); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

// --- Split Queries ---

func SetAutoSplit(txID string, userID int, amount int, date string, note string) error {
	_, err := DB.Exec(`
		INSERT INTO expense_splits (actual_transaction_id, user_id, amount_owed, auto_created, expense_date, expense_note) 
		VALUES (?, ?, ?, 1, ?, ?)
		ON CONFLICT(actual_transaction_id, user_id) DO UPDATE SET 
			amount_owed=excluded.amount_owed, auto_created=excluded.auto_created,
			expense_date=excluded.expense_date, expense_note=excluded.expense_note
	`, txID, userID, amount, date, note)
	return err
}

func ClearSplitsForTx(txID string) error {
	_, err := DB.Exec("DELETE FROM expense_splits WHERE actual_transaction_id = ?", txID)
	return err
}

func SetSplit(txID string, userID int, amount int, date string, note string) error {
	_, err := DB.Exec(`
		INSERT INTO expense_splits (actual_transaction_id, user_id, amount_owed, auto_created, expense_date, expense_note) 
		VALUES (?, ?, ?, 0, ?, ?)
		ON CONFLICT(actual_transaction_id, user_id) DO UPDATE SET 
			amount_owed=excluded.amount_owed, auto_created=0,
			expense_date=excluded.expense_date, expense_note=excluded.expense_note
	`, txID, userID, amount, date, note)
	return err
}

func GetAllSplits() ([]ExpenseSplit, error) {
	rows, err := DB.Query("SELECT id, actual_transaction_id, user_id, amount_owed, auto_created, expense_date, expense_note FROM expense_splits")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var splits []ExpenseSplit
	for rows.Next() {
		var s ExpenseSplit
		var autoCreated int
		if err := rows.Scan(&s.ID, &s.ActualTransactionID, &s.UserID, &s.AmountOwed, &autoCreated, &s.ExpenseDate, &s.ExpenseNote); err != nil {
			return nil, err
		}
		s.AutoCreated = autoCreated == 1
		splits = append(splits, s)
	}
	return splits, nil
}

func GetSplitsForUser(userID int) ([]ExpenseSplit, error) {
	rows, err := DB.Query("SELECT id, actual_transaction_id, user_id, amount_owed, auto_created, expense_date, expense_note FROM expense_splits WHERE user_id = ?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var splits []ExpenseSplit
	for rows.Next() {
		var s ExpenseSplit
		var autoCreated int
		if err := rows.Scan(&s.ID, &s.ActualTransactionID, &s.UserID, &s.AmountOwed, &autoCreated, &s.ExpenseDate, &s.ExpenseNote); err != nil {
			return nil, err
		}
		s.AutoCreated = autoCreated == 1
		splits = append(splits, s)
	}
	return splits, nil
}
