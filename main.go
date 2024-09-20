package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"

	"flag"

	"github.com/peterbourgon/ff/v3"

	"github.com/AnomalyFi/hypersdk/pubsub"
	hrpc "github.com/AnomalyFi/hypersdk/rpc"

	"github.com/AnomalyFi/nodekit-l1/types"
	"github.com/AnomalyFi/nodekit-seq/rpc"
	"github.com/AnomalyFi/nodekit-seq/sequencer"

	"github.com/ava-labs/avalanchego/ids"

	ethbind "github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/ethereum/go-ethereum/crypto"
)

// Environment variables beginning with this prefix can be used to instantiate command line flags
const ENV_PREFIX = "NODEKIT_COMMITMENT"

// TODO fix this
var (
	fs             = flag.NewFlagSet("commitment", flag.ContinueOnError)
	sequencerAddr  = fs.String("seq-addr", "http://3.215.71.153:9650/ext/bc/LGbUJJgok9VA4gKSZ2Fw7XA42jtsXub6xj2kBG4mPJipmeG5z", "address of NodeKit SEQ")
	gethAddr       = fs.String("geth-addr", "https://devnet.nodekit.xyz", "address of the geth node")
	contractAddr   = fs.String("contract-addr", "0x5FbDB2315678afecb367f032d93F642f64180aa3", "address of the sequencer contract")
	contractWallet = fs.String("contract-wallet", "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80", "owner of the sequencer contract")
	vm_id          = fs.Int64("chain-id-l1", 32382, "Chain ID of L1")
	chain_id       = fs.String("chain-id", "LGbUJJgok9VA4gKSZ2Fw7XA42jtsXub6xj2kBG4mPJipmeG5z", "Chain ID of SEQ instance")
)

func main() {
	if err := ff.Parse(fs, os.Args[1:], ff.WithEnvVarPrefix(ENV_PREFIX)); err != nil {
		panic(err)
	}

	log.Println("Starting commitment manager")
	log.Printf("sequencerAddr: %s\n", *sequencerAddr)
	log.Printf("gethAddr: %s\n", *gethAddr)
	log.Printf("contractAddr: %s\n", *contractAddr)
	log.Printf("contractWallet: %s\n", *contractWallet)
	log.Printf("vm_id: %d\n", *vm_id)
	log.Printf("chain_id: %s\n", *chain_id)

	conn, err := ethclient.Dial(*gethAddr)
	if err != nil {
		panic(fmt.Errorf("unable to connect to eth l1: %v", err))
	}

	sequencerContractTest, err := sequencer.NewSequencer(ethcommon.HexToAddress(*contractAddr), conn)
	if err != nil {
		panic(fmt.Errorf("sequencer contract error: %v", err))
	}

	// function call on `instance`. Retrieves pending name
	maxBlocks, err := sequencerContractTest.MAXBLOCKS(&ethbind.CallOpts{Pending: true})
	if err != nil {
		log.Fatalf("Failed to retrieve max blocks: %v", err)
	}
	fmt.Println("max blocks:", maxBlocks)

	// logfile, err := os.Create("app.log")

	// if err != nil {
	// 	log.Fatal(err)
	// }

	// defer logfile.Close()
	// log.SetOutput(logfile)

	Commit(maxBlocks, sequencerContractTest, conn)

}

func Commit(maxBlocks *big.Int, seq *sequencer.Sequencer, client *ethclient.Client) {
	for {
		if err := SyncRequest(maxBlocks, seq, client); err != nil {
			fmt.Printf("Failed to Sync %v\n", err)

			// Wait to avoid spam
			time.Sleep(1 * time.Second)
		}
	}

}

