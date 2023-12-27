package spanlint_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/jjti/go-spanlint"
)

func Test(t *testing.T) {
	t.Parallel()

	for dir, config := range map[string]*spanlint.Config{
		"base": spanlint.DefaultConfig,
		"enableall": {
			EnableSetStatusCheck:   true,
			EnableRecordErrorCheck: true,
		},
	} {
		dir := dir
		t.Run(dir, func(t *testing.T) {
			analysistest.Run(t, "testdata/"+dir, spanlint.NewAnalyzer(config))
		})
	}
}
