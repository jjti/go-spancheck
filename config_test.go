package spancheck

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_parseChecks(t *testing.T) {
	t.Parallel()

	for flag, tc := range map[string]struct {
		checks []Check
		err    error
	}{
		"": {
			err: errNoChecks,
		},
		"unknown": {
			err: errInvalidCheck,
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

			checks, err := parseChecks(flag)
			if tc.err != nil {
				r.ErrorIs(err, tc.err)
				return
			}
			r.NoError(err)
			r.Equal(tc.checks, checks)
		})
	}
}
