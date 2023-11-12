// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package communitytokendeployer

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// CommunityTokenDeployerDeploymentSignature is an auto generated low-level Go binding around an user-defined struct.
type CommunityTokenDeployerDeploymentSignature struct {
	Signer   common.Address
	Deployer common.Address
	V        uint8
	R        [32]byte
	S        [32]byte
}

// CommunityTokenDeployerTokenConfig is an auto generated low-level Go binding around an user-defined struct.
type CommunityTokenDeployerTokenConfig struct {
	Name    string
	Symbol  string
	BaseURI string
}

// CommunityTokenDeployerMetaData contains all meta data concerning the CommunityTokenDeployer contract.
var CommunityTokenDeployerMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_registry\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_ownerTokenFactory\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_masterTokenFactory\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[],\"name\":\"CommunityTokenDeployer_AlreadyDeployed\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"CommunityTokenDeployer_EqualFactoryAddresses\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"CommunityTokenDeployer_InvalidDeployerAddress\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"CommunityTokenDeployer_InvalidDeploymentRegistryAddress\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"CommunityTokenDeployer_InvalidDeploymentSignature\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"CommunityTokenDeployer_InvalidSignerKeyOrCommunityAddress\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"CommunityTokenDeployer_InvalidTokenFactoryAddress\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"CommunityTokenDeployer_InvalidTokenMetadata\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidShortString\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"str\",\"type\":\"string\"}],\"name\":\"StringTooLong\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"DeployMasterToken\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"DeployOwnerToken\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"DeploymentRegistryAddressChange\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"EIP712DomainChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"MasterTokenFactoryAddressChange\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"OwnerTokenFactoryAddressChange\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferStarted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"DEPLOYMENT_SIGNATURE_TYPEHASH\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"DOMAIN_SEPARATOR\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"acceptOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"symbol\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"baseURI\",\"type\":\"string\"}],\"internalType\":\"structCommunityTokenDeployer.TokenConfig\",\"name\":\"_ownerToken\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"symbol\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"baseURI\",\"type\":\"string\"}],\"internalType\":\"structCommunityTokenDeployer.TokenConfig\",\"name\":\"_masterToken\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"signer\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"deployer\",\"type\":\"address\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"internalType\":\"structCommunityTokenDeployer.DeploymentSignature\",\"name\":\"_signature\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"_signerPublicKey\",\"type\":\"bytes\"}],\"name\":\"deploy\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"deploymentRegistry\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"eip712Domain\",\"outputs\":[{\"internalType\":\"bytes1\",\"name\":\"fields\",\"type\":\"bytes1\"},{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"version\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"chainId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"verifyingContract\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"salt\",\"type\":\"bytes32\"},{\"internalType\":\"uint256[]\",\"name\":\"extensions\",\"type\":\"uint256[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"masterTokenFactory\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ownerTokenFactory\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"pendingOwner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_deploymentRegistry\",\"type\":\"address\"}],\"name\":\"setDeploymentRegistryAddress\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_masterTokenFactory\",\"type\":\"address\"}],\"name\":\"setMasterTokenFactoryAddress\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_ownerTokenFactory\",\"type\":\"address\"}],\"name\":\"setOwnerTokenFactoryAddress\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x6101606040523480156200001257600080fd5b50604051620020df380380620020df833981016040819052620000359162000372565b6040518060400160405280601681526020017f436f6d6d756e697479546f6b656e4465706c6f79657200000000000000000000815250604051806040016040528060018152602001603160f81b815250620000a06000836200023a60201b62000c0e1790919060201c565b61012052620000bd8160016200023a602090811b62000c0e17901c565b61014052815160208084019190912060e052815190820120610100524660a0526200014b60e05161010051604080517f8b73c3c69bb8fe3d512ecc4cf759cc79239f7b179b0ffacaa9a75d522b39400f60208201529081019290925260608201524660808201523060a082015260009060c00160405160208183030381529060405280519060200120905090565b60805250503060c0526200015f336200028a565b6001600160a01b03831662000187576040516333a066b160e11b815260040160405180910390fd5b6001600160a01b0382161580620001a557506001600160a01b038116155b15620001c457604051633b901c6960e11b815260040160405180910390fd5b806001600160a01b0316826001600160a01b031603620001f757604051631b08426360e11b815260040160405180910390fd5b600480546001600160a01b039485166001600160a01b031991821617909155600580549385169382169390931790925560068054919093169116179055620005a2565b60006020835110156200025a576200025283620002b4565b905062000284565b8262000271836200030060201b62000c411760201c565b906200027e908262000461565b5060ff90505b92915050565b600380546001600160a01b0319169055620002b18162000303602090811b62000c4417901c565b50565b600080829050601f81511115620002eb578260405163305a27a960e01b8152600401620002e291906200052d565b60405180910390fd5b8051620002f8826200057d565b179392505050565b90565b600280546001600160a01b038381166001600160a01b0319831681179093556040519116919082907f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e090600090a35050565b80516001600160a01b03811681146200036d57600080fd5b919050565b6000806000606084860312156200038857600080fd5b620003938462000355565b9250620003a36020850162000355565b9150620003b36040850162000355565b90509250925092565b634e487b7160e01b600052604160045260246000fd5b600181811c90821680620003e757607f821691505b6020821081036200040857634e487b7160e01b600052602260045260246000fd5b50919050565b601f8211156200045c57600081815260208120601f850160051c81016020861015620004375750805b601f850160051c820191505b81811015620004585782815560010162000443565b5050505b505050565b81516001600160401b038111156200047d576200047d620003bc565b62000495816200048e8454620003d2565b846200040e565b602080601f831160018114620004cd5760008415620004b45750858301515b600019600386901b1c1916600185901b17855562000458565b600085815260208120601f198616915b82811015620004fe57888601518255948401946001909101908401620004dd565b50858210156200051d5787850151600019600388901b60f8161c191681555b5050505050600190811b01905550565b600060208083528351808285015260005b818110156200055c578581018301518582016040015282016200053e565b506000604082860101526040601f19601f8301168501019250505092915050565b80516020808301519190811015620004085760001960209190910360031b1b16919050565b60805160a05160c05160e051610100516101205161014051611ae2620005fd60003960006103cb015260006103a101526000610da201526000610d7a01526000610cd501526000610cff01526000610d290152611ae26000f3fe608060405234801561001057600080fd5b50600436106100f55760003560e01c80639ff02d1811610097578063c663109211610066578063c663109214610252578063e30c397814610272578063f2fde38b14610290578063f8851475146102a357600080fd5b80639ff02d18146101c5578063a53b2bdb146101d8578063a825483c146101eb578063b0f95f281461021257600080fd5b806379ba5097116100d357806379ba509714610164578063830c26261461016c57806384b0196e1461018c5780638da5cb5b146101a757600080fd5b80633644e515146100fa57806362457f5914610115578063715018a61461015a575b600080fd5b6101026102b6565b6040519081526020015b60405180910390f35b6005546101359073ffffffffffffffffffffffffffffffffffffffff1681565b60405173ffffffffffffffffffffffffffffffffffffffff909116815260200161010c565b6101626102c5565b005b6101626102d9565b6006546101359073ffffffffffffffffffffffffffffffffffffffff1681565b610194610393565b60405161010c979695949392919061148a565b60025473ffffffffffffffffffffffffffffffffffffffff16610135565b6101626101d336600461156b565b610437565b6101626101e636600461156b565b6104fb565b6101027fdd91c30357aafeb2792b5f0facbd83995943c1ea113a906ebbeb58bfeb27dfc281565b6102256102203660046115d6565b6105bf565b6040805173ffffffffffffffffffffffffffffffffffffffff93841681529290911660208301520161010c565b6004546101359073ffffffffffffffffffffffffffffffffffffffff1681565b60035473ffffffffffffffffffffffffffffffffffffffff16610135565b61016261029e36600461156b565b610a9a565b6101626102b136600461156b565b610b4a565b60006102c0610cbb565b905090565b6102cd610df3565b6102d76000610e74565b565b600354339073ffffffffffffffffffffffffffffffffffffffff168114610387576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602960248201527f4f776e61626c6532537465703a2063616c6c6572206973206e6f74207468652060448201527f6e6577206f776e6572000000000000000000000000000000000000000000000060648201526084015b60405180910390fd5b61039081610e74565b50565b6000606080828080836103c67f000000000000000000000000000000000000000000000000000000000000000083610ea5565b6103f17f00000000000000000000000000000000000000000000000000000000000000006001610ea5565b604080516000808252602082019092527f0f000000000000000000000000000000000000000000000000000000000000009b939a50919850469750309650945092509050565b61043f610df3565b73ffffffffffffffffffffffffffffffffffffffff811661048c576040517f772038d200000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b600580547fffffffffffffffffffffffff00000000000000000000000000000000000000001673ffffffffffffffffffffffffffffffffffffffff83169081179091556040517f17ae1257210039eb267be68929104e6c28fc9ebb9dc6aaa84be39b45eb6f376790600090a250565b610503610df3565b73ffffffffffffffffffffffffffffffffffffffff8116610550576040517f772038d200000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b600680547fffffffffffffffffffffffff00000000000000000000000000000000000000001673ffffffffffffffffffffffffffffffffffffffff83169081179091556040517f184513c31b135bda32c81b8586d52ad5bdbc7b7e4ec5847eee48374ee07e8e4890600090a250565b600080806105d0602086018661156b565b73ffffffffffffffffffffffffffffffffffffffff1614806105f157508251155b15610628576040517fb5709c7f00000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b33610639604086016020870161156b565b73ffffffffffffffffffffffffffffffffffffffff1614610686576040517f3137f20700000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b60045460009073ffffffffffffffffffffffffffffffffffffffff16637db6a4e46106b4602088018861156b565b6040517fffffffff0000000000000000000000000000000000000000000000000000000060e084901b16815273ffffffffffffffffffffffffffffffffffffffff90911660048201526024016020604051808303816000875af115801561071f573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610743919061172b565b73ffffffffffffffffffffffffffffffffffffffff1614610790576040517f1f58274700000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b61079984610f49565b6107cf576040517f18f68bed00000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b60055460009073ffffffffffffffffffffffffffffffffffffffff1663cc3c79756107fa8980611748565b61080760208c018c611748565b61081460408e018e611748565b338c6040518963ffffffff1660e01b81526004016108399897969594939291906117fd565b6020604051808303816000875af1158015610858573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061087c919061172b565b60405190915073ffffffffffffffffffffffffffffffffffffffff8216907f6f60871ce1ae7c2bc82a9fca785cdf029fa7c9984afe96eaa106d1b7b19c632290600090a260065460009073ffffffffffffffffffffffffffffffffffffffff1663cc3c79756108eb8980611748565b6108f860208c018c611748565b61090560408e018e611748565b604080516020810182526000815290517fffffffff0000000000000000000000000000000000000000000000000000000060e08a901b168152610952979695949392918c916004016117fd565b6020604051808303816000875af1158015610971573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610995919061172b565b60405190915073ffffffffffffffffffffffffffffffffffffffff8216907f1464afef6e77413c9c3201405b55530340d684e2a19f3a9d83bc604d4aa3a25590600090a260045473ffffffffffffffffffffffffffffffffffffffff1663a7a95840610a04602089018961156b565b6040517fffffffff0000000000000000000000000000000000000000000000000000000060e084901b16815273ffffffffffffffffffffffffffffffffffffffff91821660048201529085166024820152604401600060405180830381600087803b158015610a7257600080fd5b505af1158015610a86573d6000803e3d6000fd5b509395509193505050505b94509492505050565b610aa2610df3565b6003805473ffffffffffffffffffffffffffffffffffffffff83167fffffffffffffffffffffffff00000000000000000000000000000000000000009091168117909155610b0560025473ffffffffffffffffffffffffffffffffffffffff1690565b73ffffffffffffffffffffffffffffffffffffffff167f38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e2270060405160405180910390a350565b610b52610df3565b73ffffffffffffffffffffffffffffffffffffffff8116610b9f576040517f6740cd6200000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b600480547fffffffffffffffffffffffff00000000000000000000000000000000000000001673ffffffffffffffffffffffffffffffffffffffff83169081179091556040517f8f3be421db34ad7dfa5c7fb9391b363b444007c7b26f0a22c58aad6e130b935e90600090a250565b6000602083511015610c2a57610c2383611040565b9050610c3b565b81610c358482611914565b5060ff90505b92915050565b90565b6002805473ffffffffffffffffffffffffffffffffffffffff8381167fffffffffffffffffffffffff0000000000000000000000000000000000000000831681179093556040519116919082907f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e090600090a35050565b60003073ffffffffffffffffffffffffffffffffffffffff7f000000000000000000000000000000000000000000000000000000000000000016148015610d2157507f000000000000000000000000000000000000000000000000000000000000000046145b15610d4b57507f000000000000000000000000000000000000000000000000000000000000000090565b6102c0604080517f8b73c3c69bb8fe3d512ecc4cf759cc79239f7b179b0ffacaa9a75d522b39400f60208201527f0000000000000000000000000000000000000000000000000000000000000000918101919091527f000000000000000000000000000000000000000000000000000000000000000060608201524660808201523060a082015260009060c00160405160208183030381529060405280519060200120905090565b60025473ffffffffffffffffffffffffffffffffffffffff1633146102d7576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015260640161037e565b600380547fffffffffffffffffffffffff000000000000000000000000000000000000000016905561039081610c44565b606060ff8314610eb857610c2383611097565b818054610ec490611878565b80601f0160208091040260200160405190810160405280929190818152602001828054610ef090611878565b8015610f3d5780601f10610f1257610100808354040283529160200191610f3d565b820191906000526020600020905b815481529060010190602001808311610f2057829003601f168201915b50505050509050610c3b565b600080610fd97fdd91c30357aafeb2792b5f0facbd83995943c1ea113a906ebbeb58bfeb27dfc2610f7d602086018661156b565b610f8d604087016020880161156b565b60408051602081019490945273ffffffffffffffffffffffffffffffffffffffff92831690840152166060820152608001604051602081830303815290604052805190602001206110d6565b9050610fff610fee6060850160408601611a2e565b82906060860135608087013561111e565b73ffffffffffffffffffffffffffffffffffffffff16611022602085018561156b565b73ffffffffffffffffffffffffffffffffffffffff16149392505050565b600080829050601f8151111561108457826040517f305a27a900000000000000000000000000000000000000000000000000000000815260040161037e9190611a51565b805161108f82611a64565b179392505050565b606060006110a483611146565b604080516020808252818301909252919250600091906020820181803683375050509182525060208101929092525090565b6000610c3b6110e3610cbb565b836040517f19010000000000000000000000000000000000000000000000000000000000008152600281019290925260228201526042902090565b600080600061112f87878787611187565b9150915061113c81611273565b5095945050505050565b600060ff8216601f811115610c3b576040517fb3512b0c00000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b6000807f7fffffffffffffffffffffffffffffff5d576e7357a4501ddfe92f46681b20a08311156111be5750600090506003610a91565b6040805160008082526020820180845289905260ff881692820192909252606081018690526080810185905260019060a0016020604051602081039080840390855afa158015611212573d6000803e3d6000fd5b50506040517fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0015191505073ffffffffffffffffffffffffffffffffffffffff811661126657600060019250925050610a91565b9660009650945050505050565b600081600481111561128757611287611aa6565b0361128f5750565b60018160048111156112a3576112a3611aa6565b0361130a576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601860248201527f45434453413a20696e76616c6964207369676e61747572650000000000000000604482015260640161037e565b600281600481111561131e5761131e611aa6565b03611385576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601f60248201527f45434453413a20696e76616c6964207369676e6174757265206c656e67746800604482015260640161037e565b600381600481111561139957611399611aa6565b03610390576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602260248201527f45434453413a20696e76616c6964207369676e6174757265202773272076616c60448201527f7565000000000000000000000000000000000000000000000000000000000000606482015260840161037e565b6000815180845260005b8181101561144c57602081850181015186830182015201611430565b5060006020828601015260207fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0601f83011685010191505092915050565b7fff00000000000000000000000000000000000000000000000000000000000000881681526000602060e0818401526114c660e084018a611426565b83810360408501526114d8818a611426565b6060850189905273ffffffffffffffffffffffffffffffffffffffff8816608086015260a0850187905284810360c0860152855180825283870192509083019060005b818110156115375783518352928401929184019160010161151b565b50909c9b505050505050505050505050565b73ffffffffffffffffffffffffffffffffffffffff8116811461039057600080fd5b60006020828403121561157d57600080fd5b813561158881611549565b9392505050565b6000606082840312156115a157600080fd5b50919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b6000806000808486036101008112156115ee57600080fd5b853567ffffffffffffffff8082111561160657600080fd5b61161289838a0161158f565b9650602088013591508082111561162857600080fd5b61163489838a0161158f565b955060a07fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc08401121561166657600080fd5b60408801945060e088013592508083111561168057600080fd5b828801925088601f84011261169457600080fd5b82359150808211156116a8576116a86115a7565b604051601f83017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0908116603f011681019082821181831017156116ee576116ee6115a7565b816040528381528a602085870101111561170757600080fd5b83602086016020830137600060208583010152809550505050505092959194509250565b60006020828403121561173d57600080fd5b815161158881611549565b60008083357fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe184360301811261177d57600080fd5b83018035915067ffffffffffffffff82111561179857600080fd5b6020019150368190038213156117ad57600080fd5b9250929050565b8183528181602085013750600060208284010152600060207fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0601f840116840101905092915050565b60a08152600061181160a083018a8c6117b4565b828103602084015261182481898b6117b4565b905082810360408401526118398187896117b4565b905073ffffffffffffffffffffffffffffffffffffffff8516606084015282810360808401526118698185611426565b9b9a5050505050505050505050565b600181811c9082168061188c57607f821691505b6020821081036115a1577f4e487b7100000000000000000000000000000000000000000000000000000000600052602260045260246000fd5b601f82111561190f57600081815260208120601f850160051c810160208610156118ec5750805b601f850160051c820191505b8181101561190b578281556001016118f8565b5050505b505050565b815167ffffffffffffffff81111561192e5761192e6115a7565b6119428161193c8454611878565b846118c5565b602080601f831160018114611995576000841561195f5750858301515b7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff600386901b1c1916600185901b17855561190b565b6000858152602081207fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe08616915b828110156119e2578886015182559484019460019091019084016119c3565b5085821015611a1e57878501517fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff600388901b60f8161c191681555b5050505050600190811b01905550565b600060208284031215611a4057600080fd5b813560ff8116811461158857600080fd5b6020815260006115886020830184611426565b805160208083015191908110156115a1577fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff60209190910360031b1b16919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052602160045260246000fdfea164736f6c6343000811000a",
}

