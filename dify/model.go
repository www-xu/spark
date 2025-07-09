package dify

type InvokeWorkflowResponse struct {
	WorkflowRunID string       `json:"workflow_run_id"`
	TaskID        string       `json:"task_id"`
	Data          WorkflowData `json:"data"`
}

type WorkflowData struct {
	ID          string         `json:"id"`
	WorkflowID  string         `json:"workflow_id"`
	Status      string         `json:"status"`
	Outputs     map[string]any `json:"outputs"`
	Error       *string        `json:"error"`
	ElapsedTime float64        `json:"elapsed_time"`
	TotalTokens int            `json:"total_tokens"`
	TotalSteps  int            `json:"total_steps"`
	CreatedAt   int64          `json:"created_at"`
	FinishedAt  int64          `json:"finished_at"`
}
