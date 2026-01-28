// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

interface IUniswapV2Router02 {
    /// @notice 获取 Uniswap V2 工厂合约地址
    /// @return 工厂合约的地址
    function factory() external pure returns (address);

    /// @notice 获取 WETH（Wrapped ETH）地址
    /// @return WETH 代币的地址
    function WETH() external pure returns (address);

    /// @notice 向流动性池添加流动性（ERC20-ERC20）
    /// @dev 用户需先授权 Router 使用代币
    /// @param tokenA 第一个代币地址
    /// @param tokenB 第二个代币地址
    /// @param amountADesired 期望添加的 tokenA 数量
    /// @param amountBDesired 期望添加的 tokenB 数量
    /// @param amountAMin 可接受的 tokenA 最小数量（滑点保护）
    /// @param amountBMin 可接受的 tokenB 最小数量（滑点保护）
    /// @param to 流动性代币（LP Token）接收地址
    /// @param deadline 交易过期时间戳
    /// @return amountA 实际添加的 tokenA 数量
    /// @return amountB 实际添加的 tokenB 数量
    /// @return liquidity 获得的 LP Token 数量
    function addLiquidity(
        address tokenA,
        address tokenB,
        uint amountADesired,
        uint amountBDesired,
        uint amountAMin,
        uint amountBMin,
        address to,
        uint deadline
    ) external returns (uint amountA, uint amountB, uint liquidity);

    /// @notice 向流动性池添加流动性（ERC20-ETH）
    /// @dev ETH 通过 msg.value 传入，自动转换为 WETH。参数含义同上
    /// @param token ERC20 代币地址
    /// @return amountToken 实际添加的 token 数量
    /// @return amountETH 实际添加的 ETH 数量
    /// @return liquidity 获得的 LP Token 数量
    function addLiquidityETH(
        address token,
        uint amountTokenDesired,
        uint amountTokenMin,
        uint amountETHMin,
        address to,
        uint deadline
    ) external payable returns (uint amountToken, uint amountETH, uint liquidity);

    /// @notice 从流动性池移除流动性（ERC20-ERC20）
    /// @dev 销毁 LP Token，按比例取回两个代币
    /// @param liquidity 要移除的 LP Token 数量
    /// @return amountA 实际取回的 tokenA 数量
    /// @return amountB 实际取回的 tokenB 数量
    function removeLiquidity(
        address tokenA,
        address tokenB,
        uint liquidity,
        uint amountAMin,
        uint amountBMin,
        address to,
        uint deadline
    ) external returns (uint amountA, uint amountB);

    /// @notice 从流动性池移除流动性（ERC20-ETH）
    /// @dev 销毁 LP Token，WETH 自动转换为 ETH 发送。参数含义同上
    /// @return amountToken 实际取回的 token 数量
    /// @return amountETH 实际取回的 ETH 数量
    function removeLiquidityETH(
        address token,
        uint liquidity,
        uint amountTokenMin,
        uint amountETHMin,
        address to,
        uint deadline
    ) external returns (uint amountToken, uint amountETH);

    /// @notice 从流动性池移除流动性（ERC20-ERC20），使用 Permit 省略授权步骤
    /// @dev 使用 EIP-712 Permit 签名代替单独的授权交易，节省 gas
    /// @param approveMax 是否批准最大数量（v, r, s 为签名参数）
    /// @return amountA 实际取回的 tokenA 数量
    /// @return amountB 实际取回的 tokenB 数量
    function removeLiquidityWithPermit(
        address tokenA,
        address tokenB,
        uint liquidity,
        uint amountAMin,
        uint amountBMin,
        address to,
        uint deadline,
        bool approveMax, uint8 v, bytes32 r, bytes32 s
    ) external returns (uint amountA, uint amountB);

    /// @notice 从流动性池移除流动性（ERC20-ETH），使用 Permit 省略授权步骤
    /// @dev 同上，适用于 ETH 池。参数含义同上
    /// @return amountToken 实际取回的 token 数量
    /// @return amountETH 实际取回的 ETH 数量
    function removeLiquidityETHWithPermit(
        address token,
        uint liquidity,
        uint amountTokenMin,
        uint amountETHMin,
        address to,
        uint deadline,
        bool approveMax, uint8 v, bytes32 r, bytes32 s
    ) external returns (uint amountToken, uint amountETH);

    /// @notice 精确输入换输出（ERC20-ERC20）
    /// @dev 指定输入数量，尽可能多地换取输出代币
    /// @param amountIn 精确输入的代币数量
    /// @param amountOutMin 可接受的输出最小数量（滑点保护）
    /// @param path 交易路径数组，path[0]为输入代币，path[path.length-1]为输出代币
    /// @param to 输出代币接收地址
    /// @param deadline 交易过期时间戳
    /// @return amounts 每个步骤的输入/输出数量数组
    function swapExactTokensForTokens(
        uint amountIn,
        uint amountOutMin,
        address[] calldata path,
        address to,
        uint deadline
    ) external returns (uint[] memory amounts);

    /// @notice 精确输出换输入（ERC20-ERC20）
    /// @dev 指定输出数量，尽可能少地使用输入代币
    /// @param amountOut 精确需要的输出数量
    /// @param amountInMax 可接受的输入最大数量（滑点保护）
    /// @return amounts 每个步骤的输入/输出数量数组
    function swapTokensForExactTokens(
        uint amountOut,
        uint amountInMax,
        address[] calldata path,
        address to,
        uint deadline
    ) external returns (uint[] memory amounts);

