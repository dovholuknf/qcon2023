package main

import (
	"github.com/dovholuknf/qcon2023/shared/common"
	"github.com/dovholuknf/qcon2023/shared/openziti"
	"github.com/dovholuknf/qcon2023/shared/spire"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"log"
)

func main() {
	opts := workloadapi.WithClientOptions(workloadapi.WithAddr(common.SocketPath))

	httpServer := common.CreateServer()

	jwt, _ := spire.FetchJwt(common.SpiffeServerId, opts)

	ln := openziti.CreateOpenZitiListener(jwt, "openziti-only-service")
	log.Printf("Starting server secured by OpenZiti on the OpenZiti overlay, no open port\n")
	if err := httpServer.Serve(ln); err != nil {
		log.Fatal(err)
	}
}
