package spancheck

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_parseChecks(t *testing.T) {
	t.Parallel()

	for flag, tc := range map[string]struct {
		checks []Check
	}{
		"": {
			checks: []Check{},
		},
		"unknown": {
			checks: []Check{},
		},
		"end": {
			checks: []Check{EndCheck},
		},
		"end,record-error": {
			checks: []Check{EndCheck, RecordErrorCheck},
		},
		"end,record-error,set-status": {
			checks: []Check{EndCheck, RecordErrorCheck, SetStatusCheck},
		},
	} {
		flag, tc := flag, tc
		t.Run(flag, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			checks := parseChecks(strings.Split(flag, ","))
			r.Equal(tc.checks, checks)
		})
	}
}
