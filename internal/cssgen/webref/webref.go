package webref

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type Data struct {
	Properties  map[string]*Property
	Features    map[string]string
	Descriptors map[string]map[string]string
	Selectors   []string
	NamedColors []string
	Values      map[string]string
}

type Property struct {
	Name   string
	Values []string
}

type document struct {
	AtRules    []atRule `json:"atrules"`
	Functions  []item   `json:"functions"`
	Properties []item   `json:"properties"`
	Selectors  []item   `json:"selectors"`
	Types      []item   `json:"types"`
}

type atRule struct {
	Name        string `json:"name"`
	Descriptors []item `json:"descriptors"`
}

type item struct {
	Name   string `json:"name"`
	Syntax string `json:"syntax"`
}

func Load(dir string) (*Data, error) {
	file := filepath.Join(dir, "css.json")
	raw, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("webref: %w", err)
	}
	var doc document
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("webref: %s: %w", file, err)
	}
	values := doc.values()
	descriptors := doc.descriptors()
	features := descriptors["media"]
	if features == nil {
		features = map[string]string{}
	}
	return &Data{
		Properties:  doc.properties(),
		Features:    features,
		Descriptors: descriptors,
		Selectors:   doc.selectorNames(),
		NamedColors: alternatives(values["named-color"]),
		Values:      values,
	}, nil
}

func (d document) properties() map[string]*Property {
	properties := map[string]*Property{}
	for _, entry := range d.Properties {
		if entry.Name == "" || entry.Syntax == "" {
			continue
		}
		property := properties[entry.Name]
		if property == nil {
			property = &Property{Name: entry.Name}
			properties[entry.Name] = property
		}
		property.Values = append(property.Values, entry.Syntax)
	}
	return properties
}

func (d document) descriptors() map[string]map[string]string {
	rules := map[string]map[string]string{}
	for _, rule := range d.AtRules {
		name := strings.TrimPrefix(rule.Name, "@")
		if rules[name] == nil {
			rules[name] = map[string]string{}
		}
		for _, descriptor := range rule.Descriptors {
			if descriptor.Name != "" && descriptor.Syntax != "" {
				rules[name][descriptor.Name] = descriptor.Syntax
			}
		}
	}
	return rules
}

func (d document) selectorNames() []string {
	var names []string
	for _, entry := range d.Selectors {
		if entry.Name != "" {
			names = append(names, entry.Name)
		}
	}
	slices.Sort(names)
	return slices.Compact(names)
}

func (d document) values() map[string]string {
	values := map[string]string{}
	for _, entry := range slices.Concat(d.Types, d.Functions) {
		if entry.Name == "" || entry.Syntax == "" {
			continue
		}
		if _, defined := values[entry.Name]; !defined {
			values[entry.Name] = entry.Syntax
		}
	}
	return values
}

func alternatives(syntax string) []string {
	var keywords []string
	for _, part := range strings.Split(syntax, "|") {
		if keyword := strings.TrimSpace(part); keyword != "" {
			keywords = append(keywords, keyword)
		}
	}
	return keywords
}
