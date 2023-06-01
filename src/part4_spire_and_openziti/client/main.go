package main

import (
	"fmt"
	"github.com/dovholuknf/qcon2023/shared/common"
	"github.com/dovholuknf/qcon2023/shared/openziti"
	"github.com/dovholuknf/qcon2023/shared/spire"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"os"
)

func main() {
	baseURL := common.CreateBaseUrlForClient(443, "https", "openziti.spire.ziti")
	mathUrl := common.CreateUrlForClient(baseURL, os.Args[1], os.Args[2], os.Args[3])
	opts := workloadapi.WithClientOptions(workloadapi.WithAddr(common.SocketPath))
	if len(os.Args) > 4 && os.Args[4] == "showcurl" {
		spire.WriteKeyAndCertToFiles()
		fmt.Println("This is the equivalent curl echo'ed from bash:")
		fmt.Printf("\n  echo Response: $(curl -sk --cert ./cert.pem --key ./key.pem '%s')\n\n", mathUrl)
		fmt.Println("  Of course you know - this won't _actually_ work without OpenZiti, right?")
		fmt.Println("  If you want it to work, provision and enroll an identity in a locally running tunneler")
		fmt.Println("  Notice how you have to supply a key and cert in this example??? Very cool!\n")
	}

	jwt, _ := spire.FetchJwt(common.SpiffeServerId, opts)
	openziti.SecureDefaultHttpClientWithSpireAndOpenZiti(jwt, opts)

	common.CallTheApi(mathUrl)
}
