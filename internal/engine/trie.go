package engine

// MatchResult classifies how a key sequence matches a trie.
type MatchResult int

const (
	MatchNone MatchResult = iota
	MatchPrefix
	MatchExact
)

// TrieNode is a node in a prefix tree keyed by key strings (e.g. "g", "ArrowLeft").
type TrieNode struct {
	children map[string]*TrieNode
	value    any
}

// Trie supports multi-key sequence registration and prefix/exact matching.
type Trie struct {
	root *TrieNode
}

func (t *Trie) insertNode(keys []string) *TrieNode {
	if t.root == nil {
		t.root = &TrieNode{}
	}
	node := t.root
	for _, k := range keys {
		if node.children == nil {
			node.children = make(map[string]*TrieNode)
		}
		next, ok := node.children[k]
		if !ok {
			next = &TrieNode{}
			node.children[k] = next
		}
		node = next
	}
	return node
}

// Insert associates a value with the terminal key sequence.
func (t *Trie) Insert(keys []string, value any) {
	node := t.insertNode(keys)
	node.value = value
}

// Match walks keys; returns Exact with value at a terminal, Prefix if more keys are needed,
// or None if the path does not exist.
func (t *Trie) Match(keys []string) (MatchResult, any) {
	if len(keys) == 0 {
		return MatchNone, nil
	}
	if t.root == nil {
		return MatchNone, nil
	}
	node := t.root
	for _, k := range keys {
		if node.children == nil {
			return MatchNone, nil
		}
		next, ok := node.children[k]
		if !ok {
			return MatchNone, nil
		}
		node = next
	}
	if node.value != nil {
		return MatchExact, node.value
	}
	if len(node.children) > 0 {
		return MatchPrefix, nil
	}
	return MatchNone, nil
}
