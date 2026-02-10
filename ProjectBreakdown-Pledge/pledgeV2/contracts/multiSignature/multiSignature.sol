// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "./multiSignatureClient.sol";

library whiteListAddress {
    
    function addWhiteListAddress(address[] storage whiteList, address temp) internal {
        if (!isEligibleAddress(whiteList, temp)) {
            whiteList.push(temp);
        }
    }

    function removeWhiteListAddress(address[] storage whiteList, address temp) internal returns (bool) {
        uint256 len = whiteList.length;
        uint256 i = 0;
        for (;i<len;i++) {
            if (whiteList[i] == temp) break;
        }
        if (i<len) {
            if (i!=len-1) {
                whiteList[i] = whiteList[len-1];
            }
            whiteList.pop();
            return true;
        }
        return false;
    }

    function isEligibleAddress(address[] memory whiteList, address temp) internal pure returns (bool){
        for (uint256 i = 0; i < whiteList.length; i++) {
            if (whiteList[i] == temp) {
                return true;
            }
        }
        return false;
    }
}

contract multiSignature is multiSignatureClient{
    uint256 private constant defaultIndex = 0;
    using whiteListAddress for address[];

    address[] public signatureOwners;
    uint256 public threshold;
    struct signatureInfo {
        address applicant;
        address[] signatures;
    }
    mapping(bytes32=>signatureInfo[]) public signatureMap;

    event TransferOwner(address indexed sender, address indexed oldOwner, address indexed newOwner);
    event CreateApplication(address indexed from, address indexed to, bytes32 indexed msgHash);
    event SignApplication(address indexed from, bytes32 indexed msgHash, uint256 index);
    event RevokeApplication(address indexed from, bytes32 indexed msgHash, uint256 index);

    constructor(address[] memory owners, uint256 limitedSignNum) multiSignatureClient(address(this)) public {
        require(owners.length >= limitedSignNum, "Multiple Signature: Signature threshold is greater than owner's length!");
        signatureOwners = owners;
        threshold = limitedSignNum;
    }

    function transferOwner()
}