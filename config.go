package spancheck

import (
	"flag"
	"fmt"
	"log"
	"regexp"
	"strings"
)

var (
	ignoreSetStatusCheckSignatures   = "ignore-set-status-check-signatures"
	ignoreRecordErrorCheckSignatures = "ignore-record-error-check-signatures"
)

// Config is a configuration for the spancheck analyzer.
type Config struct {
	fs flag.FlagSet

	// EnableAll enables all checks and takes precedence over other fields like
	// DisableEndCheck. Ignore*CheckSignatures still apply.
	EnableAll bool

	// DisableEndCheck enables the check for calling span.End().
	DisableEndCheck bool

	// EnableSetStatusCheck enables the check for calling span.SetStatus.
	EnableSetStatusCheck bool

	// IgnoreSetStatusCheckSignatures is a regex that, if matched, disables the
	// SetStatus check for a particular error.
	IgnoreSetStatusCheckSignatures *regexp.Regexp

	// EnableRecordErrorCheck enables the check for calling span.RecordError.
	// By default, this check is disabled.
	EnableRecordErrorCheck bool

	// IgnoreRecordErrorCheckSignatures is a regex that, if matched, disables the
	// RecordError check for a particular error.
	IgnoreRecordErrorCheckSignatures *regexp.Regexp
}

// NewConfig returns a new Config with default values and flags for cli usage.
func NewConfig() *Config {
	cfg := &Config{
		fs:                               flag.FlagSet{},
		DisableEndCheck:                  false,
		EnableAll:                        false,
		EnableSetStatusCheck:             false,
		IgnoreSetStatusCheckSignatures:   nil,
		EnableRecordErrorCheck:           false,
		IgnoreRecordErrorCheckSignatures: nil,
	}

	cfg.fs.BoolVar(&cfg.DisableEndCheck, "disable-end-check", cfg.DisableEndCheck, "disable the check for calling span.End() after span creation")
	cfg.fs.BoolVar(&cfg.EnableAll, "enable-all", cfg.EnableAll, "enable all checks, overriding individual check flags")
	cfg.fs.BoolVar(&cfg.EnableSetStatusCheck, "enable-set-status-check", cfg.EnableSetStatusCheck, "enable check for a span.SetStatus(codes.Error, msg) call when returning an error")
	ignoreSetStatusCheckSignatures := flag.String(ignoreSetStatusCheckSignatures, "", "comma-separated list of regex for function signature that disable the span.SetStatus(codes.Error, msg) check on errors")
	flag.BoolVar(&cfg.EnableRecordErrorCheck, "enable-record-error-check", cfg.EnableRecordErrorCheck, "enable check for a span.RecordError(err) call when returning an error")
	ignoreRecordErrorCheckSignatures := flag.String(ignoreRecordErrorCheckSignatures, "", "comma-separated list of regex for function signature that disable the span.RecordError(err) check on errors")
	// Set and parse the flags.

	// Parse the signatures.
	cfg.IgnoreSetStatusCheckSignatures = parseSignatures(ignoreSetStatusCheckSignatures)
	cfg.IgnoreRecordErrorCheckSignatures = parseSignatures(ignoreRecordErrorCheckSignatures)

	return cfg
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
