package schema

type NodeUpdate struct {
	Node        string `json:"device"`
	Timestamp   string `json:"timestamp"`
	Description string `json:"description"`
}
