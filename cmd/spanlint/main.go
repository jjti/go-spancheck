package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/jjti/go-spanlint"
)

func main() {
	singlechecker.Main(spanlint.Analyzer)
}
