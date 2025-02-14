package network

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type HttpClientInterface interface {
	Get(endpoint string, queryParams map[string]string) ([]byte, error)
	Post(endpoint string, payload interface{}) ([]byte, error)
	Put(endpoint string, payload interface{}) ([]byte, error)
	Delete(endpoint string) ([]byte, error)
	AddHandleFunc(params HandlerParams)
	Listen(params ListenerParams)
}

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

func (hc *HttpClient) SetAuthToken(token string) {
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
	body, err := io.ReadAll(resp.Body)
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

func (hc *HttpClient) Delete(endpoint string, payload interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodDelete, hc.BaseURL+endpoint, bytes.NewBuffer(jsonData))
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

type HandlerParams struct {
	Pattern string
	Handler func(http.ResponseWriter, []byte)
}

// Still probably needs authentication shoved in.
func AddHandleFunc(params HandlerParams) {
	http.HandleFunc(params.Pattern, func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Println("Error, there was an issue with reading the message:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()
		w.WriteHeader(http.StatusOK)
		params.Handler(w, body)
		w.Write([]byte("Message received successfully!"))
	})
}

type ListenerParams struct {
	Port    string
	Handler http.Handler
}

func Listen(params ListenerParams) {
	http.ListenAndServe(":"+params.Port, params.Handler)
}

type FakeHttpClient struct {
	listenCalled   bool
	listenerParams ListenerParams
}

func (fhc *FakeHttpClient) Get(endpoint string, queryParams map[string]string) ([]byte, error) {
	return nil, nil
}
func (fhc *FakeHttpClient) Post(endpoint string, payload interface{}) ([]byte, error) {
	return nil, nil
}
func (fhc *FakeHttpClient) Put(endpoint string, payload interface{}) ([]byte, error) { return nil, nil }
func (fhc *FakeHttpClient) Delete(endpoint string) ([]byte, error)                   { return nil, nil }
func (fhc *FakeHttpClient) AddHandleFunc(params HandlerParams)                       {}
func (fhc *FakeHttpClient) Listen(params ListenerParams) {
	fhc.listenCalled = true
	fhc.listenerParams = params
}
