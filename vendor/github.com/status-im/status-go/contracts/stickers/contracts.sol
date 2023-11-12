// This solidity file was added to the project to generate the ABI to consume
// these smart contracts:
// 0x0577215622f43a39f4bc9640806dfea9b10d2a36: StickerType
// 0x12824271339304d3a9f7e096e62a2a7e73b4a7e7: StickerMarket
// 0x110101156e8F0743948B2A61aFcf3994A8Fb172e: StickerPack

pragma solidity ^0.5.0;

contract Controlled {
    event NewController(address controller);

    address payable public controller;

    /// @notice Changes the controller of the contract
    /// @param _newController The new controller of the contract
    function changeController(address payable _newController) public;
}

/**
 * @dev Interface of the ERC165 standard, as defined in the
 * [EIP](https://eips.ethereum.org/EIPS/eip-165).
 *
 * Implementers can declare support of contract interfaces, which can then be
 * queried by others (`ERC165Checker`).
 *
 * For an implementation, see `ERC165`.
 */
interface IERC165 {
    /**
     * @dev Returns true if this contract implements the interface defined by
     * `interfaceId`. See the corresponding
     * [EIP section](https://eips.ethereum.org/EIPS/eip-165#how-interfaces-are-identified)
     * to learn more about how these ids are created.
     *
     * This function call must use less than 30 000 gas.
     */
    function supportsInterface(bytes4 interfaceId) external view returns (bool);
}

/**
 * @title ERC721 token receiver interface
 * @dev Interface for any contract that wants to support safeTransfers
 * from ERC721 asset contracts.
 */
contract IERC721Receiver {
    /**
     * @notice Handle the receipt of an NFT
     * @dev The ERC721 smart contract calls this function on the recipient
     * after a `safeTransfer`. This function MUST return the function selector,
     * otherwise the caller will revert the transaction. The selector to be
     * returned can be obtained as `this.onERC721Received.selector`. This
     * function MAY throw to revert and reject the transfer.
     * Note: the ERC721 contract address is always the message sender.
     * @param operator The address which called `safeTransferFrom` function
     * @param from The address which previously owned the token
     * @param tokenId The NFT identifier which is being transferred
     * @param data Additional data with no specified format
     * @return bytes4 `bytes4(keccak256("onERC721Received(address,address,uint256,bytes)"))`
     */
    function onERC721Received(
        address operator,
        address from,
        uint256 tokenId,
        bytes memory data
    ) public returns (bytes4);
}

/**
 * @dev Implementation of the `IERC165` interface.
 *
 * Contracts may inherit from this and call `_registerInterface` to declare
 * their support of an interface.
 */
contract ERC165 is IERC165 {
    /**
     * @dev See `IERC165.supportsInterface`.
     *
     * Time complexity O(1), guaranteed to always use less than 30 000 gas.
     */
    function supportsInterface(bytes4 interfaceId) external view returns (bool);
}


/**
 * @dev Required interface of an ERC721 compliant contract.
 */
contract IERC721 is IERC165 {
    event Transfer(
        address indexed from,
        address indexed to,
        uint256 indexed tokenId
    );
    event Approval(
        address indexed owner,
        address indexed approved,
        uint256 indexed tokenId
    );
    event ApprovalForAll(
        address indexed owner,
        address indexed operator,
        bool approved
    );

    /**
     * @dev Returns the number of NFTs in `owner`'s account.
     */
    function balanceOf(address owner) public view returns (uint256 balance);

    /**
     * @dev Returns the owner of the NFT specified by `tokenId`.
     */
    function ownerOf(uint256 tokenId) public view returns (address owner);

    /**
     * @dev Transfers a specific NFT (`tokenId`) from one account (`from`) to
     * another (`to`).
     *
     *
     *
     * Requirements:
     * - `from`, `to` cannot be zero.
     * - `tokenId` must be owned by `from`.
     * - If the caller is not `from`, it must be have been allowed to move this
     * NFT by either `approve` or `setApproveForAll`.
     */
    function safeTransferFrom(
        address from,
        address to,
        uint256 tokenId
    ) public;

    /**
     * @dev Transfers a specific NFT (`tokenId`) from one account (`from`) to
     * another (`to`).
     *
     * Requirements:
     * - If the caller is not `from`, it must be approved to move this NFT by
     * either `approve` or `setApproveForAll`.
     */
    function transferFrom(
        address from,
        address to,
        uint256 tokenId
    ) public;

    function approve(address to, uint256 tokenId) public;

    function getApproved(uint256 tokenId)
        public
        view
        returns (address operator);

    function setApprovalForAll(address operator, bool _approved) public;

    function isApprovedForAll(address owner, address operator)
        public
        view
        returns (bool);

    function safeTransferFrom(
        address from,
        address to,
        uint256 tokenId,
        bytes memory data
    ) public;
}

