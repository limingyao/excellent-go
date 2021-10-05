package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/protobuf/proto"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"regexp"
	"time"
)

var spaceRegexp = regexp.MustCompile(`[ \n\r\t\v\f]+`)

func formatBody(buffer []byte) string {
	return string(spaceRegexp.ReplaceAll(buffer, []byte(" ")))
}

type HTTPClient struct {
	timeout   time.Duration
	transport *http.Transport
}

func New(opts ...Option) (*HTTPClient, error) {
	defaultOpts := defaultOptions
	for _, o := range opts {
		o.apply(&defaultOpts)
	}

	// transport
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConns = defaultOpts.maxIdleConns
	transport.MaxIdleConnsPerHost = defaultOpts.maxIdleConns

	s := &HTTPClient{transport: transport, timeout: defaultOpts.timeout}

	return s, nil
}

func (x HTTPClient) Request(ctx context.Context, target, method string, headers map[string]string, data []byte) ([]byte, int, error) {
	request, err := http.NewRequest(method, target, bytes.NewReader(data))
	if err != nil {
		return nil, 0, err
	}
	for k, v := range headers {
		request.Header.Set(k, v)
	}
	request.WithContext(ctx)

	client := &http.Client{
		Transport: x.transport,
		Timeout:   x.timeout,
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, 0, err
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			log.Println(err)
		}
	}()

	body, err := ioutil.ReadAll(response.Body)
	return body, response.StatusCode, err
}

func (x HTTPClient) JSONPost(ctx context.Context, target string, headers map[string]string, req interface{}, rsp interface{}) (string, int, error) {
	if headers == nil {
		headers = make(map[string]string)
	}
	if _, ok := headers["Content-Type"]; !ok {
		headers["Content-Type"] = "application/json"
	}
	rv := reflect.ValueOf(rsp)
	if rsp == nil || rv.Kind() != reflect.Ptr {
		return "", 0, fmt.Errorf("rsp can't be nil or non-pointer")
	}

	buffer, err := json.Marshal(req)
	if err != nil {
		return "", 0, err
	}

	buffer, httpCode, err := x.Request(ctx, target, "POST", headers, buffer)
	if err != nil {
		return "", httpCode, err
	}

	if err := json.Unmarshal(buffer, rsp); err != nil {
		return formatBody(buffer), httpCode, err
	}

	return "", httpCode, nil
}

func (x HTTPClient) ProtoPost(ctx context.Context, target string, headers map[string]string, req proto.Message, rsp proto.Message) (string, int, error) {
	if headers == nil {
		headers = make(map[string]string)
	}
	if _, ok := headers["Content-Type"]; !ok {
		headers["Content-Type"] = "application/x-protobuf"
	}
	rv := reflect.ValueOf(rsp)
	if rsp == nil || rv.Kind() != reflect.Ptr {
		return "", 0, fmt.Errorf("rsp can't be nil or non-pointer")
	}

	buffer, err := proto.Marshal(req)
	if err != nil {
		return "", 0, err
	}

	buffer, httpCode, err := x.Request(ctx, target, "POST", headers, buffer)
	if err != nil {
		return "", httpCode, err
	}

	if err := proto.Unmarshal(buffer, rsp); err != nil {
		return formatBody(buffer), httpCode, err
	}

	return "", httpCode, nil
}

func (x HTTPClient) Post(ctx context.Context, target string, headers map[string]string, requestData []byte) ([]byte, int, error) {
	return x.Request(ctx, target, "POST", headers, requestData)
}
