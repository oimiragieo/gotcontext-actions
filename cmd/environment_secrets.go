package cmd

import (
	"strings"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

// loadEnvironmentSecretFiles parses name=path pairs into environment -> secrets maps.
func loadEnvironmentSecretFiles(specs []string) map[string]map[string]string {
	return loadEnvironmentMaps(specs, "env-secret-file", "secrets")
}

// loadEnvironmentVarFiles parses name=path pairs into environment -> vars maps.
func loadEnvironmentVarFiles(specs []string) map[string]map[string]string {
	return loadEnvironmentMaps(specs, "env-var-file", "vars")
}

func loadEnvironmentMaps(specs []string, flagName, kind string) map[string]map[string]string {
	out := map[string]map[string]string{}
	for _, spec := range specs {
		name, path, ok := strings.Cut(spec, "=")
		if !ok || name == "" || path == "" {
			log.Warnf("ignoring invalid --%s value %q (expected name=path)", flagName, spec)
			continue
		}
		vals, err := godotenv.Read(path)
		if err != nil {
			log.Warnf("failed to read environment %s file %q: %v", kind, path, err)
			continue
		}
		out[name] = vals
		log.Debugf("loaded %d %s for environment %q", len(vals), kind, name)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
