package types

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type Header struct {
	Height           uint64  `json:"height"`
	Timestamp        uint64  `json:"timestamp"`
	L1Head           uint64  `json:"l1_head"`
	TransactionsRoot NmtRoot `json:"transactions_root"`
}

func (h *Header) UnmarshalJSON(b []byte) error {
	type Dec struct {
		Height           *uint64  `json:"height"`
		Timestamp        *uint64  `json:"timestamp"`
		L1Head           *uint64  `json:"l1_head"`
		TransactionsRoot *NmtRoot `json:"transactions_root"`
	}

	var dec Dec
	if err := json.Unmarshal(b, &dec); err != nil {
		return err
	}

	if dec.Height == nil {
		return fmt.Errorf("Field height of type Header is required")
	}
	h.Height = *dec.Height

	if dec.Timestamp == nil {
		return fmt.Errorf("Field timestamp of type Header is required")
	}
	h.Timestamp = *dec.Timestamp

	if dec.L1Head == nil {
		return fmt.Errorf("Field l1_head of type Header is required")
	}
	h.L1Head = *dec.L1Head

	if dec.TransactionsRoot == nil {
		return fmt.Errorf("Field transactions_root of type Header is required")
	}
	h.TransactionsRoot = *dec.TransactionsRoot

	return nil
}

func (self *Header) Commit() Commitment {
	return NewRawCommitmentBuilder("BLOCK").
		Uint64Field("height", self.Height).
		Uint64Field("timestamp", self.Timestamp).
		Uint64Field("l1_head", self.L1Head).
		Field("transactions_root", self.TransactionsRoot.Commit()).
		Finalize()
}

// func (self *Header) Commit() Commitment {
// 	return Commitment(self.TransactionsRoot.Root)
// }

type L1BlockInfo struct {
	Number    uint64      `json:"number"`
	Timestamp U256        `json:"timestamp"`
	Hash      common.Hash `json:"hash"`
}

func (i *L1BlockInfo) UnmarshalJSON(b []byte) error {
	// Parse using pointers so we can distinguish between missing and default fields.
	type Dec struct {
		Number    *uint64      `json:"number"`
		Timestamp *U256        `json:"timestamp"`
		Hash      *common.Hash `json:"hash"`
	}

	var dec Dec
	if err := json.Unmarshal(b, &dec); err != nil {
		return err
	}

	if dec.Number == nil {
		return fmt.Errorf("Field number of type L1BlockInfo is required")
	}
	i.Number = *dec.Number

	if dec.Timestamp == nil {
		return fmt.Errorf("Field timestamp of type L1BlockInfo is required")
	}
	i.Timestamp = *dec.Timestamp

	if dec.Hash == nil {
		return fmt.Errorf("Field hash of type L1BlockInfo is required")
	}
	i.Hash = *dec.Hash

	return nil
}

func (self *L1BlockInfo) Commit() Commitment {
	return NewRawCommitmentBuilder("L1BLOCK").
		Uint64Field("number", self.Number).
		Uint256Field("timestamp", &self.Timestamp).
		FixedSizeField("hash", self.Hash[:]).
		Finalize()
}

type NmtRoot struct {
	Root Bytes `json:"root"`
}

func (r *NmtRoot) UnmarshalJSON(b []byte) error {
	// Parse using pointers so we can distinguish between missing and default fields.
	type Dec struct {
		Root *Bytes `json:"root"`
	}

	var dec Dec
	if err := json.Unmarshal(b, &dec); err != nil {
		return err
	}

	if dec.Root == nil {
		return fmt.Errorf("Field root of type NmtRoot is required")
	}
	r.Root = *dec.Root

	return nil
}

func (self *NmtRoot) Commit() Commitment {
	return NewRawCommitmentBuilder("NMTROOT").
		VarSizeField("root", self.Root).
		Finalize()
}

type BatchMerkleProof = Bytes

// A bytes type which serializes to JSON as an array, rather than a base64 string. This ensures
// compatibility with the NodeKit APIs.
type Bytes []byte

// // TODO do I want this or can I use a base64?
// func (b Bytes) MarshalJSON() ([]byte, error) {
// 	// Convert to `int` array, which serializes the way we want.
// 	ints := make([]int, len(b))
// 	for i := range b {
// 		ints[i] = int(b[i])
// 	}

// 	return json.Marshal(ints)
// }

// func (b *Bytes) UnmarshalJSON(in []byte) error {
// 	// Parse as `int` array, which deserializes the way we want.
// 	var ints []int
// 	if err := json.Unmarshal(in, &ints); err != nil {
// 		return err
// 	}

// 	// Convert back to `byte` array.
// 	*b = make([]byte, len(ints))
// 	for i := range ints {
// 		if ints[i] < 0 || 255 < ints[i] {
// 			return fmt.Errorf("byte out of range: %d", ints[i])
// 		}
// 		(*b)[i] = byte(ints[i])
// 	}

// 	return nil
// }

// A BigInt type which serializes to JSON a a hex string. This ensures compatibility with the
// NodeKit APIs.
type U256 struct {
	big.Int
}

func NewU256() *U256 {
	return new(U256)
}

func (i *U256) SetBigInt(n *big.Int) *U256 {
	i.Int.Set(n)
	return i
}

func (i *U256) SetUint64(n uint64) *U256 {
	i.Int.SetUint64(n)
	return i
}

func (i *U256) SetBytes(buf [32]byte) *U256 {
	i.Int.SetBytes(buf[:])
	return i
}

func (i U256) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmt.Sprintf("0x%s", i.Text(16)))
}

func (i *U256) UnmarshalJSON(in []byte) error {
	var s string
	if err := json.Unmarshal(in, &s); err != nil {
		return err
	}
	if _, err := fmt.Sscanf(s, "0x%x", &i.Int); err != nil {
		return err
	}
	return nil
}
