package openrtb_ext

type DspConfig struct {
	ConditionType string                           `json:"condition_type"`
	Conditions    map[string][]DspConfigValueEntry `json:"conditions"`
}

// DspConfigValueEntry represents the value entry inside the `values` map
type DspConfigValueEntry struct {
	KeyMap map[string]string `json:"key_map"`
	Place  string            `json:"place"`
}

type ExtInfytvHb struct {
	DspID        string      `json:"dsp_id"`
	CustomerID   string      `json:"customer_id"`
	TagID        string      `json:"tag_id"`
	EndpointID   string      `json:"endpoint_id"`
	DealID       string      `json:"deal_id"`
	Base         string      `json:"base"`
	Path         string      `json:"path"`
	DspType      string      `json:"dsp_type"`
	MinCpm       float64     `json:"min_cpm"`
	MaxCpm       float64     `json:"max_cpm"`
	EndpointType string      `json:"type"`
	DspConfigs   []DspConfig `json:"dsp_configs"`
	Floor        float64     `json:"floor_price"`
}
