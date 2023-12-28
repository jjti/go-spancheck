package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/jjti/go-spancheck"
)

func main() {
	singlechecker.Main(spancheck.NewAnalyzer())
}
