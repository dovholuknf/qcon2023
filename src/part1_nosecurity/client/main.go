package main

import (
	"fmt"
	"github.com/dovholuknf/qcon2023/shared/common"
	"os"
)

func main() {
	baseURL := common.CreateBaseUrlForClient(common.InsecurePort, "http", "localhost")
	mathUrl := common.CreateUrlForClient(baseURL, os.Args[1], os.Args[2], os.Args[3])
	if len(os.Args) > 4 && os.Args[4] == "showcurl" {
		fmt.Println("This is the equivalent curl echo'ed from bash:")
		fmt.Printf("\n  echo Response: $(curl -sk '%s')\n\n", mathUrl)
	}

	common.CallTheApi(mathUrl)
}
