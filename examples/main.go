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

	logging "github.com/ipfs/go-log/v2"

	p2p_database "github.com/dTelecom/p2p-realtime-database"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

var (
	ethPrivateKey = flag.String("pk", "", "ethereum wallet private key")

	loggingDebug   = flag.Bool("vvv", false, "debug mode")
	loggingInfo    = flag.Bool("vv", false, "info mode")
	loggingWarning = flag.Bool("v", false, "warning mode")

	db *p2p_database.DB
)

func main() {
	flag.Parse()

	if len(os.Args) < 2 {
		log.Fatalf("expected private key ethereum wallet as first argument: ./main")
	}

	switch {
	case *loggingWarning:
		logging.SetLogLevel("*", "warn")
	case *loggingInfo:
		logging.SetLogLevel("*", "info")
	case *loggingDebug:
		logging.SetLogLevel("*", "debug")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	defer cancel()

	logger := logging.Logger("db")
	defer logger.Sync()

	cfg := p2p_database.EnvConfig
	cfg.DatabaseName = "global"
	cfg.WalletPrivateKey = *ethPrivateKey

	var additionalDatabases = map[string]*p2p_database.DB{}
	var err error
	db, err = p2p_database.Connect(ctx, cfg, logger)
	if err != nil {
		panic(err)
	}
	additionalDatabases["global"] = db

	h := db.GetHost()
	fmt.Printf("Peer id %s\n", h.ID().String())

	fmt.Printf("> ")
	scanner := bufio.NewScanner(os.Stdin)

l:
	for scanner.Scan() {
		text := scanner.Text()
		fields := strings.Fields(text)
		if len(fields) == 0 {
			fmt.Printf("[db %s] > ", db.Name)
			continue
		}

		switch fields[0] {
		case "connect":
			_, alreadyExists := additionalDatabases[fields[1]]
			if !alreadyExists {
				cfg := p2p_database.EnvConfig
				cfg.DatabaseName = fields[1]
				cfg.WalletPrivateKey = *ethPrivateKey
				additionalDatabases[fields[1]], err = p2p_database.Connect(ctx, cfg, logger)
				if err != nil {
					fmt.Printf("error connecting %s\n", err)
				}
			}
			db = additionalDatabases[fields[1]]
		case "disconnect":
			_, alreadyExists := additionalDatabases[db.Name]
			if !alreadyExists {
				fmt.Printf("db not found")
				return
			}
			err = db.Disconnect(context.Background())
			if err != nil {
				fmt.Printf("disconnect error: %v", err)
			}
			delete(additionalDatabases, db.Name)
		case "subscribe":
			err = db.Subscribe(ctx, fields[1], func(event p2p_database.Event) {
				fmt.Printf("Got subscription event %v", event)
			})
			if err != nil {
				fmt.Printf("error %s\n", err)
			}
		case "publish":
			_, err = db.Publish(ctx, fields[1], fields[2])
			if err != nil {
				fmt.Printf("error %s\n", err)
			}
		case "peers":
			for _, p := range connectedPeers(db.GetHost()) {
				fmt.Printf("Peer [%s] %s\r\n", p.ID, p.Addrs[0].String())
			}
		case "debug":
			switch fields[1] {
			case "on":
				level := "debug"
				if len(fields) > 2 {
					level = fields[2]
				}
				logging.SetLogLevel("*", level)
			case "off":
				logging.SetLogLevel("*", "error")
			}
		case "exit", "quit":
			break l
		case "list":
			keys, err := db.List(ctx)
			if err != nil {
				fmt.Printf("error list keys %s\n", err)
			} else {
				fmt.Printf("Found %d keys\n", len(keys))
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
				fmt.Printf("[db %s] [%s] -> %s\n", db.String(), fields[1], string(val))
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
		case "ttl":
			if len(fields) < 3 {
				fmt.Println("ttl <key> <duration>")
				fmt.Println("> ")
				continue
			}
			duration, err := time.ParseDuration(fields[2])
			if err != nil {
				fmt.Printf("error parse duration: %s\n", err)
				fmt.Println("> ")
				continue
			}
			err = db.TTL(ctx, fields[1], duration)
			if err != nil {
				fmt.Printf("error ttl command: %s\n", err)
				continue
			} else {
				fmt.Println("success set ttl")
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
