package parser

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
)

const (
	// tag length constant in bytes. We define it here as it won't change, but
	// this allows us to not bring whole Babylon node as dependency.
	TagLen = 4

	// minimum unbonding output value to avoid the unbonding output value being
	// less than Bitcoin dust
	MinUnbondingOutputValue = btcutil.Amount(1000)
)

func checkPositive(value uint64) error {
	if value == 0 {
		return fmt.Errorf("value must be positive")
	}
	return nil
}

func parseTimeLockValue(timelock uint64) (uint16, error) {
	if timelock > math.MaxUint16 {
		return 0, fmt.Errorf("timelock value %d is too large. Max: %d", timelock, math.MaxUint16)
	}

	if err := checkPositive(timelock); err != nil {
		return 0, fmt.Errorf("invalid timelock value: %w", err)
	}

	return uint16(timelock), nil
}

func parseConfirmationDepthValue(confirmationDepth uint64) (uint16, error) {
	if confirmationDepth > math.MaxUint16 {
		return 0, fmt.Errorf("timelock value %d is too large. Max: %d", confirmationDepth, math.MaxUint16)
	}

	if confirmationDepth <= 1 {
		return 0, fmt.Errorf("confirmation depth value should be at least 2, got %d", confirmationDepth)
	}

	return uint16(confirmationDepth), nil
}

func parseBtcValue(value uint64) (btcutil.Amount, error) {
	if value > math.MaxInt64 {
		return 0, fmt.Errorf("value %d is too large. Max: %d", value, math.MaxInt64)
	}

	if err := checkPositive(value); err != nil {
		return 0, fmt.Errorf("invalid btc value: %w", err)
	}
	// return amount in satoshis
	return btcutil.Amount(value), nil
}

func parseUint32(value uint64) (uint32, error) {
	if value > math.MaxUint32 {
		return 0, fmt.Errorf("value %d is too large. Max: %d", value, math.MaxUint32)
	}

	if err := checkPositive(value); err != nil {
		return 0, fmt.Errorf("invalid value: %w", err)
	}

	return uint32(value), nil
}

// parseCovenantPubKeyFromHex parses public key string to btc public key
// the input should be 33 bytes
func parseCovenantPubKeyFromHex(pkStr string) (*btcec.PublicKey, error) {
	pkBytes, err := hex.DecodeString(pkStr)
	if err != nil {
		return nil, err
	}

	pk, err := btcec.ParsePubKey(pkBytes)
	if err != nil {
		return nil, err
	}

	return pk, nil
}

// either staking cap and cap height should be positive if cap height is positive
func parseCap(stakingCap, capHeight uint64) (btcutil.Amount, uint64, error) {
	if stakingCap != 0 && capHeight != 0 {
		return 0, 0, fmt.Errorf("only either of staking cap and cap height can be set")
	}

	if stakingCap == 0 && capHeight == 0 {
		return 0, 0, fmt.Errorf("either of staking cap and cap height must be set")
	}

	if stakingCap != 0 {
		parsedStakingCap, err := parseBtcValue(stakingCap)
		return parsedStakingCap, 0, err
	}

	return 0, capHeight, nil
}

type GlobalParams struct {
	Versions []*VersionedGlobalParams `json:"versions"`
}

type VersionedGlobalParams struct {
	Version           uint64   `json:"version"`
	ActivationHeight  uint64   `json:"activation_height"`
	StakingCap        uint64   `json:"staking_cap,omitempty"`
	CapHeight         uint64   `json:"cap_height,omitempty"`
	Tag               string   `json:"tag"`
	CovenantPks       []string `json:"covenant_pks"`
	CovenantQuorum    uint64   `json:"covenant_quorum"`
	UnbondingTime     uint64   `json:"unbonding_time"`
	UnbondingFee      uint64   `json:"unbonding_fee"`
	MaxStakingAmount  uint64   `json:"max_staking_amount"`
	MinStakingAmount  uint64   `json:"min_staking_amount"`
	MaxStakingTime    uint64   `json:"max_staking_time"`
	MinStakingTime    uint64   `json:"min_staking_time"`
	ConfirmationDepth uint64   `json:"confirmation_depth"`
}

type ParsedGlobalParams struct {
	Versions []*ParsedVersionedGlobalParams
}

