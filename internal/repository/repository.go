package repository

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/EmeraldLS/terminal-shop/internal/types"
	_ "github.com/lib/pq"
)

type repository struct {
	dB *sql.DB
}

const table = `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username TEXT NOT NULL,
		pub_key TEXT NOT NULL UNIQUE,
		key_type TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS products (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		price FLOAT NOT NULL,
		seller_key TEXT NOT NULL,
		FOREIGN KEY (seller_key) REFERENCES users(pub_key)
	);

`

func NewRepository(pgConnString string) *repository {

	conn, err := sql.Open("postgres", pgConnString)
	if err != nil {
		slog.Error("unable to setup database connection", "err", err)
		return nil
	}

	if err = conn.Ping(); err != nil {
		slog.Error("unable to ping database", "err", err)
		return nil
	}

	_, err = conn.Exec(table)
	if err != nil {
		slog.Error("unable to execute stmt", "err", err)
		return nil
	}

	return &repository{
		dB: conn,
	}
}

func (s *repository) GetProducts() ([]types.Product, error) {
	stmt, err := s.dB.Prepare("SELECT * FROM products")
	if err != nil {
		return nil, fmt.Errorf("unable to prepare statement: %v", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return nil, fmt.Errorf("unable to query rows: %v", err)
	}

	var products []types.Product
	for rows.Next() {
		var product types.Product
		if err = rows.Scan(&product.ID, &product.Name, &product.Price, &product.UserName, &product.SellerKey); err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}
		products = append(products, product)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	return products, nil
}

func (s *repository) SellProduct(product types.Product) error {
	stmt, err := s.dB.Prepare("INSERT INTO products (name, price, username, seller_key) VALUES ($1, $2, $3, $4)")
	if err != nil {
		return fmt.Errorf("unable to prepare statement: %v", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(product.Name, product.Price, product.UserName, product.SellerKey)
	if err != nil {
		return fmt.Errorf("unable to execute stmt: %v", err)
	}

	return nil
}

func (s *repository) Login(user types.User) (string, string, error) {
	stmt, err := s.dB.Prepare("SELECT username, pub_key FROM users WHERE pub_key=$1 AND key_type=$2")
	if err != nil {
		return "", "", fmt.Errorf("unable to prepare statement: %v", err)
	}
	defer stmt.Close()
	rows, err := stmt.Query(user.PublicKey, user.KeyType)
	if err != nil {
		return "", "", fmt.Errorf("unable to query rows: %v", err)
	}

	defer rows.Close()

	if rows.Next() {
		var username string
		var pubKey string
		err = rows.Scan(&username, &pubKey)
		if err != nil {
			return "", "", fmt.Errorf("unable to scan row: %v", err)
		}

		_, err = s.dB.Exec("UPDATE users SET updated_at=$1 WHERE username=$2", user.UpdatedAt, user.Username)
		if err != nil {
			return "", "", fmt.Errorf("unable to update user: %v", err)
		}

		return username, pubKey, nil
	} else {
		stmtInsert, err := s.dB.Prepare("INSERT INTO users (pub_key, key_type, username, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)")
		if err != nil {
			return "", "", fmt.Errorf("unable to prepare INSERT statement: %v", err)
		}
		defer stmtInsert.Close()

		_, err = stmtInsert.Exec(user.PublicKey, user.KeyType, user.Username, user.CreatedAt, user.UpdatedAt)
		if err != nil {
			return "", "", fmt.Errorf("unable to execute INSERT query: %v", err)
		}
		return user.Username, user.PublicKey, nil
	}
}

func (r *repository) GetProduct(id, name string) (*types.Product, error) {
	stmt, err := r.dB.Prepare("SELECT price, username, seller_key FROM products WHERE id=$1 AND name=$2")
	if err != nil {
		return nil, fmt.Errorf("unable to prepare statement: %v", err)
	}

	rows, err := stmt.Query(id, name)
	if err != nil {
		return nil, fmt.Errorf("unable to make query: %v", err)
	}
	defer rows.Close()

	var product types.Product

	if rows.Next() {

		err = rows.Scan(&product.Price, &product.UserName, &product.SellerKey)
		if err != nil {
			return nil, fmt.Errorf("unable to scan rows: %v", err)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	product.ID = id
	product.Name = name

	return &product, nil

}
