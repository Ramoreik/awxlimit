package awxlimit

import "testing"

func TestParseLimitPatternCommon(t *testing.T) {
	p, err := ParseLimitPattern("webservers:dbservers:&staging:!phoenix")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(p.Any) != 2 || p.Any[0].Raw != "webservers" || p.Any[1].Raw != "dbservers" {
		t.Fatalf("unexpected Any: %#v", p.Any)
	}
	if len(p.All) != 1 || p.All[0].Raw != "staging" {
		t.Fatalf("unexpected All: %#v", p.All)
	}
	if len(p.Not) != 1 || p.Not[0].Raw != "phoenix" {
		t.Fatalf("unexpected Not: %#v", p.Not)
	}
}

func TestParseLimitPatternEscapes(t *testing.T) {
	p, err := ParseLimitPattern(`~^(web|db)\:prod\,blue$:!bad`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(p.Any) != 1 || p.Any[0].Kind != KindRegex {
		t.Fatalf("expected regex token, got %#v", p.Any)
	}
	if p.Any[0].Raw != `~^(web|db)\:prod\,blue$` {
		t.Fatalf("unexpected raw: %q", p.Any[0].Raw)
	}
	if len(p.Not) != 1 || p.Not[0].Raw != "bad" {
		t.Fatalf("unexpected Not: %#v", p.Not)
	}
}
