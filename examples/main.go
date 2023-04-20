package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	p2p_database "github.com/dTelecom/p2p-database"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

var (
	nodePort = flag.Uint("p", 3500, "node port")
)

func main() {
	flag.Parse()

	fmt.Printf("Used port %d", int(*nodePort))

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	defer cancel()

	h, dht, err := p2p_database.MakeHost(ctx, int(*nodePort), false)
	if err != nil {
		panic(err)
	}

	bstr, _ := multiaddr.NewMultiaddr("/ip4/162.55.89.211/tcp/3500/p2p/QmYw5WSR9T8bg6HgcUgJwD1NZHj2h4tTvyp3ysGJFA6Gmw")
	inf, err := peer.AddrInfoFromP2pAddr(bstr)
	if err != nil {
		panic(err)
	}

	err = h.Connect(ctx, *inf)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Host id is %s\n", h.ID().String())
	h.ConnManager().TagPeer(inf.ID, "keep", 100)

	db, err := p2p_database.Connect(ctx, h, dht, []peer.AddrInfo{*inf}, "chat")

	fmt.Printf("> ")
	scanner := bufio.NewScanner(os.Stdin)

l:
	for scanner.Scan() {
		text := scanner.Text()
		fields := strings.Fields(text)
		if len(fields) == 0 {
			fmt.Printf("> ")
			continue
		}

		switch fields[0] {
		case "exit", "quit":
			break l
		case "get":
			if len(fields) < 2 {
				fmt.Println("get <key>")
				fmt.Println("> ")
				continue
			}
			val, err := db.Get(ctx, fields[1])
			if err != nil {
				fmt.Printf("error get key %s %s\n", fields[1], err)
			} else {
				fmt.Printf("[%s] -> %s\n", fields[1], string(val))
			}
		case "set":
			if len(fields) < 3 {
				fmt.Println("set <key> <val>")
				fmt.Println("> ")
				continue
			}
			err := db.Set(ctx, fields[1], fields[2])
			if err != nil {
				fmt.Printf("error set key %s %s\n", fields[1], err)
			} else {
				fmt.Printf("key %s successfully set\n", fields[1])
			}
		}
		fmt.Printf("> ")
	}

	shutdownCtx, c := context.WithTimeout(context.Background(), 15*time.Second)
	defer c()

	fmt.Println("Disconnecting...")
	err = db.Disconnect(shutdownCtx)
	if err != nil {
		panic(err)
	}
}
