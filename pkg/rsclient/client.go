package rsclient

import (
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type RsClient struct {
    BaseURL    string
    HTTPClient *http.Client
}

type SuccessResponse struct {
    Data interface{} `json:"data"`
}

//WAIT_FOR_OFER load the URL from config
func NewRsClient() *RsClient {
    return &RsClient{
        BaseURL: RsBaseURL,
        HTTPClient: &http.Client{
            Timeout: time.Minute,
        },
    }
}

func (c *RsClient) sendRequest(req *http.Request, v interface{}) (int, error) {

    req.Header.Set("Content-Type", "application/json; charset=utf-8")
    req.Header.Set("Accept", "application/json; charset=utf-8")

    res, err := c.HTTPClient.Do(req)
    if err != nil {
        return -1, err
    }

    defer res.Body.Close()

    if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
        /* WAIT_FOR_OFER Error response?
        var errRes errorResponse
        if err = json.NewDecoder(res.Body).Decode(&errRes); err == nil {
            return errors.New(errRes.Message)
        }
        */
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

