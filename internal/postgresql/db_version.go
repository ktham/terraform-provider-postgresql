package postgresql

import (
	"fmt"
	"regexp"
)

const (
	PostgreSQL  string = "PostgreSQL"
	CockroachDB string = "CockroachDB"
)

type DbVersion struct {
	DbType    string // Different editions of Postgres may return different values
	DbVersion string
}

func ParseDbVersion(version string) (DbVersion, error) {
	patterns := []string{
		`(CockroachDB CCL) v([\d+.]+) .*`,
		`(PostgreSQL) ([\d+.]+) .*`,
	}

	for _, pattern := range patterns {
		regex := regexp.MustCompile(pattern)
		matches := regex.FindStringSubmatch(version)

		if matches != nil {
			return DbVersion{
				DbType:    matches[1],
				DbVersion: matches[2],
			}, nil
		}
	}
	return DbVersion{}, fmt.Errorf("output of `SELECT VERSION();`: '%s', didn't match expected patterns", version)
}
