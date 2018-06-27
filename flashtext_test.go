package flashtext

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	extractCases = make([]*ExtractorTestCase, 0, 100)
	removeCases  = make([]*RemoveTestCase, 0, 100)
	replaceCases = make([]*ReplaceTestCase, 0, 100)
)

func TestExtract(t *testing.T) {
	Init(t)
	//case insensitive
	for _, c := range extractCases {
		p := NewKeywordProcessor()
		p.SetCaseSensitive(false)
		for cleanName, keywords := range c.KeywordDict {
			for _, keyword := range keywords {
				p.AddKeywordAndName(keyword, cleanName)
			}
		}
		res := p.ExtractKeywords(c.Sentence, &Option{Longest: true})
		resultArray := []string{}
		for _, result := range res {
			resultArray = append(resultArray, result.Keyword)
		}
		assert.EqualValues(t, c.Keywords, resultArray, "insensitive keywords should match at sentence:"+c.Sentence)
	}

	//case sensitive
	for _, c := range extractCases {
		p := NewKeywordProcessor()
		p.SetCaseSensitive(true)
		for cleanName, keywords := range c.KeywordDict {
			for _, keyword := range keywords {
				p.AddKeywordAndName(keyword, cleanName)
			}
		}
		res := p.ExtractKeywords(c.Sentence, &Option{Longest: true})
		resultArray := []string{}
		for _, result := range res {
			resultArray = append(resultArray, result.Keyword)
		}
		assert.EqualValues(t, c.KeywordsCaseSensitive, resultArray, "sensitive keywords should match at sentence:"+c.Sentence)
	}
}

func TestRemove(t *testing.T) {
	Init(t)
	//case insensitive
	for _, c := range removeCases {
		p := NewKeywordProcessor()
		p.SetCaseSensitive(false)
		for cleanName, keywords := range c.KeywordDict {
			for _, keyword := range keywords {
				p.AddKeywordAndName(keyword, cleanName)
			}
		}
		for _, keywords := range c.RemoveKeywordDict {
			p.RemoveKeywords(keywords...)
		}
		res := p.ExtractKeywords(c.Sentence, &Option{Longest: true})
		resultArray := []string{}
		for _, result := range res {
			resultArray = append(resultArray, result.Keyword)
		}
		assert.EqualValues(t, c.Keywords, resultArray, "insensitive keywords should match at sentence:"+c.Sentence)
	}

	//case sensitive
	for _, c := range removeCases {
		p := NewKeywordProcessor()
		p.SetCaseSensitive(true)
		for cleanName, keywords := range c.KeywordDict {
			for _, keyword := range keywords {
				p.AddKeywordAndName(keyword, cleanName)
			}
		}
		for _, keywords := range c.RemoveKeywordDict {
			p.RemoveKeywords(keywords...)
		}
		res := p.ExtractKeywords(c.Sentence, &Option{Longest: true})
		resultArray := []string{}
		for _, result := range res {
			resultArray = append(resultArray, result.Keyword)
		}
		assert.EqualValues(t, c.KeywordsCaseSensitive, resultArray, "sensitive keywords should match at sentence:"+c.Sentence)
	}
}

func TestReplace(t *testing.T) {
	InitReplace(t)
	//case insensitive
	for _, c := range replaceCases {
		p := NewKeywordProcessor()
		p.SetCaseSensitive(false)
		for cleanName, keywords := range c.KeywordDict {
			for _, keyword := range keywords {
				p.AddKeywordAndName(keyword, cleanName)
			}
		}
		replaced, res := p.ReplaceKeywords(c.Sentence, &Option{Longest: true})
		resultArray := []string{}
		for _, result := range res {
			resultArray = append(resultArray, result.Keyword)
		}
		assert.EqualValues(t, c.Keywords, resultArray, "insensitive keywords should match at sentence:"+c.Sentence)
		assert.EqualValues(t, c.Expected, replaced, c.Explaination)
	}

	//case sensitive
	for _, c := range replaceCases {
		p := NewKeywordProcessor()
		p.SetCaseSensitive(true)
		for cleanName, keywords := range c.KeywordDict {
			for _, keyword := range keywords {
				p.AddKeywordAndName(keyword, cleanName)
			}
		}
		replaced, res := p.ReplaceKeywords(c.Sentence, &Option{Longest: true})
		resultArray := []string{}
		for _, result := range res {
			resultArray = append(resultArray, result.Keyword)
		}
		assert.EqualValues(t, c.KeywordsCaseSensitive, resultArray, "sensitive keywords should match at sentence:"+c.Sentence)
		assert.EqualValues(t, c.ExpectedCaseSensitive, replaced, c.Explaination)
	}
}

// read test_cases json files
func Init(t *testing.T) {
	err := json.Unmarshal(helperLoadBytes(t, "extracts.json"), &extractCases)
	if err != nil {
		t.Fatal(err)
	}
}

func InitReplace(t *testing.T) {
	err := json.Unmarshal(helperLoadBytes(t, "replaces.json"), &replaceCases)
	if err != nil {
		t.Fatal(err)
	}
}

func helperLoadBytes(t *testing.T, name string) []byte {
	path := filepath.Join("test_cases", name) // relative path
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}

type ExtractorTestCase struct {
	Sentence              string
	KeywordDict           map[string][]string `json:"keyword_dict"`
	Explaination          string
	Keywords              []string
	KeywordsCaseSensitive []string `json:"keywords_case_sensitive"`
}

type RemoveTestCase struct {
	Sentence              string              `json:"sentence"`
	KeywordDict           map[string][]string `json:"remove_keyword_dict"`
	RemoveKeywordDict     map[string][]string `json:"remove_keyword_dict"`
	Keywords              []string            `json:"keywords"`
	KeywordsCaseSensitive []string            `json:"keywords_case_sensitive"`
}

type ReplaceTestCase struct {
	Sentence              string
	KeywordDict           map[string][]string `json:"keyword_dict"`
	Explaination          string
	Keywords              []string
	KeywordsCaseSensitive []string `json:"keywords_case_sensitive"`
	Expected              string   `json:"expected"`
	ExpectedCaseSensitive string   `json:"expected_case_sensitive"`
}