/**
 * @title ERC-721 Non-Fungible Token Standard, optional enumeration extension
 * @dev See https://eips.ethereum.org/EIPS/eip-721
 */
contract IERC721Enumerable is IERC721 {
    function totalSupply() public view returns (uint256);

    function tokenOfOwnerByIndex(address owner, uint256 index)
        public
        view
        returns (uint256 tokenId);

    function tokenByIndex(uint256 index) public view returns (uint256);
}

/**
 * @title ERC-721 Non-Fungible Token Standard, optional metadata extension
 * @dev See https://eips.ethereum.org/EIPS/eip-721
 */
contract IERC721Metadata is IERC721 {
    function name() external view returns (string memory);

    function symbol() external view returns (string memory);

    function tokenURI(uint256 tokenId) external view returns (string memory);
}

contract TokenClaimer {
    event ClaimedTokens(
        address indexed _token,
        address indexed _controller,
        uint256 _amount
    );

    function claimTokens(address _token) external;
}

/**
 * @title ERC721 Non-Fungible Token Standard basic implementation
 * @dev see https://eips.ethereum.org/EIPS/eip-721
 */
contract ERC721 is ERC165, IERC721 {
    /**
     * @dev Gets the balance of the specified address.
     * @param owner address to query the balance of
     * @return uint256 representing the amount owned by the passed address
     */
    function balanceOf(address owner) public view returns (uint256);

    /**
     * @dev Gets the owner of the specified token ID.
     * @param tokenId uint256 ID of the token to query the owner of
     * @return address currently marked as the owner of the given token ID
     */
    function ownerOf(uint256 tokenId) public view returns (address);

    /**
     * @dev Approves another address to transfer the given token ID
     * The zero address indicates there is no approved address.
     * There can only be one approved address per token at a given time.
     * Can only be called by the token owner or an approved operator.
     * @param to address to be approved for the given token ID
     * @param tokenId uint256 ID of the token to be approved
     */
    function approve(address to, uint256 tokenId) public;

    /**
     * @dev Gets the approved address for a token ID, or zero if no address set
     * Reverts if the token ID does not exist.
     * @param tokenId uint256 ID of the token to query the approval of
     * @return address currently approved for the given token ID
     */
    function getApproved(uint256 tokenId) public view returns (address);

    /**
     * @dev Sets or unsets the approval of a given operator
     * An operator is allowed to transfer all tokens of the sender on their behalf.
     * @param to operator address to set the approval
     * @param approved representing the status of the approval to be set
     */
    function setApprovalForAll(address to, bool approved) public;

    /**
     * @dev Tells whether an operator is approved by a given owner.
     * @param owner owner address which you want to query the approval of
     * @param operator operator address which you want to query the approval of
     * @return bool whether the given operator is approved by the given owner
     */
    function isApprovedForAll(address owner, address operator)
        public
        view
        returns (bool);

    /**
     * @dev Transfers the ownership of a given token ID to another address.
     * Usage of this method is discouraged, use `safeTransferFrom` whenever possible.
     * Requires the msg.sender to be the owner, approved, or operator.
     * @param from current owner of the token
     * @param to address to receive the ownership of the given token ID
     * @param tokenId uint256 ID of the token to be transferred
     */
    function transferFrom(
        address from,
        address to,
        uint256 tokenId
    ) public;

    /**
     * @dev Safely transfers the ownership of a given token ID to another address
     * If the target address is a contract, it must implement `onERC721Received`,
     * which is called upon a safe transfer, and return the magic value
     * `bytes4(keccak256("onERC721Received(address,address,uint256,bytes)"))`; otherwise,
     * the transfer is reverted.
     * Requires the msg.sender to be the owner, approved, or operator
     * @param from current owner of the token
     * @param to address to receive the ownership of the given token ID
     * @param tokenId uint256 ID of the token to be transferred
     */
    function safeTransferFrom(
        address from,
        address to,
        uint256 tokenId
    ) public;

    /**
     * @dev Safely transfers the ownership of a given token ID to another address
     * If the target address is a contract, it must implement `onERC721Received`,
     * which is called upon a safe transfer, and return the magic value
     * `bytes4(keccak256("onERC721Received(address,address,uint256,bytes)"))`; otherwise,
     * the transfer is reverted.
     * Requires the msg.sender to be the owner, approved, or operator
     * @param from current owner of the token
     * @param to address to receive the ownership of the given token ID
     * @param tokenId uint256 ID of the token to be transferred
     * @param _data bytes data to send along with a safe transfer check
     */
    function safeTransferFrom(
        address from,
        address to,
        uint256 tokenId,
        bytes memory _data
    ) public;
}

