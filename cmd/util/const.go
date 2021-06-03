package util

import v1 "k8s.io/api/core/v1"

const (
	RUNAI_QUEUE_LABEL = "runai/queue"
	CliCommand        = "runai-cli-command"

	AllStatuses = v1.PodPhase("")
)
