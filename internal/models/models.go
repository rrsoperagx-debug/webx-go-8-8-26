
package models

type Metric struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	Value     float64 `json:"value"`
	Labels    string  `json:"labels"`
	Timestamp string  `json:"timestamp"`
}

type IngestPayload struct {
	Name   string   `json:"name"`
	Value  float64  `json:"value"`
	Labels *string  `json:"labels,omitempty"`
}

type Device struct {
	Name       string `json:"name"`
	Resolution string `json:"resolution"`
	Category   string `json:"category"`
	Support    string `json:"support"`
}

type Browser struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Engine      string `json:"engine"`
	Support     string `json:"support"`
	MarketShare string `json:"market_share"`
}

type MetricsView struct {
	CPU           float64 `json:"cpu"`
	Mem           float64 `json:"mem"`
	Users         float64 `json:"users"`
	RPS           float64 `json:"rps"`
	RequestsTotal float64 `json:"requests_total"`
	ErrorsTotal   float64 `json:"errors_total"`
}
