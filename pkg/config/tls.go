package config

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"

	"github.com/pkg/errors"
)

var clientCipherSuites = []uint16{
	tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
}

// TLSData tls 配置
type TLSData struct {
	CA   []byte
	Key  []byte
	Cert []byte
}

// TLSDataFromFiles 从证书文件加载tls配置
func TLSConfigFromFiles(caPath, certPath, keyPath, tlsPassword string, skipTLSVerify bool) (*tls.Config, error) {
	var (
		ca, cert, key []byte
		err           error
	)
	if caPath != "" {
		if ca, err = ioutil.ReadFile(caPath); err != nil {
			return nil, err
		}
	}
	if certPath != "" {
		if cert, err = ioutil.ReadFile(certPath); err != nil {
			return nil, err
		}
	}
	if keyPath != "" {
		if key, err = ioutil.ReadFile(keyPath); err != nil {
			return nil, err
		}
	}
	if ca == nil && cert == nil && key == nil {
		return nil, nil
	}
	tlsData := &TLSData{
		CA:   ca,
		Cert: cert,
		Key:  key,
	}

	if tlsData == nil && !skipTLSVerify {
		// there is no specific tls config
		return nil, nil
	}
	tlsconfig := &tls.Config{
		// Prefer TLS1.2 as the client minimum
		MinVersion:   tls.VersionTLS12,
		CipherSuites: clientCipherSuites,
	}
	if tlsData != nil && tlsData.CA != nil {
		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(tlsData.CA) {
			return nil, errors.New("failed to retrieve context tls info: ca.pem seems invalid")
		}
		tlsconfig.RootCAs = certPool
	}
	if tlsData != nil && tlsData.Key != nil && tlsData.Cert != nil {
		keyBytes := tlsData.Key
		pemBlock, _ := pem.Decode(keyBytes)
		if pemBlock == nil {
			return nil, errors.New("no valid private key found")
		}

		var err error
		if x509.IsEncryptedPEMBlock(pemBlock) {
			keyBytes, err = x509.DecryptPEMBlock(pemBlock, []byte(""))
			if err != nil {
				return nil, errors.Wrap(err, "private key is encrypted, but could not decrypt it")
			}
			keyBytes = pem.EncodeToMemory(&pem.Block{Type: pemBlock.Type, Bytes: keyBytes})
		}

		x509cert, err := tls.X509KeyPair(tlsData.Cert, keyBytes)
		if err != nil {
			return nil, errors.Wrap(err, "failed to retrieve context tls info")
		}
		tlsconfig.Certificates = []tls.Certificate{x509cert}
	}
	if skipTLSVerify {
		tlsconfig.InsecureSkipVerify = true
	}

	return tlsconfig, nil
}

