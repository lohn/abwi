package ado

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtracking"

	"github.com/lohn/abwi/internal/config"
)

// adoScope is the OAuth scope of the Azure DevOps service (its well-known
// application ID).
const adoScope = "499b84ac-1321-427f-aa17-267ca6975798/.default"

// Client bundles the SDK clients with resolved org/project context.
type Client struct {
	Conn    *azuredevops.Connection
	WIT     workitemtracking.Client
	Org     string // organization URL without trailing slash
	Project string
	Format  string // FormatMarkdown or FormatHTML
}

// NewClient builds an authenticated client from the resolved config.
func NewClient(ctx context.Context, cfg *config.Config) (*Client, error) {
	if cfg.Org == "" {
		return nil, errors.New("organization URL is not set: use --org, ABWI_ORG, or the org config key")
	}
	if cfg.Project == "" {
		return nil, errors.New("project is not set: use --project, ABWI_PROJECT, or the project config key")
	}
	conn, err := newConnection(ctx, cfg.Org, cfg.Auth)
	if err != nil {
		return nil, err
	}
	wit, err := workitemtracking.NewClient(ctx, conn)
	if err != nil {
		return nil, err
	}
	return &Client{
		Conn:    conn,
		WIT:     wit,
		Org:     strings.TrimRight(cfg.Org, "/"),
		Project: cfg.Project,
		Format:  cfg.Format,
	}, nil
}

func newConnection(ctx context.Context, orgURL, auth string) (*azuredevops.Connection, error) {
	if auth == "pat" {
		pat := os.Getenv("ABWI_PAT")
		if pat == "" {
			pat = os.Getenv("AZURE_DEVOPS_EXT_PAT")
		}
		if pat == "" {
			return nil, errors.New("auth is \"pat\" but neither ABWI_PAT nor AZURE_DEVOPS_EXT_PAT is set")
		}
		return azuredevops.NewPatConnection(orgURL, pat), nil
	}
	cred, err := azidentity.NewAzureCLICredential(nil)
	if err != nil {
		return nil, err
	}
	tok, err := cred.GetToken(ctx, policy.TokenRequestOptions{Scopes: []string{adoScope}})
	if err != nil {
		return nil, fmt.Errorf("acquiring an Entra ID token failed (run `az login` first): %w", err)
	}
	return &azuredevops.Connection{
		AuthorizationString:     "Bearer " + tok.Token,
		BaseUrl:                 strings.ToLower(strings.TrimRight(orgURL, "/")),
		SuppressFedAuthRedirect: true,
	}, nil
}
