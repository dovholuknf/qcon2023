package main

import (
	"context"
	"crypto/x509"
	"fmt"
	"github.com/dovholuknf/qcon2023/shared/common"
	"github.com/openziti/edge-api/rest_management_api_client"
	"github.com/openziti/edge-api/rest_model"
	"github.com/openziti/edge-api/rest_util"
	"log"
	"net/http"
	"os"

	"github.com/openziti/edge-api/rest_management_api_client/identity"
)

const (
	defaultProvider = "auth0"
)

var client *rest_management_api_client.ZitiEdgeManagement

func main() {
	var err error
	zitiAdminUsername := os.Getenv("OPENZITI_USER")
	zitiAdminPassword := os.Getenv("OPENZITI_PWD")
	ctrlAddress := os.Getenv("OPENZITI_CTRL")

	// Authenticate with the controller
	caCerts, err := rest_util.GetControllerWellKnownCas(ctrlAddress)
	if err != nil {
		log.Fatal(err)
	}
	caPool := x509.NewCertPool()
	for _, ca := range caCerts {
		caPool.AddCert(ca)
	}
	client, err = rest_util.NewEdgeManagementClientWithUpdb(zitiAdminUsername, zitiAdminPassword, ctrlAddress, caPool)
	if err != nil {
		log.Fatal(err)
	}

	svr := &http.Server{}
	mux := http.NewServeMux()
	mux.Handle("/add-me-to-openziti", http.HandlerFunc(addToOpenZiti))
	svr.Handler = mux
	ln := common.CreateUnderlayListener(common.InsecurePort)
	log.Printf("Starting insecure server on %d\n", common.InsecurePort)
	if err := svr.Serve(ln); err != nil {
		log.Fatal(err)
	}
}

func addToOpenZiti(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "Invalid input. email query param not provided", http.StatusBadRequest)
		return
	}

	oidcProvider := r.URL.Query().Get("oidcProvider")
	if oidcProvider == "" {
		log.Printf("oidcProvider not provided. using default: %s", defaultProvider)
		oidcProvider = defaultProvider
	}

	log.Printf("inputs: %s %s", email, oidcProvider)
	_ = createIdentity(email, email, rest_model.IdentityTypeUser, false)
	_, _ = fmt.Fprintf(w, "Result: %s", "now in a browser go to https://browzer.clint.demo.openziti.org/ and use google auth")
}

func createIdentity(name string, email string,
	identType rest_model.IdentityType, isAdmin bool) *identity.CreateIdentityCreated {
	authPolicyId := os.Getenv("OPENZITI_AUTH_POLICY_ID")
	attrs := &rest_model.Attributes{"docker.whale.dialers"}

	i := &rest_model.IdentityCreate{
		AuthPolicyID:              &authPolicyId,
		ExternalID:                &email,
		IsAdmin:                   &isAdmin,
		Name:                      &name,
		RoleAttributes:            attrs,
		ServiceHostingCosts:       nil,
		ServiceHostingPrecedences: nil,
		Tags:                      nil,
		Type:                      &identType,
	}
	p := identity.NewCreateIdentityParams()
	p.Identity = i
	p.Context = context.Background()

	searchParam := identity.NewListIdentitiesParams()
	filter := "name contains \"" + email + "\""
	searchParam.Filter = &filter
	id, err := client.Identity.ListIdentities(searchParam, nil)
	if err != nil {
		fmt.Println(err)
	}

	if id != nil && len(id.Payload.Data) > 0 {
		delParam := identity.NewDeleteIdentityParams()
		delParam.ID = *id.Payload.Data[0].ID
		_, err := client.Identity.DeleteIdentity(delParam, nil)
		if err != nil {
			fmt.Println(err)
		}
	}
	ident, err := client.Identity.CreateIdentity(p, nil)
	if err != nil {
		fmt.Println(err)
	}

	return ident
}
