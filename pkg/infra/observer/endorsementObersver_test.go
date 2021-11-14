package observer_test

import (
	"tape/pkg/infra/basic"
	"tape/pkg/infra/observer"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

var _ = Describe("EndorsementObersver", func() {

	It("Should work with number limit", func() {
		envs := make(chan *basic.TracingEnvelope, 1024)
		finishCh := make(chan struct{})
		logger := log.New()
		instance := observer.CreateEndorseObserver(envs, 2, finishCh, logger)

		go instance.Start()

		envs <- &basic.TracingEnvelope{}
		Consistently(finishCh).ShouldNot(BeClosed())
		envs <- &basic.TracingEnvelope{}
		Eventually(finishCh).Should(BeClosed())
	})

	It("Should work with number limit", func() {
		envs := make(chan *basic.TracingEnvelope, 1024)
		finishCh := make(chan struct{})
		logger := log.New()
		instance := observer.CreateEndorseObserver(envs, 1, finishCh, logger)

		go instance.Start()

		envs <- &basic.TracingEnvelope{}
		Eventually(finishCh).Should(BeClosed())
	})

	It("Should work without number limit", func() {
		envs := make(chan *basic.TracingEnvelope, 1024)
		finishCh := make(chan struct{})
		logger := log.New()
		instance := observer.CreateEndorseObserver(envs, 0, finishCh, logger)

		go instance.Start()

		envs <- &basic.TracingEnvelope{}
		Consistently(finishCh).ShouldNot(BeClosed())
		envs <- &basic.TracingEnvelope{}
		Eventually(finishCh).ShouldNot(BeClosed())
	})

})
