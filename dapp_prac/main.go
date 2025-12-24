package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"golang.org/x/crypto/sha3"
)

func main() {
	alchemyRpcUrl := "https://eth-sepolia.g.alchemy.com/v2/3KcnR4Q-vakMtRYSMPTck"
	client, err := ethclient.Dial(alchemyRpcUrl)
	if err != nil {
		log.Fatal(err)
	}
	// blockNumber := big.NewInt(5671744)
	// block := findBlock(client, blockNumber)
	// findTransaction(client, block)
	// findReceipt(client, blockNumber)
	// createNewWallet()
	createTransaction(client)
}

// 代币转账
func tokenTransfer(client *ethclient.Client) {
	// Account1 private key:
	privateKey, err := crypto.HexToECDSA("private key")
	if err != nil {
		log.Fatal(err)
	}
	//
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}
	//
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}
	//
	value := big.NewInt(0)
	// 预估gas费
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	// 代币接收方地址
	toAddress := common.HexToAddress("0x4592d8f8d7b001e72cb26a73e4fa1806a51ac79d")
	// 代币合约地址
	tokenAddress := common.HexToAddress("0x28b149020d2152179873ec60bed6bf7cd705775d")
	// 生成函数签名的 Keccak256 哈希
	transferFnSignature := []byte("transfer(address,uint256)")
	hash := sha3.NewLegacyKeccak256()
	hash.Write(transferFnSignature)
	methodID := hash.Sum(nil)[:4]
	fmt.Println(hexutil.Encode(methodID)) // 0xa9059cbb
	// 代币接收方地址 左填充到32字节
	paddedAddress := common.LeftPadBytes(toAddress.Bytes(), 32)
	fmt.Println(hexutil.Encode(paddedAddress))
	// 1,000 个代币，在 big.Int 中格式化为 wei。
	amount := new(big.Int)
	amount.SetString("1000000000000000000000", 10) // 1000 tokens
	// 代币量左填充到 32 个字节。
	paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)
	fmt.Println(hexutil.Encode(paddedAmount))
	// 将方法 ID，填充后的地址和填后的转账量，接到将成为我们数据字段的字节片。
	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)
	// 估算完成交易所需的估计燃气上限
	gasLimit, err := client.EstimateGas(context.Background(), ethereum.CallMsg{
		To:   &tokenAddress,
		Data: data,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(gasLimit)
	// 生成未签名以太坊事务
	tx := types.NewTransaction(nonce, tokenAddress, value, gasLimit, gasPrice, data)
	// 使用发件人的私钥对事务进行签名
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatal(err)
	}
	// 发送交易
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("tx sent: %s", signedTx.Hash().Hex())
}

// ETH转账
func createTransaction(client *ethclient.Client) {
	// Account1 private key:
	privateKey, err := crypto.HexToECDSA("private key")
	if err != nil {
		log.Fatal(err)
	}
	// 从私钥派生发送账户的公共地址
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	// 生成账户交易的随机数
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}
	// gas
	value := big.NewInt(100000000000000000) // in wei (0.1 eth)
	gasLimit := uint64(21000)               // 转账的燃气上限为"21000"单位
	// gasPrice := big.NewInt(30000000000)      // 打包交易的燃气价格为30 Gwei
	gasPrice, err := client.SuggestGasPrice(context.Background()) // 根据'x'个先前块来获得平均燃气价格。
	if err != nil {
		log.Fatal(err)
	}
	// 接收方
	// TestAccount2 public key: 0x72a99330b6872F1713E02D68Fb6e71De7a03f780
	toAddress := common.HexToAddress("0x72a99330b6872F1713E02D68Fb6e71De7a03f780")
	// 生成未签名以太坊事务
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, nil)
	// 使用发件人的私钥对事务进行签名
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatal(err)
	}
	// 调用 SendTransaction 来将已签名的事务广播到整个网络
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("tx sent: %s", signedTx.Hash().Hex()) // tx sent: 0x42d14fa6f4a9f9ccfe898761bd26bef738d2c71c99bb9453f36f61a007df12bb
}

// 创建新钱包
func createNewWallet() {
	// 1. 生成随机的ECDSA私钥
	// 使用椭圆曲线数字签名算法(ECDSA)生成私钥，这是以太坊钱包的核心
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}

	// 2. 将私钥转换为字节数组格式
	// 私钥在内存中以32字节的形式存储
	privateKeyBytes := crypto.FromECDSA(privateKey)
	// 打印私钥的十六进制表示，去掉"0x"前缀
	fmt.Println(hexutil.Encode(privateKeyBytes)[2:]) // 去掉0x

	// 3. 从私钥中提取公钥
	// 椭圆曲线中，私钥可以计算出对应的公钥
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	// 4. 从公钥生成以太坊地址
	// 以太坊地址是公钥经过Keccak-256哈希后，取最后20字节的结果
	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	fmt.Println(address)

	// 5. 手动演示地址生成过程（与上面PubkeyToAddress相同的结果）
	// 将公钥转换为字节数组格式
	// 未压缩的公钥以65字节存储，第一个字节为0x04前缀
	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
	// 打印公钥的十六进制表示，去掉"0x04"前缀（前缀用4个字符表示）
	fmt.Println("from pubKey:", hexutil.Encode(publicKeyBytes)[4:]) // 去掉0x04
	// 创建Keccak-256哈希函数
	hash := sha3.NewLegacyKeccak256()
	// 对公钥进行哈希计算（去掉前缀0x04）
	hash.Write(publicKeyBytes[1:])
	// 打印完整的32字节哈希值
	fmt.Println("full:", hexutil.Encode(hash.Sum(nil)[:]))
	// 以太坊地址是哈希值的后20位（32-12=20），即取[12:32]部分
	fmt.Println(hexutil.Encode(hash.Sum(nil)[12:])) // 原长32位，截去12位，保留后20位
}

