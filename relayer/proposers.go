package relayer

import (
	"context"
	"time"

	"github.com/AnomalyFi/hypersdk/rpc"
	pvm "github.com/ava-labs/avalanchego/vms/platformvm"
)

func (e *Exe) UpdatePBlockHeight(pCli pvm.Client) {
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

func (e *Exe) UpdateBlockHeight(rcli *rpc.JSONRPCClient) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		_, height, _, err := rcli.Accepted(context.Background())
		if err == nil {
			e.Height = height // update P Height
		}
	}
}
func (e *Exe) UpdateProposers(rcli *rpc.JSONRPCClient) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	ctx := context.Background()
	for {
		<-ticker.C // wait for 9 secs to pass.
		//@todo rules for block height & verification
		height := e.Height - e.Height%12
		pHeight := e.PHeight - e.PHeight%12
		proposrs, err := rcli.GetOrchestrator(ctx, pHeight, height)
		if err == nil {
			e.NextProposers[height] = proposrs
		}
	}
}

func (e *Exe) IsProposer(height uint64, rcli *rpc.JSONRPCClient) bool { // delay in block commit hash received & we are matching against wrong proposer window
	modHeight := height - height%12
	_, exists := e.NextProposers[modHeight]
	if !exists {
		e.ProposerUpdate(modHeight, rcli)
	}
	len := uint64(len(*e.NextProposers[modHeight]))
	for i := uint64(0); i < len; i++ {
		if (*e.NextProposers[modHeight])[i] == e.NodeId {
			return true
		}
	}
	return false
}

func (e *Exe) ProposerUpdate(modHeight uint64, rcli *rpc.JSONRPCClient) {
	pHeight := e.PHeight - e.PHeight%12
	proposrs, err := rcli.GetOrchestrator(context.Background(), pHeight, modHeight)
	if err == nil {
		e.NextProposers[modHeight] = proposrs
	}
}
