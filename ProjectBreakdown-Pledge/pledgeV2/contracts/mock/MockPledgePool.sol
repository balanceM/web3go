// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "../library/SafeTransfer.sol";
import "../interface/IDebtToken.sol";
import "../interface/IBscPledgeOracle.sol";
import "../interface/IUniswapV2Router02.sol";
import "../multiSignature/multiSignatureClient.sol";

contract MockPledgePool is ReentrancyGuard, Ownable, SafeTransfer, multiSignatureClient {

    using SafeERC20 for IERC20;

    uint256 constant internal calDecimal = 1e18;
    uint256 constant internal baseDecimal = 1e8;
    uint256 public minAmount = 100e18;

    uint256 constant baseYear = 365 days;

    enum PoolState{MATCH, EXECUTION, FINISH, LIQUIDATION, UNDONE}
    PoolState constant defaultChoice = PoolState.MATCH;

    bool public globalPaused = false;
    address public swapRouter;
    address payable public feeAddress;
    IBscPledgeOracle public oracle;
    uint256 public lendFee;
    uint256 public borrowFee;

    struct PoolBaseInfo {
        uint256 settleTime;
        uint256 endTime;
        uint256 interestRate;
        uint256 maxSupply;
        uint256 lendSupply;
        uint256 borrowSupply;
        uint256 martgageRate;
        address lendToken;
        address borrowToken;
        PoolState state;
        IDebtToken spCoin;
        IDebtToken jpCoin;
        uint256 autoLiquidateThreshold;
    }
    PoolBaseInfo[] public poolBaseInfo;

    struct PoolDataInfo {
        uint256 settleAmountLend;
        uint256 settleAmountBorrow;
        uint256 finishAmountLend;
        uint256 finishAmountBorrow;
        uint256 liquidateAmountLend;
        uint256 liquidateAmountBorrow;
    }
    PoolDataInfo[] public poolDataInfo;

    struct BorrowInfo {
        uint256 stakeAmount;
        uint256 refundAmount;
        bool hasNoRefund;
        bool hasNoClaim;
    }
    mapping(address => mapping(uint256 => BorrowInfo)) public userBorrowInfo;

    struct LendInfo {
        uint256 stakeAmount;
        uint256 refundAmount;
        bool hasNoRefund;
        bool hasNoClaim;
    }
    mapping(address => mapping(uint256 => LendInfo)) public userLendInfo;

    event DepositLend(address indexed from, address indexed token, uint256 amount, uint256 mintAmount);
    event RefundLend(address indexed from, address indexed token, uint256 refund);
    event ClaimLend(address indexed from, address indexed token, uint256 amount);
    event WithdrawLend(address indexed from, address indexed token, uint256 amount);
    event DepositBorrow(address indexed from, address indexed token, uint256 amount, uint256 mintAmount);
    event RefundBorrow(address indexed from, address indexed token, uint256 refund);
    event ClaimBorrow(address indexed from, address indexed token, uint256 amount);
    event WithdrawBorrow(address indexed from, address indexed token, uint256 amount, uint256 burnAmount);
    event Swap(address indexed fromCoin, address indexed toCoin, uint256 fromValue, uint256 toValue);
    event EmergencyBorrowWithdrawal(address indexed from, address indexed token, uint256 amount);
    event EmergencyLendWithdrawal(address indexed from, address indexed token, uint256 amount);
    event StateChange(uint256 indexed pid, uint256 indexed beforeState, uint256 indexed afterState);
    event SetFee(uint256 newLendFee, uint256 newBorrowFee);
    event SetSwapRouterAddress(address indexed oldSwapRouterAddress, address indexed newSwapRouterAddress);
    event SetFeeAddress(address indexed oldFeeAddress, address indexed newFeeAddress);
    event SetMinAmount(uint256 indexed oldMinAmount, uint256 indexed newMinAmount);

    constructor(address _oracle, address _swapRouter, address payable _feeAddress, address _multiSignature) Ownable(msg.sender) multiSignatureClient(_multiSignature) {
        oracle = IBscPledgeOracle(_oracle);
        swapRouter = _swapRouter;
        feeAddress = _feeAddress;
        lendFee = 0;
        borrowFee = 0;
    }

    // Mock helper function - bypass validCall modifier
    function setPoolState(uint256 _pid, PoolState _newState) external onlyOwner {
        PoolState oldState = poolBaseInfo[_pid].state;
        poolBaseInfo[_pid].state = _newState;
        emit StateChange(_pid, uint256(oldState), uint256(_newState));
    }

    function setPoolDataInfo(uint256 _pid, PoolDataInfo memory _data) external onlyOwner {
        poolDataInfo[_pid] = _data;
    }

    function setGlobalPaused(bool _paused) external onlyOwner {
        globalPaused = _paused;
    }

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

    function poolLength() external view returns (uint256) {
        return poolBaseInfo.length;
    }

    function getPoolState(uint256 _pid) external view returns (PoolState) {
        return poolBaseInfo[_pid].state;
    }

    function depositLend(uint256 _pid, uint256 _stakeAmount) external payable nonReentrant notPause timeBefore(_pid) stateMatch(_pid) {
        PoolBaseInfo storage pool = poolBaseInfo[_pid];
        LendInfo storage lendInfo = userLendInfo[msg.sender][_pid];
        require(_stakeAmount > 0, "depositLend: stake amount must be greater than zero");
        require(_stakeAmount <= pool.maxSupply - pool.lendSupply, "depositLend: stake amount must be less than max supply");
        uint256 amount = getPayableAmount(pool.lendToken, _stakeAmount);
        require(amount > minAmount, "depositLend: amount must be greater than min amount");
        
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

        uint256 userShare = lendInfo.stakeAmount * calDecimal / pool.lendSupply;
        uint256 refundAmount = (pool.lendSupply - data.settleAmountLend) * userShare / calDecimal;
        _redeem(payable(msg.sender), pool.lendToken, refundAmount);
        lendInfo.hasNoRefund = true;
        lendInfo.refundAmount += refundAmount;
        emit RefundLend(msg.sender, pool.lendToken, refundAmount);
    }

    function claimLend(uint256 _pid) external nonReentrant notPause timeAfter(_pid) stateNotMatchUndone(_pid) {
        PoolBaseInfo storage pool = poolBaseInfo[_pid];
        PoolDataInfo storage data = poolDataInfo[_pid];
        LendInfo storage lendInfo = userLendInfo[msg.sender][_pid];
        require(lendInfo.stakeAmount > 0, "claimLend: stakeAmount is zero");
        require(!lendInfo.hasNoClaim,"claimLend: can't claim again");
        
        uint256 userShare = lendInfo.stakeAmount * calDecimal / pool.lendSupply; 
        uint256 totalSpAmount = data.settleAmountLend;
        uint256 spAmount = totalSpAmount * userShare / calDecimal; 
        pool.spCoin.mint(msg.sender, spAmount); 
        lendInfo.hasNoClaim = true;
        emit ClaimLend(msg.sender, pool.borrowToken, spAmount); 
    }

    function withdrawLend(uint256 _pid, uint256 _spAmount) external nonReentrant notPause stateFinishLiquidation(_pid) {
        PoolBaseInfo storage pool = poolBaseInfo[_pid];
        PoolDataInfo storage data = poolDataInfo[_pid];
        pool.spCoin.burn(msg.sender, _spAmount);
        uint256 totalSpAmount = data.settleAmountLend;
        uint256 spShare = _spAmount * calDecimal / totalSpAmount;
        uint256 redeemAmount;
        
        if (pool.state == PoolState.FINISH) {
            require(block.timestamp > pool.endTime, "withdrawLend: less than end time");
            redeemAmount = data.finishAmountLend * spShare / calDecimal;
            _redeem(payable(msg.sender), pool.lendToken, redeemAmount);
            emit WithdrawLend(msg.sender, pool.lendToken, redeemAmount);
        }
        if (pool.state == PoolState.LIQUIDATION) {
            require(block.timestamp > pool.settleTime, "withdrawLend: less than settle time");
            redeemAmount = data.liquidateAmountLend * spShare / calDecimal;
            _redeem(payable(msg.sender), pool.lendToken, redeemAmount);
            emit WithdrawLend(msg.sender, pool.lendToken, redeemAmount);
        }
    }

    function emergencyLendWithdrawal(uint256 _pid) external nonReentrant notPause stateUndone(_pid) {
        PoolBaseInfo storage pool = poolBaseInfo[_pid];
        require(pool.lendSupply > 0,"emergencLend: not withdrawal"); 
        LendInfo storage lendInfo = userLendInfo[msg.sender][_pid]; 
        require(lendInfo.stakeAmount > 0, "refundLend: not pledged"); 
        require(!lendInfo.hasNoRefund, "refundLend: again refund"); 
        _redeem(payable(msg.sender), pool.lendToken, lendInfo.stakeAmount); 
        lendInfo.hasNoRefund = true; 
        emit EmergencyLendWithdrawal(msg.sender, pool.lendToken, lendInfo.stakeAmount); 
    }

    function depositBorrow(uint256 _pid, uint256 _stakeAmount) external payable nonReentrant notPause timeBefore(_pid) stateMatch(_pid) {
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

        uint256 userShare = borrowInfo.stakeAmount * calDecimal / pool.borrowSupply;
        uint256 refundAmount = (pool.borrowSupply - data.settleAmountBorrow) * userShare / calDecimal;
        _redeem(payable(msg.sender), pool.borrowToken, refundAmount);
        borrowInfo.hasNoRefund = true;
        borrowInfo.refundAmount += refundAmount;
        emit RefundBorrow(msg.sender, pool.lendToken, refundAmount);
    }

    function claimBorrow(uint256 _pid) external nonReentrant notPause timeAfter(_pid) stateNotMatchUndone(_pid) {
        PoolBaseInfo storage pool = poolBaseInfo[_pid];
        PoolDataInfo storage data = poolDataInfo[_pid];
        BorrowInfo storage borrowInfo = userBorrowInfo[msg.sender][_pid];
        require(borrowInfo.stakeAmount > 0, "claimBorrow: stakeAmount is zero");
        require(!borrowInfo.hasNoClaim, "claimBorrow: can't claim again");
        
        uint256 userShare = borrowInfo.stakeAmount * calDecimal / pool.borrowSupply;
        uint256 totalSpAmount = data.settleAmountBorrow;
        uint256 jpAmount = totalSpAmount * userShare / calDecimal;
        pool.jpCoin.mint(msg.sender, jpAmount);
        borrowInfo.hasNoClaim = true;
        emit ClaimBorrow(msg.sender, pool.borrowToken, jpAmount);
    }

    function withDrawBorrow(uint256 _pid, uint256 _jpAmount) external nonReentrant notPause stateFinishLiquidation(_pid) {
        PoolBaseInfo storage pool = poolBaseInfo[_pid];
        PoolDataInfo storage data = poolDataInfo[_pid];
        require(_jpAmount > 0, "withDrawBorrow: jpAmount must be greater than zero");
        pool.jpCoin.burn(msg.sender, _jpAmount);
        uint256 totalJpAmount = data.settleAmountBorrow;
        uint256 userShare = _jpAmount * calDecimal / totalJpAmount;
        uint256 redeemAmount;
        
        if (pool.state == PoolState.FINISH) {
            require(block.timestamp > pool.endTime, "withdrawBorrow: less than settle time");
            redeemAmount = data.finishAmountBorrow * userShare / calDecimal;
            _redeem(payable(msg.sender), pool.borrowToken, redeemAmount);
            emit WithdrawBorrow(msg.sender, pool.borrowToken, redeemAmount, _jpAmount);
        }
        if (pool.state == PoolState.LIQUIDATION) {
            require(block.timestamp > pool.settleTime, "withdrawBorrow: less than settle time");
            redeemAmount = data.liquidateAmountBorrow * userShare / calDecimal;
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
        _redeem(payable(msg.sender), pool.borrowToken, borrowInfo.stakeAmount);
        borrowInfo.hasNoRefund = true;
        emit EmergencyBorrowWithdrawal(msg.sender, pool.borrowToken, borrowInfo.stakeAmount);
    }

    function settle(uint256 _pid) public timeAfter(_pid) stateMatch(_pid) {
        PoolBaseInfo storage pool = poolBaseInfo[_pid];
        PoolDataInfo storage data = poolDataInfo[_pid];
        
        if (pool.lendSupply > 0 && pool.borrowSupply > 0) {
            uint256[2] memory prices = getUnderlyingPriceView(_pid);
            uint256 lendPrice = prices[0];
            uint256 borrowPrice = prices[1];
            uint256 matchBorrowAmount = pool.lendSupply * lendPrice * pool.martgageRate / borrowPrice / baseDecimal;
            
            if (matchBorrowAmount > pool.borrowSupply) {
                uint256 matchLendAmount = pool.borrowSupply * borrowPrice * calDecimal / lendPrice / calDecimal * baseDecimal / pool.martgageRate;
                data.settleAmountBorrow = pool.borrowSupply;
                data.settleAmountLend = matchLendAmount;
            } else {
                data.settleAmountBorrow = matchBorrowAmount;
                data.settleAmountLend = pool.lendSupply;
            }
            pool.state = PoolState.EXECUTION;
            emit StateChange(_pid, uint256(PoolState.MATCH), uint256(PoolState.EXECUTION));
        } else {
            pool.state = PoolState.UNDONE;
            data.settleAmountLend = pool.lendSupply;
            data.settleAmountBorrow = pool.borrowSupply;
            emit StateChange(_pid,uint256(PoolState.MATCH), uint256(PoolState.EXECUTION));
        }
    }

    function finish(uint256 _pid) public {
        PoolBaseInfo storage pool = poolBaseInfo[_pid];
        PoolDataInfo storage data = poolDataInfo[_pid];
        require(block.timestamp > pool.endTime, "finish: less than end time");
        require(pool.state == PoolState.EXECUTION, "finish: state is not execution");
        
        data.finishAmountLend = data.settleAmountLend * 110 / 100;
        data.finishAmountBorrow = data.settleAmountBorrow;
        
        pool.state = PoolState.FINISH;
        emit StateChange(_pid, uint256(PoolState.EXECUTION), uint256(PoolState.FINISH));
    }

    function liquidate(uint256 _pid) public {
        PoolDataInfo storage data = poolDataInfo[_pid]; 
        PoolBaseInfo storage pool = poolBaseInfo[_pid]; 
        require(block.timestamp > pool.settleTime, "liquidate: less than settle time"); 
        require(pool.state == PoolState.EXECUTION,"liquidate: state must be execution"); 
        
        data.liquidateAmountLend = data.settleAmountLend * 90 / 100;
        data.liquidateAmountBorrow = data.settleAmountBorrow * 95 / 100;
        
        pool.state = PoolState.LIQUIDATION;
        emit StateChange(_pid,uint256(PoolState.EXECUTION), uint256(PoolState.LIQUIDATION));
    }

    function getUnderlyingPriceView(uint256 _pid) public view returns (uint256[2] memory) {
        PoolBaseInfo storage pool = poolBaseInfo[_pid];
        uint256[] memory assets = new uint256[](2);
        assets[0] = uint256(pool.lendSupply);
        assets[1] = uint256(pool.borrowSupply);
        uint256[] memory prices = oracle.getPrices(assets);
        return [prices[0], prices[1]];
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