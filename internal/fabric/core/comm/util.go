/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package comm

import (
	"crypto/x509"
	"encoding/pem"
)

// AddPemToCertPool adds PEM-encoded certs to a cert pool
func AddPemToCertPool(pemCerts []byte, pool *x509.CertPool) error {
	certs, _, err := pemToX509Certs(pemCerts)
	if err != nil {
		return err
	}
	for _, cert := range certs {
		pool.AddCert(cert)
	}
	return nil
}

// parse PEM-encoded certs
func pemToX509Certs(pemCerts []byte) ([]*x509.Certificate, []string, error) {
	var certs []*x509.Certificate
	var subjects []string

	// it's possible that multiple certs are encoded
	for len(pemCerts) > 0 {
		var block *pem.Block
		block, pemCerts = pem.Decode(pemCerts)
		if block == nil {
			break
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, []string{}, err
		}

		certs = append(certs, cert)
		subjects = append(subjects, string(cert.RawSubject))
	}

	return certs, subjects, nil
}
