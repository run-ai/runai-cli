git config --global url."https://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:x-oauth-basic@github.com/run-ai".insteadOf "https://github.com/run-ai" || true
GO111MODULE=on ${GENERAL_BUILD_OPTIONS} go test ./... -v -tags test

