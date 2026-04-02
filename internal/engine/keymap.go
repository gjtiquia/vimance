package engine

import "unicode/utf8"

// maxKeymapDepth limits recursive nmap expansion to avoid infinite loops.
const maxKeymapDepth = 50

// KeymapEntry is the RHS of a mapping plus whether expansion is recursive (nmap vs nnoremap).
type KeymapEntry struct {
	RHS       []string
	Recursive bool
}

// KeymapTable maps LHS key sequences to remappings (vim :nmap-style).
type KeymapTable struct {
	trie Trie
}

// Set registers or overwrites a mapping at lhs.
func (t *KeymapTable) Set(lhs []string, rhs []string, recursive bool) {
	if len(lhs) == 0 {
		return
	}
	t.trie.Insert(lhs, KeymapEntry{RHS: rhs, Recursive: recursive})
}

// Lookup classifies lhs against registered mappings.
func (t *KeymapTable) Lookup(keys []string) (MatchResult, *KeymapEntry) {
	mr, v := t.trie.Match(keys)
	if mr != MatchExact {
		return mr, nil
	}
	e, ok := v.(KeymapEntry)
	if !ok {
		return MatchNone, nil
	}
	return MatchExact, &e
}

// Delete removes a mapping at lhs. Returns false if no mapping existed.
func (t *KeymapTable) Delete(lhs []string) bool {
	return t.trie.Delete(lhs)
}

// ParseKeys parses a mapping string into key tokens.
// Single runes become one token each. Sequences like <Escape> or <Ctrl+r> become one token (contents inside brackets).
func ParseKeys(s string) []string {
	if s == "" {
		return nil
	}
	var out []string
	for i := 0; i < len(s); {
		if s[i] == '<' {
			rest := s[i+1:]
			j := 0
			for j < len(rest) && rest[j] != '>' {
				j++
			}
			if j >= len(rest) || rest[j] != '>' {
				// unclosed '<' — literal '<'
				out = append(out, "<")
				i++
				continue
			}
			out = append(out, rest[:j])
			i += 1 + j + 1
			continue
		}
		r, sz := utf8.DecodeRuneInString(s[i:])
		if sz == 0 {
			break
		}
		out = append(out, string(r))
		i += sz
	}
	return out
}
