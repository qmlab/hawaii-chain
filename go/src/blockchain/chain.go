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
	bc.Blocks = append(bc.Blocks, &pb.Block{Index: 0})
	return bc
}

// AddTransaction creates a new transactions and put it into open tx list
func (bc *BlockChain) AddTransaction(r string, v float32) {
	// Validate the sender account value
	// Create transaction into open transaction
	tx := pb.Transaction{
		Sender:    bc.Usr.Addr,
		Recipient: r,
		Val:       v,
	}

	bc.OpenTxs = append(bc.OpenTxs, &tx)
}

// MineBlock adds open transactions to the blockchain after validation
func (bc *BlockChain) MineBlock() {
	lastBlock := bc.Blocks[len(bc.Blocks)-1]
	lastProof := POW(lastBlock.Proof, int(bc.Difficulty))
	proof := POW(lastProof, int(bc.Difficulty))
	bc.RWMutex.Lock()
	defer bc.RWMutex.Unlock()
	bc.AddNewBlock(proof, lastBlock)
	bc.OpenTxs = []*pb.Transaction{}
}

func (bc *BlockChain) AddNewTransaction(sender, recipient string, val float32) {
	tx := &pb.Transaction{
		Sender:    sender,
		Recipient: recipient,
		Val:       val,
		Timestamp: time.Now().UnixNano(),
	}

	tx.Id, _ = utils.Hash(tx)
	bc.RWMutex.Lock()
	defer bc.RWMutex.Unlock()
	bc.OpenTxs = append(bc.OpenTxs, tx)
}

func (bc *BlockChain) AddNewBlock(proof int64, lastBlock *pb.Block) {
	block := &pb.Block{
		Index:     lastBlock.Index + 1,
		Timestamp: time.Now().UnixNano(),
		Proof:     proof,
		PrevHash:  lastBlock.Hash,
	}

	txs := merkle.NewPatriciaTrie()
	for _, tx := range bc.OpenTxs {
		if data, err := proto.Marshal(tx); err != nil {
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
func POW(lastProof int64, difficulty int) int64 {
	var proof int64 = 0
	for !IsValidProof(lastProof, proof, difficulty) {
		proof++
	}

	return proof
}

func IsValidProof(lastProof, proof int64, difficulty int) bool {
	guess := fmt.Sprintf("%d%d", lastProof, proof)
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
