package spancheck_test

import (
	"testing"

	"github.com/jjti/go-spancheck"
	"golang.org/x/tools/go/analysis/analysistest"
)

func Test(t *testing.T) {
	testdata := analysistest.TestData()

	analysistest.Run(t, testdata, spancheck.Analyzer)
}
