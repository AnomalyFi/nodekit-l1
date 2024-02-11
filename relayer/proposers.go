package relayer

import (
	"context"
	"time"

	"github.com/AnomalyFi/hypersdk/rpc"
)

func (e *Exe) UpdateProposers(rcli *rpc.JSONRPCClient) {
	ticker := time.NewTicker(9 * time.Second)
	defer ticker.Stop()
	ctx := context.Background()
	for {
		<-ticker.C // wait for 9 secs to pass.
		proposrs, err := rcli.GetOrchestrator(ctx, e.PHeight, e.PHeight)
		if err == nil {
			e.NextProposers = proposrs
		}
	}
}

func (e *Exe) IsProposer() bool {
	len := uint64(len(*e.NextProposers))
	for i := uint64(0); i < len; i++ {
		if (*e.NextProposers)[i] == e.NodeId {
			return true
		}
	}
	return false
}