// CommunityTokenDeployerABI is the input ABI used to generate the binding from.
// Deprecated: Use CommunityTokenDeployerMetaData.ABI instead.
var CommunityTokenDeployerABI = CommunityTokenDeployerMetaData.ABI

// CommunityTokenDeployerBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use CommunityTokenDeployerMetaData.Bin instead.
var CommunityTokenDeployerBin = CommunityTokenDeployerMetaData.Bin

// DeployCommunityTokenDeployer deploys a new Ethereum contract, binding an instance of CommunityTokenDeployer to it.
func DeployCommunityTokenDeployer(auth *bind.TransactOpts, backend bind.ContractBackend, _registry common.Address, _ownerTokenFactory common.Address, _masterTokenFactory common.Address) (common.Address, *types.Transaction, *CommunityTokenDeployer, error) {
	parsed, err := CommunityTokenDeployerMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(CommunityTokenDeployerBin), backend, _registry, _ownerTokenFactory, _masterTokenFactory)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &CommunityTokenDeployer{CommunityTokenDeployerCaller: CommunityTokenDeployerCaller{contract: contract}, CommunityTokenDeployerTransactor: CommunityTokenDeployerTransactor{contract: contract}, CommunityTokenDeployerFilterer: CommunityTokenDeployerFilterer{contract: contract}}, nil
}

