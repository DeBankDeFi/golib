package httplib

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"time"

	"github.com/DeBankOps/golib/syserror"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

type BasicAuth struct {
	Username string
	Password string
}

type RequestArgs struct {
	// TraceID for tracing ecosystem, optional.
	TraceID string

	// Request URL, required.
	URL string

	// Host, required
	Host string

	// Request Headers, optional.
	Headers map[string]string

	// Request params, optional.
	Params map[string]string

	// Request Body, optional.
	Body interface{}

	ExpectedStatusCode []int

	JSONResult interface{} `json:"-"`

	BytesResult *bytes.Buffer `json:"-"`

	ResponseHeaders map[string][]string

	BasicAuth    *BasicAuth
	ProtobufType bool

	ReqHandle func(req *http.Request)
}

// HTTPClient 对http client的抽象
// TODO: https support
type HTTPClient struct {
	// http请求超时时间(默认30s)
	timeout time.Duration

	// TLS的配置, 如果该项非空, 则使用HTTPS进行请求(可选)
	tlsConfig *tls.Config

	protoMarshaler jsonpb.Marshaler
}

// NewHTTPClient 新建一个默认配置的http client
func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		timeout:        30 * time.Second,
		tlsConfig:      nil,
		protoMarshaler: jsonpb.Marshaler{OrigName: true},
	}
}

// NewHTTPSClient 新建一个默认配置的https client
// 如果tlsConfig为nil, 将退化成http请求
func NewHTTPSClient(tlsConfig *tls.Config) *HTTPClient {
	return &HTTPClient{
		timeout:   30 * time.Second,
		tlsConfig: tlsConfig,
	}
}

// SetTimeout 设置HTTP请求超时时间，默认30s
// 0表示无超时
func (c *HTTPClient) SetTimeout(v time.Duration) *HTTPClient {
	c.timeout = v
	return c
}

func (c *HTTPClient) validate(args *RequestArgs) (err error) {
	if len(args.TraceID) <= 0 {
		err = errors.New("traceID is required in http client")
		goto FINISH
	}
	if len(args.URL) <= 0 {
		err = errors.New("url is required in http client")
		goto FINISH
	}
FINISH:
	if err != nil {
		return syserror.New(args.TraceID, "VALIDATE_HTTP_REQUEST_ARGS", err.Error(), map[string]interface{}{
			"RequestArgs": args,
		})
	}
	return nil
}

// setHeaders 设置HTTP头部字段
func (c *HTTPClient) setHeaders(req *http.Request, args *RequestArgs) {
	if len(args.Host) > 0 {
		req.Host = args.Host
	}
	for k, v := range args.Headers {
		req.Header.Set(k, v)
	}
}

func (c *HTTPClient) setBasicAuth(req *http.Request, args *RequestArgs) {
	if args.BasicAuth != nil {
		req.SetBasicAuth(args.BasicAuth.Username, args.BasicAuth.Password)
	}
}

func (c *HTTPClient) handleRequest(req *http.Request, args *RequestArgs) {
	if args.ReqHandle != nil {
		args.ReqHandle(req)
	}
}

// setParams 设置HTTP请求参数（以url编码方式设置）
func (c *HTTPClient) setParams(req *http.Request, args *RequestArgs) {
	if len(args.Params) == 0 {
		return
	}
	query := req.URL.Query()
	for k, v := range args.Params {
		query.Add(k, v)
	}
	req.URL.RawQuery = query.Encode()
}

// genBody 生成请求Body
func (c *HTTPClient) genBody(args *RequestArgs) (io.Reader, error) {
	var err error
	var body *bytes.Buffer

	if args.Body == nil {
		return nil, nil
	}
	b, ok := args.Body.([]byte)
	if !ok {
		if args.ProtobufType {
			data, err := c.protoMarshaler.MarshalToString(args.Body.(proto.Message))
			if err != nil {
				return nil, syserror.New(args.TraceID, "PROTO_JSON_MARSHAL", err.Error(), nil)
			}
			b = []byte(data)
		} else {
			if b, err = json.Marshal(args.Body); err != nil {
				return nil, syserror.New(args.TraceID, "JSON_MARSHAL", err.Error(), nil)
			}
		}
	}
	if len(b) > 0 {
		body = bytes.NewBuffer(b)
	}
	return body, nil
}

// Get 发送HTTP Get请求, 但在发送请求之前，需要对 request 做处理，此处理函数逻辑是调用者定义的
func (c *HTTPClient) Get(ctx context.Context, args *RequestArgs) error {
	req, err := http.NewRequest("GET", args.URL, nil)
	if err != nil {
		return syserror.New(args.TraceID, "NEW_HTTP_REQUEST", err.Error(), map[string]interface{}{
			"URL": args.URL,
		})
	}
	return c.doRequest(ctx, req, args)
}

