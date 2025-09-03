module github.com/grafana/shared-workflows/actions/cleanup-stale-branches

go 1.24.3

require github.com/bradleyfalzon/ghinstallation/v2 v2.16.0

require github.com/go-logfmt/logfmt v0.5.1 // indirect

require (
	github.com/go-kit/log v0.2.1
	github.com/golang-jwt/jwt/v4 v4.5.2 // indirect
	github.com/google/go-github/v72 v72.0.0 // indirect
	github.com/google/go-github/v74 v74.0.0
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/pkg/errors v0.9.1
	golang.org/x/oauth2 v0.30.0
)
