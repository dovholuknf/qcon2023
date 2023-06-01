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
	ctx := context.Background()
	opts := workloadapi.WithClientOptions(workloadapi.WithAddr(common.SocketPath))
	ctx = context.WithValue(ctx, "workloadApiOpts", opts)

	httpServer := common.CreateServer(ctx)
	jwt, _ := spire.FetchJwt(common.SpiffeServerId, opts)
	ln := openziti.CreateOpenZitiListener(jwt, "openziti-only-service")
	log.Printf("Starting server secured by OpenZiti on the OpenZiti overlay, no open port\n")
	if err := httpServer.Serve(ln); err != nil {
		log.Fatal(err)
	}
}
