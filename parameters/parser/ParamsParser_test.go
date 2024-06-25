package parser_test

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/babylonchain/networks/parameters/parser"
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
)

func generateInitParams(t *testing.T, r *rand.Rand) *parser.VersionedGlobalParams {
	var pks []string

	quorum := r.Intn(10) + 1

	for i := 0; i < quorum+1; i++ {
		privkey, err := btcec.NewPrivateKey()
		require.NoError(t, err)
		pks = append(pks, hex.EncodeToString(privkey.PubKey().SerializeCompressed()))
	}

	gp := parser.VersionedGlobalParams{
		Version:           0,
		ActivationHeight:  1,
		StakingCap:        uint64(r.Int63n(int64(initialCapMin)) + int64(initialCapMin)),
		CapHeight:         0,
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
		// 1/3 chance to have a time-based cap
		if r.Intn(3) == 0 {
			next.CapHeight = next.ActivationHeight + uint64(r.Int63n(100)+100)
			next.StakingCap = 0
		} else {
			lastStakingCap := parser.FindLastStakingCap(versions[:i])
			next.StakingCap = lastStakingCap + uint64(r.Int63n(1000000000)+1)
		}

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
			require.Equal(t, globalParams.Versions[i].CapHeight, p.CapHeight)
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

		// If we are querying exactly by one of the activation height, we should always
		// retrieve original parameters

		params := parsedParams.GetVersionedGlobalParamsByHeight(randVersionedParams.ActivationHeight)
		require.NotNil(t, params)

		require.Equal(t, randVersionedParams.StakingCap, params.StakingCap)
		require.Equal(t, randVersionedParams.CapHeight, params.CapHeight)
		require.Equal(t, randVersionedParams.Version, params.Version)
		require.Equal(t, randVersionedParams.ActivationHeight, params.ActivationHeight)
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
			require.Equal(t, paramsBeforeRand.StakingCap, params.StakingCap)
			require.Equal(t, paramsBeforeRand.CapHeight, params.CapHeight)
			require.Equal(t, paramsBeforeRand.Version, params.Version)
			require.Equal(t, paramsBeforeRand.ActivationHeight, params.ActivationHeight)
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

func TestReadBbnTest4Params(t *testing.T) {
	// Test checking whether the global params which is used on testnet currently
	// is parsed correctly
	globalParams, err := parser.NewParsedGlobalParamsFromFile("../../bbn-test-4/parameters/global-params.json")
	require.NoError(t, err)
	require.NotNil(t, globalParams)
}

var defaultParam = parser.VersionedGlobalParams{
	Version:          0,
	ActivationHeight: 100,
	StakingCap:       400000,
	Tag:              "01020304",
	CovenantPks: []string{
		"03ffeaec52a9b407b355ef6967a7ffc15fd6c3fe07de2844d61550475e7a5233e5",
		"03a5c60c2188e833d39d0fa798ab3f69aa12ed3dd2f3bad659effa252782de3c31",
		"0359d3532148a597a2d05c0395bf5f7176044b1cd312f37701a9b4d0aad70bc5a4",
		"0357349e985e742d5131e1e2b227b5170f6350ac2e2feb72254fcc25b3cee21a18",
		"03c8ccb03c379e452f10c81232b41a1ca8b63d0baf8387e57d302c987e5abb8527",
	},
	CovenantQuorum:    3,
	UnbondingTime:     1000,
	UnbondingFee:      1000,
	MaxStakingAmount:  300000,
	MinStakingAmount:  3000,
	MaxStakingTime:    10000,
	MinStakingTime:    100,
	ConfirmationDepth: 10,
}

func TestFailGlobalParamsValidation(t *testing.T) {
	var clonedParams parser.GlobalParams
	defaultGlobalParams := parser.GlobalParams{
		Versions: []*parser.VersionedGlobalParams{&defaultParam},
	}
	// Empty versions
	jsonData := []byte(`{
		"versions": [
		]
	}`)
	fileName := createJsonFile(t, jsonData)
	// Call NewGlobalParams with the path to the temporary file
	_, err := parser.NewParsedGlobalParamsFromFile(fileName)
	assert.Equal(t, "global params must have at least one version", err.Error())

	// invalid tag length
	deepCopy(&defaultGlobalParams, &clonedParams)
	clonedParams.Versions[0].Tag = "010203"

	invalidTagJsonData, err := json.Marshal(clonedParams)
	assert.NoError(t, err, "marshalling invalid tag data should not fail")
	fileName = createJsonFile(t, invalidTagJsonData)
	_, err = parser.NewParsedGlobalParamsFromFile(fileName)
	assert.Equal(t, "invalid params with version 0: invalid tag length, expected 4, got 3", err.Error())

	// test covenant pks sizes
	var invalidCovenantPksParam parser.GlobalParams
	deepCopy(&defaultGlobalParams, &invalidCovenantPksParam)
	invalidCovenantPksParam.Versions[0].CovenantPks = []string{}

	invalidJson, err := json.Marshal(invalidCovenantPksParam)
	assert.NoError(t, err, "marshalling invalid covenant pks data should not fail")

	fileName = createJsonFile(t, invalidJson)
	_, err = parser.NewParsedGlobalParamsFromFile(fileName)
	assert.Equal(t, "invalid params with version 0: empty covenant public keys", err.Error())

	// test covenant quorum
	deepCopy(&defaultGlobalParams, &clonedParams)
	clonedParams.Versions[0].CovenantQuorum = 6

	invalidJson, err = json.Marshal(clonedParams)
	assert.NoError(t, err, "marshalling invalid covenant pks data should not fail")

	fileName = createJsonFile(t, invalidJson)
	_, err = parser.NewParsedGlobalParamsFromFile(fileName)
	assert.Equal(t, "invalid params with version 0: covenant quorum 6 cannot be more than the amount of covenants 5", err.Error())

	// test invalid convenant pks
	deepCopy(&defaultGlobalParams, &clonedParams)
	clonedParams.Versions[0].CovenantPks = []string{
		"04ffeaec52a9b407b355ef6967a7ffc15fd6c3fe07de2844d61550475e7a5233e5",
		"03a5c60c2188e833d39d0fa798ab3f69aa12ed3dd2f3bad659effa252782de3c31",
		"0359d3532148a597a2d05c0395bf5f7176044b1cd312f37701a9b4d0aad70bc5a4",
		"0357349e985e742d5131e1e2b227b5170f6350ac2e2feb72254fcc25b3cee21a18",
		"03c8ccb03c379e452f10c81232b41a1ca8b63d0baf8387e57d302c987e5abb8527",
		"03c8ccb03c379e452f10c81232b41a1ca8b63d0baf8387e57d302c987e5abb8527",
	}

	jsonData, err = json.Marshal(clonedParams)
	assert.NoError(t, err, "marshalling invalid covenant pks data should not fail")

	fileName = createJsonFile(t, jsonData)
	_, err = parser.NewParsedGlobalParamsFromFile(fileName)
	assert.Contains(t, err.Error(), "invalid covenant public key")

	// test invalid min and max staking amount
	deepCopy(&defaultGlobalParams, &clonedParams)
	clonedParams.Versions[0].MaxStakingAmount = 300
	clonedParams.Versions[0].MinStakingAmount = 400

	jsonData, err = json.Marshal(clonedParams)
	assert.NoError(t, err, "marshalling invalid staking amount data should not fail")

	fileName = createJsonFile(t, jsonData)
	_, err = parser.NewParsedGlobalParamsFromFile(fileName)
	assert.Equal(t, "invalid params with version 0: max-staking-amount 300 must be larger than or equal to min-staking-amount 400", err.Error())

	// test activation height
	deepCopy(&defaultGlobalParams, &clonedParams)
	clonedParams.Versions[0].ActivationHeight = 0

	jsonData, err = json.Marshal(clonedParams)
	assert.NoError(t, err)

	fileName = createJsonFile(t, jsonData)
	_, err = parser.NewParsedGlobalParamsFromFile(fileName)
	assert.Equal(t, "invalid params with version 0: activation_height: value must be positive", err.Error())

	// test staking cap
	deepCopy(&defaultGlobalParams, &clonedParams)
	clonedParams.Versions[0].StakingCap = 0

	jsonData, err = json.Marshal(clonedParams)
	assert.NoError(t, err)

	fileName = createJsonFile(t, jsonData)
	_, err = parser.NewParsedGlobalParamsFromFile(fileName)
	assert.Equal(t, "invalid params with version 0: invalid cap: either of staking cap and cap height must be set", err.Error())

	// test confirmation depth
	deepCopy(&defaultGlobalParams, &clonedParams)
	clonedParams.Versions[0].ConfirmationDepth = 0

	jsonData, err = json.Marshal(clonedParams)
	assert.NoError(t, err)

	fileName = createJsonFile(t, jsonData)
	_, err = parser.NewParsedGlobalParamsFromFile(fileName)
	assert.Equal(t, "invalid params with version 0: invalid confirmation_depth: confirmation depth value should be at least 2, got 0", err.Error())
}

func TestGlobalParamsSortedByActivationHeight(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	params := generateGlobalParams(r, 10)
	// We pick a random one and set its activation height to be less than its previous one
	params[5].ActivationHeight = params[4].ActivationHeight - 1

	globalParams := parser.GlobalParams{
		Versions: params,
	}

	jsonData, err := json.Marshal(globalParams)
	assert.NoError(t, err)
	fileName := createJsonFile(t, jsonData)
	_, err = parser.NewParsedGlobalParamsFromFile(fileName)
	assert.Equal(t, "invalid params with version 5. activation height cannot be overlapping between earlier and later versions", err.Error())
}

func TestGlobalParamsWithIncrementalVersions(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	params := generateGlobalParams(r, 10)
	// We pick a random one and set its activation height to be less than its previous one
	params[5].Version = params[4].Version - 1

	globalParams := parser.GlobalParams{
		Versions: params,
	}

	jsonData, err := json.Marshal(globalParams)
	assert.NoError(t, err)

	fileName := createJsonFile(t, jsonData)
	_, err = parser.NewParsedGlobalParamsFromFile(fileName)

	assert.Equal(t, "invalid params with version 3. versions should be monotonically increasing by 1", err.Error())
}

func TestGlobalParamsWithIncrementalStakingCap(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	params := generateGlobalParams(r, 10)
	// We pick a random one and set its activation height to be less than its previous one
	params[5].StakingCap = params[4].StakingCap - 1

	globalParams := parser.GlobalParams{
		Versions: params,
	}

	jsonData, err := json.Marshal(globalParams)
	assert.NoError(t, err)

	fileName := createJsonFile(t, jsonData)
	_, err = parser.NewParsedGlobalParamsFromFile(fileName)
	assert.True(t, strings.Contains(err.Error(), "invalid params with version 5. staking cap cannot be decreased in later versions"))
}

func TestGlobalParamsWithSmallStakingCap(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	params := generateGlobalParams(r, 10)
	// We pick a random one and set its activation height to be less than its previous one
	params[5].StakingCap = params[5].MaxStakingAmount - 1

	globalParams := parser.GlobalParams{
		Versions: params,
	}

	jsonData, err := json.Marshal(globalParams)
	assert.NoError(t, err)

	fileName := createJsonFile(t, jsonData)
	_, err = parser.NewParsedGlobalParamsFromFile(fileName)
	assert.Equal(t, fmt.Sprintf("invalid params with version 5: invalid staking_cap, should be larger than max_staking_amount: %d, got: %d",
		params[5].MaxStakingAmount, params[5].StakingCap), err.Error())
}

func TestGlobalParamsWithCapBothSet(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	params := generateGlobalParams(r, 10)
	// We pick a random one and set its cap height to be the activation height
	params[5].CapHeight = params[5].ActivationHeight

	globalParams := parser.GlobalParams{
		Versions: params,
	}

	jsonData, err := json.Marshal(globalParams)
	assert.NoError(t, err)

	fileName := createJsonFile(t, jsonData)
	_, err = parser.NewParsedGlobalParamsFromFile(fileName)
	assert.Equal(t, "invalid params with version 5: invalid cap: only either of staking cap and cap height can be set", err.Error())
}

func TestGlobalParamsWithUnbondingFee(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	params := generateGlobalParams(r, 10)
	// We pick a random one and set its min_staking_amount less than unbonding_fee
	params[5].MinStakingAmount = uint64(r.Int63n(int64(params[5].UnbondingFee) + int64(parser.MinUnbondingOutputValue)))

	globalParams := parser.GlobalParams{
		Versions: params,
	}

	jsonData, err := json.Marshal(globalParams)
	assert.NoError(t, err)

	fileName := createJsonFile(t, jsonData)
	_, err = parser.NewParsedGlobalParamsFromFile(fileName)
	assert.Equal(t, fmt.Sprintf("invalid params with version 5: min_staking_amount %d should not be less than unbonding fee %d plus %d",
		params[5].MinStakingAmount, params[5].UnbondingFee, parser.MinUnbondingOutputValue), err.Error())
}

func generateGlobalParams(r *rand.Rand, numOfParams int) []*parser.VersionedGlobalParams {
	var params []*parser.VersionedGlobalParams

	lastParam := defaultParam
	for i := 0; i < numOfParams; i++ {
		var param parser.VersionedGlobalParams
		deepCopy(&defaultParam, &param)
		param.ActivationHeight = lastParam.ActivationHeight + uint64(r.Intn(100))
		param.Version = uint64(i)
		param.StakingCap = lastParam.StakingCap + uint64(r.Intn(100))
		params = append(params, &param)
		lastParam = param
	}

	return params
}

func deepCopy(src, dst interface{}) error {
	// Marshal the source object to JSON.
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}

	// Unmarshal the JSON data into the destination object.
	if err := json.Unmarshal(data, dst); err != nil {
		return err
	}

	return nil
}

func createJsonFile(t *testing.T, jsonData []byte) string {
	tempFile, err := os.CreateTemp("", "params-test-*")
	require.NoError(t, err)
	defer tempFile.Close()
	_, err = tempFile.Write(jsonData)
	require.NoError(t, err)
	t.Cleanup(func() {
		os.Remove(tempFile.Name())
	})
	return tempFile.Name()
}
