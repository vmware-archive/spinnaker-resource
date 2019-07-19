package spinnaker

import (
	"crypto/tls"
	"fmt"
	"net/http"
)

type X509AuthClient struct {
	cert string
	key  string
}

func NewX509AuthClient(cert string, key string) *X509AuthClient {
	return &X509AuthClient{
		cert: cert,
		key : key,
	}
}

func (ac *X509AuthClient) GetClient(url string) (*http.Client, error)  {
	fmt.Println("oh my god")
	cert, err := tls.X509KeyPair([]byte(ac.cert), []byte(ac.key))

	if err != nil {
		return nil, err
	}
	tlsConfig := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		PreferServerCipherSuites: true,
		Certificates:             []tls.Certificate{cert},
		//TODO Do something!!
		InsecureSkipVerify: true,
	}

	tr := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	client := http.Client{Transport: tr}

	return &client, nil
}
