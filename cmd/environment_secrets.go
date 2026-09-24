package cmd

import (
	"strings"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

// loadEnvironmentSecretFiles parses name=path pairs into environment -> secrets maps.
func loadEnvironmentSecretFiles(specs []string) map[string]map[string]string {
	out := map[string]map[string]string{}
	for _, spec := range specs {
		name, path, ok := strings.Cut(spec, "=")
		if !ok || name == "" || path == "" {
			log.Warnf("ignoring invalid --env-secret-file value %q (expected name=path)", spec)
			continue
		}
		vals, err := godotenv.Read(path)
		if err != nil {
			log.Warnf("failed to read environment secret file %q: %v", path, err)
			continue
		}
		out[name] = vals
		log.Debugf("loaded %d secrets for environment %q", len(vals), name)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
