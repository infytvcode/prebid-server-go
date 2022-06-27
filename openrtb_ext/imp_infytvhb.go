package openrtb_ext

type ExtInfytvhb struct {
	DspID      string  `json:"dsp_id"`
	CustomerID string  `json:"customer_id"`
	TagID      string  `json:"tag_id"`
	Base       string  `json:"base"`
	Path       string  `json:"path"`
	DspType    string  `json:"dsp_type"`
	MinCpm     float64 `json:"min_cpm"`
	MaxCpm     float64 `json:"max_cpm"`
}
