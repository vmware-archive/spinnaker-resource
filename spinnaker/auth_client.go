package spinnaker

import (
	"net/http"
)

type AuthHttpClient interface {
	GetClient(url string) (*http.Client, error)
}
