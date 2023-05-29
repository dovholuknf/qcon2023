package main

import (
	"github.com/dovholuknf/qcon2023/shared/common"
	"log"
)

func main() {
	underlayServer := common.CreateServer(nil, nil)
	ln := common.CreateUnderlayListener(common.InsecurePort)
	log.Printf("Starting insecure server on %d\n", common.InsecurePort)
	if err := underlayServer.Serve(ln); err != nil {
		log.Fatal(err)
	}
}