    /// @notice 精确输入换输出（ETH-ERC20）
    /// @dev ETH 通过 msg.value 传入，path[0] 必须为 WETH 地址。参数含义同上
    /// @return amounts 每个步骤的输入/输出数量数组
    function swapExactETHForTokens(
        uint amountOutMin,
        address[] calldata path,
        address to,
        uint deadline
    ) external payable returns (uint[] memory amounts);

    /// @notice 精确输出换输入（ERC20-ETH）
    /// @dev 最终输出 ETH，自动将 WETH 转换为 ETH 发送。参数含义同上
    /// @return amounts 每个步骤的输入/输出数量数组
    function swapTokensForExactETH(
        uint amountOut,
        uint amountInMax,
        address[] calldata path,
        address to,
        uint deadline
    ) external returns (uint[] memory amounts);

    /// @notice 精确输入换输出（ERC20-ETH）
    /// @dev 代币换 ETH，path[path.length-1] 必须为 WETH 地址。参数含义同上
    /// @return amounts 每个步骤的输入/输出数量数组
    function swapExactTokensForETH(
        uint amountIn, uint amountOutMin, address[] calldata path, address to, uint deadline
    ) external returns (uint[] memory amounts);

    /// @notice 精确输出换输入（ETH-ERC20）
    /// @dev ETH 换代币，尽可能少使用 ETH。参数含义同上
    /// @return amounts 每个步骤的输入/输出数量数组
    function swapETHForExactTokens(
        uint amountOut, address[] calldata path, address to, uint deadline
    ) external payable returns (uint[] memory amounts);

    /// @notice 价格查询：根据池子储备量计算理论交换比例
    /// @dev 假设 0.3% 手续费，计算 amountA 可换多少 amountB
    /// @return amountB 理论可获得的 amountB 数量
    function quote(uint amountA, uint reserveA, uint reserveB) external pure returns (uint amountB);

    /// @notice 计算单跳交换的输出数量
    /// @dev 考虑 0.3% 手续费后的实际输出数量
    /// @param amountIn 输入数量
    /// @param reserveIn 输入代币储备量
    /// @param reserveOut 输出代币储备量
    /// @return amountOut 实际可获得的输出数量
    function getAmountOut(uint amountIn, uint reserveIn, uint reserveOut) external pure returns (uint amountOut);

    /// @notice 计算单跳交换的输入数量
    /// @dev 计算获得指定输出数量需要多少输入（含手续费）
    /// @return amountIn 需要的输入数量
    function getAmountIn(uint amountOut, uint reserveIn, uint reserveOut) external pure returns (uint amountIn);

    /// @notice 计算多跳交换的输出数量数组
    /// @dev 遍历 path 路径，计算每个跳的输出数量
    /// @param amountIn 初始输入数量
    /// @param path 交换路径，path[0] 为输入，path[path.length-1] 为最终输出
    /// @return amounts 每个跳的输出数量数组，amounts[amounts.length-1] 为最终输出
    function getAmountsOut(uint amountIn, address[] calldata path) external view returns (uint[] memory amounts);

    /// @notice 计算多跳交换的输入数量数组
    /// @dev 反向计算，从最终输出推算需要的输入数量
    /// @param amountOut 期望的最终输出数量
    /// @return amounts 每个跳的输入数量数组，amounts[0] 为需要的初始输入
    function getAmountsIn(uint amountOut, address[] calldata path) external view returns (uint[] memory amounts);

    /// @notice 从流动性池移除流动性（ERC20-ETH），支持转账费代币
    /// @dev 适用于 USDT、USDC 等每笔转账都会收取手续费的代币。参数含义同上
    /// @return amountETH 实际取回的 ETH 数量（因代币转账费，token 返回数量可能少于预期）
    function removeLiquidityETHSupportingFeeOnTransferTokens(
        address token,
        uint liquidity,
        uint amountTokenMin,
        uint amountETHMin,
        address to,
        uint deadline
    ) external returns (uint amountETH);

    /// @notice 从流动性池移除流动性（ERC20-ETH），使用 Permit 且支持转账费代币
    /// @dev 结合 Permit 和转账费代币支持，参数含义同上
    /// @return amountETH 实际取回的 ETH 数量
    function removeLiquidityETHWithPermitSupportingFeeOnTransferTokens(
        address token,
        uint liquidity,
        uint amountTokenMin,
        uint amountETHMin,
        address to,
        uint deadline,
        bool approveMax, uint8 v, bytes32 r, bytes32 s
    ) external returns (uint amountETH);

    /// @notice 精确输入换输出（ERC20-ERC20），支持转账费代币
    /// @dev 适用于每笔转账都收费的代币，实际收到的代币数量可能少于预期。参数含义同上
    function swapExactTokensForTokensSupportingFeeOnTransferTokens(
        uint amountIn,
        uint amountOutMin,
        address[] calldata path,
        address to,
        uint deadline
    ) external;

    /// @notice 精确输入换输出（ETH-ERC20），支持转账费代币
    /// @dev ETH 通过 msg.value 传入，适用于转账费代币。参数含义同上
    function swapExactETHForTokensSupportingFeeOnTransferTokens(
        uint amountOutMin,
        address[] calldata path,
        address to,
        uint deadline
    ) external payable;

    /// @notice 精确输入换输出（ERC20-ETH），支持转账费代币
    /// @dev 代币换 ETH，适用于转账费代币。参数含义同上
    function swapExactTokensForETHSupportingFeeOnTransferTokens(
        uint amountIn,
        uint amountOutMin,
        address[] calldata path,
        address to,
        uint deadline
    ) external;
}