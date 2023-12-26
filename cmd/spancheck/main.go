package main

import (
	"github.com/jjti/go-spancheck"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(spancheck.Analyzer)
}
