// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";

contract SafeTransfer {
    
    using SafeERC20 for IERC20;
    event Redeem(address indexed recieptor, address indexed token, uint256 amount);

    /**
     * @dev 从调用者安全转移资产到合约
     * @param token 转账的代币地址，address(0) 表示原生 ETH
     * @param amount 转账金额（仅 ERC20 有效，ETH 使用 msg.value）
     * @return 实际转账金额
     *
     * 说明：
     * - 如果 token 为 address(0)，则使用 msg.value 作为转账金额
     * - 如果 token 不为空，则从调用者转移 ERC20 代币到合约地址
     * - 使用 SafeERC20 确保 ERC20 转账的安全性
     */
    function getPayableAmount(address token, uint256 amount) internal returns (uint256) {
        if (token == address(0)) {
            amount = msg.value;
        } else if (amount > 0) {
            IERC20 oToken = IERC20(token);
            oToken.safeTransferFrom(msg.sender, address(this), amount);
        }
        return amount;
    }

    /**
     * @dev 安全赎回资产到指定接收地址
     * @param recieptor 接收资产的钱包地址（必须是 payable）
     * @param token 转账的代币地址，address(0) 表示原生 ETH
     * @param amount 转账金额
     *
     * 说明：
     * - 如果 token 为 address(0)，使用 call 方法安全转账 ETH，防止重入攻击
     * - 如果 token 不为空，使用 safeTransfer 安全转账 ERC20 代币
     * - 转账完成后触发 Redeem 事件
     */
    function _redeem(address payable recieptor, address token, uint256 amount) internal {
        if (token == address(0)) {
            // recieptor.transfer(amount);
            (bool success, ) = recieptor.call{value: amount}("");
            require(success, "ETH transfer failed");
        } else {
            IERC20 oToken = IERC20(token);
            oToken.safeTransfer(recieptor, amount);
        }
        emit Redeem(recieptor,token,amount);
    }
}
