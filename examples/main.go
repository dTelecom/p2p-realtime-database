package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	logging "github.com/ipfs/go-log/v2"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	p2p_database "github.com/dTelecom/p2p-database"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

var (
	ethPrivateKey = flag.String("pk", "", "ethereum wallet private key")
)

func main() {
	flag.Parse()

	if len(os.Args) < 2 {
		log.Fatalf("expected private key ethereum wallet as first argument: ./main")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	defer cancel()

	db, err := p2p_database.Connect(ctx, *ethPrivateKey, "chat", *logging.Logger("db"))
	if err != nil {
		panic(err)
	}

	h := db.GetHost()
	fmt.Printf("Peer id %s\n", h.ID().String())

	fmt.Printf("> ")
	scanner := bufio.NewScanner(os.Stdin)

	var quitDebugCh chan struct{}

l:
	for scanner.Scan() {
		text := scanner.Text()
		fields := strings.Fields(text)
		if len(fields) == 0 {
			fmt.Printf("> ")
			continue
		}

		switch fields[0] {
		case "debug":
			switch fields[1] {
			case "on":
				logging.SetLogLevel("*", "debug")

				quitDebugCh = make(chan struct{})
				go func() {
					for {
						select {
						default:
							fmt.Println("Peers:")
							for _, p := range connectedPeers(db.GetHost()) {
								logging.Logger("cli").Infof("Peer [%s] %s\r\n", p.ID, p.Addrs[0].String())
							}
							fmt.Println()
							time.Sleep(3 * time.Second)
						case <-quitDebugCh:
							return
						}
					}
				}()
			case "off":
				logging.SetLogLevel("*", "error")
				close(quitDebugCh)
			}
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

func connectedPeers(h host.Host) []*peer.AddrInfo {
	var pinfos []*peer.AddrInfo
	for _, c := range h.Network().Conns() {
		pinfos = append(pinfos, &peer.AddrInfo{
			ID:    c.RemotePeer(),
			Addrs: []multiaddr.Multiaddr{c.RemoteMultiaddr()},
		})
	}
	return pinfos
}
