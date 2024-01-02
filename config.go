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

	// ignoreSetStatusCheckSignatures is a regex that, if matched, disables the
	// SetStatus check for a particular error.
	ignoreSetStatusCheckSignatures *regexp.Regexp

	// IgnoreSetStatusCheckSignaturesSlice is a slice of strings that are turned into
	// the IgnoreSetStatusCheckSignatures regex.
	IgnoreSetStatusCheckSignaturesSlice []string

	// EnableRecordErrorCheck enables the check for calling span.RecordError.
	// By default, this check is disabled.
	EnableRecordErrorCheck bool

	// ignoreRecordErrorCheckSignatures is a regex that, if matched, disables the
	// RecordError check for a particular error.
	ignoreRecordErrorCheckSignatures *regexp.Regexp

	// IgnoreRecordErrorCheckSignaturesSlice is a slice of strings that are turned into
	// the IgnoreRecordErrorCheckSignatures regex.
	IgnoreRecordErrorCheckSignaturesSlice []string
}

// NewConfigFromFlags returns a new Config with default values and flags for CLI usage.
func NewConfigFromFlags() *Config {
	cfg := newDefaultConfig()

	cfg.fs = flag.FlagSet{}
	cfg.fs.BoolVar(&cfg.DisableEndCheck, "disable-end-check", cfg.DisableEndCheck, "disable the check for calling span.End() after span creation")
	cfg.fs.BoolVar(&cfg.EnableAll, "enable-all", cfg.EnableAll, "enable all checks, overriding individual check flags")
	cfg.fs.BoolVar(&cfg.EnableSetStatusCheck, "enable-set-status-check", cfg.EnableSetStatusCheck, "enable check for a span.SetStatus(codes.Error, msg) call when returning an error")
	ignoreSetStatusCheckSignatures := flag.String(ignoreSetStatusCheckSignatures, "", "comma-separated list of regex for function signature that disable the span.SetStatus(codes.Error, msg) check on errors")
	flag.BoolVar(&cfg.EnableRecordErrorCheck, "enable-record-error-check", cfg.EnableRecordErrorCheck, "enable check for a span.RecordError(err) call when returning an error")
	ignoreRecordErrorCheckSignatures := flag.String(ignoreRecordErrorCheckSignatures, "", "comma-separated list of regex for function signature that disable the span.RecordError(err) check on errors")

	cfg.ignoreSetStatusCheckSignatures = parseSignatures(*ignoreSetStatusCheckSignatures)
	cfg.ignoreRecordErrorCheckSignatures = parseSignatures(*ignoreRecordErrorCheckSignatures)

	return cfg
}

func newDefaultConfig() *Config {
	return &Config{
		DisableEndCheck:                       false,
		EnableAll:                             false,
		EnableRecordErrorCheck:                false,
		EnableSetStatusCheck:                  false,
		ignoreRecordErrorCheckSignatures:      nil,
		IgnoreRecordErrorCheckSignaturesSlice: nil,
		ignoreSetStatusCheckSignatures:        nil,
		IgnoreSetStatusCheckSignaturesSlice:   nil,
	}
}

// parseSignatures sets the Ignore*CheckSignatures regex from the string slices.
func (c *Config) parseSignatures() {
	if c.ignoreRecordErrorCheckSignatures == nil && len(c.IgnoreRecordErrorCheckSignaturesSlice) > 0 {
		c.ignoreRecordErrorCheckSignatures = createRegex(c.IgnoreRecordErrorCheckSignaturesSlice)
	}

	if c.ignoreSetStatusCheckSignatures == nil && len(c.IgnoreSetStatusCheckSignaturesSlice) > 0 {
		c.ignoreSetStatusCheckSignatures = createRegex(c.IgnoreSetStatusCheckSignaturesSlice)
	}
}

func parseSignatures(sigFlag string) *regexp.Regexp {
	if sigFlag == "" {
		return nil
	}

	sigs := []string{}
	for _, sig := range strings.Split(sigFlag, ",") {
		sig = strings.TrimSpace(sig)
		if sig == "" {
			continue
		}

		sigs = append(sigs, sig)
	}

	return createRegex(sigs)
}

func createRegex(sigs []string) *regexp.Regexp {
	if len(sigs) == 0 {
		return nil
	}

	regex := fmt.Sprintf("(%s)", strings.Join(sigs, "|"))
	regexCompiled, err := regexp.Compile(regex)
	if err != nil {
		log.Default().Print("[WARN] failed to compile regex from signature flag", "regex", regex, "err", err)
		return nil
	}

	return regexCompiled
}
