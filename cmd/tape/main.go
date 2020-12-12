package main

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"tape/pkg/infra"
)

const (
	loglevel = "TAPE_LOGLEVEL"
)

var (
	app = kingpin.New("tape", "A performance test tool for Hyperledger Fabric")

	run     = app.Command("run", "Start the tape program").Default()
	con     = run.Flag("config", "Path to config file").Required().Short('c').String()
	num     = run.Flag("number", "Number of tx for shot").Required().Short('n').Int()
	version = app.Command("version", "Show version information")
)

func main() {
	var err error

	logger := log.New()
	logger.SetLevel(log.WarnLevel)
	if customerLevel, customerSet := os.LookupEnv(loglevel); customerSet {
		if lvl, err := log.ParseLevel(customerLevel); err == nil {
			logger.SetLevel(lvl)
		}
	}

	fullCmd := kingpin.MustParse(app.Parse(os.Args[1:]))
	switch fullCmd {
	case version.FullCommand():
		fmt.Printf(infra.GetVersionInfo())
	case run.FullCommand():
		err = infra.Process(*con, *num, logger)
	default:
		err = errors.Errorf("invalid command: %s", fullCmd)
	}

	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}
	os.Exit(0)
}
