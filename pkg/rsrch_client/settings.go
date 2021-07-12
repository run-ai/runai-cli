package rsrch_client

import (
	"context"
	"fmt"
	rsrch_api "github.com/run-ai/researcher-service/server/pkg/runai/api"
	rsrch_cs "github.com/run-ai/researcher-service/server/pkg/runai/client"
	"github.com/run-ai/researcher-service/server/pkg/server/template"
	"github.com/run-ai/runai-cli/pkg/client"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
)

type JobSettingsGetOptions struct {
	Interactive bool
}

func (self JobSettingsGetOptions) jobSettingsName() string {
	if self.Interactive {
		return template.InteractiveJobSettings
	} else {
		return template.TrainingJobSettings
	}
}

func (c *RsrchClient) JobSettingsGet(ctx context.Context, options JobSettingsGetOptions) (*rsrch_api.JobSettings, error) {

	url := c.BaseURL + SettingsURL + "/" + options.jobSettingsName()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(url), nil)
	if err != nil {
		return nil, err
	}

	var res rsrch_api.JobSettings
	if _, err := c.sendRequest(req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

func GetJobSettings(ctx context.Context, options JobSettingsGetOptions) (*rsrch_api.JobSettings, error) {

	restConfig, _, err := client.GetRestConfig()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	var result *rsrch_api.JobSettings

	rs := NewRsrchClient(restConfig, SettingsListMinVersion)
	if rs != nil {
		//
		//   RS can serve the request, so send it to RS
		//
		result, err = rs.JobSettingsGet(context.TODO(), options)
	} else {
		log.Infof("researcher-service cannot serve the request, use in-house CLI for job settings")

		clientSet, err := rsrch_cs.NewCliClientFromConfig(restConfig)
		if err != nil {
			log.Errorf("Failed to create clientSet for in-house CLI admin settings: %v", err.Error())
			return nil, err
		}

		result, err = clientSet.GetAdminTemplate(ctx, options.jobSettingsName())
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}
