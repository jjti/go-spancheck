package spancheck

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"regexp"
	"slices"
	"strings"
)

// Check is a type of check that can be enabled or disabled.
type Check string

const (
	// EndCheck if enabled, checks that span.End() is called after span creation and before the function returns.
	EndCheck Check = "end"

	// SetStatusCheck if enabled, checks that `span.SetStatus(codes.Error, msg)` is called when returning an error.
	SetStatusCheck = "set-status"

	// RecordErrorCheck if enabled, checks that span.RecordError(err) is called when returning an error.
	RecordErrorCheck = "record-error"
)

var (
	// AllChecks is a list of all checks.
	AllChecks = []string{
		string(EndCheck),
		string(SetStatusCheck),
		string(RecordErrorCheck),
	}

	errNoChecks     = errors.New("no checks enabled")
	errInvalidCheck = errors.New("invalid check")
)

// Config is a configuration for the spancheck analyzer.
type Config struct {
	fs flag.FlagSet

	// EnabledChecks is a list of checks that are enabled.
	EnabledChecks []Check

	// ignoreChecksSignatures is a regex that, if matched, disables the
	// SetStatus and RecordError checks on error.
	ignoreChecksSignatures *regexp.Regexp

	// IgnoreChecksSignaturesSlice is a slice of strings that are turned into
	// the IgnoreSetStatusCheckSignatures regex.
	IgnoreChecksSignaturesSlice []string
}

// NewConfigFromFlags returns a new Config with default values and flags for CLI usage.
func NewConfigFromFlags() *Config {
	cfg := NewDefaultConfig()

	cfg.fs = flag.FlagSet{}

	// Set the list of checks to enable.
	checkDefault := []string{}
	for _, check := range cfg.EnabledChecks {
		checkDefault = append(checkDefault, string(check))
	}
	checkStrings := cfg.fs.String("checks", strings.Join(checkDefault, ","), fmt.Sprintf("comma-separated list of checks to enable (options: %v)", strings.Join(AllChecks, ", ")))
	checks, err := parseChecks(*checkStrings)
	if err != nil {
		log.Default().Fatalf("failed to parse checks: %v", err)
	}
	cfg.EnabledChecks = checks

	// Set the list of function signatures to ignore checks for.
	ignoreCheckSignatures := flag.String("ignore-check-signatures", "", "comma-separated list of regex for function signatures that disable checks on errors")
	cfg.ignoreChecksSignatures = parseSignatures(*ignoreCheckSignatures)

	return cfg
}

// NewDefaultConfig returns a new Config with default values.
func NewDefaultConfig() *Config {
	return &Config{
		EnabledChecks: []Check{EndCheck},
	}
}

// parseSignatures sets the Ignore*CheckSignatures regex from the string slices.
func (c *Config) parseSignatures() {
	if c.ignoreChecksSignatures == nil && len(c.IgnoreChecksSignaturesSlice) > 0 {
		c.ignoreChecksSignatures = createRegex(c.IgnoreChecksSignaturesSlice)
	}
}

func parseChecks(checksFlag string) ([]Check, error) {
	if checksFlag == "" {
		return nil, errNoChecks
	}

	checks := []Check{}
	for _, check := range strings.Split(checksFlag, ",") {
		check = strings.TrimSpace(check)
		if check == "" {
			continue
		}

		if !slices.Contains(AllChecks, check) {
			return nil, fmt.Errorf("%w: %s", errInvalidCheck, check)
		}

		checks = append(checks, Check(check))
	}

	if len(checks) == 0 {
		return nil, errNoChecks
	}

	return checks, nil
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
