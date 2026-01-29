// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

interface IMultiSignature{
    function getValidSignature(bytes32 msghash, uint256 lastIndex) external view returns(uint256);
}

contract multiSignatureClient {
    // 存储机制（EIP-1967 标准），使用 keccak256("contract_name.storage") 作为存储槽位，避免与其他存储冲突。
    uint256 private constant multiSignaturePosition = uint256(keccak256("org.multiSignature.storage"));
    // 
    uint256 private constant defaultIndex = 0;

    constructor(address multiSignature) {
        require(multiSignature != address(0), "multiSignatureClient : Multiple signature contract address is zero!");
        saveValue(multiSignaturePosition, uint256(uint160(multiSignature)));
    }

    function getMultiSignatureAddress() public view returns (address) {
        return address(uint160(getValue(multiSignaturePosition)));
    }

    modifier validCall() {
        checkMultiSignature();
        _;
    }

    function checkMultiSignature() internal view {
        // 获取调用时发送的 ETH 数量
        uint256 value;
        assembly {
            value := callvalue() // 等同于 msg.value，但用汇编节省 gas
        }
        // 生成消息哈希
        bytes32 msgHash = keccak256(abi.encodePacked(msg.sender, value));
        // bytes32 msgHash = keccak256(
        //     abi.encodePacked(
        //         address(this),      // 当前合约地址
        //         msg.sender,         // 调用者
        //         value,              // ETH 数量
        //         msg.data            // 完整的调用数据（可选）
        //     )
        // );
        // 获取多签合约地址
        address multiSign = getMultiSignatureAddress();
        // uint256 index = getValue(uint256(msgHash))
        // 查询多签验证结果
        uint256 newIndex = IMultiSignature(multiSign).getValidSignature(msgHash, defaultIndex);
        // 验证是否已批准
        require(newIndex > defaultIndex, "multiSignatureClient : This tx is not approved!");
    }

    function saveValue(uint256 position, uint256 value) internal {
        assembly {
            sstore(position, value)
        }
    }

    function getValue(uint256 position) internal view returns (uint256 value) {
        assembly {
            value := sload(position)
        }
    }
}