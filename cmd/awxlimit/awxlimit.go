package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	awxlimit "github.com/Ramoreik/awxlimit/pkg/awxlimit"
)

func main() {
	var (
		pattern       = flag.String("pattern", "", "Ansible limit pattern (e.g. web:db:&staging:!phoenix)")
		inventoryPath = flag.String("inventory", "", "Path to inventory JSON (hosts + groups)")
		pretty        = flag.Bool("pretty", true, "Pretty-print JSON")
	)
	flag.Parse()

	if *pattern == "" || *inventoryPath == "" {
		fmt.Fprintln(os.Stderr, "error: -pattern and -inventory are required")
		os.Exit(2)
	}

	b, err := os.ReadFile(*inventoryPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading inventory: %v", err)
		os.Exit(2)
	}

	var inv awxlimit.Inventory
	if err := json.Unmarshal(b, &inv); err != nil {
		fmt.Fprintf(os.Stderr, "error parsing inventory JSON: %v", err)
		os.Exit(2)
	}

	matched, err := awxlimit.MatchHosts(*pattern, inv)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(2)
	}

	out := struct {
		Pattern string   `json:"pattern"`
		Matched []string `json:"matched"`
	}{
		Pattern: *pattern,
		Matched: matched,
	}

	var outb []byte
	if *pretty {
		outb, err = json.MarshalIndent(out, "", "  ")
	} else {
		outb, err = json.Marshal(out)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(1)
	}
	fmt.Println(string(outb))
}
