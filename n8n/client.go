package n8n

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/www-xu/spark"
)

type Client struct {
	host            string
	authHeaderKey   string
	authHeaderValue string
	httpClient      *http.Client
}

func NewClient(host, authHeaderKey, authHeaderValue string) *Client {
	return &Client{
		host:            host,
		authHeaderKey:   authHeaderKey,
		authHeaderValue: authHeaderValue,
		httpClient:      &http.Client{},
	}
}

func (c *Client) InvokeWorkflow(ctx context.Context, workflowId string, inputs map[string]interface{}) (response *InvokeWorkflowResponse, err error) {
	requestBody := map[string]interface{}{
		"environment": spark.Env(),
		"inputs":      inputs,
	}
	requestBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s%s", c.host, fmt.Sprintf("/webhook/%s", workflowId)), bytes.NewReader(requestBytes))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set(c.authHeaderKey, c.authHeaderValue)

	rawResponse, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	responseBody, err := io.ReadAll(rawResponse.Body)
	if err != nil {
		return nil, err
	}

	if rawResponse.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("[%s] | %s", rawResponse.Status, string(responseBody)))
	}

	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		return nil, err
	}

	return response, nil
}
