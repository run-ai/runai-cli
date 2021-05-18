package rsrch_client

const (

	VersionURL  = "/api/v1/version"
    ProjectsURL = "/api/v1/projects"
	JobsURL     = "/api/v1/jobs"

    //   full url of the researcher service, e.g. http://11.22.33.44:8080
    devRsrchUrlEnvVar = "RESEARCHER_SERVICE_URL"

    //   The RS port
    rsServicePort = "32282"

    HeaderContentType = "Content-Type"
    HeaderAccept = "Accept"
    HeaderAuth = "Authorization"

    KubeConfigIdToken = "id-token"

    AuthBearerPrefix = "Bearer "

    ContentTypeApplicationJson = "application/json; charset=utf-8"
)
