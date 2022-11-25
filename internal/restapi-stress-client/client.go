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
	Client        *http.Client
	UserName      string
	Password      string
}

func NewClient(method, url string, headers map[string]string, body []byte) (*HttpClient, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConnsPerHost = 100
	transport.MaxConnsPerHost = 100
	transport.MaxIdleConnsPerHost = 100

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   time.Second * 10,
	}

	return &HttpClient{
		Method:  method,
		Url:     url,
		Headers: headers,
		Body:    body,
		Client:  httpClient,
	}, nil
}

func (cIns *HttpClient) Destory() {
	if cIns != nil {
		cIns.Client.CloseIdleConnections()
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

func (cIns *HttpClient) SetAuth(username, password string) {
	cIns.UserName = username
	cIns.Password = password
}

func (cIns *HttpClient) MakeHttpRequest() (*http.Request, error) {
	log.Printf("[DEBUG] MakeHttpRequest url=%s", cIns.Url)
	//log.Printf("[DEBUG] MakeHttpRequest body=%s", cIns.Body)

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

	if len(cIns.UserName) != 0 {
		req.SetBasicAuth(cIns.UserName, cIns.Password)
	}

	return req, err
}

func (cIns *HttpClient) MakeHttpRequestJsonBody(jsonData map[string]interface{}) (*http.Request, error) {
	log.Printf("[DEBUG] MakeHttpRequestJsonBody url=%s", cIns.Url)

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

	resp, err := cIns.Client.Do(req)
	if err != nil {
		log.Printf("[ERROR] DoHttp failed, error=%s", err)
		return nil, err
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[ERROR] ReadAll() error=%s", err)
		return nil, err
	}

	// 2xx都表示成功
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		log.Printf("[ERROR] Response StatusCode=%d, Status=%s", resp.StatusCode, resp.Status)
		return nil, ecode.Errorf(resp.StatusCode, string(respBytes))
	}

	cIns.Duration = time.Now().Sub(beginTime) / time.Millisecond
	cIns.ContentLength = resp.ContentLength

	log.Printf("[DEBUG] DoHttp StatusCode=%d respBytes=%s", resp.StatusCode, respBytes)

	defer resp.Body.Close()
	return respBytes, nil
}
