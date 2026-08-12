package omgo

// Weather contains the response from the Open-Meteo API.
// This is the main response type for both Forecast and Historical requests.
type Weather struct {
	// Location information
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Elevation float64 `json:"elevation"`

	// Timezone information
	Timezone             string `json:"timezone"`
	TimezoneAbbreviation string `json:"timezone_abbreviation"`
	UTCOffsetSeconds     int    `json:"utc_offset_seconds"`

	// Generation time for performance monitoring
	GenerationTimeMs float64 `json:"generationtime_ms"`

	// PrimaryModel identifies the first explicitly requested non-empty model.
	// When that model has suffixed data for a cadence, it backs the corresponding
	// Hourly, Minutely15, or Daily compatibility field. PrimaryModel is empty for
	// requests without an explicit non-empty model.
	PrimaryModel string `json:"-"`

	// Current weather conditions
	Current      *CurrentData  `json:"-"` // parsed separately
	CurrentUnits *CurrentUnits `json:"current_units,omitempty"`

	// Hourly data
	Hourly      *HourlyData  `json:"-"` // parsed separately
	HourlyUnits *HourlyUnits `json:"hourly_units,omitempty"`
	// HourlyByModel contains model-specific hourly data when the response uses
	// requested model suffixes. It is nil when no such fields are present.
	HourlyByModel map[string]*HourlyData `json:"-"`

	// 15-minutely data
	Minutely15      *Minutely15Data  `json:"-"` // parsed separately
	Minutely15Units *Minutely15Units `json:"minutely_15_units,omitempty"`
	// Minutely15ByModel contains model-specific 15-minute data when the
	// response uses requested model suffixes. It is nil when no such fields are
	// present.
	Minutely15ByModel map[string]*Minutely15Data `json:"-"`

	// Daily data
	Daily      *DailyData  `json:"-"` // parsed separately
	DailyUnits *DailyUnits `json:"daily_units,omitempty"`
	// DailyByModel contains model-specific daily data when the response uses
	// requested model suffixes. It is nil when no such fields are present.
	DailyByModel map[string]*DailyData `json:"-"`
}
