package omgo

// EnsembleWeather contains one Ensemble API response. Member maps are
// discovered independently for each cadence and their count varies by model.
type EnsembleWeather struct {
	Latitude             float64 `json:"latitude"`
	Longitude            float64 `json:"longitude"`
	Elevation            float64 `json:"elevation"`
	Timezone             string  `json:"timezone"`
	TimezoneAbbreviation string  `json:"timezone_abbreviation"`
	UTCOffsetSeconds     int     `json:"utc_offset_seconds"`
	GenerationTimeMs     float64 `json:"generationtime_ms"`

	// Model is the exact selector supplied to NewEnsembleRequest.
	Model string `json:"-"`

	// HourlyByMember is keyed by dynamic member identifiers. Numbered upstream
	// suffixes are preserved, while member00 is synthesized for unsuffixed
	// member-valued series. Shared time and is_day values are copied to each
	// discovered member. The map is nil when the hourly section has no data.
	HourlyByMember map[string]*HourlyData `json:"-"`
	HourlyUnits    *HourlyUnits           `json:"hourly_units,omitempty"`

	// DailyByMember is keyed independently from HourlyByMember because a
	// response may expose different members or variables for each cadence.
	// Shared time, sunrise, sunset, and daylight_duration values are copied to
	// each discovered member. The map is nil when the daily section has no data.
	DailyByMember map[string]*DailyData `json:"-"`
	DailyUnits    *DailyUnits           `json:"daily_units,omitempty"`
}
