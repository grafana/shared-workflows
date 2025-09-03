package gh

import (
	"context"
	"net/http"
	"strconv"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v74/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

// TODO: should these follow the Github API or the API that I would like to support?
type GithubClient interface {
	FetchStaleBranches()
	DeleteStaleBranches()
}

type Client struct {
	restClient *github.Client
}

// so this is authenticating via token (which should be from an env variable)
func NewGitHubClientWithTokenAuth(ctx context.Context, token string) *Client {
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	))

	return newGitHubClientWithHTTPClient(httpClient)
}

// NewGitHubClientWithAppAuth creates a new GitHub client authenticated with the given GitHub App.
// The app must be installed in the given org.
func NewGitHubClientWithAppAuth(ctx context.Context, org, appID, privateKey string) (*Client, error) {
	numericAppID, err := strconv.Atoi(appID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse the numeric GitHub app ID %q", appID)
	}

	installationID, err := getGitHubAppInstallationID(ctx, org, int64(numericAppID), []byte(privateKey))
	if err != nil {
		return nil, err
	}

	transport, err := ghinstallation.New(http.DefaultTransport, int64(numericAppID), installationID, []byte(privateKey))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create HTTP transport authenticated with GitHub app credentials")
	}

	return newGitHubClientWithHTTPClient(&http.Client{Transport: transport}), nil
}

func getGitHubAppInstallationID(ctx context.Context, org string, appID int64, privateKey []byte) (int64, error) {
	appsTransport, err := ghinstallation.NewAppsTransport(http.DefaultTransport, appID, privateKey)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to create GitHub apps transport to discover installation ID for the GitHub app %d", appID)
	}

	appClient := github.NewClient(&http.Client{Transport: appsTransport})
	installation, _, err := appClient.Apps.FindOrganizationInstallation(ctx, org)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to find the installation ID for the GitHub app %d", appID)
	}

	return installation.GetID(), nil
}

func newGitHubClientWithHTTPClient(httpClient *http.Client) *Client {
	return &Client{
		restClient: github.NewClient(httpClient),
	}
}
