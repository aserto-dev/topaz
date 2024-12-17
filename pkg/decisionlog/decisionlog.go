package decisionlog

type Config struct {
	Plugin   string                 `json:"plugin"`
	Settings map[string]interface{} `json:"settings"`
}
