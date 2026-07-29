package config

import "sort"

type Config struct {
	Source      Endpoint                  `yaml:"source"`
	Target      Endpoint                  `yaml:"target"`
	Defaults    Defaults                  `yaml:"defaults"`
	Collections map[string]CollectionRule `yaml:"collections"`
	Indexes     map[string][]IndexRule    `yaml:"indexes"`
}

type Endpoint struct {
	URIEnv string `yaml:"uriEnv"`
}

type Defaults struct {
	DryRun       bool     `yaml:"dryRun"`
	AllowDeletes bool     `yaml:"allowDeletes"`
	IgnoreFields []string `yaml:"ignoreFields"`
}

type CollectionRule struct {
	Key          []string                 `yaml:"key"`
	Mode         string                   `yaml:"mode"`
	IgnoreFields []string                 `yaml:"ignoreFields"`
	Arrays       map[string]ArrayRule     `yaml:"arrays"`
	AllowDeletes *bool                    `yaml:"allowDeletes"`
	References   map[string]ReferenceRule `yaml:"references"`
}

type ArrayRule struct {
	Strategy string   `yaml:"strategy"`
	Key      []string `yaml:"key"`
}

type ReferenceRule struct {
	Collection string   `yaml:"collection"`
	Key        []string `yaml:"key"`
	Required   bool     `yaml:"required"`
}

type IndexRule struct {
	Keys    map[string]int `yaml:"keys"`
	Options IndexOptions   `yaml:"options"`
}

type IndexOptions struct {
	Unique bool   `yaml:"unique"`
	Name   string `yaml:"name"`
}

type ResolvedConfig struct {
	Config Config
	Source ResolvedEndpoint
	Target ResolvedEndpoint
}

type ResolvedEndpoint struct {
	URI string
}

func (c Config) CollectionNames() []string {
	names := make([]string, 0, len(c.Collections))
	for name := range c.Collections {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
