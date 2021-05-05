GOPRIVATE="github.com/run-ai"
go env
GO111MODULE=on ${GENERAL_BUILD_OPTIONS} go test ./... -v
