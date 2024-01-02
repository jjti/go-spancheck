package spancheck

import (
	"flag"
	"fmt"
	"log"
	"regexp"
	"slices"
	"strings"
)

// Check is a type of check that can be enabled or disabled.
type Check int

const (
	// EndCheck if enabled, checks that span.End() is called after span creation and before the function returns.
	EndCheck Check = iota

	// SetStatusCheck if enabled, checks that `span.SetStatus(codes.Error, msg)` is called when returning an error.
	SetStatusCheck

	// RecordErrorCheck if enabled, checks that span.RecordError(err) is called when returning an error.
	RecordErrorCheck
)

func (c Check) String() string {
	switch c {
	case EndCheck:
		return "end"
	case SetStatusCheck:
		return "set-status"
	case RecordErrorCheck:
		return "record-error"
	default:
		return ""
	}
}

var (
	// Checks is a list of all checks by name.
	Checks = map[string]Check{
		EndCheck.String():         EndCheck,
		SetStatusCheck.String():   SetStatusCheck,
		RecordErrorCheck.String(): RecordErrorCheck,
	}
)

// Config is a configuration for the spancheck analyzer.
type Config struct {
	fs flag.FlagSet

	// EnabledChecks is a list of checks to enable by name.
	EnabledChecks []string

	// IgnoreChecksSignaturesSlice is a slice of strings that are turned into
	// the IgnoreSetStatusCheckSignatures regex.
	IgnoreChecksSignaturesSlice []string

	endCheckEnabled    bool
	setStatusEnabled   bool
	recordErrorEnabled bool

	// ignoreChecksSignatures is a regex that, if matched, disables the
	// SetStatus and RecordError checks on error.
	ignoreChecksSignatures *regexp.Regexp
}

// NewConfigFromFlags returns a new Config with default values and flags for CLI usage.
func NewConfigFromFlags() *Config {
	cfg := NewDefaultConfig()

	cfg.fs = flag.FlagSet{}

	// Set the list of checks to enable.
	checkOptions := []string{}
	for check := range Checks {
		checkOptions = append(checkOptions, check)
	}
	checkStrings := cfg.fs.String("checks", "end", fmt.Sprintf("comma-separated list of checks to enable (options: %v)", strings.Join(checkOptions, ", ")))
	cfg.EnabledChecks = strings.Split(*checkStrings, ",")

	// Set the list of function signatures to ignore checks for.
	ignoreCheckSignatures := flag.String("ignore-check-signatures", "", "comma-separated list of regex for function signatures that disable checks on errors")
	cfg.ignoreChecksSignatures = parseSignatures(*ignoreCheckSignatures)

	return cfg
}

// NewDefaultConfig returns a new Config with default values.
func NewDefaultConfig() *Config {
	return &Config{
		EnabledChecks: []string{EndCheck.String()},
	}
}

// finalize parses checks and signatures from the public string slices of Config.
func (c *Config) finalize() {
	c.parseSignatures()

	checks := parseChecks(c.EnabledChecks)
	c.endCheckEnabled = slices.Contains(checks, EndCheck)
	c.setStatusEnabled = slices.Contains(checks, SetStatusCheck)
	c.recordErrorEnabled = slices.Contains(checks, RecordErrorCheck)
}

// parseSignatures sets the Ignore*CheckSignatures regex from the string slices.
func (c *Config) parseSignatures() {
	if c.ignoreChecksSignatures == nil && len(c.IgnoreChecksSignaturesSlice) > 0 {
		c.ignoreChecksSignatures = createRegex(c.IgnoreChecksSignaturesSlice)
	}
}

func parseChecks(checksSlice []string) []Check {
	if len(checksSlice) == 0 {
		return nil
	}

	checks := []Check{}
	for _, check := range checksSlice {
		checkName := strings.TrimSpace(check)
		if checkName == "" {
			continue
		}

		check, ok := Checks[checkName]
		if !ok {
			continue
		}

		checks = append(checks, check)
	}

	return checks
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
