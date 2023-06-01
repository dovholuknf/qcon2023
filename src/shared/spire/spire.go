package spire

import (
	"context"
	"crypto/tls"
	"encoding/pem"
	"fmt"
	"github.com/dovholuknf/qcon2023/shared/common"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/svid/jwtsvid"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type authenticator struct {
	jwtSource *workloadapi.JWTSource
	audiences []string
}

func FetchJwt(audience string, opts workloadapi.SourceOption) (string, error) {
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	jwtSource, err := workloadapi.NewJWTSource(ctx, opts)
	if err != nil {
		return "", fmt.Errorf("unable to create JWTSource: %w", err)
	}
	svid, err := jwtSource.FetchJWTSVID(ctx, jwtsvid.Params{
		Audience: audience,
	})
	if err != nil {
		return "", err
	}
	return svid.Marshal(), nil
}

func (a *authenticator) authenticateClient(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		fields := strings.Fields(req.Header.Get("Authorization"))
		if len(fields) != 2 || fields[0] != "Bearer" {
			log.Print("Malformed header")
			http.Error(w, "Invalid or unsupported authorization header", http.StatusUnauthorized)
			return
		}

		token := fields[1]
		log.Printf("JWT: %s", token)
		// Parse and validate token against fetched bundle from jwtSource,
		// an alternative is using `workloadapi.ValidateJWTSVID` that will
		// attest against SPIRE on each call and validate token
		svid, err := jwtsvid.ParseAndValidate(token, a.jwtSource, a.audiences)
		if err != nil {
			log.Printf("Invalid token: %v\n", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		req = req.WithContext(withSVIDClaims(req.Context(), svid.Claims))
		expectedId := common.SpiffeClientId
		if svid.Claims["sub"] != expectedId {
			log.Printf("sub mismatch. expected: %s, got %s", expectedId, svid.Claims["sub"])
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, req)
	})
}

type svidClaimsKey struct{}

func withSVIDClaims(ctx context.Context, claims map[string]interface{}) context.Context {
	return context.WithValue(ctx, svidClaimsKey{}, claims)
}

func ConfigureForMutualTLS(ctx context.Context, server *http.Server, opts workloadapi.SourceOption) {
	source, err := workloadapi.NewX509Source(ctx, opts)
	if err != nil {
		panic(err)
	}
	// Create a `tls.Config` to allow mTLS connections, and verify that presented certificate has SPIFFE ID `spiffe://example.org/client`
	clientID := spiffeid.RequireFromString(common.SpiffeClientId)
	tlsConfig := tlsconfig.MTLSServerConfig(source, source, tlsconfig.AuthorizeID(clientID))

	server.TLSConfig = tlsConfig
}

func CreateSpireMTLS(ctx context.Context, opts workloadapi.SourceOption) *tls.Config {
	source, err := workloadapi.NewX509Source(ctx, opts)
	if err != nil {
		panic(err)
	}
	serverID := spiffeid.RequireFromString(common.SpiffeServerId)
	return tlsconfig.MTLSClientConfig(source, source, tlsconfig.AuthorizeID(serverID))
}

func SecureDefaultHttpClientWithSpireMTLS(ctx context.Context, opts workloadapi.SourceOption) {
	t := &http.Transport{
		TLSClientConfig: CreateSpireMTLS(ctx, opts),
	}

	http.DefaultClient.Transport = t
}

func WriteKeyAndCertToFiles() {
	svid, _ := workloadapi.FetchX509SVID(context.Background(), workloadapi.WithAddr(common.SocketPath))
	c, k, _ := svid.MarshalRaw()
	keyFile, _ := os.Create("./key.pem")
	_ = pem.Encode(keyFile, &pem.Block{Type: "PRIVATE KEY", Bytes: k})
	certFile, _ := os.Create("./cert.pem")
	_ = pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: c})
}

func SecureWithSpireTLS(ctx context.Context, opts workloadapi.SourceOption) *tls.Config {
	x509Source, err := workloadapi.NewX509Source(ctx, opts)
	if err != nil {
		panic(err)
	}
	return tlsconfig.TLSServerConfig(x509Source)
}

func SecureWithSpireJwt(ctx context.Context, handlerFunc http.HandlerFunc) http.Handler {
	opts := ctx.Value("workloadApiOpts").(workloadapi.SourceOption)
	jwtSource, err := workloadapi.NewJWTSource(ctx, opts)
	if err != nil {
		log.Printf("unable to create JWTSource: %w", err)
		panic(err)
	}

	auth := &authenticator{
		jwtSource: jwtSource,
		audiences: []string{common.SpiffeServerId},
	}
	return auth.authenticateClient(handlerFunc)
}
