package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	p2p_database "github.com/dTelecom/p2p-database"
	ipfslite "github.com/hsanjuan/ipfs-lite"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	defer cancel()

	h, err := p2p_database.MakeHost(3500, true)
	if err != nil {
		panic(err)
	}

	db, err := p2p_database.Connect(ctx, h, ipfslite.DefaultBootstrapPeers(), "chat")
	err = db.Set(ctx, "key", "value")
	if err != nil {
		panic(err)
	}
	err = db.Set(ctx, "foo", "bar")
	if err != nil {
		panic(err)
	}

	v, err := db.Get(ctx, "key")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Value key %v\n", v)
	v2, err := db.Get(ctx, "foo")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Value foo %v\n", v2)

	<-ctx.Done()

	shutdownCtx, c := context.WithTimeout(context.Background(), 15*time.Second)
	defer c()

	fmt.Println("Disconnecting...")
	err = db.Disconnect(shutdownCtx)
	if err != nil {
		panic(err)
	}
}
