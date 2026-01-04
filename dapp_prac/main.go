package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"math/big"
	"strings"
	"time"

	token "github.com/balanceM/web3study/dapp_prac/erc20"
	store "github.com/balanceM/web3study/dapp_prac/store"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
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
	// createTransaction(client)
	// getBalance(client)
	// getTokenBalance(client)
	// subscribeBlock(client)
	// execContract(client)
	// execContractByABI(client)
	// filterQuery(client)
}

// 订阅事件
func subscribeEvent(client *ethclient.Client) {
	contractAddress := common.HexToAddress("0x2958d15bc5b64b11Ec65e623Ac50C198519f8742")
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddress},
	}
	// 通过 channel 接收事件
	logs := make(chan types.Log)
	sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		log.Fatal(err)
	}
	contractAbi, err := abi.JSON(strings.NewReader(string(StoreABI)))
	if err != nil {
		log.Fatal(err)
	}
	for {
		select {
		case err := <-sub.Err():
			log.Fatal(err)
		case vLog := <-logs:
			fmt.Println(vLog.BlockHash.Hex())
			fmt.Println(vLog.BlockNumber)
			fmt.Println(vLog.TxHash.Hex())
			event := struct {
				Key   [32]byte
				Value [32]byte
			}{}
			err := contractAbi.UnpackIntoInterface(&event, "ItemSet", vLog.Data)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(common.Bytes2Hex(event.Key[:]))
			fmt.Println(common.Bytes2Hex(event.Value[:]))
			var topics []string
			for i := range vLog.Topics {
				topics = append(topics, vLog.Topics[i].Hex())
			}
			fmt.Println("topics[0]=", topics[0])
			if len(topics) > 1 {
				fmt.Println("index topic:", topics[1:])
			}
		}
	}
}

// 查询事件
func filterQuery(client *ethclient.Client) {
	contractAddress := common.HexToAddress("0x2958d15bc5b64b11Ec65e623Ac50C198519f8742")
	// 读取特定区块所有日志
	query := ethereum.FilterQuery{
		// BlockHash
		FromBlock: big.NewInt(6920583),
		// ToBlock:   big.NewInt(2394201),
		Addresses: []common.Address{
			contractAddress,
		},
		// Topics: [][]common.Hash{
		//  {},
		//  {},
		// },
	}
	// 接收查询并将返回所有的匹配事件日志
	logs, err := client.FilterLogs(context.Background(), query)
	if err != nil {
		log.Fatal(err)
	}
	// 解码日志
	contractAbi, err := abi.JSON(strings.NewReader(StoreABI))
	if err != nil {
		log.Fatal(err)
	}
	for _, vLog := range logs {
		fmt.Println(vLog.BlockHash.Hex())
		fmt.Println(vLog.BlockNumber)
		fmt.Println(vLog.TxHash.Hex())
		event := struct {
			Key   [32]byte
			Value [32]byte
		}{}
		err := contractAbi.UnpackIntoInterface(&event, "ItemSet", vLog.Data)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(common.Bytes2Hex(event.Key[:]))
		fmt.Println(common.Bytes2Hex(event.Value[:]))
		var topics []string
		for i := range vLog.Topics {
			topics = append(topics, vLog.Topics[i].Hex())
		}
		fmt.Println("topics[0]=", topics[0])
		if len(topics) > 1 {
			fmt.Println("indexed topics:", topics[1:])
		}
	}
}

