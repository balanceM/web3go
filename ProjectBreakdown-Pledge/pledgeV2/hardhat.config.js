// 注释：引入 Hardhat 集成工具箱（包含 ethers、chai、web3 等常用插件）
// require("@nomicfoundation/hardhat-toolbox");

// 注释：引入 Hardhat Ignition 声明式合约部署系统
// Ignition 用于管理和部署复杂的智能合约系统
require("@nomicfoundation/hardhat-ignition");

// 注释：引入 ethers.js 集成插件
// 提供 ethers 对象，用于与以太坊网络交互（部署合约、发送交易等）
// 新版本，替代已废弃的 @nomiclabs/hardhat-ethers
require("@nomicfoundation/hardhat-ethers");

// 注释：引入 Web3.js 集成插件
// 提供 web3 对象，是另一个以太坊交互库
// 与 ethers.js 功能类似，但 API 不同，选择其中一个使用即可
require("@nomiclabs/hardhat-web3");

// 注释：引入 dotenv 配置管理
// 从 .env 文件加载环境变量到 process.env
// 用于管理敏感信息（如私钥、API 密钥等）
require("dotenv").config();

/** @type import('hardhat/config').HardhatUserConfig */
module.exports = {
  // 注释：Mocha 测试框架配置
  // Mocha 是 Hardhat 默认的测试运行器
  mocha: {
    // 注释：设置测试超时时间为 100 秒
    // 防止测试因网络延迟等原因超时失败
    timeout: 100000
  },

  // 注释：网络配置
  // 定义部署和测试时使用的区块链网络
  networks: {
    // 注释：BSC (Binance Smart Chain) 测试网配置
    // 用于在 BSC 测试环境部署和测试合约
    bscTestnet: {
      // 注释：RPC 节点地址
      // 使用 Binance 官方提供的免费测试网节点
      url: "https://data-seed-prebsc-1-s1.binance.org:8545",

      // 注释：部署账户私钥列表
      // 从 .env 文件读取，避免将私钥提交到代码仓库
      accounts: [process.env.PRIVATE_KEY],
    },

    // 注释：Ethereum Sepolia 测试网配置
    // 用于在以太坊测试环境部署和测试合约
    sepolia: {
      // 注释：RPC 节点地址
      // 使用 Infura 提供的 Ethereum Sepolia 测试网节点
      // Infura 是专业的区块链基础设施服务商
      url: "https://sepolia.infura.io/v3/d8ed0bd1de8242d998a1405b6932ab33",

      // 注释：部署账户私钥列表
      accounts: [process.env.PRIVATE_KEY],
    },
  },

  // 注释：Solidity 编译器配置
  // 定义如何编译智能合约
  solidity: {
    // 注释：全局优化器配置
    // 启用优化可以减少合约部署时的 gas 消耗
    optimizer: {
      enabled: true,    // 注释：启用代码优化
      runs: 50,         // 注释：预计合约调用次数
                        // 值越大，优化后执行 gas 越低，但部署 gas 越高
                        // 50 是常用折中值，适合频繁调用的合约
    },

    // 注释：编译器版本列表
    // 支持同时编译多个 Solidity 版本的合约
    compilers: [
      {
        // 注释：Solidity 0.4.18 版本
        version: "0.4.18",
        settings: {
          // 注释：EVM 版本设置为 berlin
          // berlin 是以太坊的硬分叉版本号
          evmVersion: "berlin"
        }
      },
      {
        // 注释：Solidity 0.5.16 版本
        version: "0.5.16",
        settings: {
          evmVersion: "berlin"
        }
      },
      {
        // 注释：Solidity 0.6.6 版本
        version: "0.6.6",
        settings: {
          evmVersion: "berlin",
          optimizer: {
            // 注释：针对此版本的单独优化设置
            enabled: true,
            runs: 1    // 注释：runs=1 适合调用较少的合约
                       // 优化侧重于减少部署 gas
          },
        }
      },
      {
        // 注释：Solidity 0.6.12 版本
        version: "0.6.12",
        settings: {
          optimizer: {
            enabled: true,
            runs: 1
          },
          // 注释：可切换到 shanghai 版本（最新的 EVM 版本）
          // evmVersion: "shanghai"
        }
      },
      {
        version: "0.8.0",
        settings: {
          optimizer: {
            enabled: true,
            runs: 1
          },
          // evmVersion: "shanghai"
        }
      }
    ]
  }
};
