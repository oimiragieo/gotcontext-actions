package model

import (
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// EnvironmentName returns the deployment environment name if set.
func (j *Job) EnvironmentName() string {
	if j == nil {
		return ""
	}
	switch j.RawEnvironment.Kind {
	case yaml.ScalarNode:
		var name string
		if decodeNode(j.RawEnvironment, &name) {
			return name
		}
	case yaml.MappingNode:
		var val struct {
			Name string `yaml:"name"`
		}
		if decodeNode(j.RawEnvironment, &val) {
			return val.Name
		}
	}
	return ""
}

// GetContinueOnError returns whether the job should continue on error.
func (j *Job) GetContinueOnError() bool {
	if j == nil || j.ContinueOnError == "" {
		return false
	}
	v, err := strconv.ParseBool(j.ContinueOnError)
	if err != nil {
		// Expression values are interpolated before parse by callers; treat unknown as false.
		return false
	}
	return v
}

// PermissionsSummary returns a short string describing declared permissions.
func PermissionsSummary(node yaml.Node) string {
	switch node.Kind {
	case yaml.ScalarNode:
		var s string
		if decodeNode(node, &s) {
			return s
		}
	case yaml.MappingNode:
		var m map[string]string
		if decodeNode(node, &m) {
			parts := make([]string, 0, len(m))
			for k, v := range m {
				parts = append(parts, fmt.Sprintf("%s:%s", k, v))
			}
			return strings.Join(parts, ",")
		}
	}
	return ""
}
