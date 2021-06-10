git config --global url."https://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:x-oauth-basic@github.com/run-ai".insteadOf "https://github.com/run-ai" || true

GO111MODULE=on ${GENERAL_BUILD_OPTIONS} go test ./... -v -tags test

