package rsrch_client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

//   WAIT_FOR_OFER we intend to take this struct from researcher-ui repository, still working on it
type DeletedJob struct {
	Name    string `json:"name"`
	Project string `json:"project"`
}

//   WAIT_FOR_OFER we intend to take this struct from researcher-ui repository, still working on it
type DeletedJobStatus struct {
	Name  string `json:"name"`
	Ok    bool   `json:"ok"`
	Error *Error `json:"error"`
}

func (c *RsrchClient) JobDelete(ctx context.Context, jobs []DeletedJob) ([]DeletedJobStatus, error) {

	url := c.BaseURL + JobsURL

	body, _ := json.Marshal(jobs)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf(url), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	res := make([]DeletedJobStatus, 0, len(jobs))
	if _, err := c.sendRequest(req, &res); err != nil {
		return nil, err
	}

	return res, nil
}
