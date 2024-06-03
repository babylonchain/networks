package parser_test

import (
	"encoding/hex"
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/stretchr/testify/require"

	"github.com/babylonchain/parameters/parser"
)

func addRandomSeedsToFuzzer(f *testing.F, num uint) {
	// Seed based on the current time
	r := rand.New(rand.NewSource(time.Now().Unix()))
	var idx uint
	for idx = 0; idx < num; idx++ {
		f.Add(r.Int63())
	}
}

var (
	initialCapMin, _ = btcutil.NewAmount(100)
	tag              = hex.EncodeToString([]byte{0x01, 0x02, 0x03, 0x04})
	quorum           = 2
)

func generateInitParams(t *testing.T, r *rand.Rand) *parser.VersionedGlobalParams {
	var pks []string

	for i := 0; i < quorum+1; i++ {
		privkey, err := btcec.NewPrivateKey()
		require.NoError(t, err)
		pks = append(pks, hex.EncodeToString(privkey.PubKey().SerializeCompressed()))
	}

	gp := parser.VersionedGlobalParams{
		Version:           0,
		ActivationHeight:  0,
		StakingCap:        uint64(r.Int63n(int64(initialCapMin)) + int64(initialCapMin)),
		Tag:               tag,
		CovenantPks:       pks,
		CovenantQuorum:    uint64(quorum),
		UnbondingTime:     uint64(r.Int63n(100) + 100),
		UnbondingFee:      uint64(r.Int63n(100000) + 100000),
		MaxStakingAmount:  uint64(r.Int63n(100000000) + 100000000),
		MinStakingAmount:  uint64(r.Int63n(1000000) + 1000000),
		MaxStakingTime:    math.MaxUint16,
		MinStakingTime:    uint64(r.Int63n(10000) + 10000),
		ConfirmationDepth: uint64(r.Int63n(10) + 2),
	}

	return &gp
}

func genValidGlobalParam(
	t *testing.T,
	r *rand.Rand,
	num uint32,
) *parser.GlobalParams {
	require.True(t, num > 0)

	initParams := generateInitParams(t, r)

	if num == 1 {
		return &parser.GlobalParams{
			Versions: []*parser.VersionedGlobalParams{initParams},
		}
	}

	var versions []*parser.VersionedGlobalParams
	versions = append(versions, initParams)

	for i := 1; i < int(num); i++ {
		prev := versions[i-1]
		next := generateInitParams(t, r)
		next.ActivationHeight = prev.ActivationHeight + uint64(r.Int63n(100)+100)
		next.Version = prev.Version + 1
		next.StakingCap = prev.StakingCap + uint64(r.Int63n(1000000000)+1)
		versions = append(versions, next)
	}

	return &parser.GlobalParams{
		Versions: versions,
	}
}

// PROPERTY: Every valid global params should be parsed successfully
func FuzzParseValidParams(f *testing.F) {
	addRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		numVersions := uint32(r.Int63n(50) + 10)
		globalParams := genValidGlobalParam(t, r, numVersions)
		parsedParams, err := parser.ParseGlobalParams(globalParams)
		require.NoError(t, err)
		require.NotNil(t, parsedParams)
		require.Len(t, parsedParams.Versions, int(numVersions))
		for i, p := range parsedParams.Versions {
			require.Equal(t, globalParams.Versions[i].Version, p.Version)
			require.Equal(t, globalParams.Versions[i].ActivationHeight, p.ActivationHeight)
			require.Equal(t, globalParams.Versions[i].StakingCap, uint64(p.StakingCap))
			require.Equal(t, globalParams.Versions[i].Tag, hex.EncodeToString(p.Tag))
			require.Equal(t, globalParams.Versions[i].CovenantQuorum, uint64(p.CovenantQuorum))
			require.Equal(t, globalParams.Versions[i].UnbondingTime, uint64(p.UnbondingTime))
			require.Equal(t, globalParams.Versions[i].UnbondingFee, uint64(p.UnbondingFee))
			require.Equal(t, globalParams.Versions[i].MaxStakingAmount, uint64(p.MaxStakingAmount))
			require.Equal(t, globalParams.Versions[i].MinStakingAmount, uint64(p.MinStakingAmount))
			require.Equal(t, globalParams.Versions[i].MaxStakingTime, uint64(p.MaxStakingTime))
			require.Equal(t, globalParams.Versions[i].MinStakingTime, uint64(p.MinStakingTime))
			require.Equal(t, globalParams.Versions[i].ConfirmationDepth, uint64(p.ConfirmationDepth))
		}
	})
}

func FuzzRetrievingParametersByHeight(f *testing.F) {
	addRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		numVersions := uint32(r.Int63n(50) + 10)
		globalParams := genValidGlobalParam(t, r, numVersions)
		parsedParams, err := parser.ParseGlobalParams(globalParams)
		require.NoError(t, err)
		numOfParams := len(parsedParams.Versions)
		randParameterIndex := r.Intn(numOfParams)
		randVersionedParams := parsedParams.Versions[randParameterIndex]

		// If we are querying exactly by one of the activation height, we shuld always
		// retriveve original parameters

		params := parsedParams.GetVersionedGlobalParamsByHeight(randVersionedParams.ActivationHeight)
		require.NotNil(t, params)

		require.Equal(t, randVersionedParams.CovenantQuorum, params.CovenantQuorum)
		require.Equal(t, randVersionedParams.CovenantPks, params.CovenantPks)
		require.Equal(t, randVersionedParams.Tag, params.Tag)
		require.Equal(t, randVersionedParams.UnbondingTime, params.UnbondingTime)
		require.Equal(t, randVersionedParams.UnbondingFee, params.UnbondingFee)
		require.Equal(t, randVersionedParams.MaxStakingAmount, params.MaxStakingAmount)
		require.Equal(t, randVersionedParams.MinStakingAmount, params.MinStakingAmount)
		require.Equal(t, randVersionedParams.MaxStakingTime, params.MaxStakingTime)
		require.Equal(t, randVersionedParams.MinStakingTime, params.MinStakingTime)
		require.Equal(t, randVersionedParams.ConfirmationDepth, params.ConfirmationDepth)

		if randParameterIndex > 0 {
			// If we are querying by a height that is one before the activations height
			// of the randomly chosen parameter, we should retrieve previous parameters version
			params := parsedParams.GetVersionedGlobalParamsByHeight(randVersionedParams.ActivationHeight - 1)
			require.NotNil(t, params)
			paramsBeforeRand := parsedParams.Versions[randParameterIndex-1]
			require.Equal(t, paramsBeforeRand.CovenantQuorum, params.CovenantQuorum)
			require.Equal(t, paramsBeforeRand.CovenantPks, params.CovenantPks)
			require.Equal(t, paramsBeforeRand.Tag, params.Tag)
			require.Equal(t, paramsBeforeRand.UnbondingTime, params.UnbondingTime)
			require.Equal(t, paramsBeforeRand.UnbondingFee, params.UnbondingFee)
			require.Equal(t, paramsBeforeRand.MaxStakingAmount, params.MaxStakingAmount)
			require.Equal(t, paramsBeforeRand.MinStakingAmount, params.MinStakingAmount)
			require.Equal(t, paramsBeforeRand.MaxStakingTime, params.MaxStakingTime)
			require.Equal(t, paramsBeforeRand.MinStakingTime, params.MinStakingTime)
			require.Equal(t, paramsBeforeRand.ConfirmationDepth, params.ConfirmationDepth)
		}
	})
}
