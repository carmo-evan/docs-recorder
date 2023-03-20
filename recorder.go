package recorder

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/swaggest/openapi-go/openapi3"
)

type DocsRecorder struct {
	t         *testing.T
	handler   http.Handler
	endpoints []*DocsRecorderEndpoint
}

func NewDocsRecoder(t *testing.T, handler http.Handler) *DocsRecorder {
	return &DocsRecorder{
		t, handler, nil,
	}
}

type DocsRecorderEndpoint struct {
	method    string
	url       string
	parent    *DocsRecorder
	testCases []*DocsRecorderTestCase
}

type CheckFunc func(rr *httptest.ResponseRecorder, t *testing.T)

type DocsRecorderTestCase struct {
	name        string
	check       CheckFunc
	body        io.Reader
	queryParams url.Values
	recorder    *httptest.ResponseRecorder
	url         string
	method      string
	request     *http.Request
	OutputType  any
}

type DocsRecorderTestCaseOptions struct {
	Name        string
	Check       CheckFunc
	QueryParams url.Values
	Url         string
	Method      string
	Body        string
	OutputType  any
}

func NewDocsRecorderTestCase(opts DocsRecorderTestCaseOptions) *DocsRecorderTestCase {
	var b io.Reader
	if opts.Body != "" {
		b = strings.NewReader(opts.Body)
	}
	return &DocsRecorderTestCase{
		opts.Name,
		opts.Check,
		b,
		opts.QueryParams,
		httptest.NewRecorder(),
		opts.Url,
		opts.Method,
		nil,
		opts.OutputType,
	}
}

type DocsRecorderTarget string

const OpenAPI DocsRecorderTarget = "openapi"

func (dr *DocsRecorder) Endpoint(method, url string) *DocsRecorderEndpoint {
	endpoint := &DocsRecorderEndpoint{method, url, dr, nil}
	dr.endpoints = append(dr.endpoints, endpoint)
	return endpoint
}

func (endpoint *DocsRecorderEndpoint) Test(cases ...*DocsRecorderTestCase) error {
	for _, tc := range cases {
		u, err := url.Parse(endpoint.url)
		if err != nil {
			return err
		}
		u.RawQuery = tc.queryParams.Encode()
		tc.url = u.String()
		tc.method = endpoint.method
		req, err := http.NewRequest(endpoint.method, u.String(), tc.body)
		if err != nil {
			return err
		}
		tc.request = req
		endpoint.parent.handler.ServeHTTP(tc.recorder, req)
		endpoint.parent.t.Run(tc.name, func(t *testing.T) {
			if tc.check != nil {
				tc.check(tc.recorder, t)
			}
		})
	}
	endpoint.testCases = cases
	return nil
}

func operationExists(method, path string, s *openapi3.Spec) bool {
	method = strings.ToLower(method)
	pathItem := s.Paths.MapOfPathItemValues[path]

	if _, found := pathItem.MapOfOperationValues[method]; found {
		return true
	}
	return false
}

func (dr *DocsRecorder) Write(target DocsRecorderTarget) error {
	switch target {
	default: //default is openapi
		reflector := openapi3.Reflector{}
		reflector.Spec = &openapi3.Spec{Openapi: "3.0.3"}
		reflector.Spec.Info.
			WithTitle("Things API").
			WithVersion("1.2.3").
			WithDescription("Put something here")
		f, err := os.OpenFile("./openapi.yaml",
			os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			return err
		}
		defer f.Close()

		for _, endpoint := range dr.endpoints {
			op := &openapi3.Operation{}
			for _, tc := range endpoint.testCases {

				if err := reflector.SetRequest(op, tc.request, tc.method); err != nil {
					return err
				}
				b, _ := ioutil.ReadAll(tc.recorder.Body)

				if err := json.Unmarshal(b, &tc.OutputType); err != nil {
					return err
				}
				if err := reflector.SetJSONResponse(op, tc.OutputType, tc.recorder.Code); err != nil {
					return err
				}
			}

			if err := reflector.Spec.AddOperation(endpoint.method, endpoint.url, *op); err != nil {
				return err
			}
		}
		schema, err := reflector.Spec.MarshalYAML()
		if err != nil {
			log.Fatal(err)
		}

		if _, err := f.Write(schema); err != nil {
			return err
		}
		return nil
	}
}
