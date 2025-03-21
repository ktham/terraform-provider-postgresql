package postgresql

import (
	"fmt"
	"regexp"
)

func ParsePostgresVersion(version string) (string, error) {
	patterns := []string{
		`(PostgreSQL) ([\d+.]+) .*`,
	}

	for _, pattern := range patterns {
		regex := regexp.MustCompile(pattern)
		matches := regex.FindStringSubmatch(version)

		if matches != nil {
			return matches[2], nil
		}
	}
	return "", fmt.Errorf("output of `SELECT VERSION();`: '%s', didn't match expected patterns", version)
}
