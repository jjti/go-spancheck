package spancheck_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/jjti/go-spancheck"
)

func Test(t *testing.T) {
	t.Parallel()

	type configFactory func() *spancheck.Config

	for dir, configFactory := range map[string]configFactory{
		"base": spancheck.NewDefaultConfig,
		"disableerrorchecks": func() *spancheck.Config {
			cfg := spancheck.NewDefaultConfig()
			cfg.EnabledChecks = []string{
				spancheck.EndCheck.String(),
				spancheck.RecordErrorCheck.String(),
				spancheck.SetStatusCheck.String(),
			}
			cfg.IgnoreChecksSignaturesSlice = []string{"telemetry.Record", "recordErr"}

			return cfg
		},
		"enableall": func() *spancheck.Config {
			cfg := spancheck.NewDefaultConfig()
			cfg.EnabledChecks = []string{
				spancheck.EndCheck.String(),
				spancheck.RecordErrorCheck.String(),
				spancheck.SetStatusCheck.String(),
			}
			cfg.StartSpanMatchersSlice = append(cfg.StartSpanMatchersSlice,
				"util.TestStartTrace:opentelemetry",
				"enableall.testStartTrace:opencensus",
			)

			return cfg
		},
	} {
		dir := dir
		t.Run(dir, func(t *testing.T) {
			analysistest.Run(t, "testdata/"+dir, spancheck.NewAnalyzerWithConfig(configFactory()))
		})
	}
}
