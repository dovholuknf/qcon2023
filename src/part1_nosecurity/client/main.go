package main

import (
	"fmt"
	"github.com/dovholuknf/qcon2023/shared/common"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

func main() {
	portToUse := common.InsecurePort
	httpScheme := "http"
	baseURL := fmt.Sprintf("%s://localhost:%d/domath", httpScheme, portToUse)
	params := url.Values{}
	params.Set("input1", os.Args[1])
	params.Set("operator", os.Args[2])
	params.Set("input2", os.Args[3])
	if len(os.Args) > 4 && os.Args[4] == "showcurl" {
		fmt.Printf("This is the equivalent curl echo'ed from bash:\n  echo Response: $(curl -sk '%s?input1=%v&operator=%v&input2=%v')\n",
			baseURL,
			os.Args[1],
			url.QueryEscape(os.Args[2]),
			os.Args[3])
	}

	mathUrl := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	req, err := http.NewRequest("GET", mathUrl, nil)
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
