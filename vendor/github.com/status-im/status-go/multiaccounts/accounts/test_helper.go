package accounts

import (
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/multiaccounts/common"
)

func GetWatchOnlyAccountsForTest() []*Account {
	wo1 := &Account{
		Address: types.Address{0x11},
		Type:    AccountTypeWatch,
		Name:    "WatchOnlyAcc1",
		ColorID: common.CustomizationColorPrimary,
		Emoji:   "emoji-1",
	}
	wo2 := &Account{
		Address: types.Address{0x12},
		Type:    AccountTypeWatch,
		Name:    "WatchOnlyAcc2",
		ColorID: common.CustomizationColorPrimary,
		Emoji:   "emoji-1",
	}
	wo3 := &Account{
		Address: types.Address{0x13},
		Type:    AccountTypeWatch,
		Name:    "WatchOnlyAcc3",
		ColorID: common.CustomizationColorPrimary,
		Emoji:   "emoji-1",
	}

	return []*Account{wo1, wo2, wo3}
}

func GetProfileKeypairForTest(includeChatAccount bool, includeDefaultWalletAccount bool, includeAdditionalAccounts bool) *Keypair {
	kp := &Keypair{
		KeyUID:      "0000000000000000000000000000000000000000000000000000000000000001",
		Name:        "Profile Name",
		Type:        KeypairTypeProfile,
		DerivedFrom: "0x0001",
	}

	if includeChatAccount {
		profileAccount := &Account{
			Address:               types.Address{0x01},
			KeyUID:                kp.KeyUID,
			Wallet:                false,
			Chat:                  true,
			Type:                  AccountTypeGenerated,
			Path:                  "m/43'/60'/1581'/0'/0",
			PublicKey:             types.Hex2Bytes("0x000000001"),
			Name:                  "Profile Name",
			Operable:              AccountFullyOperable,
			ProdPreferredChainIDs: "1",
			TestPreferredChainIDs: "5",
		}
		kp.Accounts = append(kp.Accounts, profileAccount)
	}

	if includeDefaultWalletAccount {
		defaultWalletAccount := &Account{
			Address:               types.Address{0x02},
			KeyUID:                kp.KeyUID,
			Wallet:                true,
			Chat:                  false,
			Type:                  AccountTypeGenerated,
			Path:                  "m/44'/60'/0'/0/0",
			PublicKey:             types.Hex2Bytes("0x000000002"),
			Name:                  "Generated Acc 1",
			Emoji:                 "emoji-1",
			ColorID:               common.CustomizationColorPrimary,
			Hidden:                false,
			Clock:                 0,
			Removed:               false,
			Operable:              AccountFullyOperable,
			ProdPreferredChainIDs: "1",
			TestPreferredChainIDs: "5",
		}
		kp.Accounts = append(kp.Accounts, defaultWalletAccount)
		kp.LastUsedDerivationIndex = 0
	}

	if includeAdditionalAccounts {
		generatedWalletAccount1 := &Account{
			Address:               types.Address{0x03},
			KeyUID:                kp.KeyUID,
			Wallet:                false,
			Chat:                  false,
			Type:                  AccountTypeGenerated,
			Path:                  "m/44'/60'/0'/0/1",
			PublicKey:             types.Hex2Bytes("0x000000003"),
			Name:                  "Generated Acc 2",
			Emoji:                 "emoji-2",
			ColorID:               common.CustomizationColorPrimary,
			Hidden:                false,
			Clock:                 0,
			Removed:               false,
			Operable:              AccountFullyOperable,
			ProdPreferredChainIDs: "1",
			TestPreferredChainIDs: "5",
		}
		kp.Accounts = append(kp.Accounts, generatedWalletAccount1)
		kp.LastUsedDerivationIndex = 1

		generatedWalletAccount2 := &Account{
			Address:               types.Address{0x04},
			KeyUID:                kp.KeyUID,
			Wallet:                false,
			Chat:                  false,
			Type:                  AccountTypeGenerated,
			Path:                  "m/44'/60'/0'/0/2",
			PublicKey:             types.Hex2Bytes("0x000000004"),
			Name:                  "Generated Acc 3",
			Emoji:                 "emoji-3",
			ColorID:               common.CustomizationColorPrimary,
			Hidden:                false,
			Clock:                 0,
			Removed:               false,
			Operable:              AccountFullyOperable,
			ProdPreferredChainIDs: "1",
			TestPreferredChainIDs: "5",
		}
		kp.Accounts = append(kp.Accounts, generatedWalletAccount2)
		kp.LastUsedDerivationIndex = 2
	}

	return kp
}