/**
 * @title ERC-721 Non-Fungible Token Standard, full implementation interface
 * @dev See https://eips.ethereum.org/EIPS/eip-721
 */
contract IERC721Full is IERC721, IERC721Enumerable, IERC721Metadata {
    // solhint-disable-previous-line no-empty-blocks
}

/**
 * @title ERC-721 Non-Fungible Token with optional enumeration extension logic
 * @dev See https://eips.ethereum.org/EIPS/eip-721
 */
contract ERC721Enumerable is ERC165, ERC721, IERC721Enumerable {
    /**
     * @dev Gets the token ID at a given index of the tokens list of the requested owner.
     * @param owner address owning the tokens list to be accessed
     * @param index uint256 representing the index to be accessed of the requested tokens list
     * @return uint256 token ID at the given index of the tokens list owned by the requested address
     */
    function tokenOfOwnerByIndex(address owner, uint256 index)
        public
        view
        returns (uint256);

    /**
     * @dev Gets the total amount of tokens stored by the contract.
     * @return uint256 representing the total amount of tokens
     */
    function totalSupply() public view returns (uint256);

    /**
     * @dev Gets the token ID at a given index of all the tokens in this contract
     * Reverts if the index is greater or equal to the total number of tokens.
     * @param index uint256 representing the index to be accessed of the tokens list
     * @return uint256 token ID at the given index of the tokens list
     */
    function tokenByIndex(uint256 index) public view returns (uint256);
}

contract ERC721Metadata is ERC165, ERC721, IERC721Metadata {
    /**
     * @dev Gets the token name.
     * @return string representing the token name
     */
    function name() external view returns (string memory);

    /**
     * @dev Gets the token symbol.
     * @return string representing the token symbol
     */
    function symbol() external view returns (string memory);

    /**
     * @dev Returns an URI for a given token ID.
     * Throws if the token ID does not exist. May return an empty string.
     * @param tokenId uint256 ID of the token to query
     */
    function tokenURI(uint256 tokenId) external view returns (string memory);
}

/**
 * @title Full ERC721 Token
 * This implementation includes all the required and some optional functionality of the ERC721 standard
 * Moreover, it includes approve all functionality using operator terminology
 * @dev see https://eips.ethereum.org/EIPS/eip-721
 */
contract ERC721Full is ERC721, ERC721Enumerable, ERC721Metadata {

}

/**
 * @author Ricardo Guilherme Schmidt (Status Research & Development GmbH)
 */
contract StickerPack is Controlled, TokenClaimer, ERC721Full {
    mapping(uint256 => uint256) public tokenPackId; //packId
    uint256 public tokenCount; //tokens buys

    /**
     * @notice controller can generate tokens at will
     * @param _owner account being included new token
     * @param _packId pack being minted
     * @return tokenId created
     */
    function generateToken(address _owner, uint256 _packId)
        external
        returns (uint256 tokenId);

    /**
     * @notice This method can be used by the controller to extract mistakenly
     *  sent tokens to this contract.
     * @param _token The address of the token contract that you want to recover
     *  set to 0 in case you want to extract ether.
     */
    function claimTokens(address _token) external;
}

interface ApproveAndCallFallBack {
    function receiveApproval(
        address from,
        uint256 _amount,
        address _token,
        bytes calldata _data
    ) external;
}

/**
 * @author Ricardo Guilherme Schmidt (Status Research & Development GmbH)
 * StickerMarket allows any address register "StickerPack" which can be sold to any address in form of "StickerPack", an ERC721 token.
 */
