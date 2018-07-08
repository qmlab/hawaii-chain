package merkle

// This is a Go thread-safe implementation of the merkle patricia trie
// It features automatic batch compression of nodes into encoded paths to improve insert/update/delete efficiency
// The design goal SLA is to support over 5k inserts/sec and 200k gets/sec
import (
	"fmt"
	"merkle/proto"
	"strings"
	"sync"
	"utils"

	"github.com/gogo/protobuf/proto"
)

type PatriciaTrie struct {
	pb.Tree
	sync.RWMutex
}

func NewPatriciaTrie() *PatriciaTrie {
	t := &PatriciaTrie{}
	t.Root = &pb.Node{
		EncodedPaths: make(map[string]string), // shortcut paths
		Hash:         "0",                     // arbitrary hash
	}
	t.Ht = make(map[string]*pb.Node)
	t.Ht["0"] = t.Root
	t.BatchSize = 1000
	return t
}

func (t *PatriciaTrie) Count() int {
	t.RLock()
	defer t.RUnlock()
	return len(t.Ht)
}

// Print will output the trie in dfs order to stdout for debugging
func (t *PatriciaTrie) Print() {
	t.RLock()
	defer t.RUnlock()
	t.printNode(t.Root.Hash, t.Ht, 0)
}

// Serialize converts the trie to a byte array
func (t *PatriciaTrie) Serialize() ([]byte, error) {
	t.RLock()
	defer t.RUnlock()
	return proto.Marshal(&t.Tree)
}

// Deserialize converts the byte arry back to a trie
func (t *PatriciaTrie) Deserialize(bs []byte) error {
	t.Lock()
	defer t.Unlock()
	var nodes pb.Tree
	if err := proto.Unmarshal(bs, &nodes); err != nil {
		return err
	}

	t.Tree = nodes
	return nil
}

// Upsert will update or add a kv pair to the trie
func (t *PatriciaTrie) Upsert(key, val string) error {
	t.Lock()
	defer t.Unlock()
	_, err := t.upsertWithPath(t.Root, utils.ToNibbles(key), val, true)
	if int64(len(t.Ht))-t.LastCompression > t.BatchSize {
		t.compress()
	}
	return err
}

// Delete will delete a value and update/delete its corresponding branch
func (t *PatriciaTrie) Delete(key string) error {
	t.Lock()
	defer t.Unlock()
	_, err := t.upsertWithPath(t.Root, utils.ToNibbles(key), "", true)
	return err
}

// compress with fold all the nodes with Count==1 into one encoded path to save space and reduce search time
func (t *PatriciaTrie) compress() {
	t.foldNode(t.Root, ' ')
	t.LastCompression = int64(len(t.Ht))
}

// Get returns the value to the key. No duplicate is allowed.
// rtype - string, error
func (t *PatriciaTrie) Get(key string) (string, bool) {
	t.RLock()
	defer t.RUnlock()
	rst, err := t.getWithPath(t.Root, utils.ToNibbles(key))
	return rst, err == nil
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

func (t *PatriciaTrie) getWithPath(n *pb.Node, path []byte) (string, error) {
	if len(path) == 0 {
		if len(n.Val) == 0 {
			return "", fmt.Errorf("No value found")
		}

		return n.Val, nil
	}

	// get the encoded path if any
	var nibble byte
	_, nextHash, path, ok := getEncodedPath(n, path)
	if !ok {
		// get the child if exsits
		if hasChild(n, path[0]) {
			nextHash = n.Next[path[0]]
			path = path[1:]
		}
	}

	next, ok := t.Ht[nextHash]
	if len(nextHash) == 0 || !ok {
		return "", fmt.Errorf("No child at node %s for the nibble %d", n.Hash, nibble)
	}

	return t.getWithPath(next, path)
}

// getEncodedPath returns:
// encoded_path
// encoded_path_val
// new_path
// is_ok
func getEncodedPath(n *pb.Node, path []byte) (string, string, []byte, bool) {
	for ep, nh := range n.EncodedPaths {
		// Get using the shortcurt
		if strings.HasPrefix(string(path), ep) {
			return ep, nh, path[len(ep):], true
		}
	}

	return "", "", path, false
}

// upsertPath adds or updates a path of bytes as a branch to the current node
func (t *PatriciaTrie) upsertWithPath(n *pb.Node, path []byte, val string, isRoot bool) (string, error) {
	if len(path) == 0 {
		n.Val = val
	} else {
		// If node is contained in encoded path, go along the path
		// Note: no partial update of an encoded path is currently supported
		var next *pb.Node
		var newPath []byte
		epKey, epHash, newPath, isEncodedPath := getEncodedPath(n, path)
		if isEncodedPath {
			next, _ = t.Ht[epHash]
		} else {
			// Put the hash of next node to the next node
			// Build or rebuild the branch
			if !hasChild(n, path[0]) {
				next = &pb.Node{
					EncodedPaths: make(map[string]string),
				}
			} else {
				next, _ = t.Ht[n.Next[path[0]]]
			}

			newPath = path[1:]
		}

		nextHash, err := t.upsertWithPath(next, newPath, val, false)
		if err != nil {
			return "", err
		}

		if isEncodedPath {
			updateEncodedPath(n, epKey, nextHash)
		} else {
			updateChild(n, int(path[0]), nextHash)
		}
	}

	// finally, update the hash on the current node
	return t.updateHash(n, isRoot)
}

func hasChild(n *pb.Node, b byte) bool {
	return len(n.Next) > int(b) && len(n.Next[int(b)]) > 0
}

func updateChild(n *pb.Node, key int, val string) {
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

func updateEncodedPath(n *pb.Node, key, val string) {
	if len(val) == 0 {
		n.Count--
		delete(n.EncodedPaths, key)
		return
	}

	n.EncodedPaths[key] = val
}

func (t *PatriciaTrie) updateHash(n *pb.Node, isRoot bool) (string, error) {
	if n == nil {
		return "", nil
	}

	prev := n.Hash
	// If n.Count == 0, newHash will be empty
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
