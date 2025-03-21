package postgresql

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseDbVersion(t *testing.T) {
	testCases := []struct {
		testName       string
		input          string
		expectedOutput string
	}{
		{
			testName:       "PG test (Docker 15.6)",
			input:          "PostgreSQL 15.6 (Debian 15.6-1.pgdg120+2) on aarch64-unknown-linux-gnu, compiled by gcc (Debian 12.2.0-14) 12.2.0, 64-bit",
			expectedOutput: "15.6",
		},
		{
			testName:       "PG test (AWS RDS PostgreSQL 15.5)",
			input:          "PostgreSQL 15.5 on x86_64-pc-linux-gnu, compiled by gcc (GCC) 7.3.1 20180712 (Red Hat 7.3.1-12), 64-bit",
			expectedOutput: "15.5",
		},
		{
			testName:       "PG test (AWS RDS Aurora PostgreSQL 15.6)",
			input:          "PostgreSQL 15.6 on aarch64-unknown-linux-gnu, compiled by aarch64-unknown-linux-gnu-gcc (GCC) 9.5.0, 64-bit",
			expectedOutput: "15.6",
		},
		{
			testName:       "PG test (Docker 17.4)",
			input:          "PostgreSQL 17.4 (Debian 17.4-1.pgdg120+2) on aarch64-unknown-linux-gnu, compiled by gcc (Debian 12.2.0-14) 12.2.0, 64-bit",
			expectedOutput: "17.4",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			t.Parallel()

			actualOutput, err := ParsePostgresVersion(testCase.input)

			assert.NoError(t, err)
			assert.Equal(t, testCase.expectedOutput, actualOutput)
		})
	}
}
