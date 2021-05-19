package rsrch_client

import (
	"context"
	"fmt"
	"net/http"
)

//   WAIT_FOR_OFER we intend to take this struct from researcher-ui repository, still working on it
type VersionInfo struct {
	Version string `json:"version"` // 0.1.10
	Major   int    `json:"major"`   // 0
	Minor   int    `json:"minor"`   // 1
	Subver  int    `json:"subver"`  // 10
}

func NewVersionInfo(major, minor, subver int) *VersionInfo {
	return &VersionInfo{
		Version: fmt.Sprintf("%v.%v.%v", major, minor, subver),
		Major:   major,
		Minor:   minor,
		Subver:  subver,
	}
}

var (
	ProjectListMinVersion = *NewVersionInfo(0, 1, 10)
	DeleteJobMinVersion   = *NewVersionInfo(0, 1, 10)
)

func (c *RsrchClient) VersionGet(ctx context.Context) (*VersionInfo, error) {

	url := c.BaseURL + VersionURL

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(url), nil)
	if err != nil {
		return nil, err
	}

	res := VersionInfo{}
	if _, err := c.sendRequest(req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

func CompareVersion(versiona, versionb VersionInfo) int {
	if versiona.Major != versionb.Major {
		return versiona.Major - versionb.Major
	}
	if versiona.Minor != versionb.Minor {
		return versiona.Minor - versionb.Minor
	}
	if versiona.Subver != versionb.Subver {
		return versiona.Subver - versionb.Subver
	}
	return 0
}