func SyncRequest(
	maxBlocks *big.Int,
	seq *sequencer.Sequencer,
	client *ethclient.Client,
) error {

	// url := "http://3.215.71.153:9650/ext/bc/LGbUJJgok9VA4gKSZ2Fw7XA42jtsXub6xj2kBG4mPJipmeG5z"
	cli := hrpc.NewJSONRPCClient(*sequencerAddr)

	//chain_id := "LGbUJJgok9VA4gKSZ2Fw7XA42jtsXub6xj2kBG4mPJipmeG5z"

	chainId, err := ids.FromString(*chain_id)
	if err != nil {
		return err
	}

	tcli := rpc.NewJSONRPCClient(*sequencerAddr, 1337, chainId)

	//Height, BlockID, Timestamp

	//TODO may need to change
	priv, err := crypto.HexToECDSA(*contractWallet)
	if err != nil {
		log.Fatalf("Failed to convert from hex to ECDSA: %v", err)
		return err
	}

	auth, err := ethbind.NewKeyedTransactorWithChainID(priv, big.NewInt(*vm_id))
	if err != nil {
		log.Fatalf("Failed to create authorized transactor: %v", err)
		return err
	}

	//TODO! Just changed this because it was potentially causing issues on OP Stack
	//auth.GasLimit = 1_000_000
	auth.GasLimit = 600_000

	contract_block_height, err := seq.BlockHeight(nil)
	if err != nil {
		log.Fatalf("Failed to retrieve max blocks: %v", err)
		return err
	}

	currentHeight := uint64(0)

	for {
		_, height, _, err := cli.Accepted(context.TODO())
		blkHeightBig := big.NewInt(int64(height))

		if err != nil {
			log.Fatalf("Failed to retrieve last accepted block: %v", err)
		}

		if contract_block_height.Cmp(blkHeightBig) > 0 {
			fmt.Printf("Sequencer Contract at height %d, waiting for it to pass height %d\n", contract_block_height.Int64(), height)
			err := WaitForSEQ(contract_block_height.Int64(), *sequencerAddr, cli, tcli)
			if err != nil {
				log.Fatal(err)
				break
			}
		} else {
			// We are ahead of the contract, sequence the blocks which are currently ready.
			fmt.Printf("synchronizing blocks %d-%d\n", contract_block_height.Int64(), height)
			currentHeight = height
			break
		}
	}

	fmt.Println("Current Block Height:", contract_block_height)
	// fmt.Printf("Update SEQ Height: %x\n", blk.Hght)

	firstBlock := uint64(contract_block_height.Int64())

	max := 20

	res, err := tcli.GetCommitmentBlocks(context.TODO(), firstBlock, currentHeight, max)

	if err != nil {
		log.Fatalf("Failed to retrieve commitment blocks: %v", err)
		return err
	}

	fmt.Println("Created Block Batch")

	blocks := make([]sequencer.SequencerWarpBlock, 0)

	if len(res.Blocks) < max {
		max = len(res.Blocks)
	}

	for i := 0; i < max; i++ {
		//TODO Need to modify the BlockRoot so it is a commitment instead matching OP Stack. The block root itself should be a commitment to match with NMT Root

		id, _ := types.DecodeCB58(res.Blocks[i].BlockID)

		blocks = append(blocks, sequencer.SequencerWarpBlock{
			Height:     res.Blocks[i].Height,
			BlockRoot:  res.Blocks[i].BlockRoot,
			ParentRoot: res.Blocks[i].ParentRoot,
		})

		log.Println("received comm", "comm", res.Blocks[i].BlockRoot)

		header := &types.Header{
			Height:    res.Blocks[i].Height.Uint64(),
			Timestamp: uint64(res.Blocks[i].Timestamp),
			L1Head:    res.Blocks[i].L1Head,
			TransactionsRoot: types.NmtRoot{
				Root: id,
			},
		}

		log.Println("expected header", "header", header)

		comm := header.Commit()

		log.Println("expected comm", "comm", comm.Uint256().Int)

		log.Println("expected comm in bytes", "comm", comm)

		if comm.Uint256().Int.Cmp(res.Blocks[i].BlockRoot) == 0 {
			log.Println("Comms Match")

		}

	}

	tx, err := seq.NewBlocks(auth, blocks)
	if err != nil {
		fmt.Println("An error happened:", err)
		return err
	}

	//ch := make(chan *ethtypes.Transaction)
	for i := 1; i < 10; i++ {
		tx, pending, _ := client.TransactionByHash(context.TODO(), tx.Hash())
		if !pending {
			fmt.Printf("Update Sucessful: 0x%x\n", tx.Hash())

			receipt, err := client.TransactionReceipt(context.TODO(), tx.Hash())
			if err != nil {
				return err
			}
			//fmt.Println(receipt)
			if receipt.Status != uint64(1) {
				fmt.Println("Status")
				fmt.Println(receipt.Status)
				fmt.Println("Gas")
				fmt.Println(receipt.GasUsed)

				return fmt.Errorf("Tx Failed")
			}
			fmt.Println("Gas")
			fmt.Println(receipt.GasUsed)
			test := int64(2)
			t := big.NewInt(test)
			if receipt.BlockNumber.Cmp(t) <= 0 {
				return fmt.Errorf("RPC Failed")
			}
			break
		}

		time.Sleep(time.Second * 1)
	}

	return nil
}

func WaitForSEQ(height int64, uri string, cli *hrpc.JSONRPCClient, tcli *rpc.JSONRPCClient) error {

	fmt.Println("waiting on SEQ")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	parser, err := tcli.Parser(ctx)

	if err != nil {
		return err
	}

	if err != nil {
		return err
	}
	scli, err := hrpc.NewWebSocketClient(uri, hrpc.DefaultHandshakeTimeout, pubsub.MaxPendingMessages, pubsub.MaxReadMessageSize) // we write the max read
	defer scli.Close()
	if err := scli.RegisterBlocks(); err != nil {
		return err
	}

	for ctx.Err() == nil {
		blk, _, _, _, err := scli.ListenBlock(ctx, parser)
		if err != nil {
			fmt.Println("An error happened:", err)
			return err
		}
		if blk.Hght == uint64(height) {
			return nil
		}
	}

	return nil

}
