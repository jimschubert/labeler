package model

import (
	"errors"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

type (
	// Enable is a structure to hold options around enabling the labeler
	Enable struct {
		Issues       *bool `yaml:"issues,omitempty"`
		PullRequests *bool `yaml:"prs,omitempty"`
	}

	// Comments are optional comment text to be added to the issue or pull request when labels are applied
	Comments struct {
		Issues       *string `yaml:"issues,omitempty"`
		PullRequests *string `yaml:"prs,omitempty"`
	}

	// Label holds the rules around how labels will be applied
	Label struct {
		Include  []string `yaml:"include,omitempty,flow"`
		Exclude  []string `yaml:"exclude,omitempty,flow"`
		Branches []string `yaml:"branches,omitempty,flow"`
	}

	// FullConfig is the container defining how the configuration object is structured
	FullConfig struct {
		Enable   *Enable          `yaml:"enable,omitempty"`
		Comments *Comments        `yaml:"comments,omitempty"`
		Labels   map[string]Label `yaml:"labels,flow"`
	}
)

// FromBytes is used to parse bytes into the Config instance
func (f *FullConfig) FromBytes(b []byte) error {
	err := yaml.Unmarshal(b, &f)
	if err != nil {
		return err
	}

	if len(f.Labels) == 0 {
		return errors.New("full config requires labels to be defined")
	}

	return nil
}

// LabelsFor allows config implementations to determine the labels to be applied to the input strings
func (f *FullConfig) LabelsFor(text ...string) map[string]Label {
	searchable := []byte(strings.Join(text, " "))
	labels := make(map[string]Label)
	for key, values := range f.Labels {
		excluded := false
		for _, pattern := range values.Exclude {
			re := regexp.MustCompile(pattern)
			if re.Match(searchable) {
				excluded = true
				break
			}
		}

		if excluded {
			break
		}

		for _, pattern := range values.Include {
			re := regexp.MustCompile(pattern)
			if re.Match(searchable) {
				labels[key] = values
				break
			}
		}
	}
	return labels
}

// Ptr gets the pointer to an Enable object
func (e Enable) Ptr() *Enable { return &e }

// Ptr gets the pointer to a Comments object
func (c Comments) Ptr() *Comments { return &c }

// Ptr gets the pointer to a Label object
func (l Label) Ptr() *Label { return &l }

// Ptr gets the pointer to a FullConfig object
func (f FullConfig) Ptr() *FullConfig { return &f }
