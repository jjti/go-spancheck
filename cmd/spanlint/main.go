package main

import (
	"flag"

	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/jjti/go-spanlint"
)

func main() {
	config := spanlint.DefaultConfig
	flag.BoolVar(&config.DisableEndCheck, "disable-end-check", config.DisableEndCheck, "disable the check for calling span.End() after span creation")
	flag.BoolVar(&config.EnableSetStatusCheck, "enable-set-status-check", config.EnableSetStatusCheck, "enable the check for calling span.SetStatus(codes.Error, msg) when returning an error")
	flag.BoolVar(&config.EnableRecordErrorCheck, "enable-record-error-check", config.EnableRecordErrorCheck, "enable the check for calling span.RecordError(err) when returning an error")
	flag.Parse()

	singlechecker.Main(spanlint.NewAnalyzer(config))
}
