package cmdImpl

import (
	"fmt"
	"tape/pkg/infra/observer"
	"tape/pkg/infra/trafficGenerator"
	"time"

	log "github.com/sirupsen/logrus"
)

func ProcessEndorsementOnly(configPath string, num int, burst, signerNumber int, rate float64, logger *log.Logger) error {
	/*** variables ***/
	cmdConfig, err := CreateCmd(configPath, num, burst, signerNumber, rate)
	if err != nil {
		return err
	}
	defer cmdConfig.cancel()
	/*** workers ***/
	Observer_workers, Observer, err := observer.CreateEndorsementObserverWorkers(cmdConfig.Processed, cmdConfig.Ctx, cmdConfig.FinishCh, num, cmdConfig.ErrorCh, logger)
	if err != nil {
		return err
	}
	generator_workers, err := trafficGenerator.CreateGeneratorWorkers(cmdConfig.Ctx, cmdConfig.Crypto, cmdConfig.Raw, cmdConfig.Signed, cmdConfig.Envs, cmdConfig.Processed, cmdConfig.Config, num, burst, signerNumber, rate, logger, cmdConfig.ErrorCh)
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
	for {
		select {
		case err = <-cmdConfig.ErrorCh:
			return err
		case <-cmdConfig.FinishCh:
			duration := time.Since(Observer.GetTime())
			logger.Infof("Completed processing transactions.")
			fmt.Printf("tx: %d, duration: %+v, tps: %f\n", num, duration, float64(num)/duration.Seconds())
			return nil
		}
	}
}
