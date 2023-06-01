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
	baseURL := common.CreateMathUrl(common.OpenZitiPort, "http", "openziti.ziti")
	mathUrl := common.AddMathParams(baseURL, os.Args[1], os.Args[2], os.Args[3])
	opts := workloadapi.WithClientOptions(workloadapi.WithAddr(common.SocketPath))
	if len(os.Args) > 4 && os.Args[4] == "showcurl" {
		fmt.Println("This is the equivalent curl echo'ed from bash:")
		fmt.Printf("\n  echo Response: $(curl -sk '%s')\n\n", mathUrl)
		fmt.Println("  Of course you know - this won't _actually_ work without OpenZiti, right?")
		fmt.Println("  If you want it to work, provision and enroll an identity in a locally running tunneler\n")
	}

	jwt, _ := spire.FetchJwt(common.SpiffeServerId, opts)
	openziti.SecureDefaultHttpClientWithOpenZiti(jwt)

	common.CallTheApi(mathUrl)
}
