package merkle

// This is a package for the go implementation of merkle patricia trie
import (
	"fmt"
	"merkle/proto"
	"strings"
	"utils"

	"github.com/gogo/protobuf/proto"
)

type PatriciaTrie struct {
	pb.Tree
}

func NewPatriciaTrie() *PatriciaTrie {
	ht := make(map[string]*pb.Node)
	rt := &pb.Node{
		EncodedPaths: make(map[string]string),
	}
	rt.Hash = "0"
	ht[rt.Hash] = rt
	t := &PatriciaTrie{}
	t.Root = rt
	t.Ht = ht
	return t
}

func (t *PatriciaTrie) Count() int {
	return len(t.Ht)
}

// Print will output the trie in dfs order to stdout for debugging
func (t *PatriciaTrie) Print() {
	t.printNode(t.Root.Hash, t.Ht, 0)
}

// Serialize converts the trie to a byte array
func (t *PatriciaTrie) Serialize() ([]byte, error) {
	return proto.Marshal(&t.Tree)
}

// Deserialize converts the byte arry back to a trie
func (t *PatriciaTrie) Deserialize(bs []byte) error {
	var nodes pb.Tree
	if err := proto.Unmarshal(bs, &nodes); err != nil {
		return err
	}

	t.Tree = nodes
	return nil
}

// Upsert will update or add a kv pair to the trie
func (t *PatriciaTrie) Upsert(key, val string) error {
	if t.Compressed {
		return fmt.Errorf("Cannot update/insert to a compressed trie")
	}

	_, err := t.upsertWithPath(t.Root, utils.ToNibbles(key), val, false, true)
	return err
}

// Delete will delete a value and update/delete its corresponding branch
func (t *PatriciaTrie) Delete(key string) error {
	if t.Compressed {
		return fmt.Errorf("Cannot delete from a compressed trie")
	}

	_, err := t.upsertWithPath(t.Root, utils.ToNibbles(key), "", false, true)
	return err
}

// Compress with fold all the nodes with Count==1 into one encoded path to save space and reduce search time
func (t *PatriciaTrie) Compress() {
	t.foldNode(t.Root, ' ')
}

func (t *PatriciaTrie) foldNode(n *pb.Node, nibble byte) ([]byte, string) {
	seq, target := []byte{}, n.Hash
	for i, nextHash := range n.Next {
		if len(nextHash) > 0 {
			next, _ := t.Ht[nextHash]
			seq, target = t.foldNode(next, byte(i))
			if len(seq) > 0 && len(next.Val) == 0 {
				n.EncodedPaths[string(seq)] = target
				n.Next[i] = ""
			}
		}
	}
	if n.Count > 1 {
		return []byte{nibble}, n.Hash
	}

	if n.Count == 1 && len(n.Val) == 0 && t.Root != n {
		delete(t.Ht, n.Hash)
	}

	return append([]byte{nibble}, seq...), target
}

// Get returns the value to the key. No duplicate is allowed.
// rtype - string, error
func (t *PatriciaTrie) Get(key string) (string, bool) {
	rst, err := t.getWithPath(t.Root, utils.ToNibbles(key), false)
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
	var ok bool
	var nextHash string
	for ep, nh := range n.EncodedPaths {
		// Get using the shortcurt
		if strings.HasPrefix(string(path), ep) {
			ok = true
			path = path[len(ep):]
			nextHash = nh
			if len(ep)%2 == 0 {
				odd = !odd
			}
			break
		}
	}
	if !ok {
		// get the first nibble
		nextHash = n.Next[path[0]]
		path = path[1:]
	}
	next, ok := t.Ht[nextHash]
	if len(nextHash) == 0 || !ok {
		return "", fmt.Errorf("No child at node %s for the nibble %d", n.Hash, nibble)
	}

	return t.getWithPath(next, path, !odd)
}

// upsertPath adds or updates a path of bytes as a branch to the current node
func (t *PatriciaTrie) upsertWithPath(n *pb.Node, path []byte, val string, odd, isRoot bool) (string, error) {
	if len(path) == 0 {
		n.Val = val
	} else {
		// put the hash of next node to the next node
		// build or rebuild the branch
		var next *pb.Node
		if len(n.Next) <= int(path[0]) || len(n.Next[int(path[0])]) == 0 {
			next = &pb.Node{
				EncodedPaths: make(map[string]string),
			}
		} else {
			next, _ = t.Ht[n.Next[path[0]]]
		}

		nextHash, err := t.upsertWithPath(next, path[1:], val, !odd, false)
		if err != nil {
			return "", err
		}

		t.updateChild(n, int(path[0]), nextHash)
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
		delete(t.Ht, prev)
	}

	if len(newHash) > 0 || isRoot {
		t.Ht[newHash] = n
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
		for p, ns := range n.EncodedPaths {
			fmt.Printf("[Debug]encoded_path=%v\n", utils.ToInts([]byte(p)))
			t.printNode(ns, ht, lvl+1)
		}
	}
}
