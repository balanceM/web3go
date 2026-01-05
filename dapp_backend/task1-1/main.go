package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	alchemyURL := "https://eth-sepolia.g.alchemy.com/v2/3KcnR4Q-vakMtRYSMPTck"
	client, err := ethclient.Dial(alchemyURL)
	if err != nil {
		log.Fatal(err)
	}

	// 查询区块信息
	blockNumber := big.NewInt(5671744)
	block, err := client.BlockByNumber(context.Background(), blockNumber)
	if err != nil {
		log.Fatal(err)
	}
	blockTime := time.Unix(int64(block.Time()), 0)
	fmt.Println(block.Number(), block.Hash().Hex(), blockTime.Format(time.DateTime), block.Difficulty().Uint64(), block.Transactions().Len())
	// 发送交易
	// 发送方 TestAccount2 private
	privateKeyStr := "private key"
	privateKey, err := crypto.HexToECDSA(privateKeyStr)
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("failed to convert public key")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	value := big.NewInt(100000000000000000) // 转账金额0.1 eth
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}
	gasLimit := uint64(21000)
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	// 接收方 Account1公网地址
	toAddress := common.HexToAddress("0xBC97ca8CF8C8B56EA7b9D44c75bfcB02e8BC5384")
	// 签名事务
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, nil)
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatal(err)
	}
	// 向以太坊网络广播
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}
	// 输出交易哈希
	fmt.Println(signedTx.Hash().Hex())
}
