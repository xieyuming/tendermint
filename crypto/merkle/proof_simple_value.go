package merkle

import (
	"bytes"
	"fmt"

	"github.com/tendermint/tendermint/crypto/tmhash"
	tmmerkle "github.com/tendermint/tendermint/proto/crypto/merkle"
)

const ProofOpSimpleValue = "simple:v"

// SimpleValueOp takes a key and a single value as argument and
// produces the root hash.  The corresponding tree structure is
// the SimpleMap tree.  SimpleMap takes a Hasher, and currently
// Tendermint uses tmhash.  SimpleValueOp should support
// the hash function as used in tmhash.  TODO support
// additional hash functions here as options/args to this
// operator.
//
// If the produced root hash matches the expected hash, the
// proof is good.
type SimpleValueOp struct {
	// Encoded in ProofOp.Key.
	key []byte

	// To encode in ProofOp.Data
	Proof *SimpleProof `json:"simple_proof"`
}

var _ ProofOperator = SimpleValueOp{}

func NewSimpleValueOp(key []byte, proof *SimpleProof) SimpleValueOp {
	return SimpleValueOp{
		key:   key,
		Proof: proof,
	}
}

func SimpleValueOpDecoder(pop ProofOp) (ProofOperator, error) {
	if pop.Type != ProofOpSimpleValue {
		return nil, fmt.Errorf("unexpected ProofOp.Type; got %v, want %v", pop.Type, ProofOpSimpleValue)
	}
	var pbop tmmerkle.SimpleValueOp // a bit strange as we'll discard this, but it works.
	err := pbop.Unmarshal(pop.Data)
	if err != nil {
		return nil, fmt.Errorf("decoding ProofOp.Data into SimpleValueOp: %w", err)
	}

	sp, err := SimpleProofFromProto(pbop.Proof)
	if err != nil {
		return nil, err
	}
	return NewSimpleValueOp(pop.Key, sp), nil
}

func (op SimpleValueOp) ProofOp() ProofOp {
	pbval := tmmerkle.SimpleValueOp{
		Key:   op.key,
		Proof: op.Proof.ToProto(),
	}
	bz, err := pbval.Marshal()
	if err != nil {
		panic(err)
	}
	return ProofOp{
		Type: ProofOpSimpleValue,
		Key:  op.key,
		Data: bz,
	}
}

func (op SimpleValueOp) String() string {
	return fmt.Sprintf("SimpleValueOp{%v}", op.GetKey())
}

func (op SimpleValueOp) Run(args [][]byte) ([][]byte, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("expected 1 arg, got %v", len(args))
	}
	value := args[0]
	hasher := tmhash.New()
	hasher.Write(value) // does not error
	vhash := hasher.Sum(nil)

	bz := new(bytes.Buffer)
	// Wrap <op.Key, vhash> to hash the KVPair.
	encodeByteSlice(bz, op.key) // does not error
	encodeByteSlice(bz, vhash)  // does not error
	kvhash := leafHash(bz.Bytes())

	if !bytes.Equal(kvhash, op.Proof.LeafHash) {
		return nil, fmt.Errorf("leaf hash mismatch: want %X got %X", op.Proof.LeafHash, kvhash)
	}

	return [][]byte{
		op.Proof.ComputeRootHash(),
	}, nil
}

func (op SimpleValueOp) GetKey() []byte {
	return op.key
}
