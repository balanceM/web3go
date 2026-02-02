// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "../multiSignature/multiSignatureClient.sol";
import "@chainlink/contracts/src/v0.8/shared/interfaces/AggregatorV3Interface.sol";


contract BscPledgeOracle is multiSignatureClient {
    mapping(uint256 => AggregatorV3Interface) internal assetsMap;
    mapping(uint256 => uint256) internal decimalsMap;
    mapping(uint256 => uint256) internal priceMap;
    uint256 internal decimals = 1;

    constructor(address multiSignature) multiSignatureClient(multiSignature) {
    }

    function setDecimals(uint256 newDecimals) public validCall {
        decimals = newDecimals;
    }

    function setPrices(uint256[] memory assets, uint256[] memory prices) external validCall {
        require(assets.length == prices.length, "BscPledgeOracle: assets and prices length mismatch");
        uint256 len = assets.length;
        for (uint256 i=0; i<len; i++) {
            priceMap[i] = prices[i];
        }
    }

    function getPrices(uint256[] memory assets) external view returns (uint256[] memory) {
        uint256 len = assets.length;
        uint256[] memory prices = new uint256[](len);
        for (uint256 i = 0; i < len; i++) {
            prices[i] = getUnderlyingPrice(assets[i]);
        }
    }

    function getPrice(address asset) public view returns (uint256) {
        return getUnderlyingPrice(uint256(asset));
    }

    // *** 获取代币的最小底价 ***
    // 例如，usdt代币, 预言机的返回精度decimals是10**8，则price显示的大约是1*10**8（也可能是99990000）
    // 1. 当最小精度tokenDeciamls是10**6，即最小低价显示是1*10**(-6), 即0.000001
    //    按erc20的18位表示，即是 10**(18-6) = 10**12
    // 2. 当最小精度tokenDecimals是10**20，即最低价是1*10**(-20), 即0.00000000000000000001
    //    按erc20的18位表示，即是1/10**(20-18) = 0
    function getUnderlyingPrice(uint256 underlying) public view returns (uint256) {
        AggregatorV3Interface assetsPrice = assetsMap[underlying];
        if (address(assetsPrice) != address(0)){
            (, int price, , ,) = assetsPrice.latestRoundData();
            uint256 tokenDecimals = decimalsMap[underlying];
            if (tokenDecimals < 18){ // solidity不支持负数指数
                return uint256(price)/decimals*(10**(18-tokenDecimals));
            }else if (tokenDecimals > 18){ // solidity不支持负数指数
                return uint256(price)/decimals/(10**(tokenDecimals-18));
            }else{
                return uint256(price)/decimals;
            }
        }else {
            return priceMap[underlying];
        }
    }

    function setPrice(address asset,uint256 price) public validCall {
        priceMap[uint256(asset)] = price;
    }

    function setUnderlyingPrice(uint256 underlying,uint256 price) public validCall {
        require(underlying>0 , "underlying cannot be zero");
        priceMap[underlying] = price;
    }

    function setAssetsAggregator(address asset,address aggergator,uint256 _decimals) public validCall {
        assetsMap[uint256(asset)] = AggregatorV3Interface(aggergator);
        decimalsMap[uint256(asset)] = _decimals;
    }

    function setUnderlyingAggregator(uint256 underlying,address aggergator,uint256 _decimals) public validCall {
        require(underlying>0 , "underlying cannot be zero");
        assetsMap[underlying] = AggregatorV3Interface(aggergator);
        decimalsMap[underlying] = _decimals;
    }

    function getAssetsAggregator(address asset) public view returns (address,uint256) {
        return (address(assetsMap[uint256(asset)]),decimalsMap[uint256(asset)]);
    }

    function getUnderlyingAggregator(uint256 underlying) public view returns (address,uint256) {
        return (address(assetsMap[underlying]),decimalsMap[underlying]);
    }
}