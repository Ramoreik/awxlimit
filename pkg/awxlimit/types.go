package awxlimit

type Inventory struct {
	Hosts  []string `json:"hosts"`
	Groups []Group  `json:"groups"`
}

type Group struct {
	Name     string   `json:"name"`
	Hosts    []string `json:"hosts,omitempty"`
	Children []string `json:"children,omitempty"`
}
