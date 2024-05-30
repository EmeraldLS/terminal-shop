package types

import "time"

type User struct {
	Username  string    `json:"username"`
	KeyType   string    `json:"key_type"`
	PublicKey string    `json:"pub_key"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Product struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	UserName  string  `json:"username"`
	SellerKey string  `json:"seller_key"`
}

type Cart struct {
	SellerKey   string `json:"seller_key"`
	ProductID   string `json:"product_id"`
	ProductName string `json:"product_name"`
	Username    string `json:"username"`
}