contract StickerMarket is Controlled, TokenClaimer, ApproveAndCallFallBack {
    event ClaimedTokens(
        address indexed _token,
        address indexed _controller,
        uint256 _amount
    );
    event MarketState(State state);
    event RegisterFee(uint256 value);
    event BurnRate(uint256 value);

    enum State {
        Invalid,
        Open,
        BuyOnly,
        Controlled,
        Closed
    }

    State public state = State.Open;
    uint256 registerFee;
    uint256 burnRate;

    //include global var to set burn rate/percentage
    address public snt; //payment token
    StickerPack public stickerPack;
    StickerType public stickerType;

    /**
     * @dev Mints NFT StickerPack in `msg.sender` account, and Transfers SNT using user allowance
     * emit NonfungibleToken.Transfer(`address(0)`, `msg.sender`, `tokenId`)
     * @notice buy a pack from market pack owner, including a StickerPack's token in msg.sender account with same metadata of `_packId`
     * @param _packId id of market pack
     * @param _destination owner of token being brought
     * @param _price agreed price
     * @return tokenId generated StickerPack token
     */
    function buyToken(
        uint256 _packId,
        address _destination,
        uint256 _price
    ) external returns (uint256 tokenId);

    /**
     * @dev emits StickerMarket.Register(`packId`, `_urlHash`, `_price`, `_contenthash`)
     * @notice Registers to sell a sticker pack
     * @param _price cost in wei to users minting this pack
     * @param _donate value between 0-10000 representing percentage of `_price` that is donated to StickerMarket at every buy
     * @param _category listing category
     * @param _owner address of the beneficiary of buys
     * @param _contenthash EIP1577 pack contenthash for listings
     * @param _fee Fee msg.sender agrees to pay for this registration
     * @return packId Market position of Sticker Pack data.
     */
    function registerPack(
        uint256 _price,
        uint256 _donate,
        bytes4[] calldata _category,
        address _owner,
        bytes calldata _contenthash,
        uint256 _fee
    ) external returns (uint256 packId);

    /**
     * @notice MiniMeToken ApproveAndCallFallBack forwarder for registerPack and buyToken
     * @param _from account calling "approve and buy"
     * @param _value must be exactly whats being consumed
     * @param _token must be exactly SNT contract
     * @param _data abi encoded call
     */
    function receiveApproval(
        address _from,
        uint256 _value,
        address _token,
        bytes calldata _data
    ) external;

    /**
     * @notice changes market state, only controller can call.
     * @param _state new state
     */
    function setMarketState(State _state) external;

    /**
     * @notice changes register fee, only controller can call.
     * @param _value total SNT cost of registration
     */
    function setRegisterFee(uint256 _value) external;

    /**
     * @notice changes burn rate percentage, only controller can call.
     * @param _value new value between 0 and 10000
     */
    function setBurnRate(uint256 _value) external;

    /**
     * @notice controller can generate packs at will
     * @param _price cost in wei to users minting with _urlHash metadata
     * @param _donate optional amount of `_price` that is donated to StickerMarket at every buy
     * @param _category listing category
     * @param _owner address of the beneficiary of buys
     * @param _contenthash EIP1577 pack contenthash for listings
     * @return packId Market position of Sticker Pack data.
     */
    function generatePack(
        uint256 _price,
        uint256 _donate,
        bytes4[] calldata _category,
        address _owner,
        bytes calldata _contenthash
    ) external returns (uint256 packId);

    /**
     * @notice removes all market data about a marketed pack, can only be called by market controller
     * @param _packId pack being purged
     * @param _limit limits categories being purged
     */
    function purgePack(uint256 _packId, uint256 _limit) external;

    /**
     * @notice controller can generate tokens at will
     * @param _owner account being included new token
     * @param _packId pack being minted
     * @return tokenId created
     */
    function generateToken(address _owner, uint256 _packId)
        external
        returns (uint256 tokenId);

    /**
     * @notice Change controller of stickerType
     * @param _newController new controller of stickerType.
     */
    function migrate(address payable _newController) external;

    /**
     * @notice This method can be used by the controller to extract mistakenly
     *  sent tokens to this contract.
     * @param _token The address of the token contract that you want to recover
     *  set to 0 in case you want to extract ether.
     */
    function claimTokens(address _token) external;

    /**
     * @notice returns pack data of token
     * @param _tokenId user token being queried
     * @return categories, registration time and contenthash
     */
    function getTokenData(uint256 _tokenId)
        external
        view
        returns (
            bytes4[] memory category,
            uint256 timestamp,
            bytes memory contenthash
        );

    // For ABI/web3.js purposes
    // fired by StickerType
    event Register(
        uint256 indexed packId,
        uint256 dataPrice,
        bytes contenthash
    );
    // fired by StickerPack and MiniMeToken
    event Transfer(
        address indexed from,
        address indexed to,
        uint256 indexed value
    );
}

/**
 * @author Ricardo Guilherme Schmidt (Status Research & Development GmbH)
 * StickerMarket allows any address register "StickerPack" which can be sold to any address in form of "StickerPack", an ERC721 token.
 */
