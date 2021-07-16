///*
//Copyright IBM Corp. All Rights Reserved.
//
//SPDX-License-Identifier: Apache-2.0
//*/
//
package comm

import (
	"context"
	tls "github.com/tjfoc/gmsm/gmtls"
	"net"

	"github.com/pkg/errors"
	"github.com/tjfoc/gmsm/gmtls/gmcredentials"
	"google.golang.org/grpc/credentials"
)

var ErrServerHandshakeNotImplemented = errors.New("core/comm: server handshakes are not implemented with clientCreds")

type DynamicClientCredentials struct {
	TLSConfig  *tls.Config
	TLSOptions []TLSOption
}

func (dtc *DynamicClientCredentials) latestConfig() *tls.Config {
	tlsConfigCopy := dtc.TLSConfig.Clone()
	for _, tlsOption := range dtc.TLSOptions {
		tlsOption(tlsConfigCopy)
	}
	return tlsConfigCopy
}

func (dtc *DynamicClientCredentials) ClientHandshake(ctx context.Context, authority string, rawConn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	return gmcredentials.NewTLS(dtc.latestConfig()).ClientHandshake(ctx, authority, rawConn)
}

func (dtc *DynamicClientCredentials) ServerHandshake(rawConn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	return nil, nil, ErrServerHandshakeNotImplemented
}

func (dtc *DynamicClientCredentials) Info() credentials.ProtocolInfo {
	return gmcredentials.NewTLS(dtc.latestConfig()).Info()
}

func (dtc *DynamicClientCredentials) Clone() credentials.TransportCredentials {
	return gmcredentials.NewTLS(dtc.latestConfig())
}

func (dtc *DynamicClientCredentials) OverrideServerName(name string) error {
	dtc.TLSConfig.ServerName = name
	return nil
}