// 执行合约(abi 文件调用)
func execContractByABI(client *ethclient.Client) {
	contractAddr := "0x8D4141ec2b522dE5Cf42705C3010541B4B3EC24e"
	// 根据 hex 创建私钥实例
	privateKey, err := crypto.HexToECDSA("bba94a9d5e7850903209bb8caab4ea54fd243cd97324d674715d0803ca7c9e0b")
	if err != nil {
		log.Fatal(err)
	}
	// 从私钥实例获取公开地址
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
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
	contractABI, err := abi.JSON(strings.NewReader(`[{"inputs":[{"internalType":"string","name":"_version","type":"string"}],"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"bytes32","name":"key","type":"bytes32"},{"indexed":false,"internalType":"bytes32","name":"value","type":"bytes32"}],"name":"ItemSet","type":"event"},{"inputs":[{"internalType":"bytes32","name":"","type":"bytes32"}],"name":"items","outputs":[{"internalType":"bytes32","name":"","type":"bytes32"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"bytes32","name":"key","type":"bytes32"},{"internalType":"bytes32","name":"value","type":"bytes32"}],"name":"setItem","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"version","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"}]`))
	if err != nil {
		log.Fatal(err)
	}

	methodName := "setItem"
	var key [32]byte
	var value [32]byte

	copy(key[:], []byte("demo_save_key"))
	copy(value[:], []byte("demo_save_value11111"))
	input, err := contractABI.Pack(methodName, key, value)
	// 创建并签名交易
	chainID := big.NewInt(int64(11155111))
	tx := types.NewTransaction(nonce, common.HexToAddress(contractAddr), big.NewInt(0), 300000, gasPrice, input)
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
	callInput, err := contractABI.Pack("items", key)
	if err != nil {
		log.Fatal(err)
	}
	to := common.HexToAddress(contractAddr)
	callMsg := ethereum.CallMsg{
		To:   &to,
		Data: callInput,
	}
	// 解析返回值
	result, err := client.CallContract(context.Background(), callMsg, nil)
	if err != nil {
		log.Fatal(err)
	}

	var unpacked [32]byte
	err = contractABI.UnpackIntoInterface(&unpacked, "items", result)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("is value saving in contract equals to origin value:", unpacked == value)
}

// 执行合约
func execContract(client *ethclient.Client) {
	contractAddr := "0x8D4141ec2b522dE5Cf42705C3010541B4B3EC24e"
	// 创建合约实例
	storeContract, err := store.NewStore(common.HexToAddress(contractAddr), client)
	if err != nil {
		log.Fatal(err)
	}
	// 根据 hex 创建私钥实例
	privateKey, err := crypto.HexToECDSA("bba94a9d5e7850903209bb8caab4ea54fd243cd97324d674715d0803ca7c9e0b")
	if err != nil {
		log.Fatal(err)
	}
	// 准备数据
	var key [32]byte
	var value [32]byte
	copy(key[:], []byte("demo_save_key"))
	copy(value[:], []byte("demo_save_value11111"))
	// 初始化交易opt实例
	opt, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(11155111))
	if err != nil {
		log.Fatal(err)
	}
	// 调用合约方法
	tx, err := storeContract.SetItem(opt, key, value)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("tx hash:", tx.Hash().Hex())
	// 查询合约中的数据并验证
	callOpt := &bind.CallOpts{Context: context.Background()}
	valueInContract, err := storeContract.Items(callOpt, key)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("is value saving in contract equals to origin value:", valueInContract == value)
}

// 加载合约
func loadContract(client *ethclient.Client) {
	contractAddr := "0x8D4141ec2b522dE5Cf42705C3010541B4B3EC24e"
	storeContract, err := store.NewStore(common.HexToAddress(contractAddr), client)
	if err != nil {
		log.Fatal(err)
	}

	_ = storeContract
}

