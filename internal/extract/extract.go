package extract

import (
	"bufio"
	"fmt"
	"sort"
	"strings"
)

type Keyword string

const (
	KeywordReserved   Keyword = "ReservedKeyword"
	KeywordUnreserved         = "UnReservedKeyword"
	KeywordNot                = "NotKeywordToken"
	KeywordTiDB               = "TiDBKeyword"
)

func extractMiddle(str, startMarker, endMarker string) string {
	startIdx := strings.Index(str, startMarker)
	if startIdx == -1 {
		return ""
	}
	str = str[startIdx+len(startMarker):]
	endIdx := strings.Index(str, endMarker)
	if endIdx == -1 {
		return ""
	}
	return str[:endIdx]
}

func extractLines(str, startMarker, endMarker string) []string {
	var started, stopped bool
	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(str))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == startMarker {
			started = true
			continue
		}
		if started && line == endMarker {
			stopped = true
			continue
		}
		if stopped {
			continue
		}
		if started {
			lines = append(lines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	return lines
}

func extractQuotedWords(strs []string) []string {
	var words []string
	for _, str := range strs {
		word := extractMiddle(str, "\"", "\"")
		if word == "" {
			continue
		}
		words = append(words, word)
	}
	sort.Strings(words)
	return words
}

func KeywordsFromTokens(content string, keyword Keyword) []string {
	start := fmt.Sprintf("/* The following tokens belong to %s. Notice: make sure these tokens are contained in %s. */", keyword, keyword)
	lines := extractLines(content, start, "")
	return extractQuotedWords(lines)
}

func KeywordsFromCollectionDef(content string, keyword Keyword) []string {
	lines := extractLines(content, fmt.Sprintf("%s:", keyword), "")
	return extractQuotedWords(lines)
}
