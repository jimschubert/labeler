package model

import (
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

// SimpleConfig is the simplest supported config structure. See FullConfig for more functionality.
type SimpleConfig struct {
	// Comment will be applied to any issue or pull request matching the target labels
	Comment string `yaml:"comment,omitempty"`

	// Labels are keyed by the label to be applied, and valued by the array of regular expression patterns to match before applying
	Labels map[string][]string `yaml:"labels,omitempty,flow"`
}

// FromBytes parses the bytes into the SimpleConfig object
func (s *SimpleConfig) FromBytes(b []byte) error {
	return yaml.Unmarshal(b, &s)
}

// LabelsFor allows config implementations to determine the labels to be applied to the input strings
func (s *SimpleConfig) LabelsFor(text ...string) []string {
	searchable := []byte(strings.Join(text, " "))
	labels := make([]string, 0)
	for key, values := range s.Labels {
		for _, pattern := range values {
			re := regexp.MustCompile(pattern)
			if re.Match(searchable) {
				labels = append(labels, key)
			}
		}
	}
	return labels
}
