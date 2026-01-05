pragma solidity ^0.8.28;

// 计数器合约
contract Counter {
    uint256 public count;

    // 构造函数
    constructor() {
        count = 0;
    }

    // 增加计数
    function increment() public {
        count += 1;
    }
}