package awxlimit

import (
	"errors"
	"net"
	"strings"
	"unicode"
)

var (
	ErrEmptyPattern = errors.New("empty limit pattern")
)

type EntityKind string

const (
	KindUnknown      EntityKind = "unknown"
	KindHost         EntityKind = "host"
	KindGroup        EntityKind = "group"
	KindHostPattern  EntityKind = "host_pattern"
	KindGroupPattern EntityKind = "group_pattern"
	KindFile         EntityKind = "file"
	KindRegex        EntityKind = "regex"
	KindIP           EntityKind = "ip"
)

type Entity struct {
	Raw  string     `json:"raw"`
	Kind EntityKind `json:"kind"`
	Op   OpKind     `json:"op"`
}

type OpKind string

const (
	OpAny OpKind = "any"
	OpAll OpKind = "all"
	OpNot OpKind = "not"
)

type ParsedLimit struct {
	Any []Entity `json:"any"`
	All []Entity `json:"all"`
	Not []Entity `json:"not"`

	Included []Entity `json:"included"`
	Excluded []Entity `json:"excluded"`
}

func ParseLimitPattern(s string) (ParsedLimit, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return ParsedLimit{}, ErrEmptyPattern
	}

	tokens := splitTopLevelUnion(s)

	var out ParsedLimit
	for _, tok := range tokens {
		tok = strings.TrimSpace(tok)
		if tok == "" {
			continue
		}
		op, raw := peelOps(tok)
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}

		ent := Entity{
			Raw:  raw,
			Kind: guessKind(raw),
			Op:   op,
		}

		switch op {
		case OpAll:
			out.All = append(out.All, ent)
			out.Included = append(out.Included, ent)
		case OpNot:
			out.Not = append(out.Not, ent)
			out.Excluded = append(out.Excluded, ent)
		default:
			out.Any = append(out.Any, ent)
			out.Included = append(out.Included, ent)
		}
	}

	if len(out.Any) == 0 && len(out.All) == 0 && len(out.Not) == 0 {
		return ParsedLimit{}, ErrEmptyPattern
	}

	return out, nil
}

func splitTopLevelUnion(s string) []string {
	var parts []string
	var b strings.Builder

	escape := false
	for _, r := range s {
		if escape {
			b.WriteRune(r)
			escape = false
			continue
		}
		if r == '\\' {
			escape = true
			continue
		}
		if r == ':' || r == ',' {
			parts = append(parts, b.String())
			b.Reset()
			continue
		}
		b.WriteRune(r)
	}
	parts = append(parts, b.String())
	return parts
}

func peelOps(tok string) (OpKind, string) {
	tok = strings.TrimSpace(tok)
	if tok == "" {
		return OpAny, tok
	}
	switch tok[0] {
	case '&':
		return OpAll, tok[1:]
	case '!':
		return OpNot, tok[1:]
	default:
		return OpAny, tok
	}
}

func guessKind(raw string) EntityKind {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return KindUnknown
	}

	if strings.HasPrefix(raw, "@") && len(raw) > 1 {
		return KindFile
	}
	if strings.HasPrefix(raw, "~") && len(raw) > 1 {
		return KindRegex
	}
	if ip := net.ParseIP(raw); ip != nil {
		return KindIP
	}
	if strings.Contains(raw, "[") && strings.Contains(raw, "]") {
		return KindHostPattern
	}
	if strings.ContainsAny(raw, "*?") {
		return KindHostPattern
	}
	if strings.Contains(raw, ".") {
		return KindHost
	}
	for _, r := range raw {
		if unicode.IsDigit(r) {
			return KindHost
		}
	}
	if isGroupish(raw) {
		return KindGroup
	}
	return KindUnknown
}

func isGroupish(s string) bool {
	if s == "" {
		return false
	}
	r0 := rune(s[0])
	if !unicode.IsLetter(r0) && r0 != '_' {
		return false
	}
	for _, r := range s[1:] {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' {
			continue
		}
		return false
	}
	return true
}
