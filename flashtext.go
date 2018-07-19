package flashtext

import (
	"strings"
	"sync"
)

// KeywordProcessor is the processor of keyword extract
type KeywordProcessor struct {
	// dicts store the keyword => cleanName
	dicts map[string]string
	// keytrie is the trie struct
	keytrie *trie
	// caseSensitive or not
	caseSensitive bool
	// input lowerCased or not (only makes a difference when ignoring case on input
	inputLowercased bool
	// noboundaryWords default to a-zA-Z0-9_
	noboundaryWords map[rune]bool
	// lock for the map write
	sync.RWMutex
}

type ExtractResult struct {
	Keyword string
	// StartIndex is the keyword index in the sentences
	StartIndex int
}

type Option struct {
	// Longest set to true will just match the longest keyword,
	Longest bool
}

var (
	defaultOption = &Option{
		Longest: true,
	}
)

func NewKeywordProcessor() *KeywordProcessor {
	p := &KeywordProcessor{
		dicts:           make(map[string]string),
		noboundaryWords: make(map[rune]bool),
		keytrie:         NewTrie('r'),
		caseSensitive:   false,
		inputLowercased: true,
	}
	for i := 0; i < 26; i++ {
		p.AddNoBoundaryWords(rune('a' + i))
		p.AddNoBoundaryWords(rune('A' + i))
	}
	for i := 0; i < 10; i++ {
		p.AddNoBoundaryWords(rune('0' + i))
	}
	p.AddNoBoundaryWords('-')
	return p
}

func (p *KeywordProcessor) SetCaseSensitive(caseSensitive bool) {
	p.caseSensitive = caseSensitive
}

func (p *KeywordProcessor) RemoveNoBoundaryWord(noboundaryWords ...rune) {
	for _, w := range noboundaryWords {
		_, exists := p.noboundaryWords[w]
		if exists {
			delete(p.noboundaryWords, w)
		}
	}
}

func (p *KeywordProcessor) AddNoBoundaryWords(noboundaryWords ...rune) {
	for _, w := range noboundaryWords {
		p.noboundaryWords[w] = true
	}
}

func (p *KeywordProcessor) AddKeywords(keywords ...string) {
	for _, keyword := range keywords {
		p.AddKeyword(keyword, false)
	}
}

func (p *KeywordProcessor) AddKeyword(keyword string, ignoreCase bool) {
	p.Lock()
	defer p.Unlock()

	if !ignoreCase && !p.caseSensitive {
		if p.inputLowercased {
			keyword = strings.ToLower(keyword)
		} else {
			keyword = strings.ToUpper(keyword)
		}
	}
	p.keytrie.addKeyword(keyword)
}

func (p *KeywordProcessor) AddKeywordAndName(keyword string, cleanName string) {
	p.Lock()
	defer p.Unlock()

	if !p.caseSensitive {
		if p.inputLowercased {
			keyword = strings.ToLower(keyword)
		} else {
			keyword = strings.ToUpper(keyword)
		}
	}
	p.keytrie.addKeyword(keyword)
	p.dicts[keyword] = cleanName
}

func (p *KeywordProcessor) ExtractKeywords(sentence string, option ...*Option) (res []*ExtractResult) {
	extractOption := defaultOption
	if len(option) > 0 {
		extractOption = option[0]
	}
	if !p.caseSensitive {
		if p.inputLowercased {
			sentence = strings.ToLower(sentence)
		} else {
			sentence = strings.ToUpper(sentence)
		}
	}
	runes := []rune(sentence)
	size := len(runes)
	idx := 0
	begin := true
	var curTrie *trie
	for idx < size {
		curTrie = p.keytrie
		c := runes[idx]
		if _, ok := p.noboundaryWords[c]; !ok {
			idx++
			begin = true
		} else if !begin {
			idx++
		} else {
			var j = idx
			foundKeyword := ""
			for j = idx; j < size; j++ {
				c = runes[j]
				curTrie = curTrie.next[c]
				if curTrie == nil {
					break
				}
				if curTrie.word != "" && (j == size-1 || !p.noboundaryWords[runes[j+1]]) {
					foundKeyword = curTrie.word
					if !extractOption.Longest {
						res = append(res, &ExtractResult{p.dicts[foundKeyword], idx})
						idx = j
					}
				}
			}
			if foundKeyword == "" {
				idx++
			} else if extractOption.Longest {
				res = append(res, &ExtractResult{p.dicts[foundKeyword], idx})
				idx = j
			}
			begin = false
		}
	}
	return res
}

