package network

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type HttpClient struct {
	BaseURL   string
	AuthToken string
	Client    *http.Client
	SecretKey []byte
}

func NewHttpClient(baseURL string) *HttpClient {
	return &HttpClient{
		BaseURL: baseURL,
		Client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (hc *HttpClient) setAuthToken(token string) {
	hc.AuthToken = token
}

func (hc *HttpClient) generateToken() error {

	hc.AuthToken = "your_generated_token_here"
	return nil
}

func (hc *HttpClient) authenticate(req *http.Request) error {
	if hc.AuthToken == "" {
		return errors.New("no token found, authentication required")
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", hc.AuthToken))
	return nil
}

func (hc *HttpClient) handleResponse(resp *http.Response) ([]byte, error) {
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("server returned error: %d %s", resp.StatusCode, resp.Status)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (hc *HttpClient) Get(endpoint string, queryParams map[string]string) ([]byte, error) {
	url, err := url.Parse(hc.BaseURL + endpoint)
	if err != nil {
		return nil, err
	}

	q := url.Query()
	for key, value := range queryParams {
		q.Add(key, value)
	}
	url.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	if err := hc.authenticate(req); err != nil {
		return nil, err
	}

	resp, err := hc.Client.Do(req)
	if err != nil {
		return nil, err
	}

	return hc.handleResponse(resp)
}

func (hc *HttpClient) Post(endpoint string, payload interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, hc.BaseURL+endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if err := hc.authenticate(req); err != nil {
		return nil, err
	}

	resp, err := hc.Client.Do(req)
	if err != nil {
		return nil, err
	}

	return hc.handleResponse(resp)
}

func (hc *HttpClient) Put(endpoint string, payload interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPut, hc.BaseURL+endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if err := hc.authenticate(req); err != nil {
		return nil, err
	}

	resp, err := hc.Client.Do(req)
	if err != nil {
		return nil, err
	}

	return hc.handleResponse(resp)
}

func (hc *HttpClient) Delete(endpoint string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodDelete, hc.BaseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}

	if err := hc.authenticate(req); err != nil {
		return nil, err
	}

	resp, err := hc.Client.Do(req)
	if err != nil {
		return nil, err
	}

	return hc.handleResponse(resp)
}

type HandlerParams struct {
	Pattern string
	Handler func(http.ResponseWriter, *http.Request)
}

func AddHandleFunc(params HandlerParams) {
	http.HandleFunc(params.Pattern, params.Handler)

}

type ListenerParams struct {
	Port    string
	Handler http.Handler
}

func Listen(params ListenerParams) {
	http.ListenAndServe(":"+params.Port, params.Handler)
}
