package blockchain

import (
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
func NewBlockChain(addr string) *BlockChain {
	bc := &BlockChain{}
	bc.Usr = &pb.User{Addr: addr}
	bc.Blocks = append(bc.Blocks, &pb.Block{Index: 0, Proof: 0})
	return bc
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
func (bc *BlockChain) AddTransaction(recipient string, val float32) string {
	tx := &pb.Transaction{
		Sender:    bc.Usr.Addr,
		Recipient: recipient,
		Val:       val,
		Timestamp: time.Now().UnixNano(),
	}

	tx.Id, _ = utils.Hash(tx)
	bc.RWMutex.Lock()
	defer bc.RWMutex.Unlock()
	bc.OpenTxs = append(bc.OpenTxs, tx)
	return tx.Id
}

// GetTransaction retrieves the transaction from the merkle trie
func (bc *BlockChain) GetTransaction(Id string) *pb.Transaction {
	for _, block := range bc.Blocks {
		if block.Txs != nil {
			t := merkle.NewPatriciaTrie()
			t.Tree = *block.Txs
			if v, ok := t.Get(Id); ok {
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

func (bc *BlockChain) addNewBlock(proof int64) {
	lastBlock := bc.Blocks[len(bc.Blocks)-1]
	block := &pb.Block{
		Index:     lastBlock.Index + 1,
		Timestamp: time.Now().UnixNano(),
		Proof:     proof,
		PrevHash:  lastBlock.Hash,
	}

	txs := merkle.NewPatriciaTrie()
	for _, tx := range bc.OpenTxs {
		if data, err := proto.Marshal(tx); err == nil {
			txs.Upsert(tx.Id, string(data))
		}
	}

	// TODO: add all results
	// TODO: add all status

	block.Txs = &txs.Tree
	block.Hash, _ = utils.Hash(block)
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