func (p *KeywordProcessor) ReplaceKeywords(sentence string, option ...*Option) (filteredSentence string, res []*ExtractResult) {
	extractOption := defaultOption
	if len(option) > 0 {
		extractOption = option[0]
	}
	originalRunes := []rune(sentence)
	if !p.caseSensitive {
		if p.inputLowercased {
			sentence = strings.ToLower(sentence)
		} else {
			sentence = strings.ToUpper(sentence)
		}
	}
	runes := []rune(sentence)
	size := len(runes)
	idx := 0
	offset := 0
	begin := true
	var curTrie *trie
	for idx < size {
		curTrie = p.keytrie
		c := runes[idx]
		if _, ok := p.noboundaryWords[c]; !ok {
			idx++
			begin = true
		} else if !begin {
			idx++
		} else {
			var j = idx
			foundKeyword := ""
			for j = idx; j < size; j++ {
				c = runes[j]
				curTrie = curTrie.next[c]
				if curTrie == nil {
					break
				}
				if curTrie.word != "" && (j == size-1 || !p.noboundaryWords[runes[j+1]]) {
					foundKeyword = curTrie.word
					if !extractOption.Longest {
						replacement := []rune(p.dicts[foundKeyword])
						originalRunes = append(originalRunes[:idx+offset], append(replacement, originalRunes[j+1+offset:]...)...)
						offset = offset + (len(replacement) - len([]rune(foundKeyword)))
						res = append(res, &ExtractResult{p.dicts[foundKeyword], idx})
						idx = j
					}
				}
			}
			if foundKeyword == "" {
				idx++
			} else if extractOption.Longest {
				replacement := []rune(p.dicts[foundKeyword])
				originalRunes = append(originalRunes[:idx+offset], append(replacement, originalRunes[j+offset:]...)...)
				offset = offset + (len(replacement) - len([]rune(foundKeyword)))
				res = append(res, &ExtractResult{p.dicts[foundKeyword], idx})
				idx = j
			}
			begin = false
		}
	}
	return string(originalRunes), res
}

func (p *KeywordProcessor) MaskKeywords(sentence string, maskingFunction func(string) string, option ...*Option) (filteredSentence string) {
	extractOption := defaultOption
	if len(option) > 0 {
		extractOption = option[0]
	}
	originalRunes := []rune(sentence)
	if !p.caseSensitive {
		if p.inputLowercased {
			sentence = strings.ToLower(sentence)
		} else {
			sentence = strings.ToUpper(sentence)
		}
	}
	runes := []rune(sentence)
	size := len(runes)
	idx := 0
	offset := 0
	begin := true
	var curTrie *trie
	for idx < size {
		curTrie = p.keytrie
		c := runes[idx]
		if _, ok := p.noboundaryWords[c]; !ok {
			idx++
			begin = true
		} else if !begin {
			idx++
		} else {
			var j = idx
			foundKeyword := ""
			for j = idx; j < size; j++ {
				c = runes[j]
				curTrie = curTrie.next[c]
				if curTrie == nil {
					break
				}
				if curTrie.word != "" && (j == size-1 || !p.noboundaryWords[runes[j+1]]) {
					foundKeyword = curTrie.word
					if !extractOption.Longest {
						replacement := []rune(maskingFunction(foundKeyword))
						originalRunes = append(originalRunes[:idx+offset], append(replacement, originalRunes[j+1+offset:]...)...)
						offset = offset + (len(replacement) - len([]rune(foundKeyword)))
						idx = j
					}
				}
			}
			if foundKeyword == "" {
				idx++
			} else if extractOption.Longest {
				replacement := []rune(maskingFunction(foundKeyword))
				originalRunes = append(originalRunes[:idx+offset], append(replacement, originalRunes[j+offset:]...)...)
				offset = offset + (len(replacement) - len([]rune(foundKeyword)))
				idx = j
			}
			begin = false
		}
	}
	return string(originalRunes)
}

func (p *KeywordProcessor) RemoveKeywords(keywords ...string) {
	p.Lock()
	defer p.Unlock()
	for _, keyword := range keywords {
		if !p.caseSensitive {
			keyword = strings.ToLower(keyword)
		}
		p.keytrie.removeKeyword(keyword)
	}
}

func (p *KeywordProcessor) Exists(keyword string) bool {
	return p.keytrie.exists(keyword)
}
