package openrtb_ext

type ExtInfytvhb struct {
	DspID        string  `json:"dsp_id"`
	CustomerID   string  `json:"customer_id"`
	TagID        string  `json:"tag_id"`
	EndpointID   string  `json:"endpoint_id"`
	Base         string  `json:"base"`
	Path         string  `json:"path"`
	DspType      string  `json:"dsp_type"`
	MinCpm       float64 `json:"min_cpm"`
	MaxCpm       float64 `json:"max_cpm"`
	EndpointType string  `json:"type"`
	Floor        float64 `json:"floor_price"`
}
