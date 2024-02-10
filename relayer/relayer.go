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
	"fmt"
	"time"

	"github.com/AnomalyFi/hypersdk/consts"

	"github.com/AnomalyFi/hypersdk/rpc"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/crypto/bls"
	pvm "github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/avalanchego/vms/platformvm/warp"
)

type BlockWarp struct {
	PackedValidatorBytes []byte
	WarpMsg              *warp.Message
	PHeight              uint64
	SubnetWeight         uint64
}

type Exe struct {
	BlockCommitHashes map[uint64]*BlockWarp // this seems not necessary to store, as we are going to prove one block at a time.
	PHeight           uint64
	NodeId            ids.NodeID
	NextProposers     []ids.NodeID
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

func (e *Exe) UpdateBlockHeight(pCli pvm.Client) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C // wait for 1 sec to pass.
		pHeight, err := pCli.GetHeight(context.Background())
		if err == nil {
			e.PHeight = pHeight // update if we succesfully fetched pHeight
		}
	}
}

func (e *Exe) FetchAndExecute(ctx context.Context, pcli pvm.Client, rcli *rpc.JSONRPCClient, id ids.ID, subnetID string, pHeight uint64, height uint64) (*warp.Message, uint64 /*weight*/, uint64 /*signed weight*/, error) {
	var warpMessage *warp.Message
	var subnetWt, sigWt uint64
	for ctx.Err() == nil {
		warpMsg, subnetWeight, sigWeight, err := rcli.GenerateAggregateWarpSignature(ctx, id)
		if float64(sigWeight) >= float64(subnetWeight)*signature_threshold && err == nil {
			warpMessage = warpMsg
			subnetWt = subnetWeight
			sigWt = sigWeight
			break
		}
	}
	validators, err := GetCanonicalValidatorSet(ctx, pcli, subnetID, pHeight)
	if err != nil {
		return nil, 0, 0, err
	}

	validatorDataBytes := make([]byte, len(validators)*(publicKeyBytes+consts.Uint64Len))
	for _, validator := range validators {
		nVdrDataBytes := PackValidatorsData(validatorDataBytes, validator.PublicKey, validator.Weight)
		validatorDataBytes = append(validatorDataBytes, nVdrDataBytes...)
	}
	// store all the gathered values & run a process to clear & prove them
	fmt.Printf("subnet weight: %v, sig weight: %d \nwarp message: %08b", subnetWt, sigWt, warpMessage)
	e.BlockCommitHashes[height] = &BlockWarp{PackedValidatorBytes: validatorDataBytes, WarpMsg: warpMessage, PHeight: pHeight, SubnetWeight: subnetWt}
	//@todo we are left with execute
	return warpMessage, subnetWt, sigWt, nil
}

func (e *Exe) Realyer(scli *rpc.WebSocketClient, rcli *rpc.JSONRPCClient, pcli pvm.Client, subnetID string) error {

	err := scli.RegisterBlockCommitHash() // instead we can also listen to blocks
	if err != nil {
		return err
	}
	go e.UpdateBlockHeight(pcli) // async, want full func to run, lol
	ctx := context.Background()
	for ctx.Err() == nil {
		// @todo add debug logs
		height, pHeight, hash, id, err := scli.ListenBlockCommitHash(ctx)
		if err != nil {
			return err
		}
		fmt.Printf("Received Block Commit Hash of block height: %d, hash: %s", height, hash)
		// @todo make decentralisation part here -> need to submit x blocks at time y, when z is satisfied.
		// for every 10 sec, lets run a func to know when we are going to be the next windower. -> do this in main func -> async
		if e.IsProposer() {
			go e.FetchAndExecute(ctx, pcli, rcli, id, subnetID, pHeight, height)
		}
	}
	return nil
}
