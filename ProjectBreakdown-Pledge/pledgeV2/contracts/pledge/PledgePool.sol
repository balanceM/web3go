// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import "../library/SafeTransfer.sol";
import "../interface/IDebtToken.sol";
import "../interface/IBscPledgeOracle.sol";
import "../interface/IUniswapV2Router02.sol";
import "../multiSignature/multiSignatureClient.sol";


contract PledgePool is ReentrancyGuard, SafeTransfer, multiSignatureClient {

    using SafeMath for uint256;
    using SafeERC20 for IERC20;

    uint256 constant internal calDecimal = 1e18;
    uint256 constant internal baseDecimal = 1e8;
    uint256 public minAmount = 100e18;

    uint256 constant baseYear = 365 days;

    enum PoolState{MATCH, EXECUTION, FINISH, LIQUIDATION, UNDONE}
    PoolState constant defaultChoice = PoolState.MATCH;

    bool public globalPaused = false;
    // pancake swap router
    address public swapRouter;
    // receiving fee address
    address payable public feeAddress;
    // oracle address
    IBscPledgeOracle public oracle;
    // fee
    uint256 public lendFee;
    uint256 public borrowFee;

    // base info per pool
    struct PoolBaseInfo {
        uint256 settleTime; // 结算时间
        uint256 endTime; // 结束时间
        uint256 interestRate; // 池的固定利率，单位是1e8
        uint256 maxSupply; //池子的最大限额
        uint256 lendSupply; //出借方当前总存款量（lendToken总锁仓量）
        uint256 borrowSupply; //借款方当前总借款量（borrowToken总借出量）
        uint256 martgageRate; //池的抵押率，单位是1e8
        address lendToken; //出借方存入的代币地址（如BUSD/USDT）
        address borrowToken;// 借款方借出的代币地址（如BTC/ETH）
        PoolState state; //状态 'MATCH, EXECUTION, FINISH, LIQUIDATION, UNDONE'
        IDebtToken spCoin;// 出借方凭证代币
        IDebtToken jpCoin;// 借款方凭证代币
        uint256 autoLiquidateThreshold;// 自动清算阈值 (触发清算阈值)
    }
    // total base pool
    PoolBaseInfo[] public poolBaseInfo;

    // data info per pool
    struct PoolDataInfo {
        uint256 settleAmountLend;
        uint256 settleAmountBorrow;
        uint256 finishAmountLend;
        uint256 finishAmountBorrow;
        uint256 liquidateAmountLend;
        uint256 liquidateAmountBorrow;
    }
    // total data pool
    PoolDataInfo[] public poolDataInfo;


}