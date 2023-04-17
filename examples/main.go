package main

import (
	"context"

	p2p_database "github.com/dTelecom/p2p-database"
)

func main() {
	db, err := p2p_database.Connect(context.Background())
}
