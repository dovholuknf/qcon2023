package openziti

import (
	"context"
	"github.com/dovholuknf/qcon2023/shared/common"
	edge_apis "github.com/openziti/sdk-golang/edge-apis"
	"github.com/openziti/sdk-golang/ziti"
	"log"
	"net"
	"net/http"
)

func CreateOpenZitiListener(jwt, serviceName string) net.Listener {
	caPool, caErr := ziti.GetControllerWellKnownCaPool(common.OpenZitiRootUrl)
	if caErr != nil {
		panic(caErr)
	}
	credentials := edge_apis.NewJwtCredentials(jwt)
	credentials.CaPool = caPool
	cfg := &ziti.Config{
		ZtAPI:       common.OpenZitiRootUrl + "/edge/client/v1",
		Credentials: credentials,
	}
	cfg.ConfigTypes = append(cfg.ConfigTypes, "all")
	zitiCtx, ctxErr := ziti.NewContext(cfg)
	if ctxErr != nil {
		panic(ctxErr)
	}
	authErr := zitiCtx.Authenticate()
	if authErr != nil {
		panic(authErr)
	}
	ln, err := zitiCtx.Listen(serviceName)
	if err != nil {
		log.Panicf("could not bind service %s: %v", serviceName, err)
	}
	return ln
}

func SecureDefaultHttpClientWithOpenZiti(jwt string) {
	caPool, caErr := ziti.GetControllerWellKnownCaPool(common.OpenZitiRootUrl)
	if caErr != nil {
		panic(caErr)
	}

	credentials := edge_apis.NewJwtCredentials(jwt)
	credentials.CaPool = caPool
	cfg := &ziti.Config{
		ZtAPI:       common.OpenZitiRootUrl + "/edge/client/v1",
		Credentials: credentials,
	}
	cfg.ConfigTypes = append(cfg.ConfigTypes, "all")
	ctx, err := ziti.NewContext(cfg)

	if err = ctx.Authenticate(); err != nil {
		panic(err)
	}
	ziti.DefaultCollection.Add(ctx)

	zitiTransport := http.DefaultTransport.(*http.Transport).Clone() // copy default transport
	zitiTransport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		dialer := ziti.NewDialerWithFallback(ctx, nil)
		return dialer.Dial(network, addr)
	}

	http.DefaultClient.Transport = zitiTransport
}
