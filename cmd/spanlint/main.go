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
	ignoreSetStatusCheckSignatures := flag.String(ignoreSetStatusCheckSignatures, "", "comma-separated list of function signature regex that disable the span.SetStatus(codes.Error, msg) check on errors")
	flag.BoolVar(&config.EnableRecordErrorCheck, "enable-record-error-check", config.EnableRecordErrorCheck, "enable check for a span.RecordError(err) call when returning an error")
	ignoreRecordErrorCheckSignatures := flag.String(ignoreRecordErrorCheckSignatures, "", "comma-separated list of function signature regex that disable the span.RecordError(err) check on errors")
	flag.Parse()

	// Parse the signatures.
	config.IgnoreSetStatusCheckSignatures = parseSignatures(ignoreSetStatusCheckSignatures)
	config.IgnoreRecordErrorCheckSignatures = parseSignatures(ignoreRecordErrorCheckSignatures)

	// Run the analyzer.
	singlechecker.Main(spanlint.NewAnalyzer(config))
}

func parseSignatures(sigs *string) *regexp.Regexp {
	if sigs == nil || *sigs == "" {
		return nil
	}

	disableSigs := []string{}
	sigColumns := strings.Split(*sigs, ",")
	for _, sig := range sigColumns {
		sig = strings.TrimSpace(sig)
		if sig == "" {
			log.Fatalf("empty disable-error-checks-signature value: %q", sig)
		}

		disableSigs = append(disableSigs, sig)
	}

	if len(disableSigs) == 0 {
		return nil
	}

	regex := fmt.Sprintf("(%s)", strings.Join(sigColumns, "|"))
	regexCompiled, err := regexp.Compile(regex)
	if err != nil {
		log.Fatalf("failed to compile signatures: %v", err)
	}

	return regexCompiled
}
