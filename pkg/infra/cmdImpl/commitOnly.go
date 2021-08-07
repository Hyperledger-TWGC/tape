package cmdImpl

import (
	"fmt"
	"tape/pkg/infra/observer"
	"tape/pkg/infra/trafficGenerator"
	"time"

	log "github.com/sirupsen/logrus"
)

func ProcessCommitOnly(configPath string, num int, burst, signerNumber int, rate float64, logger *log.Logger) error {
	/*** variables ***/
	cmdConfig, err := CreateCmd(configPath, num, burst, signerNumber, rate)
	if err != nil {
		return err
	}
	defer cmdConfig.cancel()
	/*** workers ***/
	crypto, err := cmdConfig.Config.LoadCrypto()
	if err != nil {
		return err
	}
	Observer_workers, Observers, err := observer.CreateCommitObserverWorkers(cmdConfig.Config.Channel, cmdConfig.Config.Orderer, crypto, logger, num, cmdConfig.ErrorCh, cmdConfig.FinishCh)
	if err != nil {
		return err
	}
	generator_workers, err := trafficGenerator.CreateMockGeneratorWorkers(cmdConfig.Ctx, cmdConfig.Crypto, cmdConfig.Envs, cmdConfig.Config, num, burst, signerNumber, rate, logger, cmdConfig.ErrorCh)
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
			duration := time.Since(Observers.GetTime())
			logger.Infof("Completed processing transactions.")
			fmt.Printf("tx: %d, duration: %+v, tps: %f\n", num, duration, float64(num)/duration.Seconds())
			return nil
		}
	}
}
