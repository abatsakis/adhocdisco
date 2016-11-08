package cache

import (
	"strings"
	"testing"
	"time"
)

func TestState(t *testing.T) {
	c, tok := CreateToken()
	if c != 200 {
		t.FailNow()
	}

	defaultValidity = 5 * time.Second

	id1 := "1"
	c, _ = AddToken(tok, id1)
	if c != 200 {
		t.FailNow()
	}

	c, toks := GetToken("randomToken")
	if c != 404 {
		t.FailNow()
	}

	id2 := "2"
	c, _ = AddToken(tok, id2)
	if c != 200 {
		t.FailNow()
	}

	c, toks = GetToken(tok)
	if c != 200 {
		t.FailNow()
	}

	if !strings.Contains(toks, id1) {
		t.Error("id1 not found")
		t.FailNow()
	}

	if !strings.Contains(toks, id2) {
		t.Error("id2 not found")
		t.FailNow()
	}

	time.Sleep(defaultValidity)

	c, toks = GetToken(tok)
	if c != 200 {
		t.FailNow()
	}

	if strings.Contains(toks, id1) {
		t.Error("id1 found")
		t.FailNow()
	}
}
