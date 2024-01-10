package types

import (
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/stretchr/testify/require"
)

// TODO rewrite this with new base cases
func removeWhitespace(s string) string {
	// Split the string on whitespace then concatenate the segments
	return strings.Join(strings.Fields(s), "")
}

var ReferenceNmtRoot NmtRoot = NmtRoot{
	Root: []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

var ReferenceL1BLockInfo L1BlockInfo = L1BlockInfo{
	Number:    123,
	Timestamp: *NewU256().SetUint64(0x456),
	Hash:      common.Hash{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
}

var ReferenceHeader Header = Header{
	Height:           42,
	Timestamp:        789,
	L1Head:           124,
	TransactionsRoot: ReferenceNmtRoot,
}

func TestNodeKitTypesHeaderCommit(t *testing.T) {
	require.Equal(t, ReferenceHeader.Commit(), Commitment{69, 70, 204, 173, 194, 81, 14, 28, 209, 104, 204, 53, 32, 43, 75, 233, 35, 99, 95, 128, 155, 14, 46, 17, 217, 191, 252, 217, 29, 252, 131, 61})
}

func TestNodeKitCommitmentFromU256TrailingZero(t *testing.T) {
	comm := Commitment{209, 146, 197, 195, 145, 148, 17, 211, 52, 72, 28, 120, 88, 182, 204, 206, 77, 36, 56, 35, 3, 143, 77, 186, 69, 233, 104, 30, 90, 105, 48, 0}
	roundTrip, err := CommitmentFromUint256(comm.Uint256())
	require.Nil(t, err)
	require.Equal(t, comm, roundTrip)
}

func TestNodeKitCommitmentConversion(t *testing.T) {

	root := [32]byte{87, 222, 83, 212, 108, 217, 91, 95, 235, 22, 31, 150, 95, 39, 142, 246, 208, 183, 178, 231, 123, 55, 241, 11, 165, 78, 221, 36, 158, 232, 198, 65}

	cb58, err := EncodeCB58(root[:])

	require.Nil(t, err)

	bytes, err := DecodeCB58(cb58)

	require.Nil(t, err)

	thirtytwo := [32]byte{}

	for i, b := range bytes {
		// Bytes() returns the bytes in big endian order and we need a 32 byte array so we populate it this way
		thirtytwo[i] = b
	}

	doot := NewU256().SetBytes(thirtytwo)
	bigRoot := doot.Int

	comm_old := Commitment(root)
	//comm := Commitment{209, 146, 197, 195, 145, 148, 17, 211, 52, 72, 28, 120, 88, 182, 204, 206, 77, 36, 56, 35, 3, 143, 77, 186, 69, 233, 104, 30, 90, 105, 48, 0}
	u := NewU256().SetBigInt(&bigRoot)

	comm, err := CommitmentFromUint256(NewU256().SetBigInt(&bigRoot))

	roundTrip, err := CommitmentFromUint256(u)
	require.Nil(t, err)
	require.Equal(t, comm, comm_old)
	require.Equal(t, roundTrip, comm_old)

}

func TestNodeKitCommitmentNewConversion(t *testing.T) {

	comm_expected := Commitment{193, 98, 70, 80, 45, 4, 82, 113, 146, 158, 194, 61, 72, 64, 34, 217, 173, 46, 78, 63, 115, 159, 115, 122, 219, 58, 120, 223, 9, 52, 140, 166}

	root := [32]byte{88, 48, 50, 99, 140, 172, 117, 35, 116, 212, 26, 123, 187, 199, 189, 130, 55, 219, 55, 144, 86, 21, 30, 68, 214, 253, 157, 141, 160, 54, 5, 190}

	h := &Header{
		Height:    2539,
		Timestamp: 1703696824,
		L1Head:    252,
		TransactionsRoot: NmtRoot{
			Root: root[:],
		},
	}

	comm := h.Commit()

	require.Equal(t, comm, comm_expected)

}

func TestNodeKitCommitmentNewNewConversion(t *testing.T) {

	//Comm from OP Stack
	comm_expected := Commitment{112, 195, 203, 219, 135, 105, 81, 102, 195, 168, 68, 105, 33, 129, 251, 200, 219, 30, 22, 109, 233, 35, 152, 109, 26, 136, 17, 90, 246, 245, 172, 87}

	//Comm from Sequencer Contract
	//comm_expected := Commitment{203, 180, 93, 217, 192, 136, 98, 9, 32, 234, 226, 223, 187, 69, 158, 245, 152, 97, 108, 214, 208, 208, 2, 92, 232, 127, 198, 198, 114, 219, 153, 155}
	//Comm from OP Stack
	//
	root := [32]byte{173, 78, 146, 230, 209, 66, 80, 196, 46, 3, 57, 192, 66, 66, 230, 93, 204, 16, 129, 181, 116, 128, 123, 25, 41, 155, 143, 4, 200, 171, 161, 107}

	//Height:650 Timestamp:1703721377

	//TimestampOriginal:1703721377845

	tmstp := int64(1703721377845)
	// tmp := tmstp / 1000

	// t.Log(tmp)

	// tFun := tmp * 1000

	h := &Header{
		Height:    650,
		Timestamp: uint64(tmstp),
		L1Head:    89,
		TransactionsRoot: NmtRoot{
			Root: root[:],
		},
	}

	comm := h.Commit()

	require.Equal(t, comm, comm_expected)

}

func TestNodeKitCommitmentNewNewNewConversion(t *testing.T) {

	comm_expected := Commitment{204, 96, 250, 106, 187, 117, 152, 27, 92, 38, 7, 163, 113, 9, 19, 47, 172, 99, 150, 21, 125, 167, 34, 75, 69, 80, 60, 226, 80, 111, 73, 74}

	//2023/12/28 10:54:53 expected header header &{100 1703782468909 40 {[200 9 194 53 133 197 167 132 254 135 120 53 97 3 141 0 104 181 162 47 15 195 115 230 41 221 209 68 39 233 119 222]}}
	//2023/12/28 10:54:53 expected comm comm {false [4994558929463626058 12421937216264544843 6640003099161858863 14727046117820045339]}
	//2023/12/28 10:54:53 expected comm in bytes comm [204 96 250 106 187 117 152 27 92 38 7 163 113 9 19 47 172 99 150 21 125 167 34 75 69 80 60 226 80 111 73 74]

	//{ 1703782468909 40 100}

	bytesDude, _ := DecodeCB58("2X6huav6bqDktn5LR8TPQXPveKKLBdAjngbGpvMnjsC1icRmGi")

	newRoot, _ := ToHash256(bytesDude)

	t.Log(newRoot)

	root := [32]byte{200, 9, 194, 53, 133, 197, 167, 132, 254, 135, 120, 53, 97, 3, 141, 0, 104, 181, 162, 47, 15, 195, 115, 230, 41, 221, 209, 68, 39, 233, 119, 222}

	t.Log(root)

	testStr, _ := EncodeCB58(root[:])

	t.Log(testStr)

	//newRoot, _ := ids.FromString("2GQrkp2R78AGQRPNue6oFBD4L2rBbWs7GgbFTgJ6CTrdAKJmKH")

	require.Equal(t, newRoot, root)

	h := &Header{
		Height:    100,
		Timestamp: 1703782468909,
		L1Head:    40,
		TransactionsRoot: NmtRoot{
			Root: newRoot[:],
		},
	}

	comm := h.Commit()

	require.Equal(t, comm, comm_expected)

}
