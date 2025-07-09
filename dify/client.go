package dify

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/www-xu/spark/log"
)

type Client struct {
	host       string
	httpClient *http.Client
}

func NewClient(host string) *Client {
	return &Client{
		host:       host,
		httpClient: &http.Client{},
	}
}

func (c *Client) InvokeWorkflow(ctx context.Context, workflowId, userID string, inputs map[string]interface{}) (response *InvokeWorkflowResponse, err error) {
	requestBody := map[string]interface{}{
		"response_mode": "blocking",
		"inputs":        inputs,
		"user":          userID,
	}
	requestBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s%s", c.host, "/v1/workflows/run"), bytes.NewReader(requestBytes))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", workflowId))

	defer func() {
		if err != nil {
			log.WithContext(ctx).WithError(err).WithFields(map[string]interface{}{
				"request_body": requestBody,
				"workflow_id":  workflowId,
			}).Error("failed to invoke workflow")
		}
	}()

	rawResponse, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	responseBody, err := io.ReadAll(rawResponse.Body)
	if err != nil {
		return nil, err
	}

	if rawResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("[%s] | %s", rawResponse.Status, string(responseBody))
	}

	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		return nil, err
	}

	if response.Data.Error != nil && len(*response.Data.Error) > 0 {
		log.WithContext(ctx).WithError(errors.New(*response.Data.Error)).WithFields(map[string]interface{}{
			"response_body": responseBody,
		}).Error("failed to invoke workflow")
		return nil, errors.New(*response.Data.Error)
	}

	return response, nil
}
