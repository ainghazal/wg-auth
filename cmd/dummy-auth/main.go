package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/ainghazal/wg-auth/prng"
)

type config struct {
	addr  string
	seed  string
	peers int
}

func main() {
	cfg := &config{}

	defaultSeedStr := strconv.Itoa(int(prng.DefaultSeed))

	flag.StringVar(&cfg.addr, "address", "127.0.0.1:58120", "address for the server (default: 127.0.0.1:58120)")
	flag.StringVar(&cfg.seed, "seed", defaultSeedStr, "seed to use in the deterministic auth generation")
	flag.IntVar(&cfg.peers, "peers", 10, "number of peers in the pool")

	isServer := false
	isPeer := false

	for i := 0; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "server" {
			isServer = true
			break
		}
		if arg == "peer" {
			isPeer = true
			break
		}
	}

	if !isServer && !isPeer {
		fmt.Println("expected either 'server' or 'client'")
		os.Exit(1)
	}

	flag.Parse()
	seed, err := strconv.Atoi(cfg.seed)

	switch isServer {
	case true:
		if err != nil {
			panic(err)
		}
		server, err := prng.NewServerFromSeed(uint64(seed))
		if err != nil {
			panic(err)
		}
		server.SetIPAddress(cfg.addr)

		server.GenerateConfig(uint64(cfg.peers))
		fmt.Println(string(server.SerializeConfig()))
		os.Exit(0)

	case false:
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		n := r.Intn(cfg.peers) + 1
		peer, err := prng.NewPeerFromSeedAndNumber(prng.DefaultSeed, uint64(n))
		if err != nil {
			panic(err)
		}
		peer.EndpointIP = cfg.addr
		peerCfg := peer.SerializeConfig()
		fmt.Println(string(peerCfg))
		os.Exit(0)
	}
}
