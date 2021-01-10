// Copyright 2017 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package parser

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestAliases(t *testing.T) {
	for k, v := range aliases {
		if cmp.Equal(k, v) {
			t.Errorf("exptected k: %s to not equal v: %s", k, v)
		}
		if diff := cmp.Diff(tokenMap[k], tokenMap[v]); diff != "" {
			t.Errorf("exptected tokens to match: %s", diff)
		}
	}
}

func TestKeywordConsistent(t *testing.T) {
	data, err := ioutil.ReadFile("parser.y")
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)

	reservedKeywords := extractKeywords(content, "ReservedKeyword")
	unreservedKeywords := extractKeywords(content, "UnReservedKeyword")
	notKeywordTokens := extractKeywords(content, "NotKeywordToken")
	tidbKeywords := extractKeywords(content, "TiDBKeyword")

	keywordCount := len(reservedKeywords) + len(unreservedKeywords) + len(notKeywordTokens) + len(tidbKeywords)
	if diff := cmp.Diff(len(tokenMap)-len(aliases), keywordCount-len(windowFuncTokenMap)); diff != "" {
		t.Errorf("length tokenMap does not match keyword count: %s", diff)
	}

	unreservedCollectionDef := extractKeywordsFromCollectionDef(content, "UnReservedKeyword:")
	if diff := cmp.Diff(unreservedKeywords, unreservedCollectionDef); diff != "" {
		t.Errorf("unreserved keywords: %s", diff)
	}

	notKeywordTokensCollectionDef := extractKeywordsFromCollectionDef(content, "NotKeywordToken:")
	if diff := cmp.Diff(notKeywordTokens, notKeywordTokensCollectionDef); diff != "" {
		t.Errorf("not keyword tokens: %s", diff)
	}

	tidbKeywordsCollectionDef := extractKeywordsFromCollectionDef(content, "TiDBKeyword:")
	if diff := cmp.Diff(tidbKeywords, tidbKeywordsCollectionDef); diff != "" {
		t.Errorf("TiDB keywords: %s", diff)
	}
}

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

func extractKeywords(content, keyword string) []string {
	start := fmt.Sprintf("/* The following tokens belong to %s. Notice: make sure these tokens are contained in %s. */", keyword, keyword)
	lines := extractLines(content, start, "")
	return extractQuotedWords(lines)
}

func extractKeywordsFromCollectionDef(content, startMarker string) []string {
	lines := extractLines(content, startMarker, "")
	return extractQuotedWords(lines)
}
