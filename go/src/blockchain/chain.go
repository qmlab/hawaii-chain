package blockchain

import (
	"fmt"
	"sync"
)

// BlockChain is the chain object that includes a genesis block and all subsequent blocks
type BlockChain struct {
	Bs   []Block
	Otxs []Transaction

	UserAddr string

	m *sync.RWMutex
}

// NewBlockChain creates a new blockchain object
// addr is the user's addr
func NewBlockChain(addr string) *BlockChain {
	bc := &BlockChain{
		m:        &sync.RWMutex{},
		UserAddr: addr,
	}

	bc.Bs = append(bc.Bs, Block{Index: 0})
	return bc
}

// AddTransaction creates a new transactions and put it into open tx list
func (bc *BlockChain) AddTransaction(r string, v float64) {
	// Validate the sender account value
	// Create transaction into open transaction
	tx := Transaction{
		Sender:    bc.UserAddr,
		Recipient: r,
		Val:       v,
	}

	bc.Otxs = append(bc.Otxs, tx)
}

// MineBlock adds open transactions to the blockchain after validation
func (bc *BlockChain) MineBlock() {
	// TODO
}

// PrintBlockChain prints out the entire chain for debugging purpose
func (bc *BlockChain) PrintBlockChain() {
	fmt.Printf("[Debug]%v\n", bc.Bs)
}
