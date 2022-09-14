package main

import (
	"context"
	"fmt"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/tak1827/blockchain-tps-test/tps"
)

const (
	ERC721 = "erc721"
	EHT    = "eth"
)

var (
	Endpoint         = "http://localhost:16888/rpc" // testnet
	Timeout          = 15 * time.Second
	MaxConcurrency   = runtime.NumCPU()
	mesuringDuration = 60 * time.Second //执行数据时间
	queueSize        = 300              //队列大小
	concurrency      = 1                //并发数量
	queue            = tps.NewQueue(queueSize)
	closing          uint32
	tpsClosing       uint32
	idlingDuration   uint32
	logLevel         = tps.WARN_LEVEL // INFO_LEVEL, WARN_LEVEL, FATAL_LEVEL
	logger           = tps.NewLogger(logLevel)
	privs            = []string{
		"2E833968E5bB786Ae419c4d13189fB081Cc43bab",
	}

	model = ERC721 //压测类型

	addr_priv     = make(map[string]string, len(privs))
	erc721address = "0x0000000000000000000000000000000000000009"
)

func main() {
	// client := jsonrpc.NewRPCClient("http://localhost:16888/rpc")
	// // res, err := client1.BlockNumber(context.Background())
	// rpcRes, rpcErr := client.Call("theta.GetBlockByHeight", trpc.GetBlockByHeightArgs{
	// 	Height: tcommon.JSONUint64(378)})

	// //logger.Infof("eth_getBlockByNumber, rpcRes: %v, rpcRes.Rsult: %v", rpcRes, rpcRes.Result)
	// chainID := new(big.Int)
	// chainID.SetString("privatenet", 16)
	// result, err := ethrpc.GetBlockFromTRPCResult(chainID, rpcRes, rpcErr, true)

	// fmt.Println(err, result.Transactions)
	// client, err := ethclient.Dial("http://localhost:18888/rpc")
	// fmt.Println(err)
	// block, err:= client.BlockByNumber(context.Background(), big.NewInt(int64(378)))
	// fmt.Println(err)
	// fmt.Println(len(block.Transactions()))

	go func() {
		//停止发送交易时间
		defer atomic.AddUint32(&closing, 1)
		time.Sleep(mesuringDuration)
	}()

	go func() {
		//统计tps结束时间
		defer atomic.AddUint32(&tpsClosing, 1)
		time.Sleep(mesuringDuration * 2)
	}()

	client, err := NewClient(Endpoint)
	if err != nil {
		logger.Fatal("err NewClient: ", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	ethStressTest(&client, ctx)

	if err = tps.StartTPSMeasuring(context.Background(), &client, &tpsClosing, &idlingDuration, logger); err != nil {
		fmt.Println("err StartTPSMeasuring:", err)
		logger.Fatal("err StartTPSMeasuring: ", err)
	}
}
