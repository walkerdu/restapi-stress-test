package internal

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	ecode "github.com/walkerdu/restapi-stress-test/pkg"
)

type HttpClient struct {
	Method        string
	Url           string
	Headers       map[string]string
	Body          []byte
	Duration      time.Duration
	ContentLength int64
	ExecCounts    int64
	HttpClient    *http.Client
}

func NewClient(method, url string, headers map[string]string, body []byte) (*HttpClient, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConnsPerHost = 100
	transport.MaxConnsPerHost = 100
	transport.MaxIdleConnsPerHost = 100

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   time.Second * 3,
	}

	return &HttpClient{
		Method:     method,
		Url:        url,
		Headers:    headers,
		Body:       body,
		HttpClient: httpClient,
	}, nil
}

func (cIns *HttpClient) Destory() {
	if cIns != nil {
		cIns.HttpClient.CloseIdleConnections()
	}
}

func (cIns *HttpClient) SetMethod(method string) {
	cIns.Method = method
}

func (cIns *HttpClient) SetUrl(url string) {
	cIns.Url = url
}

func (cIns *HttpClient) SetHeaders(headers map[string]string) {
	cIns.Headers = headers
}

func (cIns *HttpClient) SetBody(body []byte) {
	cIns.Body = body
}

func (cIns *HttpClient) MakeHttpRequest() (*http.Request, error) {
	log.Printf("MakeHttpRequest url=%s", cIns.Url)

	var reader *bytes.Reader
	if cIns.Body != nil {
		reader = bytes.NewReader(cIns.Body)
	}

	req, err := http.NewRequest(cIns.Method, cIns.Url, reader)
	if err != nil {
		return nil, err
	}

	for key, value := range cIns.Headers {
		req.Header.Set(key, value)
	}
	return req, err
}

func (cIns *HttpClient) MakeHttpRequestJsonBody(jsonData map[string]interface{}) (*http.Request, error) {
	log.Printf("MakeHttpRequestJsonBody url=%s", cIns.Url)

	if jsonData == nil {
		jsonData = make(map[string]interface{})
	}

	data, err := json.Marshal(jsonData)
	if err != nil {
		return nil, err
	}

	cIns.Body = data

	req, err := cIns.MakeHttpRequest()
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-type", "application/json")

	return req, err
}

func (cIns *HttpClient) DoHttp() ([]byte, error) {
	req, err := cIns.MakeHttpRequest()
	if err != nil {
		log.Printf("[ERROR] MakeHTTPRequest failed, error=%s", err)
		return nil, err
	}

	beginTime := time.Now()

	resp, err := cIns.HttpClient.Do(req)
	if err != nil {
		log.Printf("[ERROR] DoHttp failed, error=%s", err)
		return nil, err
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, ecode.Errorf(resp.StatusCode, string(respBytes))
	}

	cIns.Duration = time.Now().Sub(beginTime)
	cIns.ContentLength = resp.ContentLength

	defer resp.Body.Close()
	return respBytes, nil
}
