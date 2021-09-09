package cmdImpl

import (
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

func Process(configPath string, num int, burst, signerNumber, parallel int, rate float64, logger *log.Logger, processmod int) error {
	/*** signal ***/
	c := make(chan os.Signal)
	/*** variables ***/
	cmdConfig, err := CreateCmd(configPath, num, burst, signerNumber, parallel, rate, logger)
	if err != nil {
		return err
	}
	defer cmdConfig.cancel()
	/*** workers ***/
	Observer_workers, Observers, err := cmdConfig.Observerfactory.CreateObserverWorkers(processmod)
	if err != nil {
		return err
	}
	generator_workers, err := cmdConfig.Generator.CreateGeneratorWorkers(processmod)
	if err != nil {
		return err
	}
	/*** start workers ***/
	for _, worker := range Observer_workers {
		go worker.Start()
	}
	for _, worker := range generator_workers {
		go worker.Start()
	}
	/*** waiting for complete ***/
	total := num * parallel
	for {
		select {
		case err = <-cmdConfig.ErrorCh:
			return err
		case <-cmdConfig.FinishCh:
			duration := time.Since(Observers.GetTime())
			logger.Infof("Completed processing transactions.")
			fmt.Printf("tx: %d, duration: %+v, tps: %f\n", total, duration, float64(total)/duration.Seconds())
			return nil
		case s := <-c:
			fmt.Printf("Stopped by signal received.", s)
			fmt.Printf("Completed processing transactions")
			fmt.Printf("If you stopped by ctrl+c and used in distrubted way")
			fmt.Printf("Please try to use dash board (https://github.com/SamYuan1990/HLF_GrafanaDashBoard) for monitoring")
			return nil
		}
	}
}
