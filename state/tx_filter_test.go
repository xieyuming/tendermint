package state_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dbm "github.com/tendermint/tm-db"

	tmrand "github.com/tendermint/tendermint/libs/rand"
	sm "github.com/tendermint/tendermint/state"
	"github.com/tendermint/tendermint/types"
)

func TestTxFilter(t *testing.T) {
	genDoc := randomGenesisDoc()
	genDoc.ConsensusParams.Block.MaxBytes = 3000
	genDoc.ConsensusParams.Evidence.MaxNum = 1

	// Max size of Txs is much smaller than size of block,
	// since we need to account for commits and evidence.
	testCases := []struct {
		tx    types.Tx
		isErr bool
	}{
		{types.Tx(tmrand.Bytes(1680)), false},
		{types.Tx(tmrand.Bytes(1853)), true},
		{types.Tx(tmrand.Bytes(3000)), true},
	}

	for i, tc := range testCases {
		stateDB := dbm.NewDB("state", "memdb", os.TempDir())
		state, err := sm.LoadStateFromDBOrGenesisDoc(stateDB, genDoc)
		require.NoError(t, err)

		f := sm.TxPreCheck(state) // current max size of a tx 1850
		if tc.isErr {
			assert.NotNil(t, f(tc.tx), "#%v", i)
		} else {
			assert.Nil(t, f(tc.tx), "#%v", i)
		}
	}
}
