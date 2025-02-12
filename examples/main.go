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
	cfg.DatabaseName = "livekit_global"
	cfg.WalletPrivateKey = *ethPrivateKey

	var err error
	db, err = p2p_database.Connect(ctx, cfg, logger)
	if err != nil {
		panic(err)
	}

	h := db.GetHost()
	fmt.Printf("Peer id %s\n", h.ID().String())

	fmt.Printf("> ")
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		text := scanner.Text()
		fields := strings.Fields(text)
		if len(fields) == 0 {
			fmt.Printf("[db %s] > ", db.Name)
			continue
		}

		switch fields[0] {
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
