// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "../multiSignature/multiSignatureClient.sol";
import "@openzeppelin/contracts/utils/structs/EnumerableSet.sol";

// 管理 Minter（铸造者）权限
contract AddressPrivileges is multiSignatureClient {
    constructor(address multiSignature) multiSignatureClient(multiSignature) {   
    }

    using EnumerableSet for EnumerableSet.AddressSet;
    EnumerableSet.AddressSet private _minters;

    function addMinter(address _addMinter) public validCall returns (bool) {
        require(_addMinter != address(0), "Token: _addMinter is zero address");
        return EnumerableSet.add(_minters, _addMinter);
    }

    function delMinter(address _delMinter) public validCall returns (bool) {
        require(_delMinter != address(0), "Token: _delMinter is zero address");
        return EnumerableSet.remove(_minters, _delMinter);
    }

    function getMinterLength() public view returns (uint256) {
        return EnumerableSet.length(_minters);
    }

    function isMinter(address account) public view returns (bool) {
        return EnumerableSet.contains(_minters, account);
    }

    function getMinter(uint256 _index) public view returns (address) {
        require(_index <= getMinterLength() - 1, "Token: _index out of bounds");
        return EnumerableSet.at(_minters, _index);
    }

    modifier onlyMinter() {
        require(isMinter(msg.sender), "Token: caller is not a minter");
        _;
    }
}