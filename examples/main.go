package main

import (
	"bufio"
	"context"
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

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	defer cancel()

	h, err := p2p_database.MakeHost(3500, false)
	if err != nil {
		panic(err)
	}

	bstr, _ := multiaddr.NewMultiaddr("/ip4/162.55.89.211/tcp/3500/p2p/QmeaDuf7zJaaUzBKRiDDFG86aLZv7XrwBJy3Uc3hdY5h5j")
	inf, err := peer.AddrInfoFromP2pAddr(bstr)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Host id is %s\n", h.ID().String())
	h.ConnManager().TagPeer(inf.ID, "keep", 100)

	db, err := p2p_database.Connect(ctx, h, []peer.AddrInfo{*inf}, "chat")

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
