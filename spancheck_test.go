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
			EnableSetStatusCheck:                  true,
			IgnoreSetStatusCheckSignaturesSlice:   []string{"telemetry.Record"},
			EnableRecordErrorCheck:                true,
			IgnoreRecordErrorCheckSignaturesSlice: []string{"telemetry.Record"},
		},
		"enableall": {
			EnableAll: true,
		},
		"enablechecks": {
			EnableSetStatusCheck:   true,
			EnableRecordErrorCheck: true,
		},
	} {
		dir := dir
		t.Run(dir, func(t *testing.T) {
			analysistest.Run(t, "testdata/"+dir, spancheck.NewAnalyzerWithConfig(config))
		})
	}
}
