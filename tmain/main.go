package main

import (
	"context"
	"log"
	"os"

	"github.com/AnomalyFi/nodekit-l1/relayer"
	"github.com/ava-labs/avalanchego/api/info"
	"github.com/ava-labs/avalanchego/ids"
)

// http://127.0.0.1:37607/ext/bc/xQuCRPCvd9CJ7ZuAFEc9ni54m3HKP4mJUo3v6xEd2AZ8ofTMf --> url format <endpoint>/ext/bc/<bc id>/<custom point>
func main() {
	url := "http://127.0.0.1:41471/ext/bc/dkf1PFrc1qXDjFVG3s6XNJPWxaC2WskYL2Fa5UYG4GiVuVBDv"
	subnetId := "BW33BFF4BVx8TXZaEtUk4iYJYa8QvMuYcsxDje32FUTNxWsXc"
	nodeUrl := "http://127.0.0.1:41471"

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
	log.Println("configuring initial data...")
	e := relayer.Exe{NodeId: nodeID, NextProposers: map[uint64]*[]ids.NodeID{}, BlockCommitHashes: map[uint64]*relayer.BlockWarp{}}
	pHeight, err := pcli.GetHeight(context.Background())
	if err != nil {
		log.Println(err)
		os.Exit(4)
	}
	_, height, _, err := rcli.Accepted(context.Background())
	if err != nil {
		log.Println(err)
		os.Exit(4)
	}

	e.PHeight = pHeight
	e.Height = height

	log.Println("Gathering initial orhestrators...")
	height -= height % 12
	pHeight -= pHeight % 12
	proposers, err := rcli.GetOrchestrator(context.Background(), pHeight, height)
	if err != nil {
		log.Println(err)
		os.Exit(4)
	}
	e.NextProposers[height] = proposers

	log.Println("starting all process...")
	go e.UpdateProposers(rcli)
	go e.UpdatePBlockHeight(pcli)
	go e.UpdateBlockHeight(rcli)
	if err := e.Realyer(scli, rcli, pcli, subnetId); err != nil {
		log.Println(err)
		os.Exit(3)
	}
}
