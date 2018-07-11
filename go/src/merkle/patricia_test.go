package merkle

import (
	"testing"
	"utils"

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

func TestUpdateWithoutCompress(t *testing.T) {
	trie := NewPatriciaTrie()
	trie.Upsert("key1", "val1")
	trie.Upsert("key1", "val2")
	rst, _ := trie.Get("key1")
	assert.Equal(t, "val2", rst)
}

func TestUpdateWithCompress(t *testing.T) {
	trie := NewPatriciaTrie()
	trie.Upsert("key1", "val1")
	trie.compress()
	trie.Upsert("key1", "val2")
	rst, _ := trie.Get("key1")
	assert.Equal(t, "val2", rst)
}

func TestDeleteWithoutcompress(t *testing.T) {
	trie := NewPatriciaTrie()
	trie.Upsert("key1", "val1")
	trie.Upsert("key2", "val2")
	trie.Delete("key1")
	rst, _ := trie.Get("key2")
	assert.Equal(t, "val2", rst)
	rst, ok := trie.Get("key1")
	assert.False(t, ok)
}

func TestDeleteWithcompress(t *testing.T) {
	trie := NewPatriciaTrie()
	trie.Upsert("key1", "val1")
	trie.Upsert("key2", "val2")
	trie.compress()
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
	trie.compress()
	rst, _ := trie.Get("key1")
	assert.Equal(t, "val1", rst)
}

func TestCompress2(t *testing.T) {
	trie := NewPatriciaTrie()
	trie.Upsert("ka1", "val1")
	trie.Upsert("ka3", "val3")
	trie.compress()
	rst, _ := trie.Get("ka1")
	assert.Equal(t, "val1", rst)
	rst, _ = trie.Get("ka3")
	assert.Equal(t, "val3", rst)
}

func BenchmarkUpsert1000(b *testing.B) {
	trie := NewPatriciaTrie()
	trie.BatchSize = 1000
	updateAll(trie, b)
}

func BenchmarkUpsertAndGet1000(b *testing.B) {
	trie := NewPatriciaTrie()
	trie.BatchSize = 1000
	keys := updateAll(trie, b)
	getAll(trie, keys)
}

func BenchmarkUpsert2000(b *testing.B) {
	trie := NewPatriciaTrie()
	trie.BatchSize = 2000
	updateAll(trie, b)
}

func BenchmarkUpsertAndGet2000(b *testing.B) {
	trie := NewPatriciaTrie()
	trie.BatchSize = 2000
	keys := updateAll(trie, b)
	getAll(trie, keys)
}

func BenchmarkUpsert4000(b *testing.B) {
	trie := NewPatriciaTrie()
	trie.BatchSize = 4000
	updateAll(trie, b)
}

func BenchmarkUpsertAndGet4000(b *testing.B) {
	trie := NewPatriciaTrie()
	trie.BatchSize = 4000
	keys := updateAll(trie, b)
	getAll(trie, keys)
}

func BenchmarkUpsert8000(b *testing.B) {
	trie := NewPatriciaTrie()
	trie.BatchSize = 8000
	updateAll(trie, b)
}

func BenchmarkUpsertAndGet8000(b *testing.B) {
	trie := NewPatriciaTrie()
	trie.BatchSize = 8000
	keys := updateAll(trie, b)
	getAll(trie, keys)
}

func updateAll(trie *PatriciaTrie, b *testing.B) []string {
	var keys []string
	for i := 0; i < b.N; i++ {
		k, v := utils.RandStringBytesMaskImprSrc(32), utils.RandStringBytesMaskImprSrc(32)
		trie.Upsert(k, v)
		keys = append(keys, k)
	}

	return keys
}

func getAll(trie *PatriciaTrie, keys []string) {
	for _, k := range keys {
		trie.Get(k)
	}
}
