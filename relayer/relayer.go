package relayer

// Path:
//
// 		subscribe to listen for every block, to get commited block root by every validator.
//		natively verify: (4/5)*validator_stake has commited for the block root.
//			enough signed weight is not available, then query blockchain for roots.
//			if a commited block root is not received, after 5 blocks query blockchain & peers for the block root
// 		concurrently call getWarpSignatures from hypersdk client & build aggregate signature from available signatures.
//		if enough weight of signatures are recieved, send to l1_relayer.
//
//		L1_relayer, batches block commitment hash roots. say 12 block commitment hash root
//			& submits to novanet zkWasm engine
//			& obtains proof of execution(of aggSignature verification of 12 blocks).
//			Submits proof to L1 for settlement.
//
//		Every client subscribes to L1_relayer for settlement confirmations.
//		client can also run their own L1_relayer or submit proof individually to their L1_contract
//
//		Do not Over Engineer. Engineer as required. Optimize performance for god-level.
//
// 		every SEQ block has an L1BlockHead attatched.-> it is only for tracking L1. -> this should not have any collusion for our relayer(any)
//

// Current Path: Single implementation: listen to all warp signatures

// build for hypersdk current & change later for anomalyFi
import (
	"context"
	"encoding/binary"
	"log"

	"github.com/AnomalyFi/hypersdk/consts"

	"github.com/AnomalyFi/hypersdk/rpc"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/crypto/bls"
	"github.com/ava-labs/avalanchego/utils/set"
	pvm "github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/avalanchego/vms/platformvm/warp"
)

type BlockWarp struct {
	PackedValidatorBytes []byte
	WarpMsg              *warp.Message
	BitSet               []byte
	PHeight              uint64
	SubnetWeight         uint64
}

type Exe struct {
	BlockCommitHashes map[uint64]*BlockWarp // this may be unnecessary -> no, necessary if a zk verify fails
	PHeight           uint64
	Height            uint64
	NodeId            ids.NodeID
	NextProposers     map[uint64]*[]ids.NodeID
}

func PackValidatorsData(initBytes []byte, PublicKey *bls.PublicKey, weight uint64) []byte {
	pbKeyBytes := bls.PublicKeyToBytes(PublicKey)
	return append(initBytes, binary.BigEndian.AppendUint64(pbKeyBytes, weight)...)
}

func GetCanonicalValidatorSet(ctx context.Context, pcli pvm.Client, subnetIDS string, pHeight uint64) ([]*warp.Validator, error) {
	subnetID, _ := ids.FromString(subnetIDS)
	vdrSet, err := pcli.GetValidatorsAt(ctx, subnetID, pHeight)
	if err != nil {
		return nil, err
	}
	validators /*totalWeight*/, _, err := rpc.GetCanonicalValidatorSet(context.Background(), vdrSet)
	if err != nil {
		return nil, err
	}
	return validators, nil
}

func filterValidators(bitBytes []byte, validators []*warp.Validator) ([]*warp.Validator, error) {
	bitSet := set.BitsFromBytes(bitBytes)
	filteredVdrs := make([]*warp.Validator, 0, len(validators))
	for i, vdr := range validators {
		if !bitSet.Contains(i) {
			continue
		}

		filteredVdrs = append(filteredVdrs, vdr)
	}
	return filteredVdrs, nil
}

func (e *Exe) FetchAndExecute(ctx context.Context, pcli pvm.Client, rcli *rpc.JSONRPCClient, id ids.ID, subnetID string, pHeight uint64, height uint64) error {
	var warpMessage *warp.Message
	var subnetWt, sigWt uint64
	var bitSet []byte
	for ctx.Err() == nil {
		warpMsg, subnetWeight, sigWeight, bSet, err := rcli.GenerateAggregateWarpSignature(ctx, id)
		if float64(sigWeight) >= float64(subnetWeight)*signature_threshold && err == nil {
			warpMessage = warpMsg
			subnetWt = subnetWeight
			sigWt = sigWeight
			bitSet = bSet
			break
		}
	}

	validators, err := GetCanonicalValidatorSet(ctx, pcli, subnetID, pHeight)
	if err != nil {
		log.Println(err)
		return err
	}

	validatorDataBytes := make([]byte, len(validators)*(publicKeyBytes+consts.Uint64Len))
	for _, validator := range validators {
		nVdrDataBytes := PackValidatorsData(validatorDataBytes, validator.PublicKey, validator.Weight)
		validatorDataBytes = nVdrDataBytes
	}
	// store all the gathered values & run a process to clear & prove them
	log.Printf("blockHeight: %d, subnet weight: %d, sig weight: %d \n", height, subnetWt, sigWt)
	e.BlockCommitHashes[height] = &BlockWarp{PackedValidatorBytes: validatorDataBytes, WarpMsg: warpMessage, BitSet: bitSet, PHeight: pHeight, SubnetWeight: subnetWt}

	//@todo process for submission
	// case - 1: pass bitset, packedValidatorBytes, message + subnetWeight
	// case - 2: pass bitset, packedValidatorBytes, bitSortedPackedValidatorBytes + subnetWeight
	filterValidators(bitSet, validators)
	// BLSAggregateSig: gives success is all signatures are valid. So, let relayer validated signatures here first & check the weight.
	// if sig comes invalid, remove the signature from aggSignature & create a new aggSig, get WEight & then pass to novanet
	// if weight is not as sufficient, ask network to sign the block again
	// send to novanet prover
	// submit to L1
	return nil
}

func (e *Exe) Realyer(scli *rpc.WebSocketClient, rcli *rpc.JSONRPCClient, pcli pvm.Client, subnetID string) error {

	err := scli.RegisterBlockCommitHash()
	if err != nil {
		return err
	}
	ctx := context.Background()
	log.Println("listening to block commit hash...")
	for ctx.Err() == nil {
		height, pHeight, _, id, err := scli.ListenBlockCommitHash(ctx)
		if err != nil {
			return err
		}
		log.Printf("Received Block Commit Hash of block height: %d", height)
		// decentralising orchestrator.
		if e.IsProposer(height, rcli) {
			log.Printf("fetching signatures and executing for block height: %d", height)
			go e.FetchAndExecute(ctx, pcli, rcli, id, subnetID, pHeight, height)
		}
	}
	return nil
}
