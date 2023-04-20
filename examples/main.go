package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"

	p2p_database "github.com/dTelecom/p2p-database"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	defer cancel()

	h, prvKey, pubKey, err := p2p_database.MakeHost(3500, true)
	if err != nil {
		panic(err)
	}

	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		panic(err)
	}

	db, err := p2p_database.Connect(ctx, h, prvKey, pubKey, "chat", ps)
	err = db.Set(ctx, "key", "value", time.Second)
	if err != nil {
		panic(err)
	}
	err = db.Set(ctx, "foo", "bar", time.Second)
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

	err = db.Disconnect(shutdownCtx)
	if err != nil {
		panic(err)
	}
}
