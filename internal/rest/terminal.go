package rest

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/EmeraldLS/terminal-shop/internal/repository"
	"github.com/EmeraldLS/terminal-shop/internal/types"
	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

var (
	helpCmd        = regexp.MustCompile(`^/help.*`)
	exitCmd        = regexp.MustCompile(`^/exit.*`)
	productListCmd = regexp.MustCompile(`^/products.*`)
	sellProductCmd = regexp.MustCompile(`^/sell\s+-name=\w+\s+-price=\d+(\.\d+)?$`)
	addToCartCmd   = regexp.MustCompile(`^/add\s+-id=([^\s]+)\s+-name=([^\s]+)`)
	cartListCmd    = regexp.MustCompile(`^/cart.*`)
	cart           = map[string][]types.Cart{}
)

func helpMsg() string {
	return `
Hello and welcome to the chat server! Please use one of the following commands:
1. /products: Show the lists of products available
2. /sell -name=<product-name> -price=<product-price>: To post a product for sale
3. /exit: To leave the application
4. /help: To display this message
5. /add -id=<product-id> -name=<product-name> add an item to cart
6. /cart: lists all the items in cart
`
}

func ProcessTerminalInput(s ssh.Session, repo repository.DB_SERVICE) {
	authKey := gossh.MarshalAuthorizedKey(s.PublicKey())
	authKeyArr := strings.Split(string(authKey), " ")

	user := types.User{
		Username:  s.User(),
		KeyType:   authKeyArr[0],
		PublicKey: authKeyArr[1],
		CreatedAt: time.Now().Local(),
		UpdatedAt: time.Now().Local(),
	}

	username, pubKey, err := repo.Login(user)
	if err != nil {
		_, _ = fmt.Fprintln(s, err)
		return
	}

	term := term.NewTerminal(s, fmt.Sprintf("\n%s > ", strings.ToUpper(username)))

	for {
		line, err := term.ReadLine()
		if err != nil {
			slog.Error("unable to read current line", "err", err)
			break
		}
		if len(line) > 0 {
			if string(line[0]) == "/" {
				switch {
				case productListCmd.MatchString(line):
					products, err := repo.GetProducts()
					if err != nil {
						_, _ = term.Write([]byte(err.Error()))
						continue
					}

					term.Write([]byte("+------+----------------+---------+----------+\n"))
					term.Write([]byte("|  ID  |      Name      |  Price  | Username |\n"))
					term.Write([]byte("+------+----------------+---------|----------+\n"))

					for _, p := range products {
						term.Write([]byte(fmt.Sprintf("| %-4s | %-14s | $%-6.2f | %-10s\n", p.ID, p.Name, p.Price, p.UserName)))
					}

					term.Write([]byte("+------+----------------+---------|----------+\n"))

				case sellProductCmd.MatchString(line):
					lineArr := strings.Split(line, " ")
					nameKV := lineArr[1]
					priceKV := lineArr[2]
					name := strings.Split(nameKV, "=")[1]
					priceStr := strings.Split(priceKV, "=")[1]
					price, err := strconv.ParseFloat(priceStr, 64)
					if err != nil {
						term.Write([]byte(fmt.Sprintf("error converting price to float64: %v\n", err)))
						continue
					}

					product := types.Product{
						Name:      name,
						Price:     price,
						UserName:  username,
						SellerKey: pubKey,
					}

					err = repo.SellProduct(product)
					if err != nil {
						term.Write([]byte(err.Error()))
						continue
					}
					term.Write([]byte("Product has been added successfully"))

				case addToCartCmd.MatchString(line):
					lineArr := strings.Split(line, " ")
					idKV := lineArr[1]
					nameKV := lineArr[2]
					productID := strings.Split(idKV, "=")[1]
					productName := strings.Split(nameKV, "=")[1]

					product, err := repo.GetProduct(productID, productName)
					if err != nil {
						term.Write([]byte(err.Error()))
						continue
					}

					cart := types.Cart{
						SellerKey:   product.SellerKey,
						ProductID:   productID,
						ProductName: productName,
						Username:    product.UserName,
					}

					addToCart(cart)
					term.Write([]byte("Item added to cart successfully"))

				case cartListCmd.MatchString(line):
					cartItems, ok := cart[user.PublicKey]
					if !ok {
						term.Write([]byte("user does not exit in cart"))
						continue
					} else if len(cart[user.PublicKey]) < 1 {
						term.Write([]byte("No item found in cart"))
						continue
					} else {
						bytes, err := json.MarshalIndent(cartItems, "", " ")
						if err != nil {
							term.Write([]byte(err.Error()))
							continue
						}

						term.Write(bytes)
					}

				case helpCmd.MatchString(line):
					term.Write([]byte(helpMsg()))

				case exitCmd.MatchString(line):
					term.Write([]byte("Exit successful\r\n"))
					s.Close()
					s.Exit(0)

				default:
					term.Write([]byte(helpMsg()))

				}
			}
		}
	}
}

func addToCart(cartItem types.Cart) {
	cart[cartItem.SellerKey] = append(cart[cartItem.SellerKey], cartItem)
}