type ParsedVersionedGlobalParams struct {
	Version           uint64
	ActivationHeight  uint64
	StakingCap        btcutil.Amount
	CapHeight         uint64
	Tag               []byte
	CovenantPks       []*btcec.PublicKey
	CovenantQuorum    uint32
	UnbondingTime     uint16
	UnbondingFee      btcutil.Amount
	MaxStakingAmount  btcutil.Amount
	MinStakingAmount  btcutil.Amount
	MaxStakingTime    uint16
	MinStakingTime    uint16
	ConfirmationDepth uint16
}

func ParseGlobalParams(p *GlobalParams) (*ParsedGlobalParams, error) {
	if len(p.Versions) == 0 {
		return nil, fmt.Errorf("global params must have at least one version")
	}
	var parsedVersions []*ParsedVersionedGlobalParams

	for i, v := range p.Versions {
		vCopy := v
		cv, err := parseVersionedGlobalParams(vCopy)

		if err != nil {
			return nil, fmt.Errorf("invalid params with version %d: %w", vCopy.Version, err)
		}

		// Check latest version
		if len(parsedVersions) > 0 {
			pv := parsedVersions[len(parsedVersions)-1]

			lastStakingCap := FindLastStakingCap(p.Versions[:i])

			if cv.Version != pv.Version+1 {
				return nil, fmt.Errorf("invalid params with version %d. versions should be monotonically increasing by 1", cv.Version)
			}
			if cv.StakingCap != 0 && cv.StakingCap < btcutil.Amount(lastStakingCap) {
				return nil, fmt.Errorf("invalid params with version %d. staking cap cannot be decreased in later versions, last non-zero staking cap: %d, got: %d",
					cv.Version, lastStakingCap, cv.StakingCap)
			}
			if cv.ActivationHeight <= pv.ActivationHeight {
				return nil, fmt.Errorf("invalid params with version %d. activation height cannot be overlapping between earlier and later versions", cv.Version)
			}
		}

		parsedVersions = append(parsedVersions, cv)
	}

	return &ParsedGlobalParams{
		Versions: parsedVersions,
	}, nil
}

