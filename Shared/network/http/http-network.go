package networkHttp

import (
	"Shared/network"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const TIMEOUT = 5000 * time.Millisecond

type NetworkHttp struct {
	network.BaseNetworkInterface
}

func NewNetworkHttp() network.NetworkInterface {
	nh := &NetworkHttp{
		BaseNetworkInterface: network.NewNetwork(func(serviceString string) network.ClientInterface {
			return newHttpClient(os.Getenv("BASE_URL_PREFIX") + serviceString + os.Getenv("BASE_URL_POSTFIX"))
		}),
	}
	return nh

}

func handleFunc(params network.HandlerParams, w http.ResponseWriter, r *http.Request) {
	fmt.Println("Handling request for: ", r.URL.Path)
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
		fmt.Println("HTTP Handle Error, there was an issue with reading the message:", err)
		responseWriterWrapper.WriteHeader(http.StatusBadRequest)
		return
	}
	if r.Method == http.MethodGet || r.Method == http.MethodDelete || r.Method == http.MethodPut {
		//decode params
		id := strings.TrimPrefix(r.URL.Path, "/"+params.Pattern)
		if id != "" {
			queryParams.Add("id", id)
		}
	}

	if r.Method == http.MethodPost || r.Method == http.MethodPut {
		body, err = io.ReadAll(r.Body)
		if err != nil {
			fmt.Println("HTTP Handle Error, there was an issue with reading the message:", err)
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
			queryParams.Add("userID", userID)
		}
	}

	go params.Handler(responseWriterWrapper, body, queryParams, r.Method)
	select {
	case <-responseWriterWrapper.finished:
		println("finished, closing channel")
		close(responseWriterWrapper.finished)
		responseWriterWrapper.channelHasClosed = true
		break
	case <-time.After(TIMEOUT):
		println("timed out, closing channel")
		if !responseWriterWrapper.channelHasClosed {
			responseWriterWrapper.ResponseWriter.WriteHeader(http.StatusRequestTimeout)
			close(responseWriterWrapper.finished)
			responseWriterWrapper.channelHasClosed = true
		}
		break
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
	rw.currentCode = statusCode
	println("Writing header: ", statusCode)
	rw.ResponseWriter.WriteHeader(statusCode)
	//check if finished is closed
	if !rw.channelHasClosed {
		rw.finished <- true
	}
}

func (rw *responseWriterWrapper) Write(data []byte) (int, error) {
	println("Writing data: ", string(data))
	int, err := rw.ResponseWriter.Write(data)
	if !rw.channelHasClosed {
		rw.finished <- true
	}
	return int, err

}

func (rw *responseWriterWrapper) Header() http.Header {
	return rw.ResponseWriter.Header()
}

func (rw *responseWriterWrapper) EncodeResponse(statusCode int, response map[string]interface{}) {
	println("Encoding response with status code: ", statusCode)
	//rw.Header().Set("Content-Type", "application/json")
	rw.currentCode = statusCode
	j, _ := json.Marshal(response)
	rw.Write(j)
}

// For Internal handlers
func (n *NetworkHttp) AddHandleFuncUnprotected(params network.HandlerParams) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleFunc(params, w, r)

	})
	http.Handle("/"+params.Pattern, handler)
}

// For Protected handlers (I.E exposed to the outside)
func (n *NetworkHttp) AddHandleFuncProtected(params network.HandlerParams) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleFunc(params, w, r)
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