contract StickerType is Controlled, TokenClaimer, ERC721Full {
    event Register(
        uint256 indexed packId,
        uint256 dataPrice,
        bytes contenthash,
        bool mintable
    );
    event PriceChanged(uint256 indexed packId, uint256 dataPrice);
    event MintabilityChanged(uint256 indexed packId, bool mintable);
    event ContenthashChanged(uint256 indexed packid, bytes contenthash);
    event Categorized(bytes4 indexed category, uint256 indexed packId);
    event Uncategorized(bytes4 indexed category, uint256 indexed packId);
    event Unregister(uint256 indexed packId);

    struct Pack {
        bytes4[] category;
        bool mintable;
        uint256 timestamp;
        uint256 price; //in "wei"
        uint256 donate; //in "percent"
        bytes contenthash;
    }

    mapping(uint256 => Pack) public packs;
    uint256 public packCount; //pack registers

    /**
     * @notice controller can generate packs at will
     * @param _price cost in wei to users minting with _urlHash metadata
     * @param _donate optional amount of `_price` that is donated to StickerMarket at every buy
     * @param _category listing category
     * @param _owner address of the beneficiary of buys
     * @param _contenthash EIP1577 pack contenthash for listings
     * @return packId Market position of Sticker Pack data.
     */
    function generatePack(
        uint256 _price,
        uint256 _donate,
        bytes4[] calldata _category,
        address _owner,
        bytes calldata _contenthash
    ) external returns (uint256 packId);

    /**
     * @notice removes all market data about a marketed pack, can only be called by market controller
     * @param _packId position to be deleted
     * @param _limit limit of categories to cleanup
     */
    function purgePack(uint256 _packId, uint256 _limit) external;

    /**
     * @notice changes contenthash of `_packId`, can only be called by controller
     * @param _packId which market position is being altered
     * @param _contenthash new contenthash
     */
    function setPackContenthash(uint256 _packId, bytes calldata _contenthash)
        external;

    /**
     * @notice This method can be used by the controller to extract mistakenly
     *  sent tokens to this contract.
     * @param _token The address of the token contract that you want to recover
     *  set to 0 in case you want to extract ether.
     */
    function claimTokens(address _token) external;

    /**
     * @notice changes price of `_packId`, can only be called when market is open
     * @param _packId pack id changing price settings
     * @param _price cost in wei to users minting this pack
     * @param _donate value between 0-10000 representing percentage of `_price` that is donated to StickerMarket at every buy
     */
    function setPackPrice(
        uint256 _packId,
        uint256 _price,
        uint256 _donate
    ) external;

    /**
     * @notice add caregory in `_packId`, can only be called when market is open
     * @param _packId pack adding category
     * @param _category category to list
     */
    function addPackCategory(uint256 _packId, bytes4 _category) external;

    /**
     * @notice remove caregory in `_packId`, can only be called when market is open
     * @param _packId pack removing category
     * @param _category category to unlist
     */
    function removePackCategory(uint256 _packId, bytes4 _category) external;

    /**
     * @notice Changes if pack is enabled for sell
     * @param _packId position edit
     * @param _mintable true to enable sell
     */
    function setPackState(uint256 _packId, bool _mintable) external;

    /**
     * @notice read available market ids in a category (might be slow)
     * @param _category listing category
     * @return array of market id registered
     */
    function getAvailablePacks(bytes4 _category)
        external
        view
        returns (uint256[] memory availableIds);

    /**
     * @notice count total packs in a category
     * @param _category listing category
     * @return total number of packs in category
     */
    function getCategoryLength(bytes4 _category)
        external
        view
        returns (uint256 size);

    /**
     * @notice read a packId in the category list at a specific index
     * @param _category listing category
     * @param _index index
     * @return packId on index
     */
    function getCategoryPack(bytes4 _category, uint256 _index)
        external
        view
        returns (uint256 packId);

    /**
     * @notice returns all data from pack in market
     * @param _packId pack id being queried
     * @return categories, owner, mintable, price, donate and contenthash
     */
    function getPackData(uint256 _packId)
        external
        view
        returns (
            bytes4[] memory category,
            address owner,
            bool mintable,
            uint256 timestamp,
            uint256 price,
            bytes memory contenthash
        );

    /**
     * @notice returns all data from pack in market
     * @param _packId pack id being queried
     * @return categories, owner, mintable, price, donate and contenthash
     */
    function getPackSummary(uint256 _packId)
        external
        view
        returns (
            bytes4[] memory category,
            uint256 timestamp,
            bytes memory contenthash
        );

    /**
     * @notice returns payment data for migrated contract
     * @param _packId pack id being queried
     * @return owner, mintable, price and donate
     */
    function getPaymentData(uint256 _packId)
        external
        view
        returns (
            address owner,
            bool mintable,
            uint256 price,
            uint256 donate
        );
}
