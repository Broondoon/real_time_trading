package networkHttp

import (
	"Shared/network"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

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

	params.Handler(w, body, queryParams, r.Method)
	//w.WriteHeader(http.StatusOK)
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
