package bitmex_request_signer

import (
	"net/http"
)

type roundTripperFunc func(r *http.Request) (*http.Response, error)

func (r roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return r(request)
}

// NewSignTripper is http.RoundTripper that intercepts requests and signs them before sending them to server
// baseRoundTripper is a real http.RoundTripper to use for actual request execution
// when baseRoundTripper = nil the http.DefaultTransport is used
//
// Example:
// 	c := http.NewClient()
func NewSignTripper(keyID, keySecret string, baseRoundTripper http.RoundTripper) http.RoundTripper {
	if baseRoundTripper == nil {
		baseRoundTripper = http.DefaultTransport
	}

	rSigner := NewRequestSigner(keyID, keySecret)

	return roundTripperFunc(
		func(r *http.Request) (*http.Response, error) {
			sR, err := rSigner.Sign(r)
			if err != nil {
				return nil, err
			}
			return baseRoundTripper.RoundTrip(sR)
		},
	)
}
