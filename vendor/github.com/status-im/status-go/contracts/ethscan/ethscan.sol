/**
 *Submitted for verification at Etherscan.io on 2021-04-07
*/

// SPDX-License-Identifier: MIT

pragma solidity ^0.8.3;

/**
 * @title An Ether or token balance scanner
 * @author Maarten Zuidhoorn
 * @author Luit Hollander
 */
abstract contract BalanceScanner {
  struct Result {
    bool success;
    bytes data;
  }

  /**
   * @notice Get the Ether balance for all addresses specified
   * @param addresses The addresses to get the Ether balance for
   * @return results The Ether balance for all addresses in the same order as specified
   */
  function etherBalances(address[] calldata addresses) external virtual view returns (Result[] memory results);

  /**
   * @notice Get the ERC-20 token balance of `token` for all addresses specified
   * @dev This does not check if the `token` address specified is actually an ERC-20 token
   * @param addresses The addresses to get the token balance for
   * @param token The address of the ERC-20 token contract
   * @return results The token balance for all addresses in the same order as specified
   */
  function tokenBalances(address[] calldata addresses, address token) external virtual view returns (Result[] memory results);

  /**
   * @notice Get the ERC-20 token balance from multiple contracts for a single owner
   * @param owner The address of the token owner
   * @param contracts The addresses of the ERC-20 token contracts
   * @return results The token balances in the same order as the addresses specified
   */
  function tokensBalance(address owner, address[] calldata contracts) external virtual view returns (Result[] memory results);

  /**
   * @notice Call multiple contracts with the provided arbitrary data
   * @param contracts The contracts to call
   * @param data The data to call the contracts with
   * @return results The raw result of the contract calls
   */
  function call(address[] calldata contracts, bytes[] calldata data) external virtual view returns (Result[] memory results);

  /**
   * @notice Call multiple contracts with the provided arbitrary data
   * @param contracts The contracts to call
   * @param data The data to call the contracts with
   * @param gas The amount of gas to call the contracts with
   * @return results The raw result of the contract calls
   */
  function call(
    address[] calldata contracts,
    bytes[] calldata data,
    uint256 gas
  ) public view virtual returns (Result[] memory results);
}