// CommunityTokenDeployer is an auto generated Go binding around an Ethereum contract.
type CommunityTokenDeployer struct {
	CommunityTokenDeployerCaller     // Read-only binding to the contract
	CommunityTokenDeployerTransactor // Write-only binding to the contract
	CommunityTokenDeployerFilterer   // Log filterer for contract events
}

// CommunityTokenDeployerCaller is an auto generated read-only Go binding around an Ethereum contract.
type CommunityTokenDeployerCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CommunityTokenDeployerTransactor is an auto generated write-only Go binding around an Ethereum contract.
type CommunityTokenDeployerTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CommunityTokenDeployerFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type CommunityTokenDeployerFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CommunityTokenDeployerSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type CommunityTokenDeployerSession struct {
	Contract     *CommunityTokenDeployer // Generic contract binding to set the session for
	CallOpts     bind.CallOpts           // Call options to use throughout this session
	TransactOpts bind.TransactOpts       // Transaction auth options to use throughout this session
}

// CommunityTokenDeployerCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type CommunityTokenDeployerCallerSession struct {
	Contract *CommunityTokenDeployerCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                 // Call options to use throughout this session
}

// CommunityTokenDeployerTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type CommunityTokenDeployerTransactorSession struct {
	Contract     *CommunityTokenDeployerTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                 // Transaction auth options to use throughout this session
}

// CommunityTokenDeployerRaw is an auto generated low-level Go binding around an Ethereum contract.
type CommunityTokenDeployerRaw struct {
	Contract *CommunityTokenDeployer // Generic contract binding to access the raw methods on
}

// CommunityTokenDeployerCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type CommunityTokenDeployerCallerRaw struct {
	Contract *CommunityTokenDeployerCaller // Generic read-only contract binding to access the raw methods on
}

// CommunityTokenDeployerTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type CommunityTokenDeployerTransactorRaw struct {
	Contract *CommunityTokenDeployerTransactor // Generic write-only contract binding to access the raw methods on
}

