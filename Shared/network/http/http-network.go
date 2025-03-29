package networkHttp

import (
	"Shared/network"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

type NetworkHttp struct {
	network.BaseNetworkInterface
	timeout time.Duration
}

func NewNetworkHttp() network.NetworkInterface {
	timeOutEnv, err := time.ParseDuration(os.Getenv("HTTP_TIMEOUT"))
	if err != nil {
		log.Println("TIMEOUT env variable not set, defaulting to 20s")
		timeOutEnv = 20000 * time.Millisecond
	}

	nh := &NetworkHttp{
		BaseNetworkInterface: network.NewNetwork(func(serviceString string) network.ClientInterface {
			return newHttpClient(os.Getenv("BASE_URL_PREFIX") + serviceString + os.Getenv("BASE_URL_POSTFIX"))
		}),
		timeout: timeOutEnv,
	}
	return nh

}

func handleFunc(params network.HandlerParams, w http.ResponseWriter, r *http.Request, timeout time.Duration) {
	responseWriterWrapper := &responseWriterWrapper{ResponseWriter: w, currentCode: http.StatusOK, finished: make(chan bool, 1), channelHasClosed: false}
	var body []byte
	var err error
	var queryParams url.Values
	queryParams, err = url.ParseQuery(r.URL.RawQuery)
	for key, value := range r.Header {
		for _, v := range value {
			queryParams.Add(key, v)
		}
	}
	if err != nil {
		log.Println("HTTP Handle Error, there was an issue with reading the message:", err)
		responseWriterWrapper.WriteHeader(http.StatusBadRequest)
		return
	}
	if r.Method == http.MethodGet || r.Method == http.MethodDelete || r.Method == http.MethodPut || r.Method == http.MethodPatch {
		//decode params
		id := strings.TrimPrefix(r.URL.Path, "/"+params.Pattern)
		if id != "" {
			queryParams.Add("id", id)
		}
	}

	if r.Method == http.MethodPost || r.Method == http.MethodPut {
		body, err = io.ReadAll(r.Body)
		if err != nil {
			log.Println("HTTP Handle Error, there was an issue with reading the message:", err)
			responseWriterWrapper.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()
	}

	// The type assertion here is failing; r.Context().Value(userIDKey) returns a uint.
	// So we need to change that.
	if r.Context().Value(userIDKey) != nil {
		// queryParams.Add("userID", r.Context().Value(userIDKey).(string))
		if userID, ok := r.Context().Value(userIDKey).(uint); ok {
			queryParams.Add("userID", fmt.Sprintf("%d", userID)) // Convert to string
		} else if userID, ok := r.Context().Value(userIDKey).(string); ok {
			if _, err := uuid.Parse(strings.TrimSpace(userID)); err != nil {
				log.Println("Network: UserID is not a valid UUID")
				responseWriterWrapper.WriteHeader(http.StatusInternalServerError)
				return
			}
			queryParams.Add("userID", userID)
		} else {
			log.Println("Network: UserID is Unknown type")
			responseWriterWrapper.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	go params.Handler(responseWriterWrapper, body, queryParams, r.Method)
	timer := time.NewTimer(timeout)
	select {
	case <-responseWriterWrapper.finished:
		log.Println("Request Finished: ", r.URL.String())
		close(responseWriterWrapper.finished)
		responseWriterWrapper.channelHasClosed = true
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		return
	case <-timer.C:
		if !responseWriterWrapper.channelHasClosed {
			responseWriterWrapper.ResponseWriter.WriteHeader(http.StatusRequestTimeout)
			close(responseWriterWrapper.finished)
			responseWriterWrapper.channelHasClosed = true
		}
		log.Println("HTTP Handle Error, request timed out: ", r.URL.String())
		return
	}
	//w.WriteHeader(http.StatusOK)
}

type responseWriterWrapper struct {
	http.ResponseWriter
	currentCode      int
	finished         chan bool
	channelHasClosed bool
}

func (rw *responseWriterWrapper) WriteHeader(statusCode int) {
	if !rw.CheckCompleted() {
		rw.currentCode = statusCode
		rw.ResponseWriter.WriteHeader(statusCode)
	}
	//check if finished is closed
	if !rw.channelHasClosed {
		rw.finished <- true
	}
}

func (rw *responseWriterWrapper) Write(data []byte) (int, error) {
	int := 0
	var err error
	if !rw.CheckCompleted() {
		int, err = rw.ResponseWriter.Write(data)
	}
	if !rw.channelHasClosed {
		rw.finished <- true
	}
	return int, err

}

func (rw *responseWriterWrapper) Header() http.Header {
	return rw.ResponseWriter.Header()
}

func (rw *responseWriterWrapper) EncodeResponse(statusCode int, response map[string]interface{}) {
	//rw.Header().Set("Content-Type", "application/json")
	if !rw.CheckCompleted() {
		if statusCode != http.StatusOK {
			rw.ResponseWriter.WriteHeader(statusCode)
		}
		j, _ := json.Marshal(response)
		rw.Write(j)
	}
}

func (rw *responseWriterWrapper) CheckCompleted() bool {
	return rw.channelHasClosed
}

func (rw *responseWriterWrapper) GetStatusCode() int {
	return rw.currentCode
}

// For Internal handlers
func (n *NetworkHttp) AddHandleFuncUnprotected(params network.HandlerParams) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleFunc(params, w, r, n.timeout)

	})
	http.Handle("/"+params.Pattern, handler)
}

// For Protected handlers (I.E exposed to the outside)
func (n *NetworkHttp) AddHandleFuncProtected(params network.HandlerParams) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleFunc(params, w, r, n.timeout)
	})
	//To reable after testing is done.
	protectedHandler := TokenAuthMiddleware(handler)
	http.Handle("/"+params.Pattern, protectedHandler)
}

// type ListenerParams struct {
// 	Handler http.Handler //can be nil
// }

func (n *NetworkHttp) Listen() {
	http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}
