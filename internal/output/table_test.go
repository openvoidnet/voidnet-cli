package output

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

type tableRow struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type singleEnvelope struct {
	Apps []tableRow `json:"apps"`
}

type mixedEnvelope struct {
	Records []tableRow `json:"records"`
	Page    pageInfo   `json:"pagination"`
}

type pageInfo struct {
	Total int `json:"total"`
}

type multiList struct {
	A []tableRow `json:"a"`
	B []tableRow `json:"b"`
}

func TestPrintTableSingle_UnwrapsListEnvelopes(t *testing.T) {
	var buf bytes.Buffer
	single := singleEnvelope{Apps: []tableRow{{ID: "1", Name: "x"}}}
	if err := printTableSingle(&buf, reflect.ValueOf(single)); err != nil {
		t.Fatalf("single envelope must render: %v", err)
	}
	if !strings.Contains(buf.String(), "ID") || !strings.Contains(buf.String(), "x") {
		t.Fatalf("single envelope must show columns+row, got %q", buf.String())
	}
	buf.Reset()
	mixed := mixedEnvelope{Records: []tableRow{{ID: "2", Name: "y"}}, Page: pageInfo{Total: 1}}
	if err := printTableSingle(&buf, reflect.ValueOf(mixed)); err != nil {
		t.Fatalf("mixed envelope must render records: %v", err)
	}
	if !strings.Contains(buf.String(), "y") {
		t.Fatalf("mixed envelope must show record row, got %q", buf.String())
	}
	buf.Reset()
	if err := printTableSingle(&buf, reflect.ValueOf(multiList{})); err == nil {
		t.Fatal("multi-list envelope must still fail loud, not guess")
	}
}
