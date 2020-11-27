package bitmex_request_signer

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type RequestSignerTestSuite struct {
	suite.Suite
	apiKey    string
	apiSecret string

	signer *RequestSigner
}

func TestRequestSigner(t *testing.T) {
	suite.Run(t, new(RequestSignerTestSuite))
}

func (s *RequestSignerTestSuite) SetupTest() {
	s.apiKey = "LAqUlngMIQkIUjXMUreyu3qn"
	s.apiSecret = "chNOOS4KvNXR_Xq4k4c9qsfoKWvnDecLATCRlcBwyKDYnWgO"

	s.signer = NewRequestSigner(s.apiKey, s.apiSecret)
}

func (s *RequestSignerTestSuite) TestExpireGenerator() {
	s.Require().NotNil(s.signer.expireGenerator)

	nowStr := strconv.FormatInt(time.Now().Add(time.Millisecond).Unix(), 10)
	s.Require().Greater(s.signer.expireGenerator(), nowStr)
}

func (s *RequestSignerTestSuite) TestHeaders() {
	r := s.generateRequest(http.MethodGet, "/asdf", "", nil)

	signed, err := s.signer.Sign(r)
	s.Require().NoError(err)

	assertHeader := func(sr *http.Request, header string) {
		s.Assert().NotEmptyf(sr.Header[header], "'%s' header was not set on signed request", header)
	}

	assertHeader(signed, "api-key")
	assertHeader(signed, "api-expires")
	assertHeader(signed, "api-signature")
}

func (s *RequestSignerTestSuite) TestDoubleSignature() {
	const (
		sigVal  = "pregenerated-signature"
		bodyVal = "body-value"
	)

	r := s.generateRequest(http.MethodGet, "/asdf", bodyVal, nil)
	r.Header["api-signature"] = []string{sigVal}

	signed, err := s.signer.Sign(r)
	s.Require().NoError(err)
	s.Require().Equal(r, signed, "it seems signer returned copy of original request when no modification needed")
	s.checkBody(bodyVal, signed)
}

func (s *RequestSignerTestSuite) TestSign() {
	type testCase struct {
		method   string
		uri      string
		query    map[string]string
		body     string
		expire   string
		expected string
	}

	testCases := map[string]testCase{
		"simple_get": {
			method:   "GET",
			uri:      "/api/v1/instrument",
			expire:   "1518064236",
			expected: "c7682d435d0cfe87c16098df34ef2eb5a549d4c5a3c2b1f0f77b8af73423bf00",
		},
		"get_w_params": {
			method:   "GET",
			uri:      "/api/v1/instrument",
			query:    map[string]string{"filter": "{\"symbol\": \"XBTM15\"}"},
			expire:   "1518064237",
			expected: "e2f422547eecb5b3cb29ade2127e21b858b235b386bfa45e1c1756eb3383919f",
		},
		"post": {
			method:   "POST",
			uri:      "/api/v1/order",
			expire:   "1518064238",
			body:     "{\"symbol\":\"XBTM15\",\"price\":219.0,\"clOrdID\":\"mm_bitmex_1a/oemUeQ4CAJZgP3fjHsA\",\"orderQty\":98}",
			expected: "1749cd2ccae4aa49048ae09f0b95110cee706e0944e6a14ad0b3a8cb45bd336b",
		},
	}

	runTest := func(params testCase) {
		r := s.generateRequest(params.method, params.uri, params.body, params.query)
		s.setExpire(params.expire)
		sR, err := s.signer.Sign(r)
		s.Require().NoError(err)

		s.checkBody(params.body, sR)
		s.Assert().Equal([]string{params.expected}, sR.Header["api-signature"], "calculated request signature does not metch expected value")
	}

	for name, params := range testCases {
		s.Run(name, func() { runTest(params) })
	}
}

func (s *RequestSignerTestSuite) checkBody(expect string, r *http.Request) {
	if r.Body == nil {
		s.Require().Emptyf(expect, "request's body is empty while we expect data %s", expect)
		return
	}

	rBody, err := ioutil.ReadAll(r.Body)
	s.Require().NoError(err)
	s.Require().Equal([]byte(expect), rBody, "the body read from request differs from the data expected")
}

func (s *RequestSignerTestSuite) generateRequest(method, uri string, body string, query map[string]string) *http.Request {
	var bodyReader io.ReadCloser
	if body != "" {
		buf := bytes.NewBufferString(body)
		bodyReader = ioutil.NopCloser(buf)
	}

	r, err := http.NewRequest(method, uri, bodyReader)
	s.Require().NoError(err)

	//if r.Method == http.MethodPost {
	//	r.Header.Set("Content-Type", "test/plain")
	//}

	if len(query) != 0 {

		q := r.URL.Query()
		for k, v := range query {
			q.Add(k, v)
		}
		r.URL.RawQuery = q.Encode()

	}

	return r
}

func (s *RequestSignerTestSuite) setExpire(v string) {
	s.signer.expireGenerator = func() string { return v }
}
