package main

import (
	"context"
	"github.com/dovholuknf/qcon2023/shared/common"
	"github.com/dovholuknf/qcon2023/shared/spire"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"log"
)

func main() {
	opts := workloadapi.WithClientOptions(workloadapi.WithAddr(common.SocketPath))

	httpServer := common.CreateServer()
	spire.ConfigureForMutualTLS(context.Background(), httpServer, opts)

	ln := common.CreateUnderlayListener(common.SpireSecuredPort)
	log.Printf("Starting server secured by SPIRE on %d\n", common.SpireSecuredPort)
	if err := httpServer.ServeTLS(ln, "", ""); err != nil {
		log.Fatal(err)
	}
}
