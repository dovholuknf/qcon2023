package main

import (
	"context"
	"fmt"
	"github.com/dovholuknf/qcon2023/shared/common"
	"github.com/dovholuknf/qcon2023/shared/openziti"
	"github.com/dovholuknf/qcon2023/shared/spire"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

func main() {
	portToUse := 443
	httpScheme := "https"
	baseURL := fmt.Sprintf("%s://openziti.spire.ziti:%d/domath", httpScheme, portToUse)
	params := url.Values{}
	params.Set("input1", os.Args[1])
	params.Set("operator", os.Args[2])
	params.Set("input2", os.Args[3])
	opts := workloadapi.WithClientOptions(workloadapi.WithAddr(common.SocketPath))
	jwt, _ := spire.FetchJwt(common.SpiffeServerId, opts)
	if len(os.Args) > 4 && os.Args[4] == "showcurl" {
		fmt.Printf("This is the equivalent curl echo'ed from bash:\n  echo Response: $(curl -sk -H \"Authorization: Bearer %s\" '%s?input1=%v&operator=%v&input2=%v')\n",
			jwt,
			baseURL,
			os.Args[1],
			url.QueryEscape(os.Args[2]),
			os.Args[3])
		fmt.Println("\n    Of course you know - this won't _actually_ work, right?")
		fmt.Println("    If you want it to work, provision and enroll an identity in a locally running tunneler\n")
	}

	mathUrl := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	transport := openziti.CreateZitifiedTransport(jwt)
	tlsConfig := spire.CreateSpiffeEnabledTlsConfig(context.Background(), opts)
	transport.TLSClientConfig = tlsConfig
	http.DefaultTransport = transport

	req, err := http.NewRequest("GET", mathUrl, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwt))
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
