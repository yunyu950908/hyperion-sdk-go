package hyperion

import (
	"crypto/sha3"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

const deriveObjectAddressFromSeed = byte(0xfe)

type aptosAccountAddress [32]byte

func swapArgumentAddress(currency string) (string, error) {
	if !isCoinType(currency) {
		return currency, nil
	}
	return coinTypeToFAMetadataAddress(currency)
}

// coinTypeToFAMetadataAddress mirrors aptos-tool@0.0.11 Token.faTypeCalculate.
// The upstream SDK converts swap function arguments from coin types to FA
// metadata addresses, while keeping the original coin type for type selection.
func coinTypeToFAMetadataAddress(coinType string) (string, error) {
	parts := strings.Split(coinType, "::")
	if len(parts) < 3 {
		return "", errors.New("coin type must be address::module::name")
	}
	if isAptosCoinType(parts) {
		return "0xa", nil
	}

	contractAddress, err := parseAptosAccountAddress(parts[0])
	if err != nil {
		return "", fmt.Errorf("coin type address: %w", err)
	}

	shortContractAddress := strings.Join([]string{
		contractAddress.String(),
		parts[1],
		parts[2],
	}, "::")
	creator, err := parseAptosAccountAddress("0xa")
	if err != nil {
		return "", err
	}
	faType := createObjectAddress(creator, []byte(shortContractAddress))
	return faType.String(), nil
}

func isAptosCoinType(parts []string) bool {
	if len(parts) < 3 || parts[1] != "aptos_coin" || parts[2] != "AptosCoin" {
		return false
	}
	address, err := parseAptosAccountAddress(parts[0])
	return err == nil && address.LongString() == "0x0000000000000000000000000000000000000000000000000000000000000001"
}

// parseAptosAccountAddress implements the AccountAddress.from parsing behavior
// needed by aptos-tool: relaxed short special addresses plus long addresses that
// are missing no more than four leading hex characters.
func parseAptosAccountAddress(input string) (aptosAccountAddress, error) {
	var out aptosAccountAddress
	parsed := strings.TrimPrefix(input, "0x")
	if parsed == "" {
		return out, errors.New("hex string is too short")
	}
	if len(parsed) > 64 {
		return out, errors.New("hex string is too long")
	}

	padded := strings.Repeat("0", 64-len(parsed)) + parsed
	bytes, err := hex.DecodeString(padded)
	if err != nil {
		return out, fmt.Errorf("invalid hex characters: %w", err)
	}
	copy(out[:], bytes)

	if len(parsed) < 60 && !out.isSpecial() {
		return out, errors.New("hex string is too short for a non-special address")
	}
	return out, nil
}

func (a aptosAccountAddress) String() string {
	if a.isSpecial() {
		return fmt.Sprintf("0x%x", a[31])
	}
	return a.LongString()
}

func (a aptosAccountAddress) LongString() string {
	return "0x" + hex.EncodeToString(a[:])
}

func (a aptosAccountAddress) isSpecial() bool {
	for _, value := range a[:31] {
		if value != 0 {
			return false
		}
	}
	return a[31] < 16
}

// createObjectAddress mirrors @aptos-labs/ts-sdk createObjectAddress for named
// objects: sha3_256(creator_address_bcs || seed || 0xfe).
func createObjectAddress(creator aptosAccountAddress, seed []byte) aptosAccountAddress {
	hash := sha3.New256()
	hash.Write(creator[:])
	hash.Write(seed)
	hash.Write([]byte{deriveObjectAddressFromSeed})

	var out aptosAccountAddress
	copy(out[:], hash.Sum(nil))
	return out
}
