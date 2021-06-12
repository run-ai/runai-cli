package rsrch_client

import (
	"context"
	"encoding/json"
	"fmt"
	rsrch_server "github.com/run-ai/researcher-service/server/pkg/runai/api"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"
)

type RsrchClient struct {
	BaseURL    string
	authToken  string
	HTTPClient *http.Client
}

//
//    store the version of the researcher service. this helps us to figure out which
//    requests it supports
//
var rsVersion *rsrch_server.VersionInfo

type SuccessResponse struct {
	Data interface{} `json:"data"`
}

//
//   Creates RS client for sending REST requests
//   Parameters
//		rest.Config - k8s configuration used for creating the RS client
//		...VersionInfo - one or more minimal versions that we expect RS to comply with
//					for example, if we know that job deletion is supported only from version 0.1.10 of RS
//					then the function will receive this minimal version as parameter and will make sure that
//					RS complies with this version. otherwise, it will not return a RS client
//	Returns
//		Pointer to the RsrchClient to use for comminucating with RS, or nil if RS is not available or
//		does not comply with the versions that we expect it to comply with.
//
func NewRsrchClient(restConfig *rest.Config, mandatoryMinVersion rsrch_server.VersionInfo, additionalMinVersions ...rsrch_server.VersionInfo) *RsrchClient {

	//
	//   need to determine the URL to RS (researcher service)
	//
	rsUrl := &url.URL{}

	//
	//   for testing/debugging, allow the developer to specify RS URL
	//
	devRsrchUrl := os.Getenv(devRsrchUrlEnvVar)
	if devRsrchUrl != "" {
		rsUrl, _ = url.Parse(devRsrchUrl)
	} else {
		//
		//   in production, take it from the kubernetes config, but change the port to
		//   the port of the RS
		//
		mainUrl, err := url.Parse(restConfig.Host)
		if err != nil {
			log.Fatal(err)
		}
		host, _, _ := net.SplitHostPort(mainUrl.Host)
		rsUrl = &url.URL{Scheme: "http", Host: host + ":" + rsServicePort}
	}

	result := &RsrchClient{
		BaseURL: rsUrl.String(),
		HTTPClient: &http.Client{
			Timeout: time.Minute,
		},
	}

	if restConfig.AuthProvider != nil {
		result.authToken = restConfig.AuthProvider.Config[KubeConfigIdToken]
	}

	//
	//    if we did not investigate for the RS version yet, do it now
	//
	if rsVersion == nil {
		var err error
		rsVersion, err = result.VersionGet(context.TODO())
		if err != nil {
			log.Infof("Failed to obtain researcher-service version: %v", err.Error())
			rsVersion = NewVersionInfo(0, 0, 0)
		}
	}

	//
	//    make sure that the RS version complies with the minimal set of versions that we require
	//
	for _, minVersion := range append(additionalMinVersions, mandatoryMinVersion) {
		if CompareVersion(*rsVersion, minVersion) < 0 {
			if !IsZeroVersion(*rsVersion) {
				log.Warningf("researcher-service version %v < minimal required version %v\n", rsVersion.Version, minVersion.Version)
			}
			return nil
		}
	}

	return result
}

//
//    Send a request to the Researcher Service
//    Parameters
//       - Pointer to the request object
//       - Pointer to the response data, where the caller wants to receive the result
//    Returns
//       - Http Status Code (-1, if no error)
//       - error
//
func (c *RsrchClient) sendRequest(req *http.Request, v interface{}) (int, error) {

	req.Header.Set(HeaderContentType, ContentTypeApplicationJson)
	req.Header.Set(HeaderAccept, ContentTypeApplicationJson)
	if c.authToken != "" {
		req.Header.Set(HeaderAuth, AuthBearerPrefix+c.authToken)
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return -1, err
	}

	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		return res.StatusCode, fmt.Errorf("HTTP status code: %d", res.StatusCode)
	}

	fullResponse := SuccessResponse{
		Data: v,
	}

	if err = json.NewDecoder(res.Body).Decode(&fullResponse); err != nil {
		return -1, err
	}

	return res.StatusCode, nil
}