func parseVersionedGlobalParams(p *VersionedGlobalParams) (*ParsedVersionedGlobalParams, error) {
	tag, err := hex.DecodeString(p.Tag)

	if err != nil {
		return nil, fmt.Errorf("invalid tag: %w", err)
	}

	if len(tag) != TagLen {
		return nil, fmt.Errorf("invalid tag length, expected %d, got %d", TagLen, len(tag))
	}

	if len(p.CovenantPks) == 0 {
		return nil, fmt.Errorf("empty covenant public keys")
	}
	if p.CovenantQuorum > uint64(len(p.CovenantPks)) {
		return nil, fmt.Errorf("covenant quorum %d cannot be more than the amount of covenants %d", p.CovenantQuorum, len(p.CovenantPks))
	}

	quorum, err := parseUint32(p.CovenantQuorum)
	if err != nil {
		return nil, fmt.Errorf("invalid covenant quorum: %w", err)
	}

	var covenantKeys []*btcec.PublicKey
	for _, covPk := range p.CovenantPks {
		pk, err := parseCovenantPubKeyFromHex(covPk)
		if err != nil {
			return nil, fmt.Errorf("invalid covenant public key %s: %w", covPk, err)
		}

		covenantKeys = append(covenantKeys, pk)
	}

	maxStakingAmount, err := parseBtcValue(p.MaxStakingAmount)

	if err != nil {
		return nil, fmt.Errorf("invalid max_staking_amount: %w", err)
	}

	minStakingAmount, err := parseBtcValue(p.MinStakingAmount)

	if err != nil {
		return nil, fmt.Errorf("invalid min_staking_amount: %w", err)
	}

	// NOTE: Allow config when max-staking-amount is equal tomin-staking-amount, as then
	// we can configure a fixed staking amount
	if maxStakingAmount < minStakingAmount {
		return nil, fmt.Errorf("max-staking-amount %d must be larger than or equal to min-staking-amount %d", maxStakingAmount, minStakingAmount)
	}

	ubTime, err := parseTimeLockValue(p.UnbondingTime)
	if err != nil {
		return nil, fmt.Errorf("invalid unbonding_time: %w", err)
	}

	ubFee, err := parseBtcValue(p.UnbondingFee)
	if err != nil {
		return nil, fmt.Errorf("invalid unbonding_fee: %w", err)
	}

	if minStakingAmount < ubFee+MinUnbondingOutputValue {
		return nil, fmt.Errorf("min_staking_amount %d should not be less than unbonding fee %d plus %d",
			minStakingAmount, ubFee, MinUnbondingOutputValue)
	}

	maxStakingTime, err := parseTimeLockValue(p.MaxStakingTime)
	if err != nil {
		return nil, fmt.Errorf("invalid max_staking_time: %w", err)
	}

	minStakingTime, err := parseTimeLockValue(p.MinStakingTime)
	if err != nil {
		return nil, fmt.Errorf("invalid min_staking_time: %w", err)
	}

	// NOTE: Allow config when max-staking-time is equal to min-staking-time, as then
	// we can configure a fixed staking time.
	if maxStakingTime < minStakingTime {
		return nil, fmt.Errorf("max-staking-time %d must be larger than or equal to min-staking-time %d", maxStakingTime, minStakingTime)
	}

	confirmationDepth, err := parseConfirmationDepthValue(p.ConfirmationDepth)
	if err != nil {
		return nil, fmt.Errorf("invalid confirmation_depth: %w", err)
	}

	if err := checkPositive(p.ActivationHeight); err != nil {
		return nil, fmt.Errorf("activation_height: %w", err)
	}

	stakingCap, capHeight, err := parseCap(p.StakingCap, p.CapHeight)
	if err != nil {
		return nil, fmt.Errorf("invalid cap: %w", err)
	}

	if stakingCap != 0 && stakingCap < maxStakingAmount {
		return nil, fmt.Errorf("invalid staking_cap, should be larger than max_staking_amount: %d, got: %d",
			maxStakingAmount, stakingCap)
	}

	return &ParsedVersionedGlobalParams{
		Version:           p.Version,
		ActivationHeight:  p.ActivationHeight,
		StakingCap:        stakingCap,
		CapHeight:         capHeight,
		Tag:               tag,
		CovenantPks:       covenantKeys,
		CovenantQuorum:    quorum,
		UnbondingTime:     ubTime,
		UnbondingFee:      ubFee,
		MaxStakingAmount:  maxStakingAmount,
		MinStakingAmount:  minStakingAmount,
		MaxStakingTime:    maxStakingTime,
		MinStakingTime:    minStakingTime,
		ConfirmationDepth: confirmationDepth,
	}, nil
}

// GetVersionedGlobalParamsByHeight return the parsed versioned global params which
// are applicable at the given BTC btcHeight. If there in no versioned global params
// applicable at the given btcHeight, it will return nil.
func (g *ParsedGlobalParams) GetVersionedGlobalParamsByHeight(btcHeight uint64) *ParsedVersionedGlobalParams {
	// Iterate the list in reverse (i.e. decreasing ActivationHeight)
	// and identify the first element that has an activation height below
	// the specified BTC height.
	for i := len(g.Versions) - 1; i >= 0; i-- {
		paramsVersion := g.Versions[i]
		if paramsVersion.ActivationHeight <= btcHeight {
			return paramsVersion
		}
	}
	return nil
}

// FindLastStakingCap finds the last staking cap that is not zero
// it returns zero if not non-zero value is found
func FindLastStakingCap(prevVersions []*VersionedGlobalParams) uint64 {
	numPrevVersions := len(prevVersions)
	if len(prevVersions) == 0 {
		return 0
	}

	for i := numPrevVersions - 1; i >= 0; i-- {
		if prevVersions[i].StakingCap > 0 {
			return prevVersions[i].StakingCap
		}
	}

	return 0
}

func NewParsedGlobalParamsFromFile(filePath string) (*ParsedGlobalParams, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return NewParsedGlobalParamsFromBytes(data)
}

func NewParsedGlobalParamsFromBytes(data []byte) (*ParsedGlobalParams, error) {
	var globalParams GlobalParams
	err := json.Unmarshal(data, &globalParams)
	if err != nil {
		return nil, err
	}

	parsedGlobalParams, err := ParseGlobalParams(&globalParams)

	if err != nil {
		return nil, err
	}

	return parsedGlobalParams, nil
}
