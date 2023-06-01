package main

import (
	"context"
	"github.com/dovholuknf/qcon2023/shared/common"
	"github.com/dovholuknf/qcon2023/shared/openziti"
	"github.com/dovholuknf/qcon2023/shared/spire"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"log"
)

func main() {
	opts := workloadapi.WithClientOptions(workloadapi.WithAddr(common.SocketPath))

	httpServer := common.CreateServer()
	spire.ConfigureForMutualTLS(context.Background(), httpServer, opts)
	jwt, _ := spire.FetchJwt(common.SpiffeServerId, opts)

	ln := openziti.CreateOpenZitiListener(jwt, "openziti-and-spire-service")
	log.Printf("Starting server secured by SPIRE and OpenZiti on the OpenZiti overlay, no open port\n")
	if err := httpServer.ServeTLS(ln, "", ""); err != nil {
		log.Fatal(err)
	}
}
