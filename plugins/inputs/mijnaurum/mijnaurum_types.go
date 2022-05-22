package mijnaurum

type ActualsResponse struct {
	Actuals []struct {
		Source       string        `json:"source"`
		Type         string        `json:"type"`
		Measurements []interface{} `json:"measurements"`
		Baseline     float64       `json:"baseline"`
		ThisDay      struct {
			Value float64 `json:"value"`
			Cost  float64 `json:"cost"`
		} `json:"thisDay"`
		ThisWeek struct {
			Value float64 `json:"value"`
			Cost  float64 `json:"cost"`
		} `json:"thisWeek"`
		ThisMonth struct {
			Value float64 `json:"value"`
			Cost  float64 `json:"cost"`
		} `json:"thisMonth"`
		ThisYear struct {
			Value float64 `json:"value"`
			Cost  float64 `json:"cost"`
		} `json:"thisYear"`
	} `json:"actuals"`
}

type AuthenticationResponse struct {
	UserId string `json:"userId"`
}

type Source struct {
	Source     string `json:"source"`
	LocationID string `json:"locationId"`
	Type       string `json:"type"`
	Unit       string `json:"unit"`
	RateUnit   string `json:"rateUnit"`
	IsDefault  bool   `json:"isDefault"`
	MeterID    string `json:"meterId"`
}

type SourcesResponse struct {
	Sources         []Source `json:"sources"`
	ServerAddresses []struct {
		ServerAddress string `json:"serverAddress"`
	} `json:"serverAddresses"`
}
