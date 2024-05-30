package repository

import "github.com/EmeraldLS/terminal-shop/internal/types"

type DB_SERVICE interface {
	GetProducts() ([]types.Product, error)
	SellProduct(types.Product) error
	Login(types.User) (string, string, error)
	GetProduct(id, name string) (*types.Product, error)
}
