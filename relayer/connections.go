package relayer

import (
	"github.com/AnomalyFi/hypersdk/pubsub"
	"github.com/AnomalyFi/hypersdk/rpc"
	pvm "github.com/ava-labs/avalanchego/vms/platformvm"
)

func Initialize(url string, nodeURL string) (*rpc.WebSocketClient, *rpc.JSONRPCClient, pvm.Client, error) {
	scli, err := rpc.NewWebSocketClient(url, rpc.DefaultHandshakeTimeout, pubsub.MaxPendingMessages, pubsub.MaxReadMessageSize)
	if err != nil {
		return nil, nil, nil, err
	}
	rcli := rpc.NewJSONRPCClient(url)
	pcli := pvm.NewClient(nodeURL)
	return scli, rcli, pcli, nil
}
