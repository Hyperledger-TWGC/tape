package observer_test

import (
	"sync"

	"github.com/hyperledger-twgc/tape/pkg/infra/basic"
	"github.com/hyperledger-twgc/tape/pkg/infra/observer"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

var _ = Describe("EndorsementObersver", func() {

	BeforeEach(func() {
		log.New()

	})

	It("Should work with number limit", func() {
		envs := make(chan *basic.TracingEnvelope, 1024)
		finishCh := make(chan struct{})
		logger := log.New()
		var once sync.Once
		instance := observer.CreateEndorseObserver(envs, 2, finishCh, &once, logger)

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
		var once sync.Once
		instance := observer.CreateEndorseObserver(envs, 1, finishCh, &once, logger)

		go instance.Start()

		envs <- &basic.TracingEnvelope{}
		Eventually(finishCh).Should(BeClosed())
	})

	It("Should work without number limit", func() {
		envs := make(chan *basic.TracingEnvelope, 1024)
		finishCh := make(chan struct{})
		logger := log.New()
		var once sync.Once
		instance := observer.CreateEndorseObserver(envs, 0, finishCh, &once, logger)

		go instance.Start()

		envs <- &basic.TracingEnvelope{}
		Consistently(finishCh).ShouldNot(BeClosed())
		envs <- &basic.TracingEnvelope{}
		Eventually(finishCh).ShouldNot(BeClosed())
	})

})
