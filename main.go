package main

import (
	"flag"
	"os"

	"github.com/pohzipohzi/piper/internal"
)

func main() {
	var (
		flagC string
		flagD string
		flagO bool
	)
	flag.StringVar(&flagC, "c", "", "the command to run")
	flag.StringVar(&flagD, "d", "", "(optional) the command to diff against")
	flag.BoolVar(&flagO, "o", false, "(optional) show output only")
	flag.Parse()
	if flagC == "" {
		flag.Usage()
		return
	}

	os.Exit(internal.Run(flagC, flagD, flagO))
}
