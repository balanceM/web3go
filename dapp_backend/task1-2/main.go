package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	client, err := ethclient.Dial("https://eth-sepolia.g.alchemy.com/v2/3KcnR4Q-vakMtRYSMPTck")
	if err != nil {
		panic(err)
	}
	chainID, err := client.NetworkID(context.Background())
	// 部署计数器合约
	contractAddrStr := deployContract(client, chainID) // 0x9Ab5D4D9A0491f691241fAbB935A095867966656
	// contractAddrStr := "0x9Ab5D4D9A0491f691241fAbB935A095867966656"
	// 增加合约计数器的值
	incrementCounter(client, chainID, contractAddrStr)
}

// 增加合约计数器的值
func incrementCounter(client *ethclient.Client, chainID *big.Int, contractAddrStr string) {
	toContract := common.HexToAddress(contractAddrStr)
	//
	privateKeyStr := "private key"
	privateKey, err := crypto.HexToECDSA(privateKeyStr)
	if err != nil {
		log.Fatal(err)
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("error casting public key to ECDSA")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	// 使用地址获取地址的 nonce 值
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}
	// 估算 gas 价格
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	// 准备交易 calldata
	contractABI, err := abi.JSON(strings.NewReader(`[{"inputs":[],"stateMutability":"nonpayable","type":"constructor"},{"inputs":[],"name":"count","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"increment","outputs":[],"stateMutability":"nonpayable","type":"function"}]`))
	if err != nil {
		log.Fatal(err)
	}

	methodName := "increment"
	input, err := contractABI.Pack(methodName)
	// 创建并签名交易
	tx := types.NewTransaction(nonce, toContract, big.NewInt(0), 300000, gasPrice, input)
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatal(err)
	}
	// 发送签名好的交易
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Transaction sent: %s\n", signedTx.Hash().Hex())
	_, err = waitForReceipt(client, signedTx.Hash())
	if err != nil {
		log.Fatal(err)
	}
	// 创建 call 查询
	callInput, err := contractABI.Pack("count")
	if err != nil {
		log.Fatal(err)
	}
	callMsg := ethereum.CallMsg{
		To:   &toContract,
		Data: callInput,
	}
	// 解析返回值
	result, err := client.CallContract(context.Background(), callMsg, nil)
	if err != nil {
		log.Fatal(err)
	}
	var count *big.Int
	err = contractABI.UnpackIntoInterface(&count, "count", result)
	// count, err := contractABI.Unpack("count", result)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("New count: ", count)
}

// 部署合约（仅 client）
func deployContract(client *ethclient.Client, chainID *big.Int) string {
	// Counter合约的bin字节码
	contractBytecode := "6080604052348015600e575f5ffd5b505f5f8190555061012f806100225f395ff3fe6080604052348015600e575f5ffd5b50600436106030575f3560e01c806306661abd146034578063d09de08a14604e575b5f5ffd5b603a6056565b604051604591906089565b60405180910390f35b6054605b565b005b5f5481565b60015f5f828254606a919060cd565b92505081905550565b5f819050919050565b6083816073565b82525050565b5f602082019050609a5f830184607c565b92915050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b5f60d5826073565b915060de836073565b925082820190508082111560f35760f260a0565b5b9291505056fea2646970667358221220b29f56d123c65b7209c3ea69b25d373ff9ef811e80376bd8e365999e634c728f64736f6c63430008210033"
	//
	privateKey, err := crypto.HexToECDSA("private key")
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("error casting public key to ECDSA")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	// 获取nonce
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}
	// 获取建议的gas价格
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// 解码合约字节码
	data, err := hex.DecodeString(contractBytecode)
	if err != nil {
		log.Fatal(err)
	}
	// 创建交易
	tx := types.NewContractCreation(nonce, big.NewInt(0), 3000000, gasPrice, data)
	// 签名交易
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatal(err)
	}
	// 发送交易
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Transaction sent: %s\n", signedTx.Hash().Hex())
	// 等待交易被挖矿
	receipt, err := waitForReceipt(client, signedTx.Hash())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Contract deployed at: %s\n", receipt.ContractAddress.Hex())
	return receipt.ContractAddress.Hex()
}
func waitForReceipt(client *ethclient.Client, txHash common.Hash) (*types.Receipt, error) {
	for {
		receipt, err := client.TransactionReceipt(context.Background(), txHash)
		if err == nil {
			return receipt, nil
		}
		if err != ethereum.NotFound {
			return nil, err
		}
		// 等待一段时间后再次查询
		time.Sleep(1 * time.Second)
	}
}
