package merkle

// This is a package for the go implementation of merkle patricia trie
import (
	"fmt"
	"merkle/proto"
	"utils"

	"github.com/gogo/protobuf/proto"
)

type PatriciaTrie struct {
	Root       *pb.Node
	Compressed bool                // if compressed with encoded path for memory optimization, cannot update/insert/delete node
	ht         map[string]*pb.Node // hash table of hash->node
}

func NewPatriciaTrie() *PatriciaTrie {
	ht := make(map[string]*pb.Node)
	rt := &pb.Node{}
	rt.Hash = "0"
	ht[rt.Hash] = rt
	return &PatriciaTrie{rt, false, ht}
}

// Print will output the trie in dfs order to stdout for debugging
func (t *PatriciaTrie) Print() {
	t.printNode(t.Root.Hash, t.ht, 0)
}

// Serialize converts the trie to a byte array
func (t *PatriciaTrie) Serialize() ([]byte, error) {
	nodes := &pb.Tree{
		Root: t.Root,
	}
	for _, n := range t.ht {
		nodes.List = append(nodes.List, n)
	}

	return proto.Marshal(nodes)
}

// Deserialize converts the byte arry back to a trie
func (t *PatriciaTrie) Deserialize(bs []byte) error {
	var nodes pb.Tree
	if err := proto.Unmarshal(bs, &nodes); err != nil {
		return err
	}

	t.Root = nodes.Root
	t.ht = make(map[string]*pb.Node)
	for _, n := range nodes.List {
		t.ht[n.Hash] = n
	}

	return nil
}

// Upsert will update or add a kv pair to the trie
func (t *PatriciaTrie) Upsert(key, val string) error {
	if t.Compressed {
		return fmt.Errorf("Cannot update/insert to a compressed trie")
	}

	_, err := t.upsertWithPath(t.Root, []byte(key), val, false, true)
	return err
}

// Delete will delete a value and update/delete its corresponding branch
func (t *PatriciaTrie) Delete(key string) error {
	if t.Compressed {
		return fmt.Errorf("Cannot delete from a compressed trie")
	}

	_, err := t.upsertWithPath(t.Root, []byte(key), "", false, true)
	return err
}

// Compress with fold all the nodes with Count==1 into one encoded path to save space and reduce search time
func (t *PatriciaTrie) Compress() {
	t.foldNode(t.Root, ' ')
}

func (t *PatriciaTrie) foldNode(n *pb.Node, nibble byte) ([]byte, string) {
	seq, target := []byte{}, ""
	for i, nextHash := range n.Next {
		next, _ := t.ht[nextHash]
		seq, target = t.foldNode(next, byte(i))
		if len(seq) > 0 {
			n.EncodedPaths[t.encodePath(seq)] = target
			n.Next[i] = ""
		}
	}
	if n.Count == 1 && len(n.Val) == 0 && t.Root != n {
		delete(t.ht, n.Hash)
		return append(seq, nibble), target
	}

	return []byte{}, n.Hash
}

func (t *PatriciaTrie) encodePath(bs []byte) string {
	var rst []byte
	var cur int
	for i := 0; i < len(bs); i++ {
		if i%2 == 0 {
			cur = int(bs[i])
		} else {
			cur += int(bs[i]) * 16
			rst = append([]byte{byte(cur)}, rst...)
		}
	}
	if len(bs)%2 == 1 {
		rst = append([]byte{byte(cur)}, rst...)
	}

	return string(rst)
}

// Get returns the value to the key. No duplicate is allowed.
// rtype - string, error
func (t *PatriciaTrie) Get(key string) (string, bool) {
	rst, err := t.getWithPath(t.Root, []byte(key), false)
	return rst, err == nil
}

func (t *PatriciaTrie) getWithPath(n *pb.Node, path []byte, odd bool) (string, error) {
	if len(path) == 0 {
		if len(n.Val) == 0 {
			return "", fmt.Errorf("No value found")
		}

		return n.Val, nil
	}

	// get the encoded path if any
	var nibble byte
	nextHash, ok := n.EncodedPaths[string(path)]
	if ok {
		path = []byte{}
	} else {
		// get the first nibble
		nibble, path = t.firstNibble(path, odd)
		nextHash = n.Next[int(nibble)]
	}

	next, ok := t.ht[nextHash]
	if len(nextHash) == 0 || !ok {
		return "", fmt.Errorf("No child at node %s for the nibble %d", n.Hash, nibble)
	}

	return t.getWithPath(next, path, !odd)
}

func (t *PatriciaTrie) firstNibble(path []byte, odd bool) (byte, []byte) {
	var b byte
	if odd {
		b = path[0] % 16
		path = path[1:]
	} else {
		b = path[0] / 16
		path[0] = path[0] % 16
	}

	return b, path
}

// upsertPath adds or updates a path of bytes as a branch to the current node
func (t *PatriciaTrie) upsertWithPath(n *pb.Node, path []byte, val string, odd, isRoot bool) (string, error) {
	if len(path) == 0 {
		n.Val = val
	} else {
		// get the first nibble
		nibble, path := t.firstNibble(path, odd)

		// put the hash of next node to the next node
		// build or rebuild the branch
		var next *pb.Node
		if len(n.Next) <= int(nibble) || len(n.Next[int(nibble)]) == 0 {
			next = &pb.Node{}
		} else {
			next, _ = t.ht[n.Next[nibble]]
		}

		nextHash, err := t.upsertWithPath(next, path, val, !odd, false)
		if err != nil {
			return "", err
		}

		t.updateChild(n, int(nibble), nextHash)
	}

	// finally, update the hash on the current node
	return t.updateHash(n, isRoot)
}

func (t *PatriciaTrie) updateChild(n *pb.Node, key int, val string) {
	for len(n.Next) <= key {
		n.Next = append(n.Next, "")
	}

	if len(val) == 0 {
		n.Count--
	} else if len(n.Next[key]) == 0 {
		n.Count++
	}

	n.Next[key] = val
}

func (t *PatriciaTrie) updateHash(n *pb.Node, isRoot bool) (string, error) {
	if n == nil {
		return "", nil
	}

	prev := n.Hash
	newHash, err := utils.GetHash(n)
	if err != nil {
		return "", err
	}
	if len(prev) > 0 {
		delete(t.ht, prev)
	}

	if len(newHash) > 0 || isRoot {
		t.ht[newHash] = n
	}

	n.Hash = newHash
	return newHash, nil
}

func (t *PatriciaTrie) printNode(hash string, ht map[string]*pb.Node, lvl int) {
	if n, ok := ht[hash]; ok {
		bs := []byte{}
		for i := lvl; i > 0; i-- {
			bs = append(bs, '-')
		}

		fmt.Printf("[Debug]%s%v\n", string(bs), *n)
		for _, ns := range n.Next {
			if len(ns) > 0 {
				t.printNode(ns, ht, lvl+1)
			}
		}
	}
}