// 部署合约（仅 client）
func deployContract(client *ethclient.Client) {
	// store合约的字节码
	contractBytecode := "608060405234801561000f575f80fd5b5060405161087538038061087583398181016040528101906100319190610193565b805f908161003f91906103e7565b50506104b6565b5f604051905090565b5f80fd5b5f80fd5b5f80fd5b5f80fd5b5f601f19601f8301169050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b6100a58261005f565b810181811067ffffffffffffffff821117156100c4576100c361006f565b5b80604052505050565b5f6100d6610046565b90506100e2828261009c565b919050565b5f67ffffffffffffffff8211156101015761010061006f565b5b61010a8261005f565b9050602081019050919050565b8281835e5f83830152505050565b5f610137610132846100e7565b6100cd565b9050828152602081018484840111156101535761015261005b565b5b61015e848285610117565b509392505050565b5f82601f83011261017a57610179610057565b5b815161018a848260208601610125565b91505092915050565b5f602082840312156101a8576101a761004f565b5b5f82015167ffffffffffffffff8111156101c5576101c4610053565b5b6101d184828501610166565b91505092915050565b5f81519050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f600282049050600182168061022857607f821691505b60208210810361023b5761023a6101e4565b5b50919050565b5f819050815f5260205f209050919050565b5f6020601f8301049050919050565b5f82821b905092915050565b5f6008830261029d7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff82610262565b6102a78683610262565b95508019841693508086168417925050509392505050565b5f819050919050565b5f819050919050565b5f6102eb6102e66102e1846102bf565b6102c8565b6102bf565b9050919050565b5f819050919050565b610304836102d1565b610318610310826102f2565b84845461026e565b825550505050565b5f90565b61032c610320565b6103378184846102fb565b505050565b5b8181101561035a5761034f5f82610324565b60018101905061033d565b5050565b601f82111561039f5761037081610241565b61037984610253565b81016020851015610388578190505b61039c61039485610253565b83018261033c565b50505b505050565b5f82821c905092915050565b5f6103bf5f19846008026103a4565b1980831691505092915050565b5f6103d783836103b0565b9150826002028217905092915050565b6103f0826101da565b67ffffffffffffffff8111156104095761040861006f565b5b6104138254610211565b61041e82828561035e565b5f60209050601f83116001811461044f575f841561043d578287015190505b61044785826103cc565b8655506104ae565b601f19841661045d86610241565b5f5b828110156104845784890151825560018201915060208501945060208101905061045f565b868310156104a1578489015161049d601f8916826103b0565b8355505b6001600288020188555050505b505050505050565b6103b2806104c35f395ff3fe608060405234801561000f575f80fd5b506004361061003f575f3560e01c806348f343f31461004357806354fd4d5014610073578063f56256c714610091575b5f80fd5b61005d600480360381019061005891906101d7565b6100ad565b60405161006a9190610211565b60405180910390f35b61007b6100c2565b604051610088919061029a565b60405180910390f35b6100ab60048036038101906100a691906102ba565b61014d565b005b6001602052805f5260405f205f915090505481565b5f80546100ce90610325565b80601f01602080910402602001604051908101604052809291908181526020018280546100fa90610325565b80156101455780601f1061011c57610100808354040283529160200191610145565b820191905f5260205f20905b81548152906001019060200180831161012857829003601f168201915b505050505081565b8060015f8481526020019081526020015f20819055507fe79e73da417710ae99aa2088575580a60415d359acfad9cdd3382d59c80281d48282604051610194929190610355565b60405180910390a15050565b5f80fd5b5f819050919050565b6101b6816101a4565b81146101c0575f80fd5b50565b5f813590506101d1816101ad565b92915050565b5f602082840312156101ec576101eb6101a0565b5b5f6101f9848285016101c3565b91505092915050565b61020b816101a4565b82525050565b5f6020820190506102245f830184610202565b92915050565b5f81519050919050565b5f82825260208201905092915050565b8281835e5f83830152505050565b5f601f19601f8301169050919050565b5f61026c8261022a565b6102768185610234565b9350610286818560208601610244565b61028f81610252565b840191505092915050565b5f6020820190508181035f8301526102b28184610262565b905092915050565b5f80604083850312156102d0576102cf6101a0565b5b5f6102dd858286016101c3565b92505060206102ee858286016101c3565b9150509250929050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f600282049050600182168061033c57607f821691505b60208210810361034f5761034e6102f8565b5b50919050565b5f6040820190506103685f830185610202565b6103756020830184610202565b939250505056fea26469706673582212205aae308f77654b000c9d222eff2d9f2bd2ac18d990b10774842e4309d4e3e15664736f6c634300081a0033"
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
	fmt.Printf("Transaction sent: %s\n", signedTx.Hash().Hex())
	// 等待交易被挖矿
	receipt, err := waitForReceipt(client, signedTx.Hash())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Contract deployed at: %s\n", receipt.ContractAddress.Hex())
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

// 部署合约( 使用abigen)
func deployContractByAbigen(client *ethclient.Client) {
	//
	privateKey, err := crypto.HexToECDSA("private key")
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	chainId, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	//
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainId)
	if err != nil {
		log.Fatal(err)
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = uint64(30000)
	auth.GasPrice = gasPrice

	input := "1.0"
	address, tx, instance, err := store.DeployStore(auth, client, input)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(address.Hex())
	fmt.Println(tx.Hash().Hex())

	_ = instance
}

// 订阅区块
func subscribeBlock(client *ethclient.Client) {
	// alchemyRpcUrl := "wss://eth-sepolia.g.alchemy.com/v2/3KcnR4Q-vakMtRYSMPTck"
	headers := make(chan *types.Header)
	// 订阅新区块头
	sub, err := client.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		log.Fatal(err)
	}

	// 订阅将推送新的区块头事件到我们的通道
	for {
		select {
		case err := <-sub.Err():
			log.Fatal(err)
		case header := <-headers:
			fmt.Println(header.Hash().Hex()) // 打印新区块号
			//
			block, err := client.BlockByHash(context.Background(), header.Hash())
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(block.Hash().Hex())
			fmt.Println(block.Number().Uint64())
			fmt.Println(block.Time())
			fmt.Println(block.Number())
			fmt.Println(len(block.Transactions()))
		}
	}
}

