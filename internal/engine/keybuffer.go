package engine

// KeyBuffer accumulates key strings for multi-key sequences (e.g. "g" then "g" for gg).
type KeyBuffer struct {
	keys []string
}

func (kb *KeyBuffer) Append(key string) {
	kb.keys = append(kb.keys, key)
}

func (kb *KeyBuffer) Reset() {
	kb.keys = kb.keys[:0]
}

func (kb *KeyBuffer) Keys() []string {
	return kb.keys
}

func (kb *KeyBuffer) Len() int {
	return len(kb.keys)
}
