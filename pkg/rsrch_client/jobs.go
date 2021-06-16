package rsrch_client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	rsrch_server "github.com/run-ai/researcher-service/server/pkg/runai/api"
)

func (c *RsrchClient) JobDelete(ctx context.Context, jobs []rsrch_server.ResourceID) ([]rsrch_server.JobActionStatus, error) {

	url := c.BaseURL + JobsURL

	body, _ := json.Marshal(jobs)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf(url), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	res := make([]rsrch_server.JobActionStatus, 0, len(jobs))
	if _, err := c.sendRequest(req, &res); err != nil {
		return nil, err
	}

	return res, nil
}

// func (c *RsrchClient) JobSuspend(ctx context.Context, jobs []rsrch_server.ResourceID) ([]rsrch_server.JobActionStatus, error) {

// 	url := c.BaseURL + JobsURL

// 	body, _ := json.Marshal(jobs)

// 	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf(url), bytes.NewReader(body))
// 	if err != nil {
// 		return nil, err
// 	}

// 	res := make([]rsrch_server.JobActionStatus, 0, len(jobs))
// 	if _, err := c.sendRequest(req, &res); err != nil {
// 		return nil, err
// 	}

// 	return res, nil
// }
