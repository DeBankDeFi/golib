package httplib_test

import (
	"bytes"
	"context"
	"crypto/tls"
	"testing"

	"github.com/DeBankOps/golib/httplib"
)

func TestHttpGet(t *testing.T) {
	result := bytes.NewBuffer(nil)
	err := httplib.NewHTTPClient().Get(context.Background(), &httplib.RequestArgs{
		TraceID: "fakeID",
		URL:     "http://www.google.cn",
		Headers: map[string]string{
			"X-NAME": "sdsd",
		},
		Params: map[string]string{
			"id": "dsds",
		},
		BytesResult: result,
	})
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log("OK")
}

func TestHttpResponseHeader(t *testing.T) {
	result := bytes.NewBuffer(nil)
	respHeaders := make(map[string][]string, 0)
	err := httplib.NewHTTPClient().Get(context.Background(), &httplib.RequestArgs{
		TraceID: "fakeID",
		URL:     "http://www.google.cn",
		Headers: map[string]string{
			"X-NAME": "sdsd",
		},
		Params: map[string]string{
			"id": "dsds",
		},
		BytesResult:     result,
		ResponseHeaders: respHeaders,
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	if len(respHeaders) == 0 {
		t.Fatal("get zero length response headers")
	}
	t.Log("OK")
}

func TestHttpPost(t *testing.T) {
	result := bytes.NewBuffer(nil)
	err := httplib.NewHTTPClient().Post(context.Background(), &httplib.RequestArgs{
		TraceID: "fakeID",
		URL:     "http://www.google.cn",
		Headers: map[string]string{
			"X-NAME": "sdsd",
		},
		Params: map[string]string{
			"id": "dsds",
		},
		Body:               []byte("jdkdjsfkjds"),
		ExpectedStatusCode: []int{200, 405},
		BytesResult:        result,
	})
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log("OK")
}

func TestHttpsGet(t *testing.T) {
	result := bytes.NewBuffer(nil)
	err := httplib.NewHTTPSClient(
		&tls.Config{InsecureSkipVerify: true},
	).Get(context.Background(), &httplib.RequestArgs{
		TraceID: "fakeID",
		URL:     "https://www.google.cn",
		Headers: map[string]string{
			"X-NAME": "sdsd",
		},
		Params: map[string]string{
			"id": "dsds",
		},
		BytesResult: result,
	})
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log("OK")
}

func TestHttpsPost(t *testing.T) {
	result := bytes.NewBuffer(nil)
	err := httplib.NewHTTPSClient(
		&tls.Config{InsecureSkipVerify: true},
	).Post(context.Background(), &httplib.RequestArgs{
		TraceID: "fakeID",
		URL:     "https://www.google.cn",
		Headers: map[string]string{
			"X-NAME": "sdsd",
		},
		Params: map[string]string{
			"id": "dsds",
		},
		Body:               []byte("jdkdjsfkjds"),
		ExpectedStatusCode: []int{200, 405},
		BytesResult:        result,
	})
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log("OK")
}
