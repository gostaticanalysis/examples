package typeparam_test

import (
	"testing"

	"github.com/gostaticanalysis/examples/typeparam"
	"github.com/gostaticanalysis/testutil"
	"golang.org/x/tools/go/analysis/analysistest"
)

// TestAnalyzer is a test for Analyzer.
func TestAnalyzer(t *testing.T) {
	testdata := testutil.WithModules(t, analysistest.TestData(), nil)
	analysistest.Run(t, testdata, typeparam.Analyzer, "a")
}
