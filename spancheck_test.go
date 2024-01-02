package spancheck_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/jjti/go-spancheck"
)

func Test(t *testing.T) {
	t.Parallel()

	for dir, config := range map[string]*spancheck.Config{
		"base": spancheck.NewConfigFromFlags(),
		"disableerrorchecks": {
			EnabledChecks: []spancheck.Check{
				spancheck.EndCheck,
				spancheck.RecordErrorCheck,
				spancheck.SetStatusCheck,
			},
			IgnoreChecksSignaturesSlice: []string{"telemetry.Record", "recordErr"},
		},
		"enableall": {
			EnabledChecks: []spancheck.Check{
				spancheck.EndCheck,
				spancheck.RecordErrorCheck,
				spancheck.SetStatusCheck,
			},
		},
	} {
		dir := dir
		t.Run(dir, func(t *testing.T) {
			analysistest.Run(t, "testdata/"+dir, spancheck.NewAnalyzerWithConfig(config))
		})
	}
}
