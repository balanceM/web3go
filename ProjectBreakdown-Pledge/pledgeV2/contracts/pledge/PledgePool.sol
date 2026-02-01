// SPDX-License-Identifier: MIT
pragma solidity ^0.8.12;

import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import "../library/SafeTransfer.sol";
import "../interface/IDebtToken.sol";
import "../interface/IBscPledgeOracle.sol";
import "../interface/IUniswapV2Router02.sol";
import "../multiSignature/multiSignatureClient.sol";

contract PledgePool is ReentrancyGuard, SafeTransfer, multiSignatureClient {

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
        uint256 martgageRate; //池的抵押率，单位是1e8 (质押率 = 抵押物总价值 / 借款总价值，质押率必须大于1)
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

    // user pool borrow info
    struct BorrowInfo {
        uint256 stakeAmount; // 抵押金额
        uint256 refundAmount; // 可退款金额
        bool hasNoRefund; // 是否已经退款
        bool hasNoClaim; // 是否已经赎回
    }
    // Info of each user that stakes tokens.  {user.address : {pool.index : user.borrowInfo}}
    mapping(address => mapping(uint256 => BorrowInfo)) public userBorrowInfo;

    // user pool lend info
    struct LendInfo {
        uint256 stakeAmount;
        uint256 refundAmount;
        bool hasNoRefund;
        bool hasNoClaim;
    }
    // Info of each user that stakes tokens.  {user.address : {pool.index : user.lendInfo}}
    mapping(address => mapping(uint256 => LendInfo)) public userLendInfo;

    // 出借方借入
    event DepositLend(address indexed from, address indexed token, uint256 amount, uint256 mintAmount);
    // 出借方退款
    event RefundLend(address indexed from, address indexed token, uint256 refund);
    // 出借方索赔
    event ClaimLend(address indexed from, address indexed token, uint256 amount);
    // 出借方提现
    event WithdrawLend(address indexed from, address indexed token, uint256 amount);
    // 借款方借出
    event DepositBorrow(address indexed from, address indexed token, uint256 amount, uint256 mintAmount);
    // 借款方退款
    event RefundBorrow(address indexed from, address indexed token, uint256 refund);
    // 借款方索赔
    event ClaimBorrow(address indexed from, address indexed token, uint256 amount);
    // 借款方提现，from是提取者地址，token是提取的代币地址，amount是提取的数量，burnAmount是销毁的凭证代币数量
    event WithdrawBorrow(address indexed from, address indexed token, uint256 amount, uint256 burnAmount);
    // 交换，fromCoin是交换前的币种地址，toCoin是交换后的币种地址，fromValue是交换前的数量，toValue是交换后的数量
    event Swap(address indexed fromCoin, address indexed toCoin, uint256 fromValue, uint256 toValue);
    // 借款方紧急提取，from是提取者地址，token是提取的代币地址，amount是提取的数量
    event EmergencyBorrowWithdrawal(address indexed from, address indexed token, uint256 amount);
    // 出借方紧急提取，from是提取者地址，token是提取的代币地址，amount是提取的数量
    event EmergencyLendWithdrawal(address indexed from, address indexed token, uint256 amount);
    // 状态变更，pid是池子id，beforeState是变更前的状态，afterState是变更后的状态
    event StateChange(uint256 indexed pid, uint256 indexed beforeState, uint256 indexed afterState);
    // 设置费用, newLendFee是新的出借费率，newBorrowFee是新的借款费率
    event SetFee(uint256 newLendFee, uint256 newBorrowFee);
    // 设置交换路由器地址，oldSwapAddress是旧的交换地址，newSwapAddress是新的交换地址
    event SetSwapRouterAddress(address indexed oldSwapRouterAddress, address indexed newSwapRouterAddress);
    // 设置费用地址，oldFeeAddress是旧的费用地址，newFeeAddress是新的费用地址
    event SetFeeAddress(address indexed oldFeeAddress, address indexed newFeeAddress);
    // 设置最小数量事件，oldMinAmount是旧的最小数量，newMinAmount是新的最小数量
    event SetMinAmount(uint256 indexed oldMinAmount, uint256 indexed newMinAmount);

    constructor(address _oracle, address _swapRouter, address payable _feeAddress, address _multiSignature) multiSignatureClient(_multiSignature) {
        require(_oracle != address(0), "oracle address is zero");
        require(_swapRouter != address(0), "swap router address is zero");
        require(_feeAddress != address(0), "fee address is zero");

        oracle = IBscPledgeOracle(_oracle);
        swapRouter = _swapRouter;
        feeAddress = _feeAddress;
        lendFee = 0;
        borrowFee = 0;
    }

    function setFee(uint256 _lendFee, uint256 _borrowFee) validCall external {
        lendFee = _lendFee;
        borrowFee = _borrowFee;
        emit SetFee(_lendFee, _borrowFee);
    }

    function setSwapRouterAddress(address _swapRouter) validCall external {
        require(_swapRouter != address(0), "new swap router address is zero");
        emit SetSwapRouterAddress(swapRouter, _swapRouter);
        swapRouter = _swapRouter;
    }

    function setFeeAddress(address payable _feeAddress) validCall external {
        require(_feeAddress != address(0), "new fee address is zero");
        emit SetFeeAddress(feeAddress, _feeAddress);
        feeAddress = _feeAddress;
    }

    function setMinAmount(uint256 _minAmount) validCall external {
        require(_minAmount > 0, "new min amount is zero");
        emit SetMinAmount(minAmount, _minAmount);
        minAmount = _minAmount;
    }

    function poolLength() external view returns (uint256) {
        return poolBaseInfo.length;
    }

    // 创建池子
    function createPoolInfo(uint256 _settleTime, uint256 _endTime, uint256 _instrerestRate,
                            uint256 _maxSupply, uint256 _martgageRate, address _lendToken, address _borrowToken,
                            address _spToken, address _jpToken, uint256 _autoLiquidateThreshold) validCall public {
        require(_settleTime > block.timestamp, "createPool: settle time must be greater than current time");
        require(_endTime > _settleTime, "createPool: end time must be greater than settle time");
        require(_jpToken != address(0), "createPool: jpToken address is zero");
        require(_spToken != address(0), "createPool: spToken address is zero");

        poolBaseInfo.push(PoolBaseInfo({
            settleTime: _settleTime,
            endTime: _endTime,
            interestRate: _instrerestRate,
            maxSupply: _maxSupply,
            lendSupply: 0,
            borrowSupply: 0,
            martgageRate: _martgageRate,
            lendToken: _lendToken,
            borrowToken: _borrowToken,
            state: defaultChoice,
            spCoin: IDebtToken(_spToken),
            jpCoin: IDebtToken(_jpToken),
            autoLiquidateThreshold: _autoLiquidateThreshold
        }));

        poolDataInfo.push(PoolDataInfo({
            settleAmountLend: 0,
            settleAmountBorrow: 0,
            finishAmountLend: 0,
            finishAmountBorrow: 0,
            liquidateAmountLend: 0,
            liquidateAmountBorrow: 0
        }));
    }

    // 获取池子状态
    function getPoolState(uint256 _pid) external view returns (PoolState) {
        return poolBaseInfo[_pid].state;
    }

    // 
    function depositLend(uint256 _pid, uint256 _stakeAmount) external payable nonReentrant notPause timeBefore(_pid) stateMatch(_pid) {
        PoolBaseInfo storage pool = poolBaseInfo[_pid];
        LendInfo storage lendInfo = userLendInfo[msg.sender][_pid];
        require(_stakeAmount > 0, "depositLend: stake amount must be greater than zero");
        require(_stakeAmount <= pool.maxSupply - pool.lendSupply, "depositLend: stake amount must be less than max supply");
        uint256 amount = getPayableAmount(pool.lendToken, _stakeAmount);
        require(amount > minAmount, "depositLend: amount must be greater than min amount");
        // 
        lendInfo.hasNoClaim = false;
        lendInfo.hasNoRefund = false;
        lendInfo.stakeAmount += amount;
        pool.lendSupply += amount;
        emit DepositLend(msg.sender, pool.lendToken, _stakeAmount, amount);
    }

    function refundLend(uint256 _pid) external nonReentrant notPause timeAfter(_pid) stateNotMatchUndone(_pid) {
        PoolBaseInfo storage pool = poolBaseInfo[_pid];
        PoolDataInfo storage data = poolDataInfo[_pid];
        LendInfo storage lendInfo = userLendInfo[msg.sender][_pid];
        require(lendInfo.stakeAmount > 0, "refundLend: not pledged");
        require(pool.lendSupply - data.settleAmountLend > 0, "refundLend: not refund");
        require(!lendInfo.hasNoRefund, "refundLend: repeat refund");

        // 用户份额 = 当前质押金额 / 总金额
        uint256 userShare = lendInfo.stakeAmount * calDecimal / pool.lendSupply;
        // refundAmount = 总退款金额 * 用户份额
        uint256 refundAmount = (pool.lendSupply - data.settleAmountLend) * userShare / calDecimal;
        // 退款操作
        _redeem(payable(msg.sender), pool.lendToken, refundAmount);
        // 更新用户信息
        lendInfo.hasNoRefund = true;
        lendInfo.refundAmount += refundAmount;
        // 退款事件记录
        emit RefundLend(msg.sender, pool.lendToken, refundAmount);
    }

    function claimLend(uint256 _pid) external nonReentrant notPause timeAfter(_pid) stateNotMatchUndone(_pid) {
        PoolBaseInfo storage pool = poolBaseInfo[_pid]; // 获取池的基本信息
        PoolDataInfo storage data = poolDataInfo[_pid]; // 获取池的数据信息
        LendInfo storage lendInfo = userLendInfo[msg.sender][_pid]; // 获取用户的借款信息
        // 金额限制
        require(lendInfo.stakeAmount > 0, "claimLend: stakeAmount is zero"); // 需要用户的质押金额大于0
        require(!lendInfo.hasNoClaim,"claimLend: can't claim again"); // 用户不能再次领取
        // 用户份额 = 当前质押金额 / 总金额
        uint256 userShare = lendInfo.stakeAmount * calDecimal / pool.lendSupply; 
        // totalSpAmount = settleAmountLend
        uint256 totalSpAmount = data.settleAmountLend; // 总的Sp金额等于借款结算金额
        // 用户 sp 金额 = totalSpAmount * 用户份额
        uint256 spAmount = totalSpAmount * userShare / calDecimal; 
        // 铸造 sp token
        pool.spCoin.mint(msg.sender, spAmount); 
        // 更新领取标志
        lendInfo.hasNoClaim = true;
        emit ClaimLend(msg.sender, pool.borrowToken, spAmount); // 触发领取借款事件
    }

    function withdrawLend(uint256 _pid, uint256 _spAmount) external nonReentrant notPause stateFinishLiquidation(_pid) {
        PoolBaseInfo storage pool = poolBaseInfo[_pid];
        PoolDataInfo storage data = poolDataInfo[_pid];
        // 销毁spCoin
        pool.spCoin.burn(msg.sender, _spAmount);
        // 计算销毁份额
        uint256 totalSpAmount = data.settleAmountLend;
        uint256 spShare = _spAmount * calDecimal / totalSpAmount;
        // 完成
        if (pool.state == PoolState.FINISH) {
            require(block.timestamp > pool.endTime, "withdrawLend: less than end time");
            // 赎回金额 = finishAmountLend * 销毁份额
            uint256 redeemAmount = data.finishAmountLend * spShare / calDecimal;
            // 退款动作
            _redeem(payable(msg.sender), pool.lendToken, redeemAmount);
            emit WithdrawLend(msg.sender, pool.lendToken, redeemAmount);
        }
        // 清算
        if (pool.state == PoolState.LIQUIDATION) {
            require(block.timestamp > pool.settleTime, "withdrawLend: less than settle time");
            // 赎回金额 = liquidateAmountLend * 销毁份额
            uint256 redeemAmount = data.liquidateAmountLend * spShare / calDecimal;
            // 退款动作
            _redeem(payable(msg.sender), pool.lendToken, redeemAmount);
            emit WithdrawLend(msg.sender, pool.lendToken, redeemAmount);
        }
    }

    function emergencyLendWithdrawal(uint256 _pid) external nonReentrant notPause stateUndone(_pid) {
        PoolBaseInfo storage pool = poolBaseInfo[_pid];
        require(pool.lendSupply > 0,"emergencLend: not withdrawal"); // 要求贷款供应大于0
        LendInfo storage lendInfo = userLendInfo[msg.sender][_pid]; // 获取用户的贷款信息
        require(lendInfo.stakeAmount > 0, "refundLend: not pledged"); // 要求质押金额大于0
        require(!lendInfo.hasNoRefund, "refundLend: again refund"); // 要求没有退款
        // 退款操作
        _redeem(payable(msg.sender), pool.lendToken, lendInfo.stakeAmount); // 执行赎回操作
        // 更新用户信息
        lendInfo.hasNoRefund = true; // 设置没有退款为真
        emit EmergencyLendWithdrawal(msg.sender, pool.lendToken, lendInfo.stakeAmount); // 触发紧急贷款提款事件
    }

    function depositBorrow(uint256 _pid, uint256 _stakeAmount) external payable nonReentrant notPause timeBefore(_pid) stateMatch(_pid) {
        // 
        PoolBaseInfo storage pool = poolBaseInfo[_pid];
        BorrowInfo storage borrowInfo = userBorrowInfo[msg.sender][_pid];
        uint256 amount = getPayableAmount(pool.borrowToken, _stakeAmount);
        require(amount > minAmount, "depositBorrow: amount must be greater than min amount");

        borrowInfo.hasNoClaim = false;
        borrowInfo.hasNoRefund = false;
        borrowInfo.stakeAmount += amount;
        pool.borrowSupply += amount;

        emit DepositBorrow(msg.sender, pool.borrowToken, _stakeAmount, amount);
    }

    function refundBorrow(uint256 _pid) external nonReentrant notPause timeAfter(_pid) stateNotMatchUndone(_pid) {
        PoolBaseInfo storage pool = poolBaseInfo[_pid];
        PoolDataInfo storage data = poolDataInfo[_pid];
        BorrowInfo storage borrowInfo = userBorrowInfo[msg.sender][_pid];
        require(borrowInfo.stakeAmount > 0, "refundBorrow: not pledged");
        require(pool.borrowSupply - data.settleAmountBorrow > 0, "refundBorrow: not refund");
        require(!borrowInfo.hasNoRefund, "refundBorrow: repeat refund");

        // 用户份额 = 当前质押金额 / 总金额
        uint256 userShare = borrowInfo.stakeAmount * calDecimal / pool.borrowSupply;
        // refundAmount = 总退款金额 * 用户份额
        uint256 refundAmount = (pool.borrowSupply - data.settleAmountBorrow) * userShare / calDecimal;
        // 退款操作
        _redeem(payable(msg.sender), pool.borrowToken, refundAmount);
        // 更新用户信息
        borrowInfo.hasNoRefund = true;
        borrowInfo.refundAmount += refundAmount;
        // 退款事件记录
        emit RefundBorrow(msg.sender, pool.lendToken, refundAmount);
    }

    function claimBorrow(uint256 _pid) external nonReentrant notPause timeAfter(_pid) stateNotMatchUndone(_pid) {
        PoolBaseInfo storage pool = poolBaseInfo[_pid];
        PoolDataInfo storage data = poolDataInfo[_pid];
        BorrowInfo storage borrowInfo = userBorrowInfo[msg.sender][_pid];
        // 金额限制
        require(borrowInfo.stakeAmount > 0, "claimBorrow: stakeAmount is zero");
        require(!borrowInfo.hasNoClaim, "claimBorrow: can't claim again");
        // 用户份额 = 当前质押金额 / 总金额
        uint256 userShare = borrowInfo.stakeAmount * calDecimal / pool.borrowSupply;
        // totalSpAmount = settleAmountBorrow
        uint256 totalSpAmount = data.settleAmountBorrow;
        // 用户jp金额
        uint256 jpAmount = totalSpAmount * userShare / calDecimal;
        // 铸造 jp token
        pool.jpCoin.mint(msg.sender, jpAmount);
        // 更新领取标志
        borrowInfo.hasNoClaim = true;
        // 
        emit ClaimBorrow(msg.sender, pool.borrowToken, jpAmount);
    }

    function withDrawBorrow(uint256 _pid, uint256 _jpAmount) external nonReentrant notPause stateFinishLiquidation(_pid) {
        PoolBaseInfo storage pool = poolBaseInfo[_pid];
        PoolDataInfo storage data = poolDataInfo[_pid];
        require(_jpAmount > 0, "withDrawBorrow: jpAmount must be greater than zero");
        pool.jpCoin.burn(msg.sender, _jpAmount);
        // 计算销毁份额
        uint256 totalJpAmount = data.settleAmountBorrow;
        uint256 userShare = _jpAmount * calDecimal / totalJpAmount;
        // 完成
        if (pool.state == PoolState.FINISH) {
            require(block.timestamp > pool.endTime, "withdrawBorrow: less than settle time");
            // 赎回金额
            uint256 redeemAmount = data.finishAmountBorrow * userShare / calDecimal;
            // 退款动作
            _redeem(payable(msg.sender), pool.borrowToken, redeemAmount);
            emit WithdrawBorrow(msg.sender, pool.borrowToken, redeemAmount, _jpAmount);
        }
        // 清算
        if (pool.state == PoolState.LIQUIDATION) {
            require(block.timestamp > pool.settleTime, "withdrawBorrow: less than settle time");
            // 赎回金额
            uint256 redeemAmount = data.liquidateAmountBorrow * userShare / calDecimal;
            // 退款动作
            _redeem(payable(msg.sender), pool.borrowToken, redeemAmount);
            emit WithdrawBorrow(msg.sender, pool.borrowToken, redeemAmount, _jpAmount);
        }
    }

    function emergencyBorrowWithdrawal(uint256 _pid) external nonReentrant notPause stateUndone(_pid) {
        PoolBaseInfo storage pool = poolBaseInfo[_pid];
        require(pool.borrowSupply > 0,"emergencBorrow: not withdrawal");
        BorrowInfo storage borrowInfo = userBorrowInfo[msg.sender][_pid];
        require(borrowInfo.stakeAmount > 0, "refundBorrow: not pledged");
        require(!borrowInfo.hasNoRefund, "refundBorrow: again refund");
        // 退款操作
        _redeem(payable(msg.sender), pool.borrowToken, borrowInfo.stakeAmount);
        // 更新用户信息
        borrowInfo.hasNoRefund = true;
        emit EmergencyBorrowWithdrawal(msg.sender, pool.borrowToken, borrowInfo.stakeAmount);
    }
    
    // 检查是否已结算
    function checkoutSettle(uint256 _pid) public view returns (bool) {
        PoolBaseInfo storage pool = poolBaseInfo[_pid];
        return block.timestamp > pool.settleTime;
    }

    // 结算
    function settle(uint256 _pid) public validCall timeAfter(_pid) stateMatch(_pid) {
        // 获取池信息
        PoolBaseInfo storage pool = poolBaseInfo[_pid];
        PoolDataInfo storage data = poolDataInfo[_pid];
        // 判断是否可匹配
        if (pool.lendSupply > 0 && pool.borrowSupply > 0) { // 可匹配，进行结算
            // 预言机获取价格
            uint256[2] memory prices = getUnderlyingPriceView(_pid);
            uint256 lendPrice = prices[0];
            uint256 borrowPrice = prices[1];
            // 当lendSupply全满足时，获取对应的borrowAmount
            // 借款价值 = 借出价值 * 质押率
            // 借款数量 = 借款价值 / 单价
            uint256 matchBorrowAmount = pool.lendSupply * lendPrice * pool.martgageRate / borrowPrice / baseDecimal;
            // borrowSupply不足matchBorrowAmount时
            if (matchBorrowAmount > pool.borrowSupply) {
                // 取borrowSupply全满足时, 获取matchLendAmount
                uint256 matchLendAmount = pool.borrowSupply * borrowPrice * calDecimal / lendPrice / calDecimal * baseDecimal / pool.martgageRate;
                data.settleAmountBorrow = pool.borrowSupply;
                data.settleAmountLend = matchLendAmount;
            } else { // borrowSupply满足matchBorrowAmount
                data.settleAmountBorrow = matchBorrowAmount;
                data.settleAmountLend = pool.lendSupply;
            }
            // 更新池状态
            pool.state = PoolState.EXECUTION;
            emit StateChange(_pid, uint256(PoolState.MATCH), uint256(PoolState.EXECUTION));
        } else { // 不可匹配，进入UNDONE状态
            pool.state = PoolState.UNDONE;
            data.settleAmountLend = pool.lendSupply;
            data.settleAmountBorrow = pool.borrowSupply;
            emit StateChange(_pid,uint256(PoolState.MATCH), uint256(PoolState.EXECUTION));
        }
    }

    function checkoutFinish(uint256 _pid) public view returns (bool) {
        PoolBaseInfo storage pool = poolBaseInfo[_pid];
        return block.timestamp > pool.endTime;
    }

    function finish(uint256 _pid) public validCall {
        // 获取基础池子信息和数据信息
        PoolBaseInfo storage pool = poolBaseInfo[_pid];
        PoolDataInfo storage data = poolDataInfo[_pid];
        // 判断
        require(block.timestamp > pool.endTime, "finish: less than end time");
        require(pool.state == PoolState.EXECUTION, "finish: state is not execution");
        // 获取借款和贷款的token
        (address token0, address token1) = (pool.borrowToken, pool.lendToken);
        // 计算时间比率 = ((结束时间 - 结算时间) * 基础小数)/365天
        uint256 timeRatio = (pool.endTime - pool.settleTime) * baseDecimal / baseYear;
        // 计算利息
        uint256 interest = data.settleAmountLend * pool.interestRate * timeRatio / baseDecimal / baseDecimal;
        // 计算贷款金额 = 结算贷款金额 + 利息
        uint256 lendAmount = data.settleAmountLend + interest;
        // 计算销售金额 = 贷款金额*（1+贷款费）
        uint256 sellAmount = lendAmount * (lendFee + baseDecimal) / baseDecimal;
        // 执行交换操作，预计要得到selAmount数量的token1，返回 提供的token0数量 和 实际得到的token1数量
        (uint256 amountSell, uint256 amountIn) = _sellExactAmount(swapRouter, token0, token1, sellAmount);
        // 验证交换后的金额是否大于等于贷款金额
        require(amountIn >= lendAmount, "finish: Slippage is too high");
        // 如果交换后的金额大于贷款金额，计算费用并赎回
        if (amountIn > lendAmount) {
            uint256 feeAmount = amountIn - lendAmount;
            // 贷款费
            _redeem(feeAddress, pool.lendToken, feeAmount);
            data.finishAmountLend = amountIn - feeAmount;
        } else {
            data.finishAmountLend = amountIn;
        }
        // 计算剩余的借款金额并赎回借款费
        uint256 remainNowAmount = data.settleAmountBorrow - amountSell;
        uint256 remainBorrowAmount = redeemFees(borrowFee, pool.borrowToken, remainNowAmount);
        data.finishAmountBorrow = remainBorrowAmount;

        // 更新池子状态为完成
        pool.state = PoolState.FINISH;
        emit StateChange(_pid, uint256(PoolState.EXECUTION), uint256(PoolState.FINISH));
    }

    function checkoutLiquidate(uint256 _pid) external view returns (bool) {
        PoolBaseInfo storage pool = poolBaseInfo[_pid];
        PoolDataInfo storage data = poolDataInfo[_pid];
        // 保证金价格
        uint256[2]memory prices = getUnderlyingPriceView(_pid); // 获取标的价格视图
        // 保证金当前价值 = 保证金数量 * 保证金价格
        uint256 borrowValueNow = data.settleAmountBorrow * (prices[1] * calDecimal / prices[0]) / calDecimal;
        // 清算阈值 = settleAmountLend*(1+autoLiquidateThreshold)
        uint256 valueThreshold = data.settleAmountLend *(baseDecimal + pool.autoLiquidateThreshold) / baseDecimal;
        return borrowValueNow < valueThreshold; // 如果保证金当前价值小于清算阈值，则返回true，否则返回false
    }

    function liquidate(uint256 _pid) public validCall {
        PoolDataInfo storage data = poolDataInfo[_pid]; // 获取池子的数据信息
        PoolBaseInfo storage pool = poolBaseInfo[_pid]; // 获取池子的基本信息
        require(block.timestamp > pool.settleTime, "liquidate: less than settle time"); // 需要当前时间大于结算时间
        require(pool.state == PoolState.EXECUTION,"liquidate: state must be execution"); // 需要池子的状态是执行状态
        // sellamount
        (address token0, address token1) = (pool.borrowToken, pool.lendToken); // 获取借款和贷款的token
        // 时间比率 = ((当前时间 - 结算时间) * 基础小数)/365天
        uint256 timeRatio = (block.timestamp - pool.settleTime) * baseDecimal / baseYear;
        // 利息 = 时间比率 * 利率 * 结算贷款金额
        uint256 interest = data.settleAmountLend * pool.interestRate * timeRatio / baseDecimal / baseDecimal;
        // 贷款金额 = 结算贷款金额 + 利息
        uint256 lendAmount = data.settleAmountLend + interest;
        // 添加贷款费用
        uint256 sellAmount = lendAmount * (lendFee + baseDecimal) / baseDecimal;
        (uint256 amountSell,uint256 amountIn) = _sellExactAmount(swapRouter,token0,token1,sellAmount); // 卖出准确的金额
        // 可能会有滑点，amountIn - lendAmount < 0;
        if (amountIn > lendAmount) {
            uint256 feeAmount = amountIn - lendAmount; // 费用金额
            // 贷款费用
            _redeem(feeAddress,pool.lendToken, feeAmount);
            data.liquidateAmountLend = amountIn - feeAmount;
        }else {
            data.liquidateAmountLend = amountIn;
        }
        // liquidationAmounBorrow  借款费用
        uint256 remainNowAmount = data.settleAmountBorrow - amountSell; // 剩余的现在的金额
        uint256 remainBorrowAmount = redeemFees(borrowFee, pool.borrowToken, remainNowAmount); // 剩余的借款金额
        data.liquidateAmountBorrow = remainBorrowAmount;
        // 更新池子状态
        pool.state = PoolState.LIQUIDATION;
         // 事件
        emit StateChange(_pid,uint256(PoolState.EXECUTION), uint256(PoolState.LIQUIDATION));
    }

    // 获取最新的预言机价格
    function getUnderlyingPriceView(uint256 _pid) public view returns (uint256[2] memory) {
        PoolBaseInfo storage pool = poolBaseInfo[_pid];
        // 创建资产数组
        uint256[] memory assets = new uint256[](2);
        // 将借款和贷款的token添加到资产数组中
        assets[0] = uint256(pool.lendSupply);
        assets[1] = uint256(pool.borrowSupply);
        // 从预言机获取价格
        uint256[] memory prices = oracle.getPrices(assets);
        return [prices[0], prices[1]];
    }

    // 期望得到_amountout数量token1
    // 返回 需要提供的token0数量 和 实际得到的token1数量
    function _sellExactAmount(address _swapRouter, address _token0, address _token1, uint256 _amountout) internal returns (uint256, uint256) {
        uint256 amountSell = _amountout > 0 ? _getAmountIn(_swapRouter, _token0, _token1, _amountout) : 0;
        return (amountSell, _swap(_swapRouter, _token0, _token1, amountSell));
    }

    // 获取需要提供的token0数量
    function _getAmountIn(address _swapRouter, address _token0, address _token1, uint256 _amountout) internal view returns (uint256) {
        IUniswapV2Router02 IUniswap = IUniswapV2Router02(_swapRouter);
        address[] memory path = _getSwapPath(swapRouter, _token0, _token1);
        uint[] memory amounts = IUniswap.getAmountsIn(_amountout, path);
        return amounts[0];
    }

    // 获取交换路径
    function _getSwapPath(address _swapRouter, address _token0, address _token1) internal pure returns (address[] memory path) {
        IUniswapV2Router02 IUniswap = IUniswapV2Router02(_swapRouter);
        path = new address[](2);
        path[0] = _token0 == address(0) ? IUniswap.WETH() : _token0;
        path[1] = _token1 == address(0) ? IUniswap.WETH() : _token1;
        return path;
    }

    function _swap(address _swapRouter, address _token0, address _token1, uint256 _amount0) internal returns (uint256) {
        if (_token0 != address(0)) {
            _safeApprove(_token0, address(_swapRouter), type(uint256).max);
        }
        if (_token1 != address(0)) {
            _safeApprove(_token1, address(_swapRouter), type(uint256).max);
        }
        IUniswapV2Router02 IUniswap = IUniswapV2Router02(_swapRouter);
        address[] memory path = _getSwapPath(_swapRouter, _token0, _token1);
        uint256[] memory amounts;
        if (_token0 == address(0)) {
            amounts = IUniswap.swapExactETHForTokens{value:_amount0}(0, path, address(this), block.timestamp+30);
        } else if (_token1 == address(0)) {
            amounts = IUniswap.swapExactTokensForETH(_amount0, 0, path, address(this), block.timestamp+30);
        } else {
            amounts = IUniswap.swapExactTokensForTokens(_amount0, 0, path, address(this), block.timestamp+30);
        }
        emit Swap(_token0, _token1, amounts[0], amounts[amounts.length-1]);
        return amounts[amounts.length-1];
    }

    function _safeApprove(address token, address to, uint256 value) internal {
        // 0x095ea7b3 是 ERC20 标准中 approve 函数的 函数选择器（Function Selector）。
        (bool success, bytes memory data) = token.call(abi.encodeWithSelector(0x095ea7b3, to, value));
        require(success && (data.length == 0 || abi.decode(data, (bool))), "!safeApprove");
    }

    function redeemFees(uint256 feeRatio,address token,uint256 amount) internal returns (uint256){
        // 计算费用，费用 = 金额 * 费率 / 基数
        uint256 fee = amount * feeRatio / baseDecimal;
        // 如果费用大于0
        if (fee > 0){
            // 从费用地址赎回相应的费用
            _redeem(feeAddress,token, fee);
        }
        // 返回金额减去费用
        return amount - fee;
    }

    // 
    modifier notPause() {
        require(globalPaused == false, "Stake has been suspended");
        _;
    }

    modifier timeBefore(uint256 _pid) {
        require(poolBaseInfo[_pid].settleTime > block.timestamp, "Current time time must be less than settle time");
        _;
    }

    modifier timeAfter(uint256 _pid) {
        require(block.timestamp > poolBaseInfo[_pid].settleTime, "Current time must be greater than settle time");
        _;
    }

    modifier stateMatch(uint256 _pid) {
        require(poolBaseInfo[_pid].state == PoolState.MATCH, "Pool state must be MATCH");
        _;
    }
    
    modifier stateUndone(uint256 _pid) {
        require(poolBaseInfo[_pid].state == PoolState.UNDONE,"state: state must be Undone");
        _;
    }

    modifier stateNotMatchUndone(uint256 _pid) {
        require(poolBaseInfo[_pid].state == PoolState.EXECUTION || poolBaseInfo[_pid].state == PoolState.FINISH || poolBaseInfo[_pid].state == PoolState.LIQUIDATION, "Pool state is not match or undone");
        _;
    }

    modifier stateFinishLiquidation(uint256 _pid) {
        require(poolBaseInfo[_pid].state == PoolState.FINISH || poolBaseInfo[_pid].state == PoolState.LIQUIDATION, "Pool state must be FINISH or LIQUIDATION");
        _;
    }
}