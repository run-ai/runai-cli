package rsrch_client

import (
    "encoding/json"
    "fmt"
    "k8s.io/client-go/rest"
    "k8s.io/klog"
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

type SuccessResponse struct {
    Data interface{} `json:"data"`
}

//
//   Creates RS client for sending REST requests
//
func NewRsrchClient(restConfig *rest.Config) *RsrchClient {

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
            klog.Fatal(err)
        }
        host, _, _ := net.SplitHostPort(mainUrl.Host)
        rsUrl = &url.URL{Scheme: "http", Host: host + ":" + rsServicePort}
    }

    return &RsrchClient{
        BaseURL:    rsUrl.String(),
        authToken:  restConfig.AuthProvider.Config[KubeConfigIdToken],
        HTTPClient: &http.Client{
            Timeout: time.Minute,
        },
    }
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
        req.Header.Set(HeaderAuth, AuthBearerPrefix + c.authToken)
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

