package rsrch_client

import (
	"context"
	"fmt"
	rsrch_server "github.com/run-ai/researcher-service/server/pkg/runai/api"
	"net/http"
)

//WAIT_FOR_OFER -> Should move to rsrch-service repository
func NewVersionInfo(major, minor, subver int) *rsrch_server.VersionInfo {
	return &rsrch_server.VersionInfo{
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

//
//   send version GET request to the RS
//
func (c *RsrchClient) VersionGet(ctx context.Context) (*rsrch_server.VersionInfo, error) {

	url := c.BaseURL + VersionURL

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(url), nil)
	if err != nil {
		return nil, err
	}

	res := rsrch_server.VersionInfo{}
	if _, err := c.sendRequest(req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

//
//    WAIT_FOR_OFER Should move to researcher service
//    compare two versions, e.g. 0.1.10 VS 0.1.11
//    returns (strcmp style)
//		0 -> the two versions are identical
//		>0 -> versionA > versionB
//		<0 -> versionA < versionB
//
func CompareVersion(versiona, versionb rsrch_server.VersionInfo) int {
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
func IsZeroVersion(version rsrch_server.VersionInfo) bool {
	return version.Major == 0 && version.Minor == 0 && version.Subver == 0
}
