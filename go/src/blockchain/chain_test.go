package blockchain

import (
	"math/rand"
	"testing"
	"utils"

	"github.com/stretchr/testify/assert"
)

func TestAddTransaction(t *testing.T) {
	bc := NewBlockChain()
	bc.AddTransaction("receiverhash", 100.0)
	assert.Equal(t, "receiverhash", bc.OpenTxs[0].Recipient)
	assert.Equal(t, 100.0, bc.OpenTxs[0].Val)
}

func TestAddBlock(t *testing.T) {
	bc := NewBlockChain()
	bc.addNewBlock(int64(100))
	bc.addNewBlock(int64(101))
	assert.Equal(t, 3, len(bc.Blocks))
}

func TestMineBlockAndGet(t *testing.T) {
	bc := NewBlockChain()
	id := bc.AddTransaction("receiverhash", 100.0)
	bc.MineBlock()
	tx := bc.GetTransaction(id)
	assert.NotNil(t, tx)
	assert.Equal(t, 100.0, tx.Val)
	assert.Equal(t, 0, len(bc.OpenTxs))
}

func TestInitial(t *testing.T) {
	bc := NewBlockChain()
	bal := bc.GetBalance("00000000000000000000000000000000")
	assert.Equal(t, 100.0, bal)
}

func TestBalance(t *testing.T) {
	bc := NewBlockChain()
	id1 := bc.AddTransaction("receiverhash", 50.0)
	id2 := bc.AddTransaction("00000000000000000000000000000001", 50.0)
	id3 := bc.AddTransaction("00000000000000000000000000000002", 50.0)
	bc.MineBlock()
	tx1, tx2, tx3 := bc.GetTransaction(id1), bc.GetTransaction(id2), bc.GetTransaction(id3)
	assert.Equal(t, "complete", tx1.Status)
	assert.Equal(t, "complete", tx2.Status)
	assert.Equal(t, "failed", tx3.Status)
	bal1, bal2, bal3 := bc.GetBalance("receiverhash"), bc.GetBalance("00000000000000000000000000000001"), bc.GetBalance("00000000000000000000000000000000")
	assert.Equal(t, 50.0, bal1)
	assert.Equal(t, 150.0, bal2)
	assert.Equal(t, 0.0, bal3)
}

func BenchmarkMining1(b *testing.B) {
	mining(1, b)
}

func BenchmarkMining2(b *testing.B) {
	mining(2, b)
}

func BenchmarkMining3(b *testing.B) {
	mining(3, b)
}

func mining(difficulty int32, b *testing.B) {
	bc := NewBlockChain()
	for i := 0; i < b.N; i++ {
		bc.AddTransaction(utils.RandStringBytesMaskImprSrc(32), rand.Float64())
	}
	bc.Difficulty = difficulty
	bc.MineBlock()
}
