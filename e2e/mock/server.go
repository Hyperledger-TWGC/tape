package mock

import (
	"google.golang.org/grpc/credentials"
)

type Server struct {
	peers   []*Peer
	orderer *Orderer
}

func NewServer(peerN int, credentials credentials.TransportCredentials) (*Server, error) {
	var txCs []chan struct{}
	var peers []*Peer

	for i := 0; i < peerN; i++ {
		txC := make(chan struct{}, 1000)
		peer, err := NewPeer(txC, credentials)
		if err != nil {
			return nil, err
		}
		peers = append(peers, peer)
		txCs = append(txCs, txC)
	}

	orderer, err := NewOrderer(txCs, credentials)
	if err != nil {
		return nil, err
	}
	return &Server{peers: peers, orderer: orderer}, nil
}

func (s *Server) Start() {
	for _, v := range s.peers {
		go v.Start()
	}
	go s.orderer.Start()
}

func (s *Server) Stop() {
	for _, v := range s.peers {
		v.Stop()
	}
	s.orderer.Stop()
}

func (s *Server) PeersAddresses() (peersAddrs []string) {
	peersAddrs = make([]string, len(s.peers))
	for k, v := range s.peers {
		peersAddrs[k] = v.Addrs()
	}
	return
}

func (s *Server) OrderAddr() string {
	return s.orderer.Addrs()
}

func (s *Server) Addresses() ([]string, string) {
	return s.PeersAddresses(), s.OrderAddr()
}
