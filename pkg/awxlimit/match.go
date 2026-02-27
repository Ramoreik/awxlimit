package awxlimit

import (
	"path"
	"regexp"
	"sort"
	"strings"
)

func MatchHosts(pattern string, inv Inventory) ([]string, error) {
	parsed, err := ParseLimitPattern(pattern)
	if err != nil {
		return nil, err
	}

	idx := newIndex(inv)

	var cur set
	if len(parsed.Any) == 0 {
		cur = idx.allHosts.clone()
	} else {
		cur = make(set)
		for _, e := range parsed.Any {
			s, err := idx.expandEntity(e)
			if err != nil {
				return nil, err
			}
			cur = cur.union(s)
		}
	}

	for _, e := range parsed.All {
		s, err := idx.expandEntity(e)
		if err != nil {
			return nil, err
		}
		cur = cur.intersect(s)
	}

	for _, e := range parsed.Not {
		s, err := idx.expandEntity(e)
		if err != nil {
			return nil, err
		}
		cur = cur.minus(s)
	}

	return cur.sorted(), nil
}

type index struct {
	hosts       []string
	allHosts    set
	groupDirect map[string]set
	groupChild  map[string][]string
}

func newIndex(inv Inventory) *index {
	hosts := make([]string, 0, len(inv.Hosts))
	seen := make(map[string]struct{}, len(inv.Hosts))
	for _, h := range inv.Hosts {
		h = strings.TrimSpace(h)
		if h == "" {
			continue
		}
		if _, ok := seen[h]; ok {
			continue
		}
		seen[h] = struct{}{}
		hosts = append(hosts, h)
	}

	groupDirect := make(map[string]set, len(inv.Groups))
	groupChild := make(map[string][]string, len(inv.Groups))
	for _, g := range inv.Groups {
		name := strings.TrimSpace(g.Name)
		if name == "" {
			continue
		}
		if _, ok := groupDirect[name]; !ok {
			groupDirect[name] = make(set)
		}
		for _, h := range g.Hosts {
			h = strings.TrimSpace(h)
			if h == "" {
				continue
			}
			groupDirect[name].add(h)
		}
		if len(g.Children) > 0 {
			children := make([]string, 0, len(g.Children))
			for _, c := range g.Children {
				c = strings.TrimSpace(c)
				if c != "" {
					children = append(children, c)
				}
			}
			groupChild[name] = children
		}
	}

	all := make(set)
	for _, h := range hosts {
		all.add(h)
	}

	return &index{
		hosts:       hosts,
		allHosts:    all,
		groupDirect: groupDirect,
		groupChild:  groupChild,
	}
}

func (idx *index) groupExists(name string) bool {
	_, ok := idx.groupDirect[name]
	if ok {
		return true
	}
	_, ok = idx.groupChild[name]
	return ok
}

func (idx *index) expandGroup(name string) set {
	name = strings.TrimSpace(name)
	if name == "" {
		return make(set)
	}

	out := make(set)
	vis := make(map[string]struct{})
	var walk func(string)
	walk = func(g string) {
		if _, ok := vis[g]; ok {
			return
		}
		vis[g] = struct{}{}
		if s, ok := idx.groupDirect[g]; ok {
			out = out.union(s)
		}
		for _, c := range idx.groupChild[g] {
			walk(c)
		}
	}
	walk(name)

	if len(idx.hosts) > 0 {
		out = out.intersect(idx.allHosts)
	}
	return out
}

func (idx *index) expandEntity(e Entity) (set, error) {
	raw := strings.TrimSpace(e.Raw)
	if raw == "" {
		return make(set), nil
	}

	if idx.groupExists(raw) {
		return idx.expandGroup(raw), nil
	}

	if strings.HasPrefix(raw, "~") && len(raw) > 1 {
		re, err := regexp.Compile(raw[1:])
		if err != nil {
			return nil, err
		}
		out := make(set)
		for g := range idx.groupDirect {
			if re.MatchString(g) {
				out = out.union(idx.expandGroup(g))
			}
		}
		for _, h := range idx.hosts {
			if re.MatchString(h) {
				out.add(h)
			}
		}
		return out, nil
	}

	if strings.ContainsAny(raw, "*?[]") {
		out := make(set)
		for g := range idx.groupDirect {
			if ok, _ := path.Match(raw, g); ok {
				out = out.union(idx.expandGroup(g))
			}
		}
		for _, h := range idx.hosts {
			if ok, _ := path.Match(raw, h); ok {
				out.add(h)
			}
		}
		return out, nil
	}

	out := make(set)
	for _, h := range idx.hosts {
		if h == raw {
			out.add(h)
			return out, nil
		}
	}

	return out, nil
}

type set map[string]struct{}

func (s set) add(v string) {
	if v == "" {
		return
	}
	s[v] = struct{}{}
}

func (s set) clone() set {
	out := make(set, len(s))
	for k := range s {
		out[k] = struct{}{}
	}
	return out
}

func (s set) union(o set) set {
	out := s.clone()
	for k := range o {
		out[k] = struct{}{}
	}
	return out
}

func (s set) intersect(o set) set {
	if len(s) == 0 || len(o) == 0 {
		return make(set)
	}
	out := make(set)
	if len(s) > len(o) {
		s, o = o, s
	}
	for k := range s {
		if _, ok := o[k]; ok {
			out[k] = struct{}{}
		}
	}
	return out
}

func (s set) minus(o set) set {
	if len(o) == 0 {
		return s.clone()
	}
	out := make(set)
	for k := range s {
		if _, ok := o[k]; !ok {
			out[k] = struct{}{}
		}
	}
	return out
}

func (s set) sorted() []string {
	out := make([]string, 0, len(s))
	for k := range s {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
