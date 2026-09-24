package cmd

import (
	"strings"
)

func (i *Input) newPlatforms() map[string]string {
	platforms := map[string]string{
		"ubuntu-latest": "node:22-bookworm-slim",
		"ubuntu-24.04":  "node:22-bookworm-slim",
		"ubuntu-22.04":  "node:20-bookworm-slim",
		"ubuntu-20.04":  "node:20-bullseye-slim",
	}

	for _, p := range i.platforms {
		pParts := strings.Split(p, "=")
		if len(pParts) == 2 {
			platforms[strings.ToLower(pParts[0])] = pParts[1]
		}
	}
	return platforms
}