func GetSeedImportedKeypair1ForTest() *Keypair {
	kp := &Keypair{
		KeyUID:      "0000000000000000000000000000000000000000000000000000000000000002",
		Name:        "Seed Imported 1",
		Type:        KeypairTypeSeed,
		DerivedFrom: "0x0002",
	}

	seedGeneratedWalletAccount1 := &Account{
		Address:   types.Address{0x21},
		KeyUID:    kp.KeyUID,
		Wallet:    false,
		Chat:      false,
		Type:      AccountTypeSeed,
		Path:      "m/44'/60'/0'/0/0",
		PublicKey: types.Hex2Bytes("0x000000021"),
		Name:      "Seed Impo 1 Acc 1",
		Emoji:     "emoji-1",
		ColorID:   common.CustomizationColorPrimary,
		Hidden:    false,
		Clock:     0,
		Removed:   false,
		Operable:  AccountFullyOperable,
	}
	kp.Accounts = append(kp.Accounts, seedGeneratedWalletAccount1)
	kp.LastUsedDerivationIndex = 0

	seedGeneratedWalletAccount2 := &Account{
		Address:   types.Address{0x22},
		KeyUID:    kp.KeyUID,
		Wallet:    false,
		Chat:      false,
		Type:      AccountTypeSeed,
		Path:      "m/44'/60'/0'/0/1",
		PublicKey: types.Hex2Bytes("0x000000022"),
		Name:      "Seed Impo 1 Acc 2",
		Emoji:     "emoji-2",
		ColorID:   common.CustomizationColorPrimary,
		Hidden:    false,
		Clock:     0,
		Removed:   false,
		Operable:  AccountFullyOperable,
	}
	kp.Accounts = append(kp.Accounts, seedGeneratedWalletAccount2)
	kp.LastUsedDerivationIndex = 1

	return kp
}

func GetSeedImportedKeypair2ForTest() *Keypair {
	kp := &Keypair{
		KeyUID:      "0000000000000000000000000000000000000000000000000000000000000003",
		Name:        "Seed Imported 2",
		Type:        KeypairTypeSeed,
		DerivedFrom: "0x0003",
	}

	seedGeneratedWalletAccount1 := &Account{
		Address:   types.Address{0x31},
		KeyUID:    kp.KeyUID,
		Wallet:    false,
		Chat:      false,
		Type:      AccountTypeSeed,
		Path:      "m/44'/60'/0'/0/0",
		PublicKey: types.Hex2Bytes("0x000000031"),
		Name:      "Seed Impo 2 Acc 1",
		Emoji:     "emoji-1",
		ColorID:   common.CustomizationColorPrimary,
		Hidden:    false,
		Clock:     0,
		Removed:   false,
		Operable:  AccountFullyOperable,
	}
	kp.Accounts = append(kp.Accounts, seedGeneratedWalletAccount1)
	kp.LastUsedDerivationIndex = 0

	seedGeneratedWalletAccount2 := &Account{
		Address:   types.Address{0x32},
		KeyUID:    kp.KeyUID,
		Wallet:    false,
		Chat:      false,
		Type:      AccountTypeSeed,
		Path:      "m/44'/60'/0'/0/1",
		PublicKey: types.Hex2Bytes("0x000000032"),
		Name:      "Seed Impo 2 Acc 2",
		Emoji:     "emoji-2",
		ColorID:   common.CustomizationColorPrimary,
		Hidden:    false,
		Clock:     0,
		Removed:   false,
		Operable:  AccountFullyOperable,
	}
	kp.Accounts = append(kp.Accounts, seedGeneratedWalletAccount2)
	kp.LastUsedDerivationIndex = 1

	return kp
}

func GetPrivKeyImportedKeypairForTest() *Keypair {
	kp := &Keypair{
		KeyUID:      "0000000000000000000000000000000000000000000000000000000000000004",
		Name:        "Priv Key Imported",
		Type:        KeypairTypeKey,
		DerivedFrom: "", // no derived from for private key imported kp
	}

	privKeyWalletAccount := &Account{
		Address:   types.Address{0x41},
		KeyUID:    kp.KeyUID,
		Wallet:    false,
		Chat:      false,
		Type:      AccountTypeKey,
		Path:      "m",
		PublicKey: types.Hex2Bytes("0x000000041"),
		Name:      "Priv Key Impo Acc",
		Emoji:     "emoji-1",
		ColorID:   common.CustomizationColorPrimary,
		Hidden:    false,
		Clock:     0,
		Removed:   false,
		Operable:  AccountFullyOperable,
	}
	kp.Accounts = append(kp.Accounts, privKeyWalletAccount)

	return kp
}

func GetProfileKeycardForTest() *Keycard {
	profileKp := GetProfileKeypairForTest(true, true, true)
	keycard1Addresses := []types.Address{}
	for _, acc := range profileKp.Accounts {
		keycard1Addresses = append(keycard1Addresses, acc.Address)
	}
	return &Keycard{
		KeycardUID:        "00000000000000000000000000000001",
		KeycardName:       "Card01",
		KeycardLocked:     false,
		AccountsAddresses: keycard1Addresses,
		KeyUID:            profileKp.KeyUID,
		Position:          0,
	}
}

