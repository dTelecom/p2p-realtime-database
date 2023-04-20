package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	p2p_database "github.com/dTelecom/p2p-database"
	ipfslite "github.com/hsanjuan/ipfs-lite"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	defer cancel()

	h, err := p2p_database.MakeHost(3500, false)
	if err != nil {
		panic(err)
	}

	bstr, _ := multiaddr.NewMultiaddr("/ip4/65.21.98.65/tcp/3500/p2p/QmdGjE3xXzxUH4TJgghe4qNNAbG3Q2cJ6g6qtHfxGn7qjT")
	inf, err := peer.AddrInfoFromP2pAddr(bstr)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Host id is %s\n", h.ID().String())

	bootstrapPeers := ipfslite.DefaultBootstrapPeers()
	bootstrapPeers = append(bootstrapPeers, *inf)

	h.ConnManager().TagPeer(inf.ID, "keep", 100)

	db, err := p2p_database.Connect(ctx, h, bootstrapPeers, "chat")
	//err = db.Set(ctx, "key", "value")
	//if err != nil {
	//	panic(err)
	//}
	//err = db.Set(ctx, "foo", "bar")
	//if err != nil {
	//	panic(err)
	//}
	//err = db.Set(ctx, "foo2", "bar2")
	//if err != nil {
	//	panic(err)
	//}

	time.Sleep(10 * time.Second)

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
