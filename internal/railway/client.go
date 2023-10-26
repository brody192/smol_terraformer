package railway

import (
	"net/http"

	_ "github.com/brody192/genqlient/generate"
	"github.com/brody192/genqlient/graphql"
)

type authedTransport struct {
	token   string
	wrapped http.RoundTripper
}

type railwayClient struct {
	graphql.Client
}

func (t *authedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.token)
	req.Header.Set("Content-Type", "application/json")
	return t.wrapped.RoundTrip(req)
}

func NewAuthedClient(token string) *railwayClient {
	httpClient := http.Client{
		Transport: &authedTransport{
			token:   token,
			wrapped: http.DefaultTransport,
		},
	}

	return &railwayClient{
		graphql.NewClient("https://backboard.railway.app/graphql/internal", &httpClient),
	}
}