func GetKeycardForSeedImportedKeypair1ForTest() *Keycard {
	seed1Kp := GetSeedImportedKeypair1ForTest()
	keycard2Addresses := []types.Address{}
	for _, acc := range seed1Kp.Accounts {
		keycard2Addresses = append(keycard2Addresses, acc.Address)
	}
	return &Keycard{
		KeycardUID:        "00000000000000000000000000000002",
		KeycardName:       "Card02",
		KeycardLocked:     false,
		AccountsAddresses: keycard2Addresses,
		KeyUID:            seed1Kp.KeyUID,
		Position:          1,
	}
}

func GetKeycardForSeedImportedKeypair2ForTest() *Keycard {
	seed2Kp := GetSeedImportedKeypair2ForTest()
	keycard4Addresses := []types.Address{}
	for _, acc := range seed2Kp.Accounts {
		keycard4Addresses = append(keycard4Addresses, acc.Address)
	}
	return &Keycard{
		KeycardUID:        "00000000000000000000000000000003",
		KeycardName:       "Card03",
		KeycardLocked:     false,
		AccountsAddresses: keycard4Addresses,
		KeyUID:            seed2Kp.KeyUID,
		Position:          2,
	}
}

func Contains[T comparable](container []T, element T, isEqual func(T, T) bool) bool {
	for _, e := range container {
		if isEqual(e, element) {
			return true
		}
	}
	return false
}

func HaveSameElements[T comparable](a []T, b []T, isEqual func(T, T) bool) bool {
	for _, v := range a {
		if !Contains(b, v, isEqual) {
			return false
		}
	}
	return true
}

func SameAccounts(expected, real *Account) bool {
	return expected.Address == real.Address &&
		expected.KeyUID == real.KeyUID &&
		expected.Wallet == real.Wallet &&
		expected.Chat == real.Chat &&
		expected.Type == real.Type &&
		expected.Path == real.Path &&
		string(expected.PublicKey) == string(real.PublicKey) &&
		expected.Name == real.Name &&
		expected.Emoji == real.Emoji &&
		expected.ColorID == real.ColorID &&
		expected.Hidden == real.Hidden &&
		expected.Clock == real.Clock &&
		expected.Removed == real.Removed &&
		expected.ProdPreferredChainIDs == real.ProdPreferredChainIDs &&
		expected.TestPreferredChainIDs == real.TestPreferredChainIDs
}

func SameAccountsIncludingPosition(expected, real *Account) bool {
	return SameAccounts(expected, real) && expected.Position == real.Position
}

func SameAccountsWithDifferentOperable(expected, real *Account, expectedOperableValue AccountOperable) bool {
	return SameAccounts(expected, real) && real.Operable == expectedOperableValue
}

func SameKeypairs(expected, real *Keypair) bool {
	same := expected.KeyUID == real.KeyUID &&
		expected.Name == real.Name &&
		expected.Type == real.Type &&
		expected.DerivedFrom == real.DerivedFrom &&
		expected.LastUsedDerivationIndex == real.LastUsedDerivationIndex &&
		expected.Clock == real.Clock &&
		len(expected.Accounts) == len(real.Accounts)

	if same {
		for i := range expected.Accounts {
			found := false
			for j := range real.Accounts {
				if SameAccounts(expected.Accounts[i], real.Accounts[j]) {
					found = true
					break
				}
			}

			if !found {
				return false
			}
		}
	}

	return same
}

func SameKeypairsWithDifferentSyncedFrom(expected, real *Keypair, ignoreSyncedFrom bool, expectedSyncedFromValue string,
	expectedOperableValue AccountOperable) bool {
	same := expected.KeyUID == real.KeyUID &&
		expected.Name == real.Name &&
		expected.Type == real.Type &&
		expected.DerivedFrom == real.DerivedFrom &&
		expected.LastUsedDerivationIndex == real.LastUsedDerivationIndex &&
		expected.Clock == real.Clock &&
		len(expected.Accounts) == len(real.Accounts)

	if same && !ignoreSyncedFrom {
		same = same && real.SyncedFrom == expectedSyncedFromValue
	}

	if same {
		for i := range expected.Accounts {
			found := false
			for j := range real.Accounts {
				if SameAccountsWithDifferentOperable(expected.Accounts[i], real.Accounts[j], expectedOperableValue) {
					found = true
					break
				}
			}

			if !found {
				return false
			}
		}
	}

	return same
}

func SameKeycards(expected, real *Keycard) bool {
	same := expected.KeycardUID == real.KeycardUID &&
		expected.KeyUID == real.KeyUID &&
		expected.KeycardName == real.KeycardName &&
		expected.KeycardLocked == real.KeycardLocked &&
		expected.Position == real.Position &&
		len(expected.AccountsAddresses) == len(real.AccountsAddresses)

	if same {
		for i := range expected.AccountsAddresses {
			found := false
			for j := range real.AccountsAddresses {
				if expected.AccountsAddresses[i] == real.AccountsAddresses[j] {
					found = true
					break
				}
			}

			if !found {
				return false
			}
		}
	}

	return same
}
