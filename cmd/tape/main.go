package main

import (
	"fmt"
	"os"

	"github.com/hyperledger-twgc/tape/pkg/infra"
	"github.com/hyperledger-twgc/tape/pkg/infra/cmdImpl"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	loglevel    = "TAPE_LOGLEVEL"
	logfilename = "Tape.log"
)

var (
	app = kingpin.New("tape", "A performance test tool for Hyperledger Fabric")

	con            = app.Flag("config", "Path to config file").Short('c').String()
	num            = app.Flag("number", "Number of tx for shot").Short('n').Int()
	rate           = app.Flag("rate", "[Optional] Creates tx rate, default 0 as unlimited").Default("0").Float64()
	burst          = app.Flag("burst", "[Optional] Burst size for Tape, should bigger than rate").Default("1000").Int()
	signerNumber   = app.Flag("signers", "[Optional] signer parallel Number for Tape, default as 5").Default("5").Int()
	parallelNumber = app.Flag("parallel", "[Optional] parallel Number for Tape, default as 1").Default("1").Int()
	prometheus     = app.Flag("prometheus", "[Optional] prometheus enable or not").Default("false").Bool()

	run = app.Command("run", "Start the tape program").Default()

	version = app.Command("version", "Show version information")

	commitOnly = app.Command("commitOnly", "Start tape with commitOnly mode, starts dummy envelop for test orderer only")

	endorsementOnly = app.Command("endorsementOnly", "Start tape with endorsementOnly mode, starts endorsement and end")

	trafficOnly = app.Command("traffic", "Start tape with traffic mode")

	observerOnly = app.Command("observer", "Start tape with observer mode")
)

func main() {
	var err error

	logger := log.New()
	logger.SetLevel(log.WarnLevel)
	file, err := os.OpenFile(logfilename, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	logger.SetOutput(file)
	if customerLevel, customerSet := os.LookupEnv(loglevel); customerSet {
		if lvl, err := log.ParseLevel(customerLevel); err == nil {
			logger.SetLevel(lvl)
		}
	}

	fullCmd := kingpin.MustParse(app.Parse(os.Args[1:]))
	switch fullCmd {
	case version.FullCommand():
		fmt.Println(cmdImpl.GetVersionInfo())
	case commitOnly.FullCommand():
		checkArgs(rate, burst, signerNumber, parallelNumber, *con, logger)
		err = cmdImpl.Process(*con, *num, *burst, *signerNumber, *parallelNumber, *rate, *prometheus, logger, infra.COMMIT)
	case endorsementOnly.FullCommand():
		checkArgs(rate, burst, signerNumber, parallelNumber, *con, logger)
		err = cmdImpl.Process(*con, *num, *burst, *signerNumber, *parallelNumber, *rate, *prometheus, logger, infra.ENDORSEMENT)
	case run.FullCommand():
		checkArgs(rate, burst, signerNumber, parallelNumber, *con, logger)
		err = cmdImpl.Process(*con, *num, *burst, *signerNumber, *parallelNumber, *rate, *prometheus, logger, infra.FULLPROCESS)
	case trafficOnly.FullCommand():
		checkArgs(rate, burst, signerNumber, parallelNumber, *con, logger)
		err = cmdImpl.Process(*con, *num, *burst, *signerNumber, *parallelNumber, *rate, *prometheus, logger, infra.TRAFFIC)
	case observerOnly.FullCommand():
		checkArgs(rate, burst, signerNumber, parallelNumber, *con, logger)
		err = cmdImpl.Process(*con, *num, *burst, *signerNumber, *parallelNumber, *rate, *prometheus, logger, infra.OBSERVER)
	default:
		err = errors.Errorf("invalid command: %s", fullCmd)
	}

	if err != nil {
		logger.Error(err)
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}

func checkArgs(rate *float64, burst, signerNumber, parallel *int, con string, logger *log.Logger) {
	if len(con) == 0 {
		os.Stderr.WriteString("tape: error: required flag --config not provided, try --help")
		os.Exit(1)
	}
	if *rate < 0 {
		os.Stderr.WriteString("tape: error: rate must be zero (unlimited) or positive number\n")
		os.Exit(1)
	}
	if *burst < 1 {
		os.Stderr.WriteString("tape: error: burst at least 1\n")
		os.Exit(1)
	}
	if *signerNumber < 1 {
		os.Stderr.WriteString("tape: error: signerNumber at least 1\n")
		os.Exit(1)
	}
	if *parallel < 1 {
		os.Stderr.WriteString("tape: error: parallel at least 1\n")
		os.Exit(1)
	}

	if int64(*rate) > int64(*burst) {
		fmt.Printf("As rate %d is bigger than burst %d, real rate is burst\n", int64(*rate), int64(*burst))
	}

	logger.Infof("Will use rate %f as send rate\n", *rate)
	logger.Infof("Will use %d as burst\n", burst)
}
