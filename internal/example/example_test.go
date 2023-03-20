package example

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	recorder "github.com/carmo-evan/docs-recorder"
)

func TestMain(t *testing.T) {
	handler := http.HandlerFunc(getWelcome)
	dr := recorder.NewDocsRecoder(t, handler)
	if err := dr.Endpoint("GET", "/").Test(
		recorder.NewDocsRecorderTestCase(recorder.DocsRecorderTestCaseOptions{
			Name:       "it returns 200 successfully",
			OutputType: &OkResponse{},
		}),
		recorder.NewDocsRecorderTestCase(recorder.DocsRecorderTestCaseOptions{
			Name:        "it returns 400",
			QueryParams: url.Values{"id": []string{"12345"}},
			Check: func(rr *httptest.ResponseRecorder, t *testing.T) {
				if rr.Code != 400 {
					t.Fatal("wrong code")
				}
			},
		}),
	); err != nil {
		t.Fatal(err)
	}
	if err := dr.Write(recorder.OpenAPI); err != nil {
		t.Fatal(err)
	}
}
