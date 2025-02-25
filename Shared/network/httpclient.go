package network

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

type NetworkInterface interface {
	AddHandleFuncUnprotected(params HandlerParams)
	AddHandleFuncProtected(params HandlerParams)
	Listen(params ListenerParams)
	MatchingEngine() HttpClientInterface
	MicroserviceTemplate() HttpClientInterface
	UserManagement() HttpClientInterface
	Authentication() HttpClientInterface
	OrderInitiator() HttpClientInterface
	OrderExecutor() HttpClientInterface
	Stocks() HttpClientInterface
	Transactions() HttpClientInterface
	UserManagementDatabase() HttpClientInterface
}

type Network struct {
	MatchingEngineService         HttpClientInterface
	MicroserviceTemplateService   HttpClientInterface
	UserManagementService         HttpClientInterface
	AuthenticationService         HttpClientInterface
	OrderInitiatorService         HttpClientInterface
	OrderExecutorService          HttpClientInterface
	StocksService                 HttpClientInterface
	TransactionsService           HttpClientInterface
	UserManagementDatabaseService HttpClientInterface
}

func (n *Network) MatchingEngine() HttpClientInterface {
	return n.MatchingEngineService
}

func (n *Network) MicroserviceTemplate() HttpClientInterface {
	return n.MicroserviceTemplateService
}

func (n *Network) UserManagement() HttpClientInterface {
	return n.UserManagementService
}

func (n *Network) Authentication() HttpClientInterface {
	return n.AuthenticationService
}

func (n *Network) OrderInitiator() HttpClientInterface {
	return n.OrderInitiatorService
}

func (n *Network) OrderExecutor() HttpClientInterface {
	return n.OrderExecutorService
}

func (n *Network) Stocks() HttpClientInterface {
	return n.StocksService
}

func (n *Network) Transactions() HttpClientInterface {
	return n.TransactionsService
}

func (n *Network) UserManagementDatabase() HttpClientInterface {
	return n.UserManagementDatabaseService
}

func NewNetwork() NetworkInterface {
	baseURL := os.Getenv("BASE_URL_PREFIX")
	baseURLPostfix := "/"
	return &Network{
		MatchingEngineService:         newHttpClient(baseURL + os.Getenv("MATCHING_ENGINE_HOST") + ":" + os.Getenv("MATCHING_ENGINE_PORT") + baseURLPostfix),
		MicroserviceTemplateService:   newHttpClient(baseURL + os.Getenv("MICROSERVICE_TEMPLATE_HOST") + ":" + os.Getenv("MICROSERVICE_TEMPLATE_PORT") + baseURLPostfix),
		UserManagementService:         newHttpClient(baseURL + os.Getenv("USER_MANAGEMENT_HOST") + ":" + os.Getenv("USER_MANAGEMENT_PORT") + baseURLPostfix),
		AuthenticationService:         newHttpClient(baseURL + os.Getenv("AUTH_HOST") + ":" + os.Getenv("AUTH_PORT") + baseURLPostfix),
		OrderInitiatorService:         newHttpClient(baseURL + os.Getenv("ORDER_INITIATOR_HOST") + ":" + os.Getenv("ORDER_INITIATOR_PORT") + baseURLPostfix),
		OrderExecutorService:          newHttpClient(baseURL + os.Getenv("ORDER_EXECUTOR_HOST") + ":" + os.Getenv("ORDER_EXECUTOR_PORT") + baseURLPostfix),
		StocksService:                 newHttpClient(baseURL + os.Getenv("STOCK_DATABASE_SERVICE_HOST") + ":" + os.Getenv("STOCK_DATABASE_SERVICE_PORT") + baseURLPostfix),
		TransactionsService:           newHttpClient(baseURL + os.Getenv("TRANSACTION_DATABASE_SERVICE_HOST") + ":" + os.Getenv("TRANSACTION_DATABASE_SERVICE_PORT") + baseURLPostfix),
		UserManagementDatabaseService: newHttpClient(baseURL + os.Getenv("USER_MANAGEMENT_DATABASE_SERVICE_HOST") + ":" + os.Getenv("USER_MANAGEMENT_DATABASE_SERVICE_PORT") + baseURLPostfix),
	}
}

type HandlerParams struct {
	Pattern string
	Handler func(http.ResponseWriter, []byte, url.Values, string)
}

func handleFunc(params HandlerParams, w http.ResponseWriter, r *http.Request) {
	// fmt.Println("Handling request for: ", r.URL.Path)
	var body []byte
	var err error
	queryParams := make(url.Values)
	queryParams, err = url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		fmt.Println("Error, there was an issue with reading the message:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if r.Method == http.MethodGet || r.Method == http.MethodDelete || r.Method == http.MethodPut {
		//decode params
		queryParams = r.URL.Query()
		id := strings.TrimPrefix(r.URL.Path, "/"+params.Pattern)
		if id != "" {
			queryParams.Add("id", id)
		}
	}

	if r.Method == http.MethodPost || r.Method == http.MethodPut {
		body, err = io.ReadAll(r.Body)
		if err != nil {
			fmt.Println("Error, there was an issue with reading the message:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()
	}

	if r.Context().Value("userID") != nil {
		queryParams.Add("userID", r.Context().Value("userID").(string))
	}

	params.Handler(w, body, queryParams, r.Method)
	//w.WriteHeader(http.StatusOK)
}

// For Internal handlers
func (n *Network) AddHandleFuncUnprotected(params HandlerParams) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleFunc(params, w, r)
	})
	http.Handle("/"+params.Pattern, handler)

}

