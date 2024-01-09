package spancheck

import (
	"strings"
	"testing"
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
			checks := parseChecks(strings.Split(flag, ","))
			if len(checks) != len(tc.checks) {
				t.Fatalf("Unexpected checks length=%d, want=%d", len(checks), len(tc.checks))
			}
			for i, check := range tc.checks {
				want := tc.checks[i]
				if check != want {
					t.Fatalf("Unexpected check=%+v, want=%+v", check, want)
				}
			}
		})
	}
}