// 【这里前提】****solcjs 编译合约，生成 ABI 文件，然后使用 abigen 工具生成 Go 绑定代码****
// 获取代币余额
// 参数: client - 以太坊客户端连接，用于与区块链交互
// 功能: 查询指定地址的ERC20代币余额，并获取代币的基本信息（名称、符号、小数位数）
func getTokenBalance(client *ethclient.Client) {
	// 1. 将代币合约地址从十六进制字符串转换为common.Address类型
	// 这是ERC20代币智能合约在以太坊网络上的部署地址
	tokenAddress := common.HexToAddress("0xfadea654ea83c00e5003d2ea15c59830b65471c0")

	// 2. 创建代币合约的绑定实例
	// 使用生成的Token绑定器与部署在指定地址的ERC20代币合约进行交互
	// token包应该是通过abigen工具从ERC20合约ABI生成的Go绑定代码

	instance, err := token.NewErc20(tokenAddress, client)
	if err != nil {
		log.Fatal(err)
	}

	// 3. 将要查询余额的以太坊地址从十六进制字符串转换为common.Address类型
	// 这是要查询代币余额的钱包地址或合约地址
	address := common.HexToAddress("0x25836239F7b632635F815689389C537133248edb")

	// 4. 调用代币合约的BalanceOf函数查询指定地址的代币余额
	// &bind.CallOpts{} 是一个空选项对象，用于配置调用参数（如从哪个区块读取数据、gas价格等）
	// 返回的bal是以太坊的最小单位wei（即原始代币数量，未考虑小数位）表示的余额
	bal, err := instance.BalanceOf(&bind.CallOpts{}, address)
	if err != nil {
		log.Fatal(err)
	}
	// 打印原始余额（以wei为单位）
	fmt.Println("wei: %s \n", bal)

	// 5. 将wei单位的余额转换为big.Float类型以便进行浮点数运算
	// big.Float可以处理高精度的浮点数计算，避免精度丢失
	fbal := new(big.Float)
	fbal.SetString(bal.String())

	// 6. 将wei转换为标准代币单位
	// 除以10的18次方（ERC20代币通常使用18位小数，类似ETH的wei转换）
	// math.Pow10(18) = 1后面跟18个0，即1000000000000000000
	// Quo方法执行除法运算，返回一个新的大浮点数
	value := new(big.Float).Quo(fbal, big.NewFloat(math.Pow10(18)))

	// 打印转换后的标准代币数量
	fmt.Println("value: %s \n", value)

	// 7. 查询代币的名称
	// 调用代币合约的name()函数，返回代币的完整名称（如"USD Coin"）
	name, err := instance.Name(&bind.CallOpts{})
	if err != nil {
		log.Fatal(err)
	}

	// 8. 查询代币的符号
	// 调用代币合约的symbol()函数，返回代币的简短符号（如"USDC"）
	symbol, err := instance.Symbol(&bind.CallOpts{})
	if err != nil {
		log.Fatal(err)
	}

	// 9. 查询代币的小数位数
	// 调用代币合约的decimals()函数，返回代币使用的小数位数
	// 大多数ERC20代币使用18位小数，但也有少数使用不同的小数位数
	decimals, err := instance.Decimals(&bind.CallOpts{})
	if err != nil {
		log.Fatal(err)
	}

	// 打印代币基本信息
	fmt.Printf("name: %s \n", name)         // 代币名称
	fmt.Printf("symbol: %s \n", symbol)     // 代币符号
	fmt.Printf("decimals: %v \n", decimals) // 小数位数
}

// 获取账户余额
func getBalance(client *ethclient.Client) {
	// TestAccount2 public address: 0x72a99330b6872F1713E02D68Fb6e71De7a03f780
	account := common.HexToAddress("0x72a99330b6872F1713E02D68Fb6e71De7a03f780")
	// blockNumber := big.NewInt(5532993)
	// balance, err := client.BalanceAt(context.Background(), account, blockNumber) //读取指定区块时的账户余额
	balance, err := client.BalanceAt(context.Background(), account, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(balance) // 1000000000000000000000
	// wei转为eth
	fbalance := new(big.Float)
	fbalance.SetString(balance.String())
	ethValue := new(big.Float).Quo(fbalance, big.NewFloat(math.Pow10(18)))
	fmt.Println(ethValue) //
	// 待处理的账户余额
	pendingBalance, err := client.PendingBalanceAt(context.Background(), account)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(pendingBalance)
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