// NewCommunityTokenDeployer creates a new instance of CommunityTokenDeployer, bound to a specific deployed contract.
func NewCommunityTokenDeployer(address common.Address, backend bind.ContractBackend) (*CommunityTokenDeployer, error) {
	contract, err := bindCommunityTokenDeployer(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &CommunityTokenDeployer{CommunityTokenDeployerCaller: CommunityTokenDeployerCaller{contract: contract}, CommunityTokenDeployerTransactor: CommunityTokenDeployerTransactor{contract: contract}, CommunityTokenDeployerFilterer: CommunityTokenDeployerFilterer{contract: contract}}, nil
}

// NewCommunityTokenDeployerCaller creates a new read-only instance of CommunityTokenDeployer, bound to a specific deployed contract.
func NewCommunityTokenDeployerCaller(address common.Address, caller bind.ContractCaller) (*CommunityTokenDeployerCaller, error) {
	contract, err := bindCommunityTokenDeployer(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &CommunityTokenDeployerCaller{contract: contract}, nil
}

// NewCommunityTokenDeployerTransactor creates a new write-only instance of CommunityTokenDeployer, bound to a specific deployed contract.
func NewCommunityTokenDeployerTransactor(address common.Address, transactor bind.ContractTransactor) (*CommunityTokenDeployerTransactor, error) {
	contract, err := bindCommunityTokenDeployer(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &CommunityTokenDeployerTransactor{contract: contract}, nil
}

// NewCommunityTokenDeployerFilterer creates a new log filterer instance of CommunityTokenDeployer, bound to a specific deployed contract.
func NewCommunityTokenDeployerFilterer(address common.Address, filterer bind.ContractFilterer) (*CommunityTokenDeployerFilterer, error) {
	contract, err := bindCommunityTokenDeployer(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &CommunityTokenDeployerFilterer{contract: contract}, nil
}

// bindCommunityTokenDeployer binds a generic wrapper to an already deployed contract.
func bindCommunityTokenDeployer(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := CommunityTokenDeployerMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_CommunityTokenDeployer *CommunityTokenDeployerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _CommunityTokenDeployer.Contract.CommunityTokenDeployerCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_CommunityTokenDeployer *CommunityTokenDeployerRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _CommunityTokenDeployer.Contract.CommunityTokenDeployerTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_CommunityTokenDeployer *CommunityTokenDeployerRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _CommunityTokenDeployer.Contract.CommunityTokenDeployerTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_CommunityTokenDeployer *CommunityTokenDeployerCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _CommunityTokenDeployer.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_CommunityTokenDeployer *CommunityTokenDeployerTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _CommunityTokenDeployer.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_CommunityTokenDeployer *CommunityTokenDeployerTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _CommunityTokenDeployer.Contract.contract.Transact(opts, method, params...)
}

// DEPLOYMENTSIGNATURETYPEHASH is a free data retrieval call binding the contract method 0xa825483c.
//
// Solidity: function DEPLOYMENT_SIGNATURE_TYPEHASH() view returns(bytes32)
func (_CommunityTokenDeployer *CommunityTokenDeployerCaller) DEPLOYMENTSIGNATURETYPEHASH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _CommunityTokenDeployer.contract.Call(opts, &out, "DEPLOYMENT_SIGNATURE_TYPEHASH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DEPLOYMENTSIGNATURETYPEHASH is a free data retrieval call binding the contract method 0xa825483c.
//
// Solidity: function DEPLOYMENT_SIGNATURE_TYPEHASH() view returns(bytes32)
func (_CommunityTokenDeployer *CommunityTokenDeployerSession) DEPLOYMENTSIGNATURETYPEHASH() ([32]byte, error) {
	return _CommunityTokenDeployer.Contract.DEPLOYMENTSIGNATURETYPEHASH(&_CommunityTokenDeployer.CallOpts)
}

// DEPLOYMENTSIGNATURETYPEHASH is a free data retrieval call binding the contract method 0xa825483c.
//
// Solidity: function DEPLOYMENT_SIGNATURE_TYPEHASH() view returns(bytes32)
func (_CommunityTokenDeployer *CommunityTokenDeployerCallerSession) DEPLOYMENTSIGNATURETYPEHASH() ([32]byte, error) {
	return _CommunityTokenDeployer.Contract.DEPLOYMENTSIGNATURETYPEHASH(&_CommunityTokenDeployer.CallOpts)
}

// DOMAINSEPARATOR is a free data retrieval call binding the contract method 0x3644e515.
//
// Solidity: function DOMAIN_SEPARATOR() view returns(bytes32)
func (_CommunityTokenDeployer *CommunityTokenDeployerCaller) DOMAINSEPARATOR(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _CommunityTokenDeployer.contract.Call(opts, &out, "DOMAIN_SEPARATOR")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DOMAINSEPARATOR is a free data retrieval call binding the contract method 0x3644e515.
//
// Solidity: function DOMAIN_SEPARATOR() view returns(bytes32)
func (_CommunityTokenDeployer *CommunityTokenDeployerSession) DOMAINSEPARATOR() ([32]byte, error) {
	return _CommunityTokenDeployer.Contract.DOMAINSEPARATOR(&_CommunityTokenDeployer.CallOpts)
}

// DOMAINSEPARATOR is a free data retrieval call binding the contract method 0x3644e515.
//
// Solidity: function DOMAIN_SEPARATOR() view returns(bytes32)
func (_CommunityTokenDeployer *CommunityTokenDeployerCallerSession) DOMAINSEPARATOR() ([32]byte, error) {
	return _CommunityTokenDeployer.Contract.DOMAINSEPARATOR(&_CommunityTokenDeployer.CallOpts)
}

// DeploymentRegistry is a free data retrieval call binding the contract method 0xc6631092.
//
// Solidity: function deploymentRegistry() view returns(address)
func (_CommunityTokenDeployer *CommunityTokenDeployerCaller) DeploymentRegistry(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _CommunityTokenDeployer.contract.Call(opts, &out, "deploymentRegistry")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// DeploymentRegistry is a free data retrieval call binding the contract method 0xc6631092.
//
// Solidity: function deploymentRegistry() view returns(address)
func (_CommunityTokenDeployer *CommunityTokenDeployerSession) DeploymentRegistry() (common.Address, error) {
	return _CommunityTokenDeployer.Contract.DeploymentRegistry(&_CommunityTokenDeployer.CallOpts)
}

// DeploymentRegistry is a free data retrieval call binding the contract method 0xc6631092.
//
// Solidity: function deploymentRegistry() view returns(address)
func (_CommunityTokenDeployer *CommunityTokenDeployerCallerSession) DeploymentRegistry() (common.Address, error) {
	return _CommunityTokenDeployer.Contract.DeploymentRegistry(&_CommunityTokenDeployer.CallOpts)
}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_CommunityTokenDeployer *CommunityTokenDeployerCaller) Eip712Domain(opts *bind.CallOpts) (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	var out []interface{}
	err := _CommunityTokenDeployer.contract.Call(opts, &out, "eip712Domain")

	outstruct := new(struct {
		Fields            [1]byte
		Name              string
		Version           string
		ChainId           *big.Int
		VerifyingContract common.Address
		Salt              [32]byte
		Extensions        []*big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Fields = *abi.ConvertType(out[0], new([1]byte)).(*[1]byte)
	outstruct.Name = *abi.ConvertType(out[1], new(string)).(*string)
	outstruct.Version = *abi.ConvertType(out[2], new(string)).(*string)
	outstruct.ChainId = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.VerifyingContract = *abi.ConvertType(out[4], new(common.Address)).(*common.Address)
	outstruct.Salt = *abi.ConvertType(out[5], new([32]byte)).(*[32]byte)
	outstruct.Extensions = *abi.ConvertType(out[6], new([]*big.Int)).(*[]*big.Int)

	return *outstruct, err

}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_CommunityTokenDeployer *CommunityTokenDeployerSession) Eip712Domain() (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	return _CommunityTokenDeployer.Contract.Eip712Domain(&_CommunityTokenDeployer.CallOpts)
}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_CommunityTokenDeployer *CommunityTokenDeployerCallerSession) Eip712Domain() (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	return _CommunityTokenDeployer.Contract.Eip712Domain(&_CommunityTokenDeployer.CallOpts)
}

// MasterTokenFactory is a free data retrieval call binding the contract method 0x830c2626.
//
// Solidity: function masterTokenFactory() view returns(address)
func (_CommunityTokenDeployer *CommunityTokenDeployerCaller) MasterTokenFactory(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _CommunityTokenDeployer.contract.Call(opts, &out, "masterTokenFactory")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// MasterTokenFactory is a free data retrieval call binding the contract method 0x830c2626.
//
// Solidity: function masterTokenFactory() view returns(address)
func (_CommunityTokenDeployer *CommunityTokenDeployerSession) MasterTokenFactory() (common.Address, error) {
	return _CommunityTokenDeployer.Contract.MasterTokenFactory(&_CommunityTokenDeployer.CallOpts)
}

// MasterTokenFactory is a free data retrieval call binding the contract method 0x830c2626.
//
// Solidity: function masterTokenFactory() view returns(address)
func (_CommunityTokenDeployer *CommunityTokenDeployerCallerSession) MasterTokenFactory() (common.Address, error) {
	return _CommunityTokenDeployer.Contract.MasterTokenFactory(&_CommunityTokenDeployer.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_CommunityTokenDeployer *CommunityTokenDeployerCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _CommunityTokenDeployer.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_CommunityTokenDeployer *CommunityTokenDeployerSession) Owner() (common.Address, error) {
	return _CommunityTokenDeployer.Contract.Owner(&_CommunityTokenDeployer.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_CommunityTokenDeployer *CommunityTokenDeployerCallerSession) Owner() (common.Address, error) {
	return _CommunityTokenDeployer.Contract.Owner(&_CommunityTokenDeployer.CallOpts)
}

// OwnerTokenFactory is a free data retrieval call binding the contract method 0x62457f59.
//
// Solidity: function ownerTokenFactory() view returns(address)
func (_CommunityTokenDeployer *CommunityTokenDeployerCaller) OwnerTokenFactory(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _CommunityTokenDeployer.contract.Call(opts, &out, "ownerTokenFactory")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// OwnerTokenFactory is a free data retrieval call binding the contract method 0x62457f59.
//
// Solidity: function ownerTokenFactory() view returns(address)
func (_CommunityTokenDeployer *CommunityTokenDeployerSession) OwnerTokenFactory() (common.Address, error) {
	return _CommunityTokenDeployer.Contract.OwnerTokenFactory(&_CommunityTokenDeployer.CallOpts)
}

// OwnerTokenFactory is a free data retrieval call binding the contract method 0x62457f59.
//
// Solidity: function ownerTokenFactory() view returns(address)
func (_CommunityTokenDeployer *CommunityTokenDeployerCallerSession) OwnerTokenFactory() (common.Address, error) {
	return _CommunityTokenDeployer.Contract.OwnerTokenFactory(&_CommunityTokenDeployer.CallOpts)
}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_CommunityTokenDeployer *CommunityTokenDeployerCaller) PendingOwner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _CommunityTokenDeployer.contract.Call(opts, &out, "pendingOwner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_CommunityTokenDeployer *CommunityTokenDeployerSession) PendingOwner() (common.Address, error) {
	return _CommunityTokenDeployer.Contract.PendingOwner(&_CommunityTokenDeployer.CallOpts)
}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_CommunityTokenDeployer *CommunityTokenDeployerCallerSession) PendingOwner() (common.Address, error) {
	return _CommunityTokenDeployer.Contract.PendingOwner(&_CommunityTokenDeployer.CallOpts)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_CommunityTokenDeployer *CommunityTokenDeployerTransactor) AcceptOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _CommunityTokenDeployer.contract.Transact(opts, "acceptOwnership")
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_CommunityTokenDeployer *CommunityTokenDeployerSession) AcceptOwnership() (*types.Transaction, error) {
	return _CommunityTokenDeployer.Contract.AcceptOwnership(&_CommunityTokenDeployer.TransactOpts)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_CommunityTokenDeployer *CommunityTokenDeployerTransactorSession) AcceptOwnership() (*types.Transaction, error) {
	return _CommunityTokenDeployer.Contract.AcceptOwnership(&_CommunityTokenDeployer.TransactOpts)
}

// Deploy is a paid mutator transaction binding the contract method 0xb0f95f28.
//
// Solidity: function deploy((string,string,string) _ownerToken, (string,string,string) _masterToken, (address,address,uint8,bytes32,bytes32) _signature, bytes _signerPublicKey) returns(address, address)
func (_CommunityTokenDeployer *CommunityTokenDeployerTransactor) Deploy(opts *bind.TransactOpts, _ownerToken CommunityTokenDeployerTokenConfig, _masterToken CommunityTokenDeployerTokenConfig, _signature CommunityTokenDeployerDeploymentSignature, _signerPublicKey []byte) (*types.Transaction, error) {
	return _CommunityTokenDeployer.contract.Transact(opts, "deploy", _ownerToken, _masterToken, _signature, _signerPublicKey)
}

// Deploy is a paid mutator transaction binding the contract method 0xb0f95f28.
//
// Solidity: function deploy((string,string,string) _ownerToken, (string,string,string) _masterToken, (address,address,uint8,bytes32,bytes32) _signature, bytes _signerPublicKey) returns(address, address)
func (_CommunityTokenDeployer *CommunityTokenDeployerSession) Deploy(_ownerToken CommunityTokenDeployerTokenConfig, _masterToken CommunityTokenDeployerTokenConfig, _signature CommunityTokenDeployerDeploymentSignature, _signerPublicKey []byte) (*types.Transaction, error) {
	return _CommunityTokenDeployer.Contract.Deploy(&_CommunityTokenDeployer.TransactOpts, _ownerToken, _masterToken, _signature, _signerPublicKey)
}

// Deploy is a paid mutator transaction binding the contract method 0xb0f95f28.
//
// Solidity: function deploy((string,string,string) _ownerToken, (string,string,string) _masterToken, (address,address,uint8,bytes32,bytes32) _signature, bytes _signerPublicKey) returns(address, address)
func (_CommunityTokenDeployer *CommunityTokenDeployerTransactorSession) Deploy(_ownerToken CommunityTokenDeployerTokenConfig, _masterToken CommunityTokenDeployerTokenConfig, _signature CommunityTokenDeployerDeploymentSignature, _signerPublicKey []byte) (*types.Transaction, error) {
	return _CommunityTokenDeployer.Contract.Deploy(&_CommunityTokenDeployer.TransactOpts, _ownerToken, _masterToken, _signature, _signerPublicKey)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_CommunityTokenDeployer *CommunityTokenDeployerTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _CommunityTokenDeployer.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_CommunityTokenDeployer *CommunityTokenDeployerSession) RenounceOwnership() (*types.Transaction, error) {
	return _CommunityTokenDeployer.Contract.RenounceOwnership(&_CommunityTokenDeployer.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_CommunityTokenDeployer *CommunityTokenDeployerTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _CommunityTokenDeployer.Contract.RenounceOwnership(&_CommunityTokenDeployer.TransactOpts)
}

// SetDeploymentRegistryAddress is a paid mutator transaction binding the contract method 0xf8851475.
//
// Solidity: function setDeploymentRegistryAddress(address _deploymentRegistry) returns()
func (_CommunityTokenDeployer *CommunityTokenDeployerTransactor) SetDeploymentRegistryAddress(opts *bind.TransactOpts, _deploymentRegistry common.Address) (*types.Transaction, error) {
	return _CommunityTokenDeployer.contract.Transact(opts, "setDeploymentRegistryAddress", _deploymentRegistry)
}

// SetDeploymentRegistryAddress is a paid mutator transaction binding the contract method 0xf8851475.
//
// Solidity: function setDeploymentRegistryAddress(address _deploymentRegistry) returns()
func (_CommunityTokenDeployer *CommunityTokenDeployerSession) SetDeploymentRegistryAddress(_deploymentRegistry common.Address) (*types.Transaction, error) {
	return _CommunityTokenDeployer.Contract.SetDeploymentRegistryAddress(&_CommunityTokenDeployer.TransactOpts, _deploymentRegistry)
}

// SetDeploymentRegistryAddress is a paid mutator transaction binding the contract method 0xf8851475.
//
// Solidity: function setDeploymentRegistryAddress(address _deploymentRegistry) returns()
func (_CommunityTokenDeployer *CommunityTokenDeployerTransactorSession) SetDeploymentRegistryAddress(_deploymentRegistry common.Address) (*types.Transaction, error) {
	return _CommunityTokenDeployer.Contract.SetDeploymentRegistryAddress(&_CommunityTokenDeployer.TransactOpts, _deploymentRegistry)
}

// SetMasterTokenFactoryAddress is a paid mutator transaction binding the contract method 0xa53b2bdb.
//
// Solidity: function setMasterTokenFactoryAddress(address _masterTokenFactory) returns()
func (_CommunityTokenDeployer *CommunityTokenDeployerTransactor) SetMasterTokenFactoryAddress(opts *bind.TransactOpts, _masterTokenFactory common.Address) (*types.Transaction, error) {
	return _CommunityTokenDeployer.contract.Transact(opts, "setMasterTokenFactoryAddress", _masterTokenFactory)
}

// SetMasterTokenFactoryAddress is a paid mutator transaction binding the contract method 0xa53b2bdb.
//
// Solidity: function setMasterTokenFactoryAddress(address _masterTokenFactory) returns()
func (_CommunityTokenDeployer *CommunityTokenDeployerSession) SetMasterTokenFactoryAddress(_masterTokenFactory common.Address) (*types.Transaction, error) {
	return _CommunityTokenDeployer.Contract.SetMasterTokenFactoryAddress(&_CommunityTokenDeployer.TransactOpts, _masterTokenFactory)
}

// SetMasterTokenFactoryAddress is a paid mutator transaction binding the contract method 0xa53b2bdb.
//
// Solidity: function setMasterTokenFactoryAddress(address _masterTokenFactory) returns()
func (_CommunityTokenDeployer *CommunityTokenDeployerTransactorSession) SetMasterTokenFactoryAddress(_masterTokenFactory common.Address) (*types.Transaction, error) {
	return _CommunityTokenDeployer.Contract.SetMasterTokenFactoryAddress(&_CommunityTokenDeployer.TransactOpts, _masterTokenFactory)
}

// SetOwnerTokenFactoryAddress is a paid mutator transaction binding the contract method 0x9ff02d18.
//
// Solidity: function setOwnerTokenFactoryAddress(address _ownerTokenFactory) returns()
func (_CommunityTokenDeployer *CommunityTokenDeployerTransactor) SetOwnerTokenFactoryAddress(opts *bind.TransactOpts, _ownerTokenFactory common.Address) (*types.Transaction, error) {
	return _CommunityTokenDeployer.contract.Transact(opts, "setOwnerTokenFactoryAddress", _ownerTokenFactory)
}

// SetOwnerTokenFactoryAddress is a paid mutator transaction binding the contract method 0x9ff02d18.
//
// Solidity: function setOwnerTokenFactoryAddress(address _ownerTokenFactory) returns()
func (_CommunityTokenDeployer *CommunityTokenDeployerSession) SetOwnerTokenFactoryAddress(_ownerTokenFactory common.Address) (*types.Transaction, error) {
	return _CommunityTokenDeployer.Contract.SetOwnerTokenFactoryAddress(&_CommunityTokenDeployer.TransactOpts, _ownerTokenFactory)
}

// SetOwnerTokenFactoryAddress is a paid mutator transaction binding the contract method 0x9ff02d18.
//
// Solidity: function setOwnerTokenFactoryAddress(address _ownerTokenFactory) returns()
func (_CommunityTokenDeployer *CommunityTokenDeployerTransactorSession) SetOwnerTokenFactoryAddress(_ownerTokenFactory common.Address) (*types.Transaction, error) {
	return _CommunityTokenDeployer.Contract.SetOwnerTokenFactoryAddress(&_CommunityTokenDeployer.TransactOpts, _ownerTokenFactory)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_CommunityTokenDeployer *CommunityTokenDeployerTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _CommunityTokenDeployer.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_CommunityTokenDeployer *CommunityTokenDeployerSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _CommunityTokenDeployer.Contract.TransferOwnership(&_CommunityTokenDeployer.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_CommunityTokenDeployer *CommunityTokenDeployerTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _CommunityTokenDeployer.Contract.TransferOwnership(&_CommunityTokenDeployer.TransactOpts, newOwner)
}

// CommunityTokenDeployerDeployMasterTokenIterator is returned from FilterDeployMasterToken and is used to iterate over the raw logs and unpacked data for DeployMasterToken events raised by the CommunityTokenDeployer contract.
type CommunityTokenDeployerDeployMasterTokenIterator struct {
	Event *CommunityTokenDeployerDeployMasterToken // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *CommunityTokenDeployerDeployMasterTokenIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CommunityTokenDeployerDeployMasterToken)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(CommunityTokenDeployerDeployMasterToken)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *CommunityTokenDeployerDeployMasterTokenIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CommunityTokenDeployerDeployMasterTokenIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CommunityTokenDeployerDeployMasterToken represents a DeployMasterToken event raised by the CommunityTokenDeployer contract.
type CommunityTokenDeployerDeployMasterToken struct {
	Arg0 common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterDeployMasterToken is a free log retrieval operation binding the contract event 0x1464afef6e77413c9c3201405b55530340d684e2a19f3a9d83bc604d4aa3a255.
//
// Solidity: event DeployMasterToken(address indexed arg0)
func (_CommunityTokenDeployer *CommunityTokenDeployerFilterer) FilterDeployMasterToken(opts *bind.FilterOpts, arg0 []common.Address) (*CommunityTokenDeployerDeployMasterTokenIterator, error) {

	var arg0Rule []interface{}
	for _, arg0Item := range arg0 {
		arg0Rule = append(arg0Rule, arg0Item)
	}

	logs, sub, err := _CommunityTokenDeployer.contract.FilterLogs(opts, "DeployMasterToken", arg0Rule)
	if err != nil {
		return nil, err
	}
	return &CommunityTokenDeployerDeployMasterTokenIterator{contract: _CommunityTokenDeployer.contract, event: "DeployMasterToken", logs: logs, sub: sub}, nil
}

// WatchDeployMasterToken is a free log subscription operation binding the contract event 0x1464afef6e77413c9c3201405b55530340d684e2a19f3a9d83bc604d4aa3a255.
//
// Solidity: event DeployMasterToken(address indexed arg0)
func (_CommunityTokenDeployer *CommunityTokenDeployerFilterer) WatchDeployMasterToken(opts *bind.WatchOpts, sink chan<- *CommunityTokenDeployerDeployMasterToken, arg0 []common.Address) (event.Subscription, error) {

	var arg0Rule []interface{}
	for _, arg0Item := range arg0 {
		arg0Rule = append(arg0Rule, arg0Item)
	}

	logs, sub, err := _CommunityTokenDeployer.contract.WatchLogs(opts, "DeployMasterToken", arg0Rule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CommunityTokenDeployerDeployMasterToken)
				if err := _CommunityTokenDeployer.contract.UnpackLog(event, "DeployMasterToken", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDeployMasterToken is a log parse operation binding the contract event 0x1464afef6e77413c9c3201405b55530340d684e2a19f3a9d83bc604d4aa3a255.
//
// Solidity: event DeployMasterToken(address indexed arg0)
func (_CommunityTokenDeployer *CommunityTokenDeployerFilterer) ParseDeployMasterToken(log types.Log) (*CommunityTokenDeployerDeployMasterToken, error) {
	event := new(CommunityTokenDeployerDeployMasterToken)
	if err := _CommunityTokenDeployer.contract.UnpackLog(event, "DeployMasterToken", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CommunityTokenDeployerDeployOwnerTokenIterator is returned from FilterDeployOwnerToken and is used to iterate over the raw logs and unpacked data for DeployOwnerToken events raised by the CommunityTokenDeployer contract.
type CommunityTokenDeployerDeployOwnerTokenIterator struct {
	Event *CommunityTokenDeployerDeployOwnerToken // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *CommunityTokenDeployerDeployOwnerTokenIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CommunityTokenDeployerDeployOwnerToken)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(CommunityTokenDeployerDeployOwnerToken)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *CommunityTokenDeployerDeployOwnerTokenIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CommunityTokenDeployerDeployOwnerTokenIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CommunityTokenDeployerDeployOwnerToken represents a DeployOwnerToken event raised by the CommunityTokenDeployer contract.
type CommunityTokenDeployerDeployOwnerToken struct {
	Arg0 common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterDeployOwnerToken is a free log retrieval operation binding the contract event 0x6f60871ce1ae7c2bc82a9fca785cdf029fa7c9984afe96eaa106d1b7b19c6322.
//
// Solidity: event DeployOwnerToken(address indexed arg0)
func (_CommunityTokenDeployer *CommunityTokenDeployerFilterer) FilterDeployOwnerToken(opts *bind.FilterOpts, arg0 []common.Address) (*CommunityTokenDeployerDeployOwnerTokenIterator, error) {

	var arg0Rule []interface{}
	for _, arg0Item := range arg0 {
		arg0Rule = append(arg0Rule, arg0Item)
	}

	logs, sub, err := _CommunityTokenDeployer.contract.FilterLogs(opts, "DeployOwnerToken", arg0Rule)
	if err != nil {
		return nil, err
	}
	return &CommunityTokenDeployerDeployOwnerTokenIterator{contract: _CommunityTokenDeployer.contract, event: "DeployOwnerToken", logs: logs, sub: sub}, nil
}

// WatchDeployOwnerToken is a free log subscription operation binding the contract event 0x6f60871ce1ae7c2bc82a9fca785cdf029fa7c9984afe96eaa106d1b7b19c6322.
//
// Solidity: event DeployOwnerToken(address indexed arg0)
func (_CommunityTokenDeployer *CommunityTokenDeployerFilterer) WatchDeployOwnerToken(opts *bind.WatchOpts, sink chan<- *CommunityTokenDeployerDeployOwnerToken, arg0 []common.Address) (event.Subscription, error) {

	var arg0Rule []interface{}
	for _, arg0Item := range arg0 {
		arg0Rule = append(arg0Rule, arg0Item)
	}

	logs, sub, err := _CommunityTokenDeployer.contract.WatchLogs(opts, "DeployOwnerToken", arg0Rule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CommunityTokenDeployerDeployOwnerToken)
				if err := _CommunityTokenDeployer.contract.UnpackLog(event, "DeployOwnerToken", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDeployOwnerToken is a log parse operation binding the contract event 0x6f60871ce1ae7c2bc82a9fca785cdf029fa7c9984afe96eaa106d1b7b19c6322.
//
// Solidity: event DeployOwnerToken(address indexed arg0)
func (_CommunityTokenDeployer *CommunityTokenDeployerFilterer) ParseDeployOwnerToken(log types.Log) (*CommunityTokenDeployerDeployOwnerToken, error) {
	event := new(CommunityTokenDeployerDeployOwnerToken)
	if err := _CommunityTokenDeployer.contract.UnpackLog(event, "DeployOwnerToken", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CommunityTokenDeployerDeploymentRegistryAddressChangeIterator is returned from FilterDeploymentRegistryAddressChange and is used to iterate over the raw logs and unpacked data for DeploymentRegistryAddressChange events raised by the CommunityTokenDeployer contract.
type CommunityTokenDeployerDeploymentRegistryAddressChangeIterator struct {
	Event *CommunityTokenDeployerDeploymentRegistryAddressChange // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *CommunityTokenDeployerDeploymentRegistryAddressChangeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CommunityTokenDeployerDeploymentRegistryAddressChange)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(CommunityTokenDeployerDeploymentRegistryAddressChange)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *CommunityTokenDeployerDeploymentRegistryAddressChangeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CommunityTokenDeployerDeploymentRegistryAddressChangeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CommunityTokenDeployerDeploymentRegistryAddressChange represents a DeploymentRegistryAddressChange event raised by the CommunityTokenDeployer contract.
type CommunityTokenDeployerDeploymentRegistryAddressChange struct {
	Arg0 common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterDeploymentRegistryAddressChange is a free log retrieval operation binding the contract event 0x8f3be421db34ad7dfa5c7fb9391b363b444007c7b26f0a22c58aad6e130b935e.
//
// Solidity: event DeploymentRegistryAddressChange(address indexed arg0)
func (_CommunityTokenDeployer *CommunityTokenDeployerFilterer) FilterDeploymentRegistryAddressChange(opts *bind.FilterOpts, arg0 []common.Address) (*CommunityTokenDeployerDeploymentRegistryAddressChangeIterator, error) {

	var arg0Rule []interface{}
	for _, arg0Item := range arg0 {
		arg0Rule = append(arg0Rule, arg0Item)
	}

	logs, sub, err := _CommunityTokenDeployer.contract.FilterLogs(opts, "DeploymentRegistryAddressChange", arg0Rule)
	if err != nil {
		return nil, err
	}
	return &CommunityTokenDeployerDeploymentRegistryAddressChangeIterator{contract: _CommunityTokenDeployer.contract, event: "DeploymentRegistryAddressChange", logs: logs, sub: sub}, nil
}

// WatchDeploymentRegistryAddressChange is a free log subscription operation binding the contract event 0x8f3be421db34ad7dfa5c7fb9391b363b444007c7b26f0a22c58aad6e130b935e.
//
// Solidity: event DeploymentRegistryAddressChange(address indexed arg0)
func (_CommunityTokenDeployer *CommunityTokenDeployerFilterer) WatchDeploymentRegistryAddressChange(opts *bind.WatchOpts, sink chan<- *CommunityTokenDeployerDeploymentRegistryAddressChange, arg0 []common.Address) (event.Subscription, error) {

	var arg0Rule []interface{}
	for _, arg0Item := range arg0 {
		arg0Rule = append(arg0Rule, arg0Item)
	}

	logs, sub, err := _CommunityTokenDeployer.contract.WatchLogs(opts, "DeploymentRegistryAddressChange", arg0Rule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CommunityTokenDeployerDeploymentRegistryAddressChange)
				if err := _CommunityTokenDeployer.contract.UnpackLog(event, "DeploymentRegistryAddressChange", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDeploymentRegistryAddressChange is a log parse operation binding the contract event 0x8f3be421db34ad7dfa5c7fb9391b363b444007c7b26f0a22c58aad6e130b935e.
//
// Solidity: event DeploymentRegistryAddressChange(address indexed arg0)
func (_CommunityTokenDeployer *CommunityTokenDeployerFilterer) ParseDeploymentRegistryAddressChange(log types.Log) (*CommunityTokenDeployerDeploymentRegistryAddressChange, error) {
	event := new(CommunityTokenDeployerDeploymentRegistryAddressChange)
	if err := _CommunityTokenDeployer.contract.UnpackLog(event, "DeploymentRegistryAddressChange", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CommunityTokenDeployerEIP712DomainChangedIterator is returned from FilterEIP712DomainChanged and is used to iterate over the raw logs and unpacked data for EIP712DomainChanged events raised by the CommunityTokenDeployer contract.
type CommunityTokenDeployerEIP712DomainChangedIterator struct {
	Event *CommunityTokenDeployerEIP712DomainChanged // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *CommunityTokenDeployerEIP712DomainChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CommunityTokenDeployerEIP712DomainChanged)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(CommunityTokenDeployerEIP712DomainChanged)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *CommunityTokenDeployerEIP712DomainChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CommunityTokenDeployerEIP712DomainChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CommunityTokenDeployerEIP712DomainChanged represents a EIP712DomainChanged event raised by the CommunityTokenDeployer contract.
type CommunityTokenDeployerEIP712DomainChanged struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterEIP712DomainChanged is a free log retrieval operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_CommunityTokenDeployer *CommunityTokenDeployerFilterer) FilterEIP712DomainChanged(opts *bind.FilterOpts) (*CommunityTokenDeployerEIP712DomainChangedIterator, error) {

	logs, sub, err := _CommunityTokenDeployer.contract.FilterLogs(opts, "EIP712DomainChanged")
	if err != nil {
		return nil, err
	}
	return &CommunityTokenDeployerEIP712DomainChangedIterator{contract: _CommunityTokenDeployer.contract, event: "EIP712DomainChanged", logs: logs, sub: sub}, nil
}

// WatchEIP712DomainChanged is a free log subscription operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_CommunityTokenDeployer *CommunityTokenDeployerFilterer) WatchEIP712DomainChanged(opts *bind.WatchOpts, sink chan<- *CommunityTokenDeployerEIP712DomainChanged) (event.Subscription, error) {

	logs, sub, err := _CommunityTokenDeployer.contract.WatchLogs(opts, "EIP712DomainChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CommunityTokenDeployerEIP712DomainChanged)
				if err := _CommunityTokenDeployer.contract.UnpackLog(event, "EIP712DomainChanged", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseEIP712DomainChanged is a log parse operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_CommunityTokenDeployer *CommunityTokenDeployerFilterer) ParseEIP712DomainChanged(log types.Log) (*CommunityTokenDeployerEIP712DomainChanged, error) {
	event := new(CommunityTokenDeployerEIP712DomainChanged)
	if err := _CommunityTokenDeployer.contract.UnpackLog(event, "EIP712DomainChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CommunityTokenDeployerMasterTokenFactoryAddressChangeIterator is returned from FilterMasterTokenFactoryAddressChange and is used to iterate over the raw logs and unpacked data for MasterTokenFactoryAddressChange events raised by the CommunityTokenDeployer contract.
type CommunityTokenDeployerMasterTokenFactoryAddressChangeIterator struct {
	Event *CommunityTokenDeployerMasterTokenFactoryAddressChange // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *CommunityTokenDeployerMasterTokenFactoryAddressChangeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CommunityTokenDeployerMasterTokenFactoryAddressChange)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(CommunityTokenDeployerMasterTokenFactoryAddressChange)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *CommunityTokenDeployerMasterTokenFactoryAddressChangeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CommunityTokenDeployerMasterTokenFactoryAddressChangeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CommunityTokenDeployerMasterTokenFactoryAddressChange represents a MasterTokenFactoryAddressChange event raised by the CommunityTokenDeployer contract.
type CommunityTokenDeployerMasterTokenFactoryAddressChange struct {
	Arg0 common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterMasterTokenFactoryAddressChange is a free log retrieval operation binding the contract event 0x184513c31b135bda32c81b8586d52ad5bdbc7b7e4ec5847eee48374ee07e8e48.
//
// Solidity: event MasterTokenFactoryAddressChange(address indexed arg0)
func (_CommunityTokenDeployer *CommunityTokenDeployerFilterer) FilterMasterTokenFactoryAddressChange(opts *bind.FilterOpts, arg0 []common.Address) (*CommunityTokenDeployerMasterTokenFactoryAddressChangeIterator, error) {

	var arg0Rule []interface{}
	for _, arg0Item := range arg0 {
		arg0Rule = append(arg0Rule, arg0Item)
	}

	logs, sub, err := _CommunityTokenDeployer.contract.FilterLogs(opts, "MasterTokenFactoryAddressChange", arg0Rule)
	if err != nil {
		return nil, err
	}
	return &CommunityTokenDeployerMasterTokenFactoryAddressChangeIterator{contract: _CommunityTokenDeployer.contract, event: "MasterTokenFactoryAddressChange", logs: logs, sub: sub}, nil
}

// WatchMasterTokenFactoryAddressChange is a free log subscription operation binding the contract event 0x184513c31b135bda32c81b8586d52ad5bdbc7b7e4ec5847eee48374ee07e8e48.
//
// Solidity: event MasterTokenFactoryAddressChange(address indexed arg0)
func (_CommunityTokenDeployer *CommunityTokenDeployerFilterer) WatchMasterTokenFactoryAddressChange(opts *bind.WatchOpts, sink chan<- *CommunityTokenDeployerMasterTokenFactoryAddressChange, arg0 []common.Address) (event.Subscription, error) {

	var arg0Rule []interface{}
	for _, arg0Item := range arg0 {
		arg0Rule = append(arg0Rule, arg0Item)
	}

	logs, sub, err := _CommunityTokenDeployer.contract.WatchLogs(opts, "MasterTokenFactoryAddressChange", arg0Rule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CommunityTokenDeployerMasterTokenFactoryAddressChange)
				if err := _CommunityTokenDeployer.contract.UnpackLog(event, "MasterTokenFactoryAddressChange", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseMasterTokenFactoryAddressChange is a log parse operation binding the contract event 0x184513c31b135bda32c81b8586d52ad5bdbc7b7e4ec5847eee48374ee07e8e48.
//
// Solidity: event MasterTokenFactoryAddressChange(address indexed arg0)
func (_CommunityTokenDeployer *CommunityTokenDeployerFilterer) ParseMasterTokenFactoryAddressChange(log types.Log) (*CommunityTokenDeployerMasterTokenFactoryAddressChange, error) {
	event := new(CommunityTokenDeployerMasterTokenFactoryAddressChange)
	if err := _CommunityTokenDeployer.contract.UnpackLog(event, "MasterTokenFactoryAddressChange", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CommunityTokenDeployerOwnerTokenFactoryAddressChangeIterator is returned from FilterOwnerTokenFactoryAddressChange and is used to iterate over the raw logs and unpacked data for OwnerTokenFactoryAddressChange events raised by the CommunityTokenDeployer contract.
type CommunityTokenDeployerOwnerTokenFactoryAddressChangeIterator struct {
	Event *CommunityTokenDeployerOwnerTokenFactoryAddressChange // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *CommunityTokenDeployerOwnerTokenFactoryAddressChangeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CommunityTokenDeployerOwnerTokenFactoryAddressChange)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(CommunityTokenDeployerOwnerTokenFactoryAddressChange)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *CommunityTokenDeployerOwnerTokenFactoryAddressChangeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CommunityTokenDeployerOwnerTokenFactoryAddressChangeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CommunityTokenDeployerOwnerTokenFactoryAddressChange represents a OwnerTokenFactoryAddressChange event raised by the CommunityTokenDeployer contract.
type CommunityTokenDeployerOwnerTokenFactoryAddressChange struct {
	Arg0 common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterOwnerTokenFactoryAddressChange is a free log retrieval operation binding the contract event 0x17ae1257210039eb267be68929104e6c28fc9ebb9dc6aaa84be39b45eb6f3767.
//
// Solidity: event OwnerTokenFactoryAddressChange(address indexed arg0)
func (_CommunityTokenDeployer *CommunityTokenDeployerFilterer) FilterOwnerTokenFactoryAddressChange(opts *bind.FilterOpts, arg0 []common.Address) (*CommunityTokenDeployerOwnerTokenFactoryAddressChangeIterator, error) {

	var arg0Rule []interface{}
	for _, arg0Item := range arg0 {
		arg0Rule = append(arg0Rule, arg0Item)
	}

	logs, sub, err := _CommunityTokenDeployer.contract.FilterLogs(opts, "OwnerTokenFactoryAddressChange", arg0Rule)
	if err != nil {
		return nil, err
	}
	return &CommunityTokenDeployerOwnerTokenFactoryAddressChangeIterator{contract: _CommunityTokenDeployer.contract, event: "OwnerTokenFactoryAddressChange", logs: logs, sub: sub}, nil
}

// WatchOwnerTokenFactoryAddressChange is a free log subscription operation binding the contract event 0x17ae1257210039eb267be68929104e6c28fc9ebb9dc6aaa84be39b45eb6f3767.
//
// Solidity: event OwnerTokenFactoryAddressChange(address indexed arg0)
func (_CommunityTokenDeployer *CommunityTokenDeployerFilterer) WatchOwnerTokenFactoryAddressChange(opts *bind.WatchOpts, sink chan<- *CommunityTokenDeployerOwnerTokenFactoryAddressChange, arg0 []common.Address) (event.Subscription, error) {

	var arg0Rule []interface{}
	for _, arg0Item := range arg0 {
		arg0Rule = append(arg0Rule, arg0Item)
	}

	logs, sub, err := _CommunityTokenDeployer.contract.WatchLogs(opts, "OwnerTokenFactoryAddressChange", arg0Rule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CommunityTokenDeployerOwnerTokenFactoryAddressChange)
				if err := _CommunityTokenDeployer.contract.UnpackLog(event, "OwnerTokenFactoryAddressChange", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnerTokenFactoryAddressChange is a log parse operation binding the contract event 0x17ae1257210039eb267be68929104e6c28fc9ebb9dc6aaa84be39b45eb6f3767.
//
// Solidity: event OwnerTokenFactoryAddressChange(address indexed arg0)
func (_CommunityTokenDeployer *CommunityTokenDeployerFilterer) ParseOwnerTokenFactoryAddressChange(log types.Log) (*CommunityTokenDeployerOwnerTokenFactoryAddressChange, error) {
	event := new(CommunityTokenDeployerOwnerTokenFactoryAddressChange)
	if err := _CommunityTokenDeployer.contract.UnpackLog(event, "OwnerTokenFactoryAddressChange", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CommunityTokenDeployerOwnershipTransferStartedIterator is returned from FilterOwnershipTransferStarted and is used to iterate over the raw logs and unpacked data for OwnershipTransferStarted events raised by the CommunityTokenDeployer contract.
type CommunityTokenDeployerOwnershipTransferStartedIterator struct {
	Event *CommunityTokenDeployerOwnershipTransferStarted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *CommunityTokenDeployerOwnershipTransferStartedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CommunityTokenDeployerOwnershipTransferStarted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(CommunityTokenDeployerOwnershipTransferStarted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *CommunityTokenDeployerOwnershipTransferStartedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CommunityTokenDeployerOwnershipTransferStartedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CommunityTokenDeployerOwnershipTransferStarted represents a OwnershipTransferStarted event raised by the CommunityTokenDeployer contract.
type CommunityTokenDeployerOwnershipTransferStarted struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferStarted is a free log retrieval operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_CommunityTokenDeployer *CommunityTokenDeployerFilterer) FilterOwnershipTransferStarted(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*CommunityTokenDeployerOwnershipTransferStartedIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _CommunityTokenDeployer.contract.FilterLogs(opts, "OwnershipTransferStarted", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &CommunityTokenDeployerOwnershipTransferStartedIterator{contract: _CommunityTokenDeployer.contract, event: "OwnershipTransferStarted", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferStarted is a free log subscription operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_CommunityTokenDeployer *CommunityTokenDeployerFilterer) WatchOwnershipTransferStarted(opts *bind.WatchOpts, sink chan<- *CommunityTokenDeployerOwnershipTransferStarted, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _CommunityTokenDeployer.contract.WatchLogs(opts, "OwnershipTransferStarted", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CommunityTokenDeployerOwnershipTransferStarted)
				if err := _CommunityTokenDeployer.contract.UnpackLog(event, "OwnershipTransferStarted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferStarted is a log parse operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_CommunityTokenDeployer *CommunityTokenDeployerFilterer) ParseOwnershipTransferStarted(log types.Log) (*CommunityTokenDeployerOwnershipTransferStarted, error) {
	event := new(CommunityTokenDeployerOwnershipTransferStarted)
	if err := _CommunityTokenDeployer.contract.UnpackLog(event, "OwnershipTransferStarted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CommunityTokenDeployerOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the CommunityTokenDeployer contract.
type CommunityTokenDeployerOwnershipTransferredIterator struct {
	Event *CommunityTokenDeployerOwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *CommunityTokenDeployerOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CommunityTokenDeployerOwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(CommunityTokenDeployerOwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *CommunityTokenDeployerOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CommunityTokenDeployerOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CommunityTokenDeployerOwnershipTransferred represents a OwnershipTransferred event raised by the CommunityTokenDeployer contract.
type CommunityTokenDeployerOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_CommunityTokenDeployer *CommunityTokenDeployerFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*CommunityTokenDeployerOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _CommunityTokenDeployer.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &CommunityTokenDeployerOwnershipTransferredIterator{contract: _CommunityTokenDeployer.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_CommunityTokenDeployer *CommunityTokenDeployerFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *CommunityTokenDeployerOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _CommunityTokenDeployer.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CommunityTokenDeployerOwnershipTransferred)
				if err := _CommunityTokenDeployer.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_CommunityTokenDeployer *CommunityTokenDeployerFilterer) ParseOwnershipTransferred(log types.Log) (*CommunityTokenDeployerOwnershipTransferred, error) {
	event := new(CommunityTokenDeployerOwnershipTransferred)
	if err := _CommunityTokenDeployer.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
