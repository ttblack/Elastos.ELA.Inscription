package main

import (
	"fmt"
	"github.com/ttblack/Elastos.ELA.Inscription/config"
	"github.com/ttblack/Elastos.ELA.Inscription/puller"
	"github.com/ttblack/Elastos.ELA.Inscription/rpc"
	"github.com/ttblack/Elastos.ELA.Inscription/store"
	"os"
	"sync"
	"time"
)

type Response struct {
	P    string `json:"p"`
	Op   string `json:"op"`
	Tick string `json:"tick"`
	Amt  string `json:"amt"`
}

func main() {
	var wg sync.WaitGroup
	cfg := config.InitConfig("./config.json")
	wg.Add(1)

	db, err := store.NewLevelDBStorage(cfg.DataDir)
	if err != nil {
		fmt.Println("open leveldb failed", "error ", err)
		os.Exit(1)
	}

	go startPuller(cfg, db)
	go rpc.StartRPCServer(cfg.ServerHttpPort, db)
	wg.Wait()
}

func startPuller(cfg *config.Config, db *store.LevelDBStorage) {
	pull, err := puller.NewPuller(cfg.Http, db)
	if err != nil {
		fmt.Println("NewPuller error", err)
		os.Exit(1)
	}

	for {
		endBlock, err := pull.Start(cfg.StartBlock)
		if err != nil {
			fmt.Println("pull block failed", "error", err, "failed block ", endBlock)
			os.Exit(1)
		}
		latestBlock, err := pull.LatestBlock()
		if err != nil {
			fmt.Println("pull LatestBlock failed", "error", err, "failed block ", endBlock)
			os.Exit(1)
		}
		if endBlock >= latestBlock {
			time.Sleep(5 * time.Second)
		}
		cfg.StartBlock = endBlock + 1
	}
}
