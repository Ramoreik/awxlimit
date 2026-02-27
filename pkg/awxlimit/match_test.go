package awxlimit

import (
	"reflect"
	"testing"
)

func TestMatchHostsUnionAllNot(t *testing.T) {
	inv := Inventory{
		Hosts: []string{"web01", "web02", "db01", "phoenix", "misc01"},
		Groups: []Group{
			{Name: "webservers", Hosts: []string{"web01", "web02"}},
			{Name: "dbservers", Hosts: []string{"db01", "phoenix"}},
			{Name: "staging", Hosts: []string{"web02", "db01", "phoenix"}},
		},
	}

	got, err := MatchHosts("webservers:dbservers:&staging:!phoenix", inv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"db01", "web02"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestMatchHostsOnlyAllMeansStartFromAll(t *testing.T) {
	inv := Inventory{
		Hosts: []string{"h1", "h2", "h3"},
		Groups: []Group{
			{Name: "blue", Hosts: []string{"h1", "h3"}},
		},
	}

	got, err := MatchHosts("&blue", inv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"h1", "h3"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestMatchHostsGlob(t *testing.T) {
	inv := Inventory{
		Hosts: []string{"web01", "web02", "db01"},
		Groups: []Group{
			{Name: "web", Hosts: []string{"web01", "web02"}},
		},
	}

	got, err := MatchHosts("web*", inv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"web01", "web02"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestMatchHostsGroupChildren(t *testing.T) {
	inv := Inventory{
		Hosts: []string{"a1", "a2", "b1"},
		Groups: []Group{
			{Name: "a", Hosts: []string{"a1"}},
			{Name: "b", Hosts: []string{"b1"}},
			{Name: "all_apps", Children: []string{"a", "b"}, Hosts: []string{"a2"}},
		},
	}

	got, err := MatchHosts("all_apps", inv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"a1", "a2", "b1"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}
