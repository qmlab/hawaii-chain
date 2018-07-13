package blockchain

import (
	"config"
	"fmt"
	"merkle"
	"proto"
	"sync"
	"time"
	"utils"

	"github.com/gogo/protobuf/proto"
)

// BlockChain is the chain object that includes a genesis block and all subsequent blocks
type BlockChain struct {
	sync.RWMutex
	pb.Chain
}

// NewBlockChain creates a new blockchain object
// addr is the user's addr
func NewBlockChain() *BlockChain {
	config.InitConfig("../config/config.json")
	bc := &BlockChain{}
	bc.Usr = &pb.User{Addr: config.Usrcfg.Address}
	bc.Difficulty = 1
	// Genesis is for initial starting block including starting balances
	genesis := initBlock()
	bc.Blocks = append(bc.Blocks, genesis)
	return bc
}

func initBlock() *pb.Block {
	genesis := &pb.Block{
		Index: 0,
		Proof: 0,
	}
	status := merkle.NewPatriciaTrie()
	for _, acc := range config.InitialAccounts {
		status.Upsert(acc.Address, fmt.Sprintf("%f", acc.Val))
	}
	genesis.Balances = &status.Tree
	return genesis
}

// MineBlock adds open transactions to the blockchain after validation
func (bc *BlockChain) MineBlock() {
	proof := pow(bc.Blocks[len(bc.Blocks)-1].Proof, int(bc.Difficulty))
	bc.RWMutex.Lock()
	defer bc.RWMutex.Unlock()
	bc.addNewBlock(proof)
	bc.OpenTxs = []*pb.Transaction{}
}

// AddTransaction creates a new transaction and add it to the open Txs list
func (bc *BlockChain) AddTransaction(recipient string, val float64) string {
	tx := &pb.Transaction{
		Sender:    bc.Usr.Addr,
		Recipient: recipient,
		Val:       val,
		Timestamp: time.Now().UnixNano(),
		Status:    "pending",
	}

	tx.Id, _ = utils.Hash(tx)
	bc.RWMutex.Lock()
	defer bc.RWMutex.Unlock()
	bc.OpenTxs = append(bc.OpenTxs, tx)
	return tx.Id
}

// GetTransaction retrieves the transaction from the merkle trie
func (bc *BlockChain) GetTransaction(id string) *pb.Transaction {
	for _, block := range bc.Blocks {
		if block.Txs != nil {
			t := merkle.NewPatriciaTrie()
			t.Tree = *block.Txs
			if v, ok := t.Get(id); ok {
				var tx pb.Transaction
				err := proto.Unmarshal([]byte(v), &tx)
				if err == nil {
					return &tx
				}
			}
		}
	}

	return nil
}

// GetBalance retrieves the transaction from the merkle trie
func (bc *BlockChain) GetBalance(acc string) float64 {
	for i := len(bc.Blocks) - 1; i >= 0; i-- {
		block := bc.Blocks[i]
		if block.Balances != nil {
			t := merkle.NewPatriciaTrie()
			t.Tree = *block.Balances
			if v, ok := t.GetFloat(acc); ok {
				return v
			}
		}
	}

	return 0.0
}

func (bc *BlockChain) addNewBlock(proof int64) {
	lastBlock := bc.Blocks[len(bc.Blocks)-1]
	block := &pb.Block{
		Index:     lastBlock.Index + 1,
		Timestamp: time.Now().UnixNano(),
		Proof:     proof,
		PrevHash:  lastBlock.Hash,
	}

	txs := merkle.NewPatriciaTrie()
	balances := merkle.NewPatriciaTrie()
	bals := make(map[string]float64) // cache the balances to memory
	for _, tx := range bc.OpenTxs {
		var sbal, rbal float64
		var ok bool
		if sbal, ok = bals[tx.Sender]; !ok {
			sbal = bc.GetBalance(tx.Sender)
		}
		if rbal, ok = bals[tx.Recipient]; !ok {
			rbal = bc.GetBalance(tx.Recipient)
		}
		if sbal >= tx.Val {
			tx.Status = "complete"
			if data, err := proto.Marshal(tx); err == nil {
				txs.Upsert(tx.Id, string(data))
				bals[tx.Sender] = sbal - tx.Val
				bals[tx.Recipient] = rbal + tx.Val
			}
		} else {
			tx.Status = "failed"
			if data, err := proto.Marshal(tx); err == nil {
				txs.Upsert(tx.Id, string(data))
			}
		}
	}
	for id, val := range bals {
		balances.UpsertFloat(id, val)
	}

	block.Txs = &txs.Tree
	block.Balances = &balances.Tree
	block.Hash = utils.HashBlock(block)
	bc.Blocks = append(bc.Blocks, block)
}

// Validation of sha256(last_proof+proof) has the first N bytes as '0's
func pow(lastProof int64, difficulty int) int64 {
	var proof int64 = 0
	for !IsValidProof(lastProof, proof, difficulty) {
		proof++
	}

	return proof
}

func IsValidProof(lastProof, proof int64, difficulty int) bool {
	guess := fmt.Sprintf("%d%d%d", lastProof, proof, time.Now().UnixNano())
	hash := utils.HashBytes([]byte(guess))
	for i := 0; i < difficulty; i++ {
		if hash[i] != '0' {
			return false
		}
	}

	return true
}

// PrintBlockChain prints out the entire chain for debugging purpose
func (bc *BlockChain) PrintBlockChain() {
	fmt.Printf("[Debug]%v\n", bc.Blocks)
}
