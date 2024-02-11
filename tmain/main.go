package main

import (
	"context"
	"log"
	"os"

	"github.com/AnomalyFi/nodekit-l1/relayer"
	"github.com/ava-labs/avalanchego/api/info"
)

// http://127.0.0.1:37607/ext/bc/xQuCRPCvd9CJ7ZuAFEc9ni54m3HKP4mJUo3v6xEd2AZ8ofTMf --> url format <endpoint>/ext/bc/<bc id>/<custom point>
func main() {
	url := "http://127.0.0.1:33943/ext/bc/Th3z6g4cRC7DdakFZtpfbH6xPRdVCyhcxKJLT5zsu1AQymKwn"
	subnetId := "BW33BFF4BVx8TXZaEtUk4iYJYa8QvMuYcsxDje32FUTNxWsXc"
	nodeUrl := "http://127.0.0.1:33943"

	scli, rcli, pcli, err := relayer.Initialize(url, nodeUrl)
	if err != nil {
		log.Println(err)
		os.Exit(2)
	}
	icli := info.NewClient(nodeUrl) // Get node id
	nodeID, _, err := icli.GetNodeID(context.Background())
	if err != nil {
		log.Println(err)
		os.Exit(4)
	}

	e := relayer.Exe{NodeId: nodeID}
	// update next orchestrator window here -> async
	go e.UpdateProposers(rcli)
	go e.UpdateBlockHeight(pcli)
	if err := e.Realyer(scli, rcli, pcli, subnetId); err != nil {
		log.Println(err)
		os.Exit(3)
	}
}
