package spancheck_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/jjti/go-spancheck"
)

func Test(t *testing.T) {
	t.Parallel()

	for dir, config := range map[string]*spancheck.Config{
		"base": spancheck.NewDefaultConfig(),
		"disableerrorchecks": {
			EnabledChecks: []string{
				spancheck.EndCheck.String(),
				spancheck.RecordErrorCheck.String(),
				spancheck.SetStatusCheck.String(),
			},
			IgnoreChecksSignaturesSlice: []string{"telemetry.Record", "recordErr"},
		},
		"enableall": {
			EnabledChecks: []string{
				spancheck.EndCheck.String(),
				spancheck.RecordErrorCheck.String(),
				spancheck.SetStatusCheck.String(),
			},
		},
	} {
		dir := dir
		t.Run(dir, func(t *testing.T) {
			analysistest.Run(t, "testdata/"+dir, spancheck.NewAnalyzerWithConfig(config))
		})
	}
}
