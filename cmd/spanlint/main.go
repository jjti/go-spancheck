package main

import (
	"flag"
	"fmt"
	"log"
	"regexp"
	"strings"

	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/jjti/go-spanlint"
)

var (
	ignoreSetStatusCheckSignatures   = "ignore-set-status-check-signatures"
	ignoreRecordErrorCheckSignatures = "ignore-record-error-check-signatures"
)

func main() {
	// Set and parse the flags.
	config := spanlint.NewDefaultConfig()
	flag.BoolVar(&config.DisableEndCheck, "disable-end-check", config.DisableEndCheck, "disable the check for calling span.End() after span creation")
	flag.BoolVar(&config.EnableAll, "enable-all", config.EnableAll, "enable all checks, overriding individual check flags")
	flag.BoolVar(&config.EnableSetStatusCheck, "enable-set-status-check", config.EnableSetStatusCheck, "enable check for a span.SetStatus(codes.Error, msg) call when returning an error")
	ignoreSetStatusCheckSignatures := flag.String(ignoreSetStatusCheckSignatures, "", "comma-separated list of regex for function signature that disable the span.SetStatus(codes.Error, msg) check on errors")
	flag.BoolVar(&config.EnableRecordErrorCheck, "enable-record-error-check", config.EnableRecordErrorCheck, "enable check for a span.RecordError(err) call when returning an error")
	ignoreRecordErrorCheckSignatures := flag.String(ignoreRecordErrorCheckSignatures, "", "comma-separated list of regex for function signature that disable the span.RecordError(err) check on errors")
	flag.Parse()

	// Parse the signatures.
	config.IgnoreSetStatusCheckSignatures = parseSignatures(ignoreSetStatusCheckSignatures)
	config.IgnoreRecordErrorCheckSignatures = parseSignatures(ignoreRecordErrorCheckSignatures)

	// Run the analyzer.
	singlechecker.Main(spanlint.NewAnalyzer(config))
}

func parseSignatures(sigFlag *string) *regexp.Regexp {
	if sigFlag == nil || *sigFlag == "" {
		return nil
	}

	sigs := []string{}
	for _, sig := range strings.Split(*sigFlag, ",") {
		sig = strings.TrimSpace(sig)
		if sig == "" {
			log.Fatalf("empty disable-error-checks-signature value: %q", sig)
		}

		sigs = append(sigs, sig)
	}

	if len(sigs) == 0 {
		return nil
	}

	regex := fmt.Sprintf("(%s)", strings.Join(sigs, "|"))
	regexCompiled, err := regexp.Compile(regex)
	if err != nil {
		log.Fatalf("failed to compile signatures: %v", err)
	}

	return regexCompiled
}
