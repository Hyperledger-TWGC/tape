package trafficGenerator_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/Hyperledger-TWGC/tape/e2e"
	"github.com/Hyperledger-TWGC/tape/pkg/infra/basic"
	"github.com/Hyperledger-TWGC/tape/pkg/infra/trafficGenerator"
)

func benchmarkProposalRandom(b *testing.B, arg string) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		trafficGenerator.ConvertString(arg)
	}
	b.StopTimer()
}

func BenchmarkProposalRandomTest1(b *testing.B) {
	benchmarkProposalRandom(b, "data")
}

func BenchmarkProposalRandomTest2(b *testing.B) {
	benchmarkProposalRandom(b, "randomString1")
}

func BenchmarkProposalRandomTest3(b *testing.B) {
	benchmarkProposalRandom(b, "{\"key\":\"randomNumber1_50\",\"key1\":\"randomNumber1_20\"}")
}

func BenchmarkProposalRandomTest4(b *testing.B) {
	benchmarkProposalRandom(b, "{\"k1\":\"uuid\",\"key2\":\"randomNumber10000_20000\",\"keys\":\"randomString10\"}")
}

func BenchmarkFackEnvelopTest(b *testing.B) {
	errorCh := make(chan error, 1000)
	envs := make(chan *basic.TracingEnvelope, 1000)
	tmpDir, _ := ioutil.TempDir("", "tape-")
	mtlsCertFile, _ := ioutil.TempFile(tmpDir, "mtls-*.crt")
	mtlsKeyFile, _ := ioutil.TempFile(tmpDir, "mtls-*.key")
	PolicyFile, _ := ioutil.TempFile(tmpDir, "policy")

	e2e.GeneratePolicy(PolicyFile)
	e2e.GenerateCertAndKeys(mtlsKeyFile, mtlsCertFile)
	mtlsCertFile.Close()
	mtlsKeyFile.Close()
	PolicyFile.Close()
	configFile, _ := ioutil.TempFile(tmpDir, "config*.yaml")
	configValue := e2e.Values{
		PrivSk:          mtlsKeyFile.Name(),
		SignCert:        mtlsCertFile.Name(),
		Mtls:            false,
		PeersAddrs:      nil,
		OrdererAddr:     "",
		CommitThreshold: 1,
		PolicyFile:      PolicyFile.Name(),
	}
	e2e.GenerateConfigFile(configFile.Name(), configValue)
	config, _ := basic.LoadConfig(configFile.Name())
	crypto, _ := config.LoadCrypto()
	fackEnvelopGenerator := &trafficGenerator.FackEnvelopGenerator{
		Num:     b.N,
		Burst:   1000,
		R:       0,
		Config:  config,
		Crypto:  crypto,
		Envs:    envs,
		ErrorCh: errorCh,
	}
	b.ReportAllocs()
	b.ResetTimer()
	go fackEnvelopGenerator.Start()
	var n int
	for n < b.N {
		<-envs
		n++
	}
	b.StopTimer()
	os.RemoveAll(tmpDir)
}