// Delete send a delete http request.
func (c *HTTPClient) Delete(ctx context.Context, args *RequestArgs) error {
	req, err := http.NewRequest(http.MethodDelete, args.URL, nil)
	if err != nil {
		return syserror.New(args.TraceID, "NEW_HTTP_REQUEST", err.Error(), map[string]interface{}{
			"URL": args.URL,
		})
	}
	return c.doRequest(ctx, req, args)
}

// Post 发送HTTP Post请求
func (c *HTTPClient) Post(ctx context.Context, args *RequestArgs) error {
	body, err := c.genBody(args)
	if err != nil {
		return syserror.Wrap(err, "generate http post body failed")
	}
	req, err := http.NewRequest("POST", args.URL, body)
	if err != nil {
		return syserror.New(args.TraceID, "NEW_HTTP_REQUEST", err.Error(), map[string]interface{}{
			"URL": args.URL,
		})
	}
	return c.doRequest(ctx, req, args)
}

func (c *HTTPClient) doRequest(ctx context.Context, req *http.Request, args *RequestArgs) error {
	c.setHeaders(req, args)
	c.setParams(req, args)
	c.setBasicAuth(req, args)
	if ctx != nil {
		req.WithContext(ctx)
	}
	c.handleRequest(req, args)

	cli := http.Client{
		Timeout: c.timeout,
	}

	// 如果TLSClientConfig非空, 将启用https进行请求
	if c.tlsConfig != nil {
		cli.Transport = &http.Transport{
			TLSClientConfig: c.tlsConfig,
		}
	}

	// 发送请求
	rsp, err := cli.Do(req)
	if err != nil {
		return syserror.New(args.TraceID, "HTTP_DO_REQUEST", err.Error(), map[string]interface{}{
			"Method":      req.Method,
			"RequestArgs": UnsafeJsonMarshal(args),
		})
	}
	defer rsp.Body.Close()

	// 读取响应Body
	body := bytes.NewBuffer(nil)
	if _, err = body.ReadFrom(rsp.Body); err != nil {
		return syserror.New(args.TraceID, "HTTP_READ_RSP_BODY", err.Error(), nil)
	}

	// 状态码校验
	var requestOK bool
	var expectedStatusCode = args.ExpectedStatusCode
	if len(expectedStatusCode) == 0 {
		expectedStatusCode = []int{http.StatusOK}
	}
	for _, code := range expectedStatusCode {
		if code == rsp.StatusCode {
			requestOK = true
			break
		}
	}
	if requestOK == false {
		return syserror.New(args.TraceID, "HTTP_STATUS_CODE_NOT_OK", "StatusCode not in expected list", map[string]interface{}{
			"StatusCode":         rsp.StatusCode,
			"ExpectedStatusCode": expectedStatusCode,
			"ResponseBody":       body.String(),
			"RequestArgs":        UnsafeJsonMarshal(args),
		})
	}

	// 结果解析
	if args.JSONResult != nil {
		if args.ProtobufType {
			if err = jsonpb.Unmarshal(body, args.JSONResult.(proto.Message)); err != nil {
				return syserror.New(args.TraceID, "PROTO_JSON_UNMARSHAL", err.Error(), map[string]interface{}{
					"JSONResultType": reflect.TypeOf(args.JSONResult).String(),
					"UnmarshalRaw":   body.String(),
					"RequestArgs":    UnsafeJsonMarshal(args),
				})
			}
		} else {
			if err = json.Unmarshal(body.Bytes(), args.JSONResult); err != nil {
				return syserror.New(args.TraceID, "JSON_UNMARSHAL", err.Error(), map[string]interface{}{
					"JSONResultType": reflect.TypeOf(args.JSONResult).String(),
					"UnmarshalRaw":   body.String(),
					"RequestArgs":    UnsafeJsonMarshal(args),
				})
			}
		}
	}
	if args.BytesResult != nil {
		if _, err = io.Copy(args.BytesResult, body); err != nil {
			return syserror.New(args.TraceID, "IO_COPY", err.Error(), map[string]interface{}{
				"RequestArgs": UnsafeJsonMarshal(args),
			})
		}
	}
	if args.ResponseHeaders != nil {
		for k, v := range rsp.Header {
			args.ResponseHeaders[k] = v
		}
	}
	return nil
}

func UnsafeJsonMarshal(v interface{}) string {
	if v != nil {
		rt, err := json.Marshal(v)
		if err == nil {
			return string(rt)
		}
		fmt.Printf("UnsafeMarshal error: " + err.Error())
	}
	return ""
}
