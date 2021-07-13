package rsrch_client

import (
	"context"
	"fmt"
	rsrch_api "github.com/run-ai/researcher-service/server/pkg/runai/api"
	"net/http"
)

type ProjectListOptions struct{}

func (c *RsrchClient) ProjectList(ctx context.Context, options *ProjectListOptions) (*[]rsrch_api.Project, error) {

	url := c.BaseURL + ProjectsURL

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(url), nil)
	if err != nil {
		return nil, err
	}

	var res []rsrch_api.Project
	if _, err := c.sendRequest(req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}
