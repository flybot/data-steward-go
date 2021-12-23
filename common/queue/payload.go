package queue

type Payload struct {
	Type    string `json:"task_type"`
	Payload string `json:"payload"`
}