// For Protected handlers (I.E exposed to the outside)
func (n *Network) AddHandleFuncProtected(params HandlerParams) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleFunc(params, w, r)
	})
	//To reable after testing is done.
	protectedHandler := TokenAuthMiddleware(handler)
	http.Handle("/"+params.Pattern, protectedHandler)
}

type ListenerParams struct {
	Handler http.Handler //can be nil
}

func (n *Network) Listen(params ListenerParams) {
	http.ListenAndServe(":"+os.Getenv("PORT"), params.Handler)
}

// HttpClientInterface is an interface for the HttpClient struct

type HttpClientInterface interface {
	Get(endpoint string, queryParams map[string]string) ([]byte, error)
	Post(endpoint string, payload interface{}) ([]byte, error)
	Put(endpoint string, payload interface{}) ([]byte, error)
	Delete(endpoint string) ([]byte, error)
}

type HttpClient struct {
	BaseURL   string
	AuthToken string
	Client    *http.Client
	SecretKey []byte
}

func newHttpClient(baseURL string) HttpClientInterface {
	return &HttpClient{
		BaseURL: baseURL,
		Client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// func (hc *HttpClient) setAuthToken() {
// 	context.
// 	token, err := GenerateToken()
// 	hc.AuthToken = token
// }

// func (hc *HttpClient) generateToken() error {

// 	hc.AuthToken = "your_generated_token_here"
// 	return nil
// }

// func GenerateToken(userID uint) (string, error) {
// 	var jwtsecret = []byte(os.Getenv("JWT_SECRET"))
// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
// 		"sub": userID,
// 		"exp": time.Now().Add(time.Hour * 1).Unix(),
// 	})

// 	return token.SignedString(jwtsecret)
// }

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
	if resp.StatusCode == http.StatusResetContent {
		return nil, errors.New("204 No Content")
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
	fmt.Printf("[DEBUG] GET Request URL: %s\n", url.String())
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

	// if err := hc.authenticate(req); err != nil {
	// 	return nil, err
	// }

	resp, err := hc.Client.Do(req)
	if err != nil {
		return nil, err
	}

	return hc.handleResponse(resp)
}

func (hc *HttpClient) Post(endpoint string, payload interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Error: ", err.Error())
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, hc.BaseURL+endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error: ", err.Error())
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	// if err := hc.authenticate(req); err != nil {
	// 	return nil, err
	// }

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
	// if err := hc.authenticate(req); err != nil {
	// 	return nil, err
	// }

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

	// if err := hc.authenticate(req); err != nil {
	// 	return nil, err
	// }

	resp, err := hc.Client.Do(req)
	if err != nil {
		return nil, err
	}

	return hc.handleResponse(resp)
}

// ExtractUserIDFromToken extracts the user ID from a JWT token
func ExtractUserIDFromToken(tokenString string) (uint, error) {
	if tokenString == "" {
		log.Println("[ExtractUserIDFromToken] Missing token in request")
		return 0, errors.New("missing token in request")
	}

	// Remove "Bearer " prefix if present
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	// Ensure JWT_SECRET is loaded
	if len(jwtSecret) == 0 {
		log.Println("[ExtractUserIDFromToken] JWT secret is missing")
		return 0, errors.New("server misconfiguration: JWT secret is missing")
	}

	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is what we expect
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Printf("[ExtractUserIDFromToken] Unexpected signing method: %v", token.Header["alg"])
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})
	if err != nil {
		log.Printf("[ExtractUserIDFromToken] Token parsing error: %v", err)
		return 0, fmt.Errorf("invalid token: %v", err)
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Println("[ExtractUserIDFromToken] Failed to parse token claims")
		return 0, errors.New("invalid claims structure in token")
	}

	// Extract userID from claims
	userID, ok := claims["sub"].(float64)
	if !ok {
		log.Println("[ExtractUserIDFromToken] Missing or malformed user ID in token")
		return 0, errors.New("missing or malformed user ID in token")
	}

	return uint(userID), nil
}

func TokenAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// tokenString := r.Header.Get("Authorization")
		// if tokenString == "" {
		// 	http.Error(w, "Unauthorized: missing token", http.StatusUnauthorized)
		// 	return
		// }
		// // Validate token and extract user ID
		// userID, err := ExtractUserIDFromToken(tokenString)
		// if err != nil {
		// 	http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
		// 	return
		// }
		// // Optionally, you can add the userID to the context:
		userID := "6fd2fc6b-9142-4777-8b30-575ff6fa2460"
		ctx := r.Context()
		ctx = context.WithValue(ctx, "userID", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
