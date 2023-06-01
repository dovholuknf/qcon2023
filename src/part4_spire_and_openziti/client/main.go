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
	server := "openziti.spire.ziti"
	baseURL := fmt.Sprintf("%s://%s:%d/domath", httpScheme, server, portToUse)
	params := url.Values{}
	params.Set("input1", os.Args[1])
	params.Set("operator", os.Args[2])
	params.Set("input2", os.Args[3])
	opts := workloadapi.WithClientOptions(workloadapi.WithAddr(common.SocketPath))
	jwt, _ := spire.FetchJwt(common.SpiffeServerId, opts)
	if len(os.Args) > 4 && os.Args[4] == "showcurl" {
		spire.WriteKeyAndCertToFiles()
		fmt.Printf("This is the equivalent curl echo'ed from bash:\n  echo Response: $(curl -sk --cert ./cert.pem --key ./key.pem -H \"Authorization: Bearer %s\" '%s?input1=%v&operator=%v&input2=%v')\n",
			jwt,
			baseURL,
			os.Args[1],
			url.QueryEscape(os.Args[2]),
			os.Args[3])
		fmt.Println("\n    Of course you know - this won't _actually_ work, right?")
		fmt.Println("    If you want it to work, provision and enroll an identity in a locally running tunneler\n")
		fmt.Println("    Notice how you had to supply a key and cert in this example too??? Very cool!\n")
	}

	mathUrl := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	transport := openziti.CreateZitifiedTransport(jwt)
	tlsConfig := spire.CreateSpireMTLS(context.Background(), opts)
	transport.TLSClientConfig = tlsConfig
	http.DefaultTransport = transport

	req, err := http.NewRequest("GET", mathUrl, nil)
	if err != nil {
		log.Fatalf("unable to create request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwt))
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
