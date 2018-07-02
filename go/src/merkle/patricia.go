package merkle

// This is a package for the go implementation of merkle patricia trie
import (
	"fmt"
	"merkle/proto"
	"utils"
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
func (t *PatriciaTrie) Get(key string) (string, error) {
	return t.getWithPath(t.Root, []byte(key), false)
}

func (t *PatriciaTrie) getWithPath(n *pb.Node, path []byte, odd bool) (string, error) {
	if len(path) == 0 {
		if len(n.Val) == 0 {
			return "", fmt.Errorf("No value found")
		}

		return n.Val, nil
	}

	// get the first nimble
	nimble, path := t.firstNimble(path, odd)
	nextHash := n.Next[nimble]
	next, ok := t.ht[nextHash]
	if len(nextHash) == 0 || !ok {
		return "", fmt.Errorf("No child at node %s for the nimble %d", n.Hash, nimble)
	}

	return t.getWithPath(next, path, !odd)
}

func (t *PatriciaTrie) firstNimble(path []byte, odd bool) (int, []byte) {
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
func (t *PatriciaTrie) upsertWithPath(n *pb.Node, path []byte, val string, odd, isRoot bool) (string, error) {
	if len(path) == 0 {
		n.Val = val
	} else {
		// get the first nimble
		nimble, path := t.firstNimble(path, odd)

		// put the hash of next node to the next node
		// build or rebuild the branch
		var next *pb.Node
		if len(n.Next) <= nimble || len(n.Next[nimble]) == 0 {
			next = &pb.Node{}
		} else {
			next, _ = t.ht[n.Next[nimble]]
		}

		nextHash, err := t.upsertWithPath(next, path, val, !odd, false)
		if err != nil {
			return "", err
		}

		t.updateChild(n, nimble, nextHash)
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
		for lvl > 0 {
			bs = append(bs, '-')
		}

		fmt.Printf("%x%v\n", bs, *n)
		for _, ns := range n.Next {
			if len(ns) > 0 {
				t.printNode(ns, ht, lvl+1)
			}
		}
	}
}
