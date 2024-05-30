package main

import (
	"log"

	"github.com/EmeraldLS/terminal-shop/internal/repository"
	"github.com/EmeraldLS/terminal-shop/internal/rest"
	"github.com/gliderlabs/ssh"
)

func main() {
	repo := repository.NewRepository("postgresql://postgres:lawrenc2003@localhost/terminal_shop?sslmode=disable")
	ssh.Handle(func(s ssh.Session) {
		rest.ProcessTerminalInput(s, repo)
	})

	publicKeyOption := ssh.PublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
		return true
	})

	log.Println("Server listening on port 2323")
	log.Fatal(ssh.ListenAndServe(":2323", nil, publicKeyOption))
}
