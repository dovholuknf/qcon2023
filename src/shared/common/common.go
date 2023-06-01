package common

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
)

const (
	InsecurePort     = 18080
	SpireSecuredPort = 18081
	OpenZitiPort     = 18082
	OpenZitiRootUrl  = "https://localhost:1280"
	SocketPath       = "unix:///tmp/spire-agent/public/api.sock"
	SpiffeClientId   = "spiffe://openziti/jwtClient"
	SpiffeServerId   = "spiffe://openziti/jwtServer"
)

func CreateServer() *http.Server {
	svr := &http.Server{}
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(index))
	mux.Handle("/domath", http.HandlerFunc(mathHandler))
	svr.Handler = mux
	return svr
}

func CreateUnderlayListener(port int) net.Listener {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}
	return ln
}

func index(w http.ResponseWriter, r *http.Request) {
	_, _ = io.WriteString(w, "Success!!!\n")
}

func mathHandler(w http.ResponseWriter, r *http.Request) {
	input1, err := strconv.ParseFloat(r.URL.Query().Get("input1"), 64)
	if err != nil {
		http.Error(w, "Invalid input1", http.StatusBadRequest)
		return
	}

	input2, err := strconv.ParseFloat(r.URL.Query().Get("input2"), 64)
	if err != nil {
		http.Error(w, "Invalid input2", http.StatusBadRequest)
		return
	}

	var result float64

	switch r.URL.Query().Get("operator") {
	case "+":
		result = input1 + input2
	case "-":
		result = input1 - input2
	case "*":
		result = input1 * input2
	case "/":
		if input2 == 0 {
			http.Error(w, "Division by zero not allowed", http.StatusBadRequest)
			return
		}
		result = input1 / input2
	default:
		http.Error(w, "Invalid operator", http.StatusBadRequest)
		return
	}

	_, _ = fmt.Fprintf(w, "Result: %.2f", result)
}

func CreateMathUrl(port int16, scheme, host string) string {
	return fmt.Sprintf("%s://%s:%d/domath", scheme, host, port)
}

func AddMathParams(baseURL, input1, operator, input2 string) string {
	params := url.Values{}
	params.Set("input1", input1)
	params.Set("operator", operator)
	params.Set("input2", input2)

	return fmt.Sprintf("%s?%s", baseURL, params.Encode())
}

func CallTheApi(mathURL string) {
	req, err := http.NewRequest("GET", mathURL, nil)
	if err != nil {
		log.Fatalf("unable to create request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Error making the request: %v", err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading the response: %v", err)
	}
	fmt.Println("Response:", string(body))
}

type HandlerSecurityFunc func(ctx context.Context, f http.HandlerFunc) http.Handler

func CreateServerWithSecFunc(ctx context.Context, secFunc HandlerSecurityFunc) *http.Server {
	svr := &http.Server{}
	mux := http.NewServeMux()
	mux.Handle("/", secFunc(ctx, index))
	mux.Handle("/domath", secFunc(ctx, mathHandler))
	svr.Handler = mux
	return svr
}
