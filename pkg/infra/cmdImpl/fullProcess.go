package cmdImpl

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hyperledger-twgc/tape/pkg/infra"
	"github.com/hyperledger-twgc/tape/pkg/infra/basic"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

func Process(configPath string, num int, burst, signerNumber, parallel int, rate float64, prometheusOpt bool, logger *log.Logger, processmod int) error {
	/*** signal ***/
	c := make(chan os.Signal, 1)

	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	/*** variables ***/
	cmdConfig, err := CreateCmd(configPath, num, burst, signerNumber, parallel, rate, logger)
	if err != nil {
		return err
	}
	defer cmdConfig.cancel()
	defer cmdConfig.Closer.Close()
	var Observer_workers []infra.Worker
	var Observers infra.ObserverWorker
	basic.SetMod(processmod)
	/*** workers ***/
	if processmod != infra.TRAFFIC {
		Observer_workers, Observers, err = cmdConfig.Observerfactory.CreateObserverWorkers(processmod)
		if err != nil {
			return err
		}
	}
	var generator_workers []infra.Worker
	if processmod != infra.OBSERVER {
		if processmod == infra.TRAFFIC {
			generator_workers, err = cmdConfig.Generator.CreateGeneratorWorkers(processmod - 1)
			if err != nil {
				return err
			}
		} else {
			generator_workers, err = cmdConfig.Generator.CreateGeneratorWorkers(processmod)
			if err != nil {
				return err
			}
		}
	}

	var transactionlatency, readlatency *prometheus.SummaryVec
	/*** start prometheus ***/
	if prometheusOpt {
		go func() {
			fmt.Println("start prometheus")
			http.Handle("/metrics", promhttp.Handler())
			server := &http.Server{Addr: ":8080", Handler: nil}
			err := server.ListenAndServe()
			if err != nil {
				cmdConfig.ErrorCh <- err
			}
		}()

		transactionlatency = prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Name:       "transaction_latency_duration",
				Help:       "Transaction latency distributions.",
				Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
			},
			[]string{"transactionlatency"},
		)

		readlatency = prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Name:       "read_latency_duration",
				Help:       "Read latency distributions.",
				Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
			},
			[]string{"readlatency"},
		)

		prometheus.MustRegister(transactionlatency)
		prometheus.MustRegister(readlatency)

		basic.InitLatencyMap(transactionlatency, readlatency, processmod, prometheusOpt)
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
			fmt.Println("For FAQ, please check https://github.com/Hyperledger-TWGC/tape/wiki/FAQ")
			return err
		case <-cmdConfig.FinishCh:
			duration := time.Since(Observers.GetTime())
			logger.Infof("Completed processing transactions.")
			fmt.Printf("tx: %d, duration: %+v, tps: %f\n", total, duration, float64(total)/duration.Seconds())
			return nil
		case s := <-c:
			fmt.Println("Stopped by signal received" + s.String())
			fmt.Println("Completed processing transactions")
			return nil
		}
	}
}
