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
	"io/ioutil"
	"testing"

	iextract "github.com/kyleconroy/sqlparse/internal/extract"

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
	reservedKeywords := iextract.KeywordsFromTokens(content, iextract.KeywordReserved)
	unreservedKeywords := iextract.KeywordsFromTokens(content, iextract.KeywordUnreserved)
	notKeywordTokens := iextract.KeywordsFromTokens(content, iextract.KeywordNot)
	tidbKeywords := iextract.KeywordsFromTokens(content, iextract.KeywordTiDB)

	keywordCount := len(reservedKeywords) + len(unreservedKeywords) + len(notKeywordTokens) + len(tidbKeywords)
	if diff := cmp.Diff(len(tokenMap)-len(aliases), keywordCount-len(windowFuncTokenMap)); diff != "" {
		t.Errorf("length tokenMap does not match keyword count: %s", diff)
	}

	unreservedCollectionDef := iextract.KeywordsFromCollectionDef(content, iextract.KeywordUnreserved)
	if diff := cmp.Diff(unreservedKeywords, unreservedCollectionDef); diff != "" {
		t.Errorf("unreserved keywords: %s", diff)
	}

	notKeywordTokensCollectionDef := iextract.KeywordsFromCollectionDef(content, iextract.KeywordNot)
	if diff := cmp.Diff(notKeywordTokens, notKeywordTokensCollectionDef); diff != "" {
		t.Errorf("not keyword tokens: %s", diff)
	}

	tidbKeywordsCollectionDef := iextract.KeywordsFromCollectionDef(content, iextract.KeywordTiDB)
	if diff := cmp.Diff(tidbKeywords, tidbKeywordsCollectionDef); diff != "" {
		t.Errorf("TiDB keywords: %s", diff)
	}
}
