package merkle

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPatriciaTrie(t *testing.T) {
	trie := NewPatriciaTrie()
	assert.NotNil(t, trie.Root)
	assert.Equal(t, trie.Root.Hash, "0")
}

func TestInsert(t *testing.T) {
	trie := NewPatriciaTrie()
	trie.Upsert("key1", "val1")
	rst, _ := trie.Get("key1")
	assert.Equal(t, "val1", rst)
}

func TestInsert2(t *testing.T) {
	trie := NewPatriciaTrie()
	trie.Upsert("key1", "val1")
	trie.Upsert("key2", "val2")
	trie.Upsert("key3", "val2")

	rst, _ := trie.Get("key1")
	assert.Equal(t, "val1", rst)
	rst, _ = trie.Get("key2")
	assert.Equal(t, "val2", rst)
	rst, _ = trie.Get("key3")
	assert.Equal(t, "val2", rst)
}

func TestUpdate(t *testing.T) {
	trie := NewPatriciaTrie()
	trie.Upsert("key1", "val1")
	trie.Upsert("key1", "val2")
	rst, _ := trie.Get("key1")
	assert.Equal(t, "val2", rst)
}

func TestDelete(t *testing.T) {
	trie := NewPatriciaTrie()
	trie.Upsert("key1", "val1")
	trie.Upsert("key2", "val2")
	trie.Delete("key1")
	rst, _ := trie.Get("key2")
	assert.Equal(t, "val2", rst)
	rst, ok := trie.Get("key1")
	assert.False(t, ok)
}

func TestSerialize(t *testing.T) {
	trie := NewPatriciaTrie()
	trie.Upsert("keys1", "val1")
	trie.Upsert("keys2", "val2")
	bs, err := trie.Serialize()
	assert.Nil(t, err)

	trie2 := NewPatriciaTrie()
	trie2.Deserialize(bs)
	rst, _ := trie2.Get("keys1")
	assert.Equal(t, "val1", rst)
	rst, _ = trie2.Get("keys2")
	assert.Equal(t, "val2", rst)
}

func TestCompress1(t *testing.T) {
	trie := NewPatriciaTrie()
	trie.Upsert("key1", "val1")
	trie.Compress()
	rst, _ := trie.Get("key1")
	assert.Equal(t, "val1", rst)
}

func TestCompress2(t *testing.T) {
	trie := NewPatriciaTrie()
	trie.Upsert("ka1", "val1")
	trie.Upsert("ka3", "val3")
	trie.Compress()
	rst, _ := trie.Get("ka1")
	assert.Equal(t, "val1", rst)
	rst, _ = trie.Get("ka3")
	assert.Equal(t, "val3", rst)
}
