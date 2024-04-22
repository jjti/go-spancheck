package main

import (
	"flag"
	"fmt"
	"strings"

	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/jjti/go-spancheck"
)

func main() {
	// Set the list of checks to enable.
	checkOptions := []string{}
	for check := range spancheck.Checks {
		checkOptions = append(checkOptions, check)
	}

	checkStrings := ""
	flag.StringVar(&checkStrings, "checks", "end", fmt.Sprintf("comma-separated list of checks to enable (options: %v)", strings.Join(checkOptions, ", ")))

	// Set the list of function signatures to ignore checks for.
	ignoreCheckSignatures := ""
	flag.StringVar(&ignoreCheckSignatures, "ignore-check-signatures", "", "comma-separated list of regex for function signatures that disable checks on errors")

	extraStartSpanSignatures := ""
	flag.StringVar(&extraStartSpanSignatures, "extra-start-span-signatures", "", "comma-separated list of regex:telemetry-type for function signatures that indicate the start of a span")

	flag.Parse()

	cfg := spancheck.NewDefaultConfig()
	cfg.EnabledChecks = strings.Split(checkStrings, ",")
	cfg.IgnoreChecksSignaturesSlice = strings.Split(ignoreCheckSignatures, ",")

	if extraStartSpanSignatures != "" {
		cfg.StartSpanMatchersSlice = append(cfg.StartSpanMatchersSlice, strings.Split(extraStartSpanSignatures, ",")...)
	}

	singlechecker.Main(spancheck.NewAnalyzerWithConfig(cfg))
}
