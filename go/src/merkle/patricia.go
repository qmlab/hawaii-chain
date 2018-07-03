package merkle

// This is a package for the go implementation of merkle patricia trie
import (
	"fmt"
	"merkle/proto"
	"utils"

	"github.com/gogo/protobuf/proto"
)

type PatriciaTrie struct {
	Root *pb.Node
	ht   map[string]*pb.Node // hash table of hash->node
}

func NewPatriciaTrie() *PatriciaTrie {
	ht := make(map[string]*pb.Node)
	rt := &pb.Node{}
	rt.Hash = "0"
	ht[rt.Hash] = rt
	return &PatriciaTrie{rt, ht}
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
	_, err := t.upsertWithPath(t.Root, []byte(key), val, false, true)
	return err
}

// Delete will delete a value and update/delete its corresponding branch
func (t *PatriciaTrie) Delete(key string) error {
	_, err := t.upsertWithPath(t.Root, []byte(key), "", false, true)
	return err
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

	// get the first nibble
	nibble, path := t.firstNibble(path, odd)
	nextHash := n.Next[nibble]
	next, ok := t.ht[nextHash]
	if len(nextHash) == 0 || !ok {
		return "", fmt.Errorf("No child at node %s for the nibble %d", n.Hash, nibble)
	}

	return t.getWithPath(next, path, !odd)
}

func (t *PatriciaTrie) firstNibble(path []byte, odd bool) (int, []byte) {
	var b byte
	if odd {
		b = path[0] % 16
		path = path[1:]
	} else {
		b = path[0] / 16
		path[0] = path[0] % 16
	}

	return int(b), path
}

// upsertPath adds or updates a path of bytes as a branch to the current node
// returns:
// 1.Hash of the current node. "" means this node is no longer needed and it will be removed by its parent.
// 2.Hex array for the encoded path. LIFO.
// 3.NextHash - for building the shortcut encoded path.
// 4.Error
func (t *PatriciaTrie) upsertWithPath(n *pb.Node, path []byte, val string, odd, isRoot bool) (string, []byte, string, error) {
	if len(path) == 0 {
		n.Val = val
	} else {
		// get the first nibble
		nibble, path := t.firstNibble(path, odd)

		// put the hash of next node to the next node
		// build or rebuild the branch
		var next *pb.Node
		if len(n.Next) <= nibble || len(n.Next[nibble]) == 0 {
			next = &pb.Node{}
		} else {
			next, _ = t.ht[n.Next[nibble]]
		}

		nextHash, bs, target, err := t.upsertWithPath(next, path, val, !odd, false)
		if err != nil {
			return "", nil, "", err
		}

		t.updateChild(n, nibble, nextHash)
	}
	// finally, update the hash on the current node
	newHash, err := t.updateHash(n, isRoot)
	if err != nil {
		return "", nil, "", err
	}

	if len(newHash) == 0 {

	}
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
