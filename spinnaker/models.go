package spinnaker

type PipelineExecution struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	BuildTime uint64 `json:"buildTime"`
	Status    string `json:"status"`
}
