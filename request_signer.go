package bitmex_request_signer

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	headerAPIKey       = "api-key"
	headerAPIExpires   = "api-expires"
	headerAPISignature = "api-signature"
)

type RequestSigner struct {
	keyID  string
	signer *Signer

	expireGenerator func() string
}

func NewRequestSigner(keyID, keySecret string) *RequestSigner {
	signer := &RequestSigner{
		keyID:  keyID,
		signer: NewSigner(keySecret),
	}
	signer.expireGenerator = func() string { return signer.expireStr(5 * time.Second) }

	return signer
}

func (s *RequestSigner) Sign(r *http.Request) (*http.Request, error) {
	if r.Header.Get(headerAPISignature) != "" || len(r.Header[headerAPISignature]) != 0 {
		// The request is already signed, no need to do it twice
		return r, nil
	}

	signedR, body, err := s.replaceRequest(r)
	if err != nil {
		return r, err
	}

	expire := s.expireGenerator()
	payload := s.stringToSign(r, expire, string(body))
	signature := s.signer.SignString(payload)

	signedR.Header[headerAPIKey] = []string{s.keyID}
	signedR.Header[headerAPIExpires] = []string{expire}
	signedR.Header[headerAPISignature] = []string{signature}

	return signedR, nil
}

func (s *RequestSigner) expireStr(timeout time.Duration) string {
	return strconv.FormatInt(
		time.Now().Add(timeout).Unix(),
		10,
	)
}

func (s *RequestSigner) stringToSign(r *http.Request, expire, body string) string {
	return strings.ToUpper(r.Method) + r.URL.RequestURI() + expire + body
}

func (s *RequestSigner) replaceRequest(r *http.Request) (copy *http.Request, body []byte, err error) {
	copy = r.Clone(r.Context())
	if r.Body == nil {
		return
	}

	body, err = ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}

	copy.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	err = r.Body.Close()

	return
}
