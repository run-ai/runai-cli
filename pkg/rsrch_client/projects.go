package rsrch_client

import (
    "context"
    "fmt"
    "net/http"
)

type ProjectListOptions struct {
    IncludeDeleted bool
}

//   WAIT_FOR_OFER we intend to take this struct from researcher-ui repository, still working on it
type Project struct {
    Name                        string       `json:"name"`
    IsDeleted                   bool         `json:"isDeleted"`
    CreatedAt                   int64        `json:"createdAt"`
    DeservedGpus                float64      `json:"deservedGpus"`
    InteractiveJobTimeLimitSecs int64        `json:"interactiveJobTimeLimitSecs"`
    TrainNodeAffinity           []string     `json:"trainNodeAffinity"`
    InteractiveNodeAffinity     []string     `json:"interactiveNodeAffinity"`
    DepartmentName              string       `json:"departmentName"`
}

type ProjectListResponse []Project

func (c *RsrchClient) ProjectList(ctx context.Context, options *ProjectListOptions) (*ProjectListResponse, error) {

    url := c.BaseURL + GetProjectsURL
    if options != nil {
        if options.IncludeDeleted {
            url = url + "?includeDeleted=true"
        }
    }

    req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(url), nil)
    if err != nil {
        return nil, err
    }

    req = req.WithContext(ctx)

    res := ProjectListResponse{}
    if _, err := c.sendRequest(req, &res); err != nil {
        return nil, err
    }

    return &res, nil
}

