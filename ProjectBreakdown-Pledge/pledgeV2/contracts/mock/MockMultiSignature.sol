// SPDX-License-Identifier: MIT

pragma solidity ^0.8.20;

contract MockMultiSignature {
    uint256 public threshold = 1;
    uint256 public defaultIndex = 1;
    bool public alwaysApproved = true;

    function getValidSignature(bytes32 msghash, uint256 lastIndex) external view returns (uint256) {
        // Always return a valid signature for testing purposes
        return alwaysApproved ? defaultIndex : 0;
    }

    function setAlwaysApproved(bool _alwaysApproved) external {
        alwaysApproved = _alwaysApproved;
    }
}