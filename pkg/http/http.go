package http

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"regexp"
	"time"

	"github.com/limingyao/excellent-go/encoding"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

var spaceRegexp = regexp.MustCompile(`[ \n\r\t\v\f]+`)

func FormatBody(buffer []byte) string {
	return string(spaceRegexp.ReplaceAll(buffer, []byte(" ")))
}

type Client struct {
	timeout   time.Duration
	transport *http.Transport
}

func New(opts ...Option) (*Client, error) {
	defaultOpts := defaultOptions
	for _, o := range opts {
		o.apply(&defaultOpts)
	}

	// transport
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConns = defaultOpts.maxIdleConns
	transport.MaxIdleConnsPerHost = defaultOpts.maxIdleConns
	if defaultOpts.tls != nil {
		transport.TLSClientConfig = defaultOpts.tls
	}
	if defaultOpts.insecureSkipVerify {
		if transport.TLSClientConfig == nil {
			transport.TLSClientConfig = &tls.Config{}
		}
		transport.TLSClientConfig.InsecureSkipVerify = true
	}

	s := &Client{transport: transport, timeout: defaultOpts.timeout}

	return s, nil
}

func (x Client) Request(
	ctx context.Context, target, method string, headers map[string]string, data []byte,
) (body []byte, response *http.Response, err error) {
	request, err := http.NewRequest(method, target, bytes.NewReader(data))
	if err != nil {
		return nil, nil, err
	}
	for k, v := range headers {
		request.Header.Set(k, v)
	}
	request = request.WithContext(ctx)

	client := &http.Client{
		Transport: x.transport,
		Timeout:   x.timeout,
	}
	response, err = client.Do(request)
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			log.WithError(err).Error()
		}
	}()

	body, err = io.ReadAll(response.Body)
	return body, response, err
}

func (x Client) JsonPost(
	ctx context.Context, target string, headers map[string]string, req interface{}, rsp interface{},
) (body []byte, code int, header map[string][]string, err error) {
	if headers == nil {
		headers = make(map[string]string)
	}
	if _, ok := headers["Content-Type"]; !ok {
		headers["Content-Type"] = encoding.MIMEJSON
	}
	rv := reflect.ValueOf(rsp)
	if rsp == nil || rv.Kind() != reflect.Ptr {
		return body, 0, nil, fmt.Errorf("rsp can't be nil or non-pointer")
	}

	buffer, err := json.Marshal(req)
	if err != nil {
		return body, 0, nil, err
	}

	body, response, err := x.Request(ctx, target, "POST", headers, buffer)
	if err != nil {
		return body, response.StatusCode, response.Header, err
	}

	if err := json.Unmarshal(buffer, rsp); err != nil {
		return body, response.StatusCode, response.Header, err
	}

	return body, response.StatusCode, response.Header, nil
}

func (x Client) ProtoPost(
	ctx context.Context, target string, headers map[string]string, req proto.Message, rsp proto.Message,
) (body []byte, code int, header map[string][]string, err error) {
	if headers == nil {
		headers = make(map[string]string)
	}
	if _, ok := headers["Content-Type"]; !ok {
		headers["Content-Type"] = encoding.MIMEPROTOBUF
	}
	rv := reflect.ValueOf(rsp)
	if rsp == nil || rv.Kind() != reflect.Ptr {
		return body, 0, nil, fmt.Errorf("rsp can't be nil or non-pointer")
	}

	buffer, err := proto.Marshal(req)
	if err != nil {
		return body, 0, nil, err
	}

	body, response, err := x.Request(ctx, target, "POST", headers, buffer)
	if err != nil {
		return body, response.StatusCode, response.Header, err
	}

	if err := proto.Unmarshal(buffer, rsp); err != nil {
		return body, response.StatusCode, response.Header, err
	}

	return body, response.StatusCode, response.Header, nil
}

func (x Client) Post(
	ctx context.Context, target string, headers map[string]string, requestData []byte,
) ([]byte, *http.Response, error) {
	return x.Request(ctx, target, "POST", headers, requestData)
}

func (x Client) Get(
	ctx context.Context, target string, headers map[string]string,
) ([]byte, *http.Response, error) {
	return x.Request(ctx, target, "GET", headers, []byte{})
}

func (x Client) Put(
	ctx context.Context, target string, headers map[string]string, requestData []byte,
) ([]byte, *http.Response, error) {
	return x.Request(ctx, target, "PUT", headers, requestData)
}
