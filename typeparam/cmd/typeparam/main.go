package main

import (
	"github.com/gostaticanalysis/examples/typeparam"
	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() { unitchecker.Main(typeparam.Analyzer) }
