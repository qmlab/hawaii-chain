package blockchain

import (
	"math/rand"
	"testing"
	"utils"

	"github.com/stretchr/testify/assert"
)

func TestAddTransaction(t *testing.T) {
	bc := NewBlockChain("ownerhash")
	bc.AddTransaction("receiverhash", 100.0)
	assert.Equal(t, "receiverhash", bc.OpenTxs[0].Recipient)
	assert.Equal(t, float32(100.0), bc.OpenTxs[0].Val)
}

func TestAddBlock(t *testing.T) {
	bc := NewBlockChain("ownerhash")
	bc.addNewBlock(int64(100))
	bc.addNewBlock(int64(101))
	assert.Equal(t, 3, len(bc.Blocks))
}

func TestMineBlockAndGet(t *testing.T) {
	bc := NewBlockChain("ownerhash")
	id := bc.AddTransaction("receiverhash", 100.0)
	bc.Difficulty = 1
	bc.MineBlock()
	tx := bc.GetTransaction(id)
	assert.NotNil(t, tx)
	assert.Equal(t, float32(100.0), tx.Val)
	assert.Equal(t, 0, len(bc.OpenTxs))
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
	bc := NewBlockChain("ownerhash")
	for i := 0; i < b.N; i++ {
		bc.AddTransaction(utils.RandStringBytesMaskImprSrc(32), rand.Float32())
	}
	bc.Difficulty = difficulty
	bc.MineBlock()
}
