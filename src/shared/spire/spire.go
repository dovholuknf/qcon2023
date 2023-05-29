package spire

import (
	"context"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/svid/jwtsvid"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"log"
	"net/http"
	"strings"
)

type authenticator struct {
	jwtSource *workloadapi.JWTSource
	audiences []string
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
		audiences: []string{"spiffe://openziti/jwtServer"},
	}
	return auth.authenticateClient(handlerFunc)
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
		expectedId := "spiffe://openziti/jwtClient"
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

func SecureWithSpire(ctx context.Context, server *http.Server) {
	opts := ctx.Value("workloadApiOpts").(workloadapi.SourceOption)
	x509Source, err := workloadapi.NewX509Source(ctx, opts)
	if err != nil {
		log.Printf("unable to create X509Source: %w", err)
		return
	}
	server.TLSConfig = tlsconfig.TLSServerConfig(x509Source)
}

func SecureDefaultHttpClientWithSpiffe(ctx context.Context, opts workloadapi.SourceOption) {
	// Create X509 source to fetch bundle certificate used to verify presented certificate from server
	x509Source, err := workloadapi.NewX509Source(ctx, opts)
	if err != nil {
		log.Fatalf("unable to create X509Source: %v", err)
	}

	// Create a `tls.Config` with configuration to allow TLS communication, and verify that presented certificate
	//from server has SPIFFE ID `spiffe://openziti/jwtServer`
	serverID := spiffeid.RequireFromString("spiffe://openziti/jwtServer")
	tlsConfig := tlsconfig.TLSClientConfig(x509Source, tlsconfig.AuthorizeID(serverID))

	t := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	http.DefaultClient.Transport = t
}
