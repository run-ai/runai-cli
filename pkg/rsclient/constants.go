package rsclient

const (

    //   WAIT_FOR_OFER need to get from researcher-service module
    GetProjectsURL = "/api/v1/projects"

    //   Developer provided RS URL, for testing/debugging
    devRsUrlEnvVar = "DEV_RS_URL"

    //   The RS port
    RsServicePort = "32282"

    HeaderContentType = "Content-Type"
    HeaderAccept = "Accept"
    ContentTypeApplicationJson = "application/json; charset=utf-8"
)
