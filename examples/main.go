package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
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
	nodePort      = flag.Uint("p", 3500, "node port")
	ethPrivateKey = flag.String("pk", "", "ethereum wallet private key")
)

func main() {
	flag.Parse()

	if len(os.Args) < 2 {
		log.Fatalf("expected private key ethereum wallet as first argument: ./main")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	defer cancel()

	h, dht, err := p2p_database.MakeHost(ctx, *ethPrivateKey, int(*nodePort), false)
	if err != nil {
		panic(err)
	}

	bstr, _ := multiaddr.NewMultiaddr("/ip4/162.55.89.211/tcp/3500/p2p/16Uiu2HAmKJTUywRaKxJ2g2trHby2GYVSvnQVUh4Jxc9fhH7UZkBY")
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
		case "list":
			keys, err := db.List(ctx)
			if err != nil {
				fmt.Printf("error list keys %s\n", err)
			} else {
				fmt.Printf("Found %d keys", len(keys))
				for _, k := range keys {
					fmt.Println(k)
				}
			}
		case "del":
			if len(fields) < 2 {
				fmt.Println("del <key>")
				fmt.Println("> ")
				continue
			}
			err := db.Remove(ctx, fields[1])
			if err != nil {
				fmt.Printf("error remove key %s\n", err)
			} else {
				fmt.Println("Successfully remove key")
			}
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