// 查询收据
func findReceipt(client *ethclient.Client, blockNumber *big.Int) {
	fmt.Println("----------- findReceipt -----------")
	// 已知区块hash，查询区块的所有收据
	blockHash := common.HexToHash("0xae713dea1419ac72b928ebe6ba9915cd4fc1ef125a606f90f5e783c47cb1a4b5")
	receiptsByHash, err := client.BlockReceipts(context.Background(), rpc.BlockNumberOrHashWithHash(blockHash, false))
	if err != nil {
		log.Fatal(err)
	}
	// 已知区块高度，查询区块的所有收据
	receiptsByNum, err := client.BlockReceipts(context.Background(), rpc.BlockNumberOrHashWithNumber(rpc.BlockNumber(blockNumber.Int64())))
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Println(*(receiptsByHash[0]) == *(receiptsByNum[0]))
	receipt1 := receiptsByHash[0]
	receipt2 := receiptsByNum[0]
	fmt.Println(
		receipt1.Status == receipt2.Status,
		receipt1.TxHash == receipt2.TxHash,
		receipt1.BlockHash == receipt2.BlockHash,
		receipt1.BlockNumber.Uint64() == receipt2.BlockNumber.Uint64(),
		receipt1.TransactionIndex == receipt2.TransactionIndex,
	)
	// 打印收据
	for _, receipt := range receiptsByHash {
		fmt.Println(receipt.Status)
		fmt.Println(receipt.Logs)
		fmt.Println(receipt.TxHash.Hex())
		fmt.Println(receipt.TransactionIndex)
		fmt.Println(receipt.ContractAddress.Hex())
		break
	}
	// 已知交易hash，查询收据
	txHash := common.HexToHash("0x20294a03e8766e9aeab58327fc4112756017c6c28f6f99c7722f4a29075601c5")
	receipt, err := client.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(receipt.Status)
	fmt.Println(receipt.Logs)
	fmt.Println(receipt.TxHash.Hex())
	fmt.Println(receipt.TransactionIndex)
	fmt.Println(receipt.ContractAddress.Hex())
}

// 查询交易
func findTransaction(client *ethclient.Client, block *types.Block) {
	fmt.Println("----------- findTransaction -----------")
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	for _, tx := range block.Transactions() {
		fmt.Println(tx.Hash().Hex())        // 0x20294a03e8766e9aeab58327fc4112756017c6c28f6f99c7722f4a29075601c5
		fmt.Println(tx.Value().String())    // 100000000000000000
		fmt.Println(tx.Gas())               // 21000
		fmt.Println(tx.GasPrice().Uint64()) // 100000000000
		fmt.Println(tx.Nonce())             // 245132
		fmt.Println(tx.Data())              // []
		fmt.Println(tx.To().Hex())          // 0x8F9aFd209339088Ced7Bc0f57Fe08566ADda3587
		if sender, err := types.Sender(types.NewEIP155Signer(chainID), tx); err == nil {
			fmt.Println("sender: ", sender.Hex()) // 0x2CdA41645F2dBffB852a605E92B185501801FC28
		} else {
			log.Fatal(err)
		}
		receipt, err := client.TransactionReceipt(context.Background(), tx.Hash())
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(receipt.Status) // 1
		fmt.Println(receipt.Logs)   // []
		break
	}

	// 已知区块hash，查询交易
	blockHash := common.HexToHash("0xae713dea1419ac72b928ebe6ba9915cd4fc1ef125a606f90f5e783c47cb1a4b5")
	count, err := client.TransactionCount(context.Background(), blockHash)
	if err != nil {
		log.Fatal(err)
	}
	for idx := uint(0); idx < count; idx++ {
		tx, err := client.TransactionInBlock(context.Background(), blockHash, idx)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(tx.Hash().Hex()) // 0x20294a03e8766e9aeab58327fc4112756017c6c28f6f99c7722f4a29075601c5
		break
	}

	// 已知交易hash，查询交易
	txHash := common.HexToHash("0x20294a03e8766e9aeab58327fc4112756017c6c28f6f99c7722f4a29075601c5")
	tx, isPending, err := client.TransactionByHash(context.Background(), txHash)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(isPending)       // false
	fmt.Println(tx.Hash().Hex()) // 0x20294a03e8766e9aeab58327fc4112756017c6c28f6f99c7722f4a29075601c5
}

// 查询区块
func findBlock(client *ethclient.Client, blockNumber *big.Int) *types.Block {
	fmt.Println("----------- findBlock -----------")
	// header, err := client.HeaderByNumber(context.Background(), blockNumber)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println(header.Number.Uint64())
	// fmt.Println(header.Time)
	// fmt.Println(header.Difficulty.Uint64())
	// fmt.Println(header.Hash().Hex())

	block, err := client.BlockByNumber(context.Background(), blockNumber)
	if err != nil {
		log.Fatal(err)
	}

	// fmt.Println(block.Number().Uint64())
	// fmt.Println(block.Time())
	// fmt.Println(block.Difficulty().Uint64())
	fmt.Println(block.Hash().Hex())
	// fmt.Println(len(block.Transactions()))
	// count, err := client.TransactionCount(context.Background(), block.Hash())
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println(count)

	return block
}
