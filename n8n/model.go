package n8n

type InvokeWorkflowResponse struct {
	Version     string `json:"version"`
	Data        string `json:"data"`
	ContentType string `json:"content_type"`
	TreeId      string `json:"tree_id"`
	TraceId     string `json:"trace_id"`
	SpanId      string `json:"span_id"`
}
