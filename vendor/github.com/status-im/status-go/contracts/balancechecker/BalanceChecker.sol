// SPDX-License-Identifier: MIT
pragma solidity >=0.4.22 <0.9.0;

// ERC20 contract interface
abstract contract Token {
    function balanceOf(address) public view virtual returns (uint);
}

contract BalanceChecker {
    function tokenBalance(
        address user,
        address token
    ) public view returns (uint) {
        return Token(token).balanceOf(user);
    }

    function balancesPerAddress(
        address user,
        address[] memory tokens
    ) public view returns (uint[] memory) {
        uint[] memory addrBalances = new uint[](
            tokens.length + 1
        );
        for (uint i = 0; i < tokens.length; i++) {
            addrBalances[i] = tokenBalance(user, tokens[i]);
        }

        addrBalances[tokens.length] = user.balance;
        return addrBalances;
    }

    function balancesHash(
        address[] calldata users,
        address[] calldata tokens
    ) external view returns (uint256, bytes32[] memory) {
        bytes32[] memory addrBalances = new bytes32[](users.length);

        for (uint i = 0; i < users.length; i++) {
            addrBalances[i] = keccak256(
                abi.encodePacked(balancesPerAddress(users[i], tokens))
            );
        }

        return (block.number, addrBalances);
    }
}
