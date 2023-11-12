// This solidity file was added to the project to generate the ABI to consume
// the smart contract deployed at 0xDB5ac1a559b02E12F29fC0eC0e37Be8E046DEF49

pragma solidity ^0.4.24;

contract Controlled {
    address public controller;

    /// @notice Changes the controller of the contract
    /// @param _newController The new controller of the contract
    function changeController(address _newController) public;
}

contract ApproveAndCallFallBack {
    function receiveApproval(
        address from,
        uint256 _amount,
        address _token,
        bytes _data
    ) public;
}

contract UsernameRegistrar is Controlled, ApproveAndCallFallBack {
    address public token;
    address public ensRegistry;
    address public resolver;
    address public parentRegistry;

    uint256 public constant releaseDelay = 365 days;
    mapping(bytes32 => Account) public accounts;
    mapping(bytes32 => SlashReserve) reservedSlashers;

    //Slashing conditions
    uint256 public usernameMinLength;
    bytes32 public reservedUsernamesMerkleRoot;

    event RegistryState(RegistrarState state);
    event RegistryPrice(uint256 price);
    event RegistryMoved(address newRegistry);
    event UsernameOwner(bytes32 indexed nameHash, address owner);

    enum RegistrarState {
        Inactive,
        Active,
        Moved
    }
    bytes32 public ensNode;
    uint256 public price;
    RegistrarState public state;
    uint256 public reserveAmount;

    struct Account {
        uint256 balance;
        uint256 creationTime;
        address owner;
    }

    struct SlashReserve {
        address reserver;
        uint256 blockNumber;
    }

    /**
     * @notice Registers `_label` username to `ensNode` setting msg.sender as owner.
     * Terms of name registration:
     * - SNT is deposited, not spent; the amount is locked up for 1 year.
     * - After 1 year, the user can release the name and receive their deposit back (at any time).
     * - User deposits are completely protected. The contract controller cannot access them.
     * - User's address(es) will be publicly associated with the ENS name.
     * - User must authorise the contract to transfer `price` `token.name()`  on their behalf.
     * - Usernames registered with less then `usernameMinLength` characters can be slashed.
     * - Usernames contained in the merkle tree of root `reservedUsernamesMerkleRoot` can be slashed.
     * - Usernames starting with `0x` and bigger then 12 characters can be slashed.
     * - If terms of the contract change—e.g. Status makes contract upgrades—the user has the right to release the username and get their deposit back.
     * @param _label Choosen unowned username hash.
     * @param _account Optional address to set at public resolver.
     * @param _pubkeyA Optional pubkey part A to set at public resolver.
     * @param _pubkeyB Optional pubkey part B to set at public resolver.
     */
    function register(
        bytes32 _label,
        address _account,
        bytes32 _pubkeyA,
        bytes32 _pubkeyB
    ) external returns (bytes32 namehash);

    /**
     * @notice Release username and retrieve locked fee, needs to be called
     * after `releasePeriod` from creation time by ENS registry owner of domain
     * or anytime by account owner when domain migrated to a new registry.
     * @param _label Username hash.
     */
    function release(bytes32 _label) external;

    /**
     * @notice update account owner, should be called by new ens node owner
     * to update this contract registry, otherwise former owner can release
     * if domain is moved to a new registry.
     * @param _label Username hash.
     **/
    function updateAccountOwner(bytes32 _label) external;

    /**
     * @notice secretly reserve the slashing reward to `msg.sender`
     * @param _secret keccak256(abi.encodePacked(namehash, creationTime, reserveSecret))
     */
    function reserveSlash(bytes32 _secret) external;

    /**
     * @notice Slash username smaller then `usernameMinLength`.
     * @param _username Raw value of offending username.
     */
    function slashSmallUsername(string _username, uint256 _reserveSecret)
        external;

    /**
     * @notice Slash username starting with "0x" and with length greater than 12.
     * @param _username Raw value of offending username.
     */
    function slashAddressLikeUsername(string _username, uint256 _reserveSecret)
        external;

    /**
     * @notice Slash username that is exactly a reserved name.
     * @param _username Raw value of offending username.
     * @param _proof Merkle proof that name is listed on merkle tree.
     */
    function slashReservedUsername(
        string _username,
        bytes32[] _proof,
        uint256 _reserveSecret
    ) external;

    /**
     * @notice Slash username that contains a non alphanumeric character.
     * @param _username Raw value of offending username.
     * @param _offendingPos Position of non alphanumeric character.
     */
    function slashInvalidUsername(
        string _username,
        uint256 _offendingPos,
        uint256 _reserveSecret
    ) external;

    /**
     * @notice Clear resolver and ownership of unowned subdomians.
     * @param _labels Sequence to erase.
     */
    function eraseNode(bytes32[] _labels) external;

    /**
     * @notice Migrate account to new registry, opt-in to new contract.
     * @param _label Username hash.
     **/
    function moveAccount(bytes32 _label, UsernameRegistrar _newRegistry)
        external;

    /**
     * @notice Activate registration.
     * @param _price The price of registration.
     */
    function activate(uint256 _price) external;

    /**
     * @notice Updates Public Resolver for resolving users.
     * @param _resolver New PublicResolver.
     */
    function setResolver(address _resolver) external;

    /**
     * @notice Updates registration price.
     * @param _price New registration price.
     */
    function updateRegistryPrice(uint256 _price) external;

    /**
     * @notice Transfer ownership of ensNode to `_newRegistry`.
     * Usernames registered are not affected, but they would be able to instantly release.
     * @param _newRegistry New UsernameRegistrar for hodling `ensNode` node.
     */
    function moveRegistry(UsernameRegistrar _newRegistry) external;

    /**
     * @notice Opt-out migration of username from `parentRegistry()`.
     * Clear ENS resolver and subnode owner.
     * @param _label Username hash.
     */
    function dropUsername(bytes32 _label) external;

    /**
     * @notice Withdraw not reserved tokens
     * @param _token Address of ERC20 withdrawing excess, or address(0) if want ETH.
     * @param _beneficiary Address to send the funds.
     **/
    function withdrawExcessBalance(address _token, address _beneficiary)
        external;

    /**
     * @notice Withdraw ens nodes not belonging to this contract.
     * @param _domainHash Ens node namehash.
     * @param _beneficiary New owner of ens node.
     **/
    function withdrawWrongNode(bytes32 _domainHash, address _beneficiary)
        external;

    /**
     * @notice Gets registration price.
     * @return Registration price.
     **/
    function getPrice() external view returns (uint256 registryPrice);

    /**
     * @notice reads amount tokens locked in username
     * @param _label Username hash.
     * @return Locked username balance.
     **/
    function getAccountBalance(bytes32 _label)
        external
        view
        returns (uint256 accountBalance);

    /**
     * @notice reads username account owner at this contract,
     * which can release or migrate in case of upgrade.
     * @param _label Username hash.
     * @return Username account owner.
     **/
    function getAccountOwner(bytes32 _label)
        external
        view
        returns (address owner);

    /**
     * @notice reads when the account was registered
     * @param _label Username hash.
     * @return Registration time.
     **/
    function getCreationTime(bytes32 _label)
        external
        view
        returns (uint256 creationTime);

    /**
     * @notice calculate time where username can be released
     * @param _label Username hash.
     * @return Exact time when username can be released.
     **/
    function getExpirationTime(bytes32 _label)
        external
        view
        returns (uint256 releaseTime);

    /**
     * @notice calculate reward part an account could payout on slash
     * @param _label Username hash.
     * @return Part of reward
     **/
    function getSlashRewardPart(bytes32 _label)
        external
        view
        returns (uint256 partReward);

    /**
     * @notice Support for "approveAndCall". Callable only by `token()`.
     * @param _from Who approved.
     * @param _amount Amount being approved, need to be equal `getPrice()`.
     * @param _token Token being approved, need to be equal `token()`.
     * @param _data Abi encoded data with selector of `register(bytes32,address,bytes32,bytes32)`.
     */
    function receiveApproval(
        address _from,
        uint256 _amount,
        address _token,
        bytes _data
    ) public;

    /**
     * @notice Continues migration of username to new registry.
     * @param _label Username hash.
     * @param _tokenBalance Amount being transfered from `parentRegistry()`.
     * @param _creationTime Time user registrated in `parentRegistry()` is preserved.
     * @param _accountOwner Account owner which migrated the account.
     **/
    function migrateUsername(
        bytes32 _label,
        uint256 _tokenBalance,
        uint256 _creationTime,
        address _accountOwner
    ) external;

    /**
     * @dev callable only by parent registry to continue migration
     * of registry and activate registration.
     * @param _price The price of registration.
     **/
    function migrateRegistry(uint256 _price) external;
}
