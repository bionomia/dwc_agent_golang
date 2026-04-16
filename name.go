// Package dwcagent parses and cleanses Darwin Core terms containing people names.
// It is a Go port of the Ruby dwc_agent gem (https://github.com/bionomia/dwc_agent).
package dwcagent

import (
	"encoding/json"
	"strings"
)

// Name holds the parsed components of a personal name, mirroring Namae::Name.
type Name struct {
	Title            *string `json:"title"`
	Appellation      *string `json:"appellation"`
	Given            *string `json:"given"`
	Particle         *string `json:"particle"`
	Family           *string `json:"family"`
	Suffix           *string `json:"suffix"`
	Nick             *string `json:"nick"`
	DroppingParticle *string `json:"dropping_particle"`
}

// Default returns an empty Name used as the sentinel "reject" value.
func Default() Name {
	return Name{}
}

// IsDefault returns true when all fields are nil.
func (n Name) IsDefault() bool {
	return n.Title == nil &&
		n.Appellation == nil &&
		n.Given == nil &&
		n.Particle == nil &&
		n.Family == nil &&
		n.Suffix == nil &&
		n.Nick == nil &&
		n.DroppingParticle == nil
}

// DisplayOrder returns "Given Particle Family" (subset of non-nil fields).
func (n Name) DisplayOrder() string {
	var parts []string
	if n.Given != nil && *n.Given != "" {
		parts = append(parts, *n.Given)
	}
	if n.Particle != nil && *n.Particle != "" {
		parts = append(parts, *n.Particle)
	}
	if n.Family != nil && *n.Family != "" {
		parts = append(parts, *n.Family)
	}
	return strings.Join(parts, " ")
}

// MarshalJSON renders a Name as a flat JSON object with null for nil fields.
func (n Name) MarshalJSON() ([]byte, error) {
	type wire struct {
		Title            interface{} `json:"title"`
		Appellation      interface{} `json:"appellation"`
		Given            interface{} `json:"given"`
		Particle         interface{} `json:"particle"`
		Family           interface{} `json:"family"`
		Suffix           interface{} `json:"suffix"`
		Nick             interface{} `json:"nick"`
		DroppingParticle interface{} `json:"dropping_particle"`
	}
	w := wire{}
	if n.Title != nil {
		w.Title = *n.Title
	}
	if n.Appellation != nil {
		w.Appellation = *n.Appellation
	}
	if n.Given != nil {
		w.Given = *n.Given
	}
	if n.Particle != nil {
		w.Particle = *n.Particle
	}
	if n.Family != nil {
		w.Family = *n.Family
	}
	if n.Suffix != nil {
		w.Suffix = *n.Suffix
	}
	if n.Nick != nil {
		w.Nick = *n.Nick
	}
	if n.DroppingParticle != nil {
		w.DroppingParticle = *n.DroppingParticle
	}
	return json.Marshal(w)
}

// strPtr returns a pointer to the trimmed string, or nil if it is empty.
func strPtr(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}