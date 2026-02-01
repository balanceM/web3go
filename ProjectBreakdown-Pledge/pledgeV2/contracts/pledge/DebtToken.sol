// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "./AddressPrivileges.sol";

contract DebtToken is ERC20, AddressPrivileges {
    constructor(string memory _name, string memory _symbol, address multiSignature)
        ERC20(_name, _symbol) AddressPrivileges(multiSignature) {
    }

    function mint(address to, uint256 _amount) public onlyMinter returns(bool) {
        _mint(to, _amount);
        return true;
    }

    function burn(address to, uint256 _amount) public onlyMinter returns(bool) {
        _burn(to, _amount);
        return true;
    }
}