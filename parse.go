package omgo

import (
	"encoding/json"
	"time"
)

// rawResponse represents the raw JSON response from the API.
// This is used as an intermediate step for parsing.
type rawResponse struct {
	Latitude             float64 `json:"latitude"`
	Longitude            float64 `json:"longitude"`
	Elevation            float64 `json:"elevation"`
	Timezone             string  `json:"timezone"`
	TimezoneAbbreviation string  `json:"timezone_abbreviation"`
	UTCOffsetSeconds     int     `json:"utc_offset_seconds"`
	GenerationTimeMs     float64 `json:"generationtime_ms"`

	Current      json.RawMessage `json:"current,omitempty"`
	CurrentUnits *CurrentUnits   `json:"current_units,omitempty"`

	Hourly      json.RawMessage `json:"hourly,omitempty"`
	HourlyUnits json.RawMessage `json:"hourly_units,omitempty"`

	Minutely15      json.RawMessage `json:"minutely_15,omitempty"`
	Minutely15Units json.RawMessage `json:"minutely_15_units,omitempty"`

	Daily      json.RawMessage `json:"daily,omitempty"`
	DailyUnits json.RawMessage `json:"daily_units,omitempty"`
}

// rawCurrent represents the raw current weather data.
type rawCurrent struct {
	Time                string       `json:"time"`
	Interval            int          `json:"interval"`
	Temperature2m       *float64     `json:"temperature_2m,omitempty"`
	RelativeHumidity2m  *float64     `json:"relative_humidity_2m,omitempty"`
	ApparentTemperature *float64     `json:"apparent_temperature,omitempty"`
	IsDay               *int         `json:"is_day,omitempty"`
	Precipitation       *float64     `json:"precipitation,omitempty"`
	Rain                *float64     `json:"rain,omitempty"`
	Showers             *float64     `json:"showers,omitempty"`
	Snowfall            *float64     `json:"snowfall,omitempty"`
	WeatherCode         *WeatherCode `json:"weather_code,omitempty"`
	CloudCover          *float64     `json:"cloud_cover,omitempty"`
	PressureMSL         *float64     `json:"pressure_msl,omitempty"`
	SurfacePressure     *float64     `json:"surface_pressure,omitempty"`
	WindSpeed10m        *float64     `json:"wind_speed_10m,omitempty"`
	WindDirection10m    *float64     `json:"wind_direction_10m,omitempty"`
	WindGusts10m        *float64     `json:"wind_gusts_10m,omitempty"`
}

// rawHourly represents the raw hourly data with time as strings.
type rawHourly struct {
	Time []string `json:"time"`
	// All other fields will be unmarshaled directly into HourlyData
}

// rawDaily represents the raw daily data with time and sun times as strings.
type rawDaily struct {
	Time    []string `json:"time"`
	Sunrise []string `json:"sunrise,omitempty"`
	Sunset  []string `json:"sunset,omitempty"`
	// All other fields will be unmarshaled directly into DailyData
}

// rawMinutely15 represents the raw 15-minutely data with time as strings.
type rawMinutely15 struct {
	Time []string `json:"time"`
	// All other fields will be unmarshaled directly into Minutely15Data
}

// parseWeatherResponse parses the API response into a Weather struct.
func parseWeatherResponse(body []byte, models []string) (*Weather, error) {
	var raw rawResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	// Load timezone for proper time parsing
	var loc *time.Location
	if raw.Timezone != "" {
		var err error
		loc, err = time.LoadLocation(raw.Timezone)
		if err != nil {
			// Fall back to UTC if timezone is invalid
			loc = time.UTC
		}
	}

	weather := &Weather{
		Latitude:             raw.Latitude,
		Longitude:            raw.Longitude,
		Elevation:            raw.Elevation,
		Timezone:             raw.Timezone,
		TimezoneAbbreviation: raw.TimezoneAbbreviation,
		UTCOffsetSeconds:     raw.UTCOffsetSeconds,
		GenerationTimeMs:     raw.GenerationTimeMs,
		CurrentUnits:         raw.CurrentUnits,
		PrimaryModel:         firstModel(models),
	}

	// Parse units after model-aware normalization. The API returns one unit
	// object per model when model-specific cadence fields are returned, while
	// the public API intentionally exposes one request-level unit object.
	var err error
	if len(raw.HourlyUnits) > 0 {
		selected, selectErr := selectModelFields(raw.HourlyUnits, models, weather.PrimaryModel)
		if selectErr != nil {
			return nil, selectErr
		}
		var units *HourlyUnits
		if err = json.Unmarshal(selected, &units); err != nil {
			return nil, err
		}
		weather.HourlyUnits = units
	}
	if len(raw.Minutely15Units) > 0 {
		selected, selectErr := selectModelFields(raw.Minutely15Units, models, weather.PrimaryModel)
		if selectErr != nil {
			return nil, selectErr
		}
		var units *Minutely15Units
		if err = json.Unmarshal(selected, &units); err != nil {
			return nil, err
		}
		weather.Minutely15Units = units
	}
	if len(raw.DailyUnits) > 0 {
		selected, selectErr := selectModelFields(raw.DailyUnits, models, weather.PrimaryModel)
		if selectErr != nil {
			return nil, selectErr
		}
		var units *DailyUnits
		if err = json.Unmarshal(selected, &units); err != nil {
			return nil, err
		}
		weather.DailyUnits = units
	}

	// Parse current weather
	if len(raw.Current) > 0 {
		current, err := parseCurrent(raw.Current, loc)
		if err != nil {
			return nil, err
		}
		weather.Current = current
	}

	// Parse hourly data
	if len(raw.Hourly) > 0 {
		hourly, byModel, err := parseHourlySection(raw.Hourly, loc, models, weather.PrimaryModel)
		if err != nil {
			return nil, err
		}
		weather.Hourly = hourly
		weather.HourlyByModel = byModel
	}

	// Parse 15-minutely data
	if len(raw.Minutely15) > 0 {
		minutely15, byModel, err := parseMinutely15Section(raw.Minutely15, loc, models, weather.PrimaryModel)
		if err != nil {
			return nil, err
		}
		weather.Minutely15 = minutely15
		weather.Minutely15ByModel = byModel
	}

	// Parse daily data
	if len(raw.Daily) > 0 {
		daily, byModel, err := parseDailySection(raw.Daily, loc, models, weather.PrimaryModel)
		if err != nil {
			return nil, err
		}
		weather.Daily = daily
		weather.DailyByModel = byModel
	}

	// When multiple models were requested but the API returned only unsuffixed
	// cadence fields, it does not identify which requested model survived. Do
	// not attribute those compatibility fields to the first requested model.
	if len(effectiveModels(models)) > 1 &&
		weather.HourlyByModel == nil &&
		weather.Minutely15ByModel == nil &&
		weather.DailyByModel == nil {
		weather.PrimaryModel = ""
	}

	return weather, nil
}

// firstModel returns the first exactly non-empty model identifier in request
// order. It intentionally does not trim, normalize, or otherwise alter it.
func firstModel(models []string) string {
	for _, model := range models {
		if model != "" {
			return model
		}
	}
	return ""
}

func effectiveModels(models []string) []string {
	if len(models) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(models))
	effective := make([]string, 0, len(models))
	for _, model := range models {
		if model == "" {
			continue
		}
		if _, ok := seen[model]; ok {
			continue
		}
		seen[model] = struct{}{}
		effective = append(effective, model)
	}
	return effective
}

// splitModelFields separates model-suffixed object fields from fields that
// are shared by all requested models. Only exact model identifiers supplied by
// the request are considered.
func splitModelFields(data json.RawMessage, models []string) (json.RawMessage, map[string]json.RawMessage, error) {
	effective := effectiveModels(models)
	if len(effective) == 0 {
		return data, nil, nil
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, nil, err
	}
	if fields == nil { // JSON null
		return data, nil, nil
	}

	sharedFields := make(map[string]json.RawMessage)
	modelFields := make(map[string]map[string]json.RawMessage)
	for key, value := range fields {
		matchedModel := ""
		matchedSuffixLength := 0
		for _, model := range effective {
			suffix := "_" + model
			if len(key) <= len(suffix) || len(suffix) <= matchedSuffixLength {
				continue
			}
			if key[len(key)-len(suffix):] == suffix {
				matchedModel = model
				matchedSuffixLength = len(suffix)
			}
		}

		if matchedModel == "" {
			sharedFields[key] = value
			continue
		}
		if modelFields[matchedModel] == nil {
			modelFields[matchedModel] = make(map[string]json.RawMessage)
		}
		baseKey := key[:len(key)-matchedSuffixLength]
		modelFields[matchedModel][baseKey] = value
	}

	if len(modelFields) == 0 {
		return data, nil, nil
	}

	shared, err := json.Marshal(sharedFields)
	if err != nil {
		return nil, nil, err
	}
	byModel := make(map[string]json.RawMessage, len(modelFields))
	for _, model := range effective {
		fieldsForModel, ok := modelFields[model]
		if !ok {
			continue
		}
		merged := make(map[string]json.RawMessage, len(sharedFields)+len(fieldsForModel))
		for key, value := range sharedFields {
			merged[key] = value
		}
		for key, value := range fieldsForModel {
			merged[key] = value
		}
		normalized, err := json.Marshal(merged)
		if err != nil {
			return nil, nil, err
		}
		byModel[model] = normalized
	}
	return shared, byModel, nil
}

func selectModelFields(data json.RawMessage, models []string, primary string) (json.RawMessage, error) {
	shared, byModel, err := splitModelFields(data, models)
	if err != nil {
		return nil, err
	}
	if selected, ok := byModel[primary]; ok {
		return selected, nil
	}
	return shared, nil
}

func parseHourlySection(data json.RawMessage, loc *time.Location, models []string, primary string) (*HourlyData, map[string]*HourlyData, error) {
	shared, byModel, err := splitModelFields(data, models)
	if err != nil {
		return nil, nil, err
	}
	if byModel == nil {
		hourly, err := parseHourly(shared, loc)
		return hourly, nil, err
	}

	parsed := make(map[string]*HourlyData, len(byModel))
	for _, model := range effectiveModels(models) {
		normalized, ok := byModel[model]
		if !ok {
			continue
		}
		hourly, err := parseHourly(normalized, loc)
		if err != nil {
			return nil, nil, err
		}
		parsed[model] = hourly
	}
	if hourly, ok := parsed[primary]; ok {
		return hourly, parsed, nil
	}
	hourly, err := parseHourly(shared, loc)
	return hourly, parsed, err
}

func parseMinutely15Section(data json.RawMessage, loc *time.Location, models []string, primary string) (*Minutely15Data, map[string]*Minutely15Data, error) {
	shared, byModel, err := splitModelFields(data, models)
	if err != nil {
		return nil, nil, err
	}
	if byModel == nil {
		minutely15, err := parseMinutely15(shared, loc)
		return minutely15, nil, err
	}

	parsed := make(map[string]*Minutely15Data, len(byModel))
	for _, model := range effectiveModels(models) {
		normalized, ok := byModel[model]
		if !ok {
			continue
		}
		minutely15, err := parseMinutely15(normalized, loc)
		if err != nil {
			return nil, nil, err
		}
		parsed[model] = minutely15
	}
	if minutely15, ok := parsed[primary]; ok {
		return minutely15, parsed, nil
	}
	minutely15, err := parseMinutely15(shared, loc)
	return minutely15, parsed, err
}

func parseDailySection(data json.RawMessage, loc *time.Location, models []string, primary string) (*DailyData, map[string]*DailyData, error) {
	shared, byModel, err := splitModelFields(data, models)
	if err != nil {
		return nil, nil, err
	}
	if byModel == nil {
		daily, err := parseDaily(shared, loc)
		return daily, nil, err
	}

	parsed := make(map[string]*DailyData, len(byModel))
	for _, model := range effectiveModels(models) {
		normalized, ok := byModel[model]
		if !ok {
			continue
		}
		daily, err := parseDaily(normalized, loc)
		if err != nil {
			return nil, nil, err
		}
		parsed[model] = daily
	}
	if daily, ok := parsed[primary]; ok {
		return daily, parsed, nil
	}
	daily, err := parseDaily(shared, loc)
	return daily, parsed, err
}

// parseCurrent parses current weather data.
func parseCurrent(data json.RawMessage, loc *time.Location) (*CurrentData, error) {
	var raw rawCurrent
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	t, err := parseDateTime(raw.Time, loc)
	if err != nil {
		return nil, err
	}

	return &CurrentData{
		Time:                t,
		Interval:            raw.Interval,
		Temperature2m:       raw.Temperature2m,
		RelativeHumidity2m:  raw.RelativeHumidity2m,
		ApparentTemperature: raw.ApparentTemperature,
		IsDay:               raw.IsDay,
		Precipitation:       raw.Precipitation,
		Rain:                raw.Rain,
		Showers:             raw.Showers,
		Snowfall:            raw.Snowfall,
		WeatherCode:         raw.WeatherCode,
		CloudCover:          raw.CloudCover,
		PressureMSL:         raw.PressureMSL,
		SurfacePressure:     raw.SurfacePressure,
		WindSpeed10m:        raw.WindSpeed10m,
		WindDirection10m:    raw.WindDirection10m,
		WindGusts10m:        raw.WindGusts10m,
	}, nil
}

// parseHourly parses hourly weather data.
func parseHourly(data json.RawMessage, loc *time.Location) (*HourlyData, error) {
	// First, parse just the time array
	var rawTime rawHourly
	if err := json.Unmarshal(data, &rawTime); err != nil {
		return nil, err
	}

	// Parse times
	times, err := parseDateTimeArray(rawTime.Time, loc)
	if err != nil {
		return nil, err
	}

	// Parse all other fields into HourlyData
	hourly := &HourlyData{}
	if err := json.Unmarshal(data, hourly); err != nil {
		return nil, err
	}

	hourly.Times = times
	return hourly, nil
}

// parseMinutely15 parses 15-minutely weather data.
func parseMinutely15(data json.RawMessage, loc *time.Location) (*Minutely15Data, error) {
	// First, parse just the time array
	var rawTime rawMinutely15
	if err := json.Unmarshal(data, &rawTime); err != nil {
		return nil, err
	}

	// Parse times
	times, err := parseDateTimeArray(rawTime.Time, loc)
	if err != nil {
		return nil, err
	}

	// Parse all other fields into Minutely15Data
	minutely15 := &Minutely15Data{}
	if err := json.Unmarshal(data, minutely15); err != nil {
		return nil, err
	}

	minutely15.Times = times
	return minutely15, nil
}

// parseDaily parses daily weather data.
func parseDaily(data json.RawMessage, loc *time.Location) (*DailyData, error) {
	// First, parse time and sun times
	var rawTime rawDaily
	if err := json.Unmarshal(data, &rawTime); err != nil {
		return nil, err
	}

	// Parse dates
	times, err := parseDateArray(rawTime.Time, loc)
	if err != nil {
		return nil, err
	}

	// Parse all other fields into DailyData
	daily := &DailyData{}
	if err := json.Unmarshal(data, daily); err != nil {
		return nil, err
	}

	daily.Times = times

	// Parse sunrise/sunset times
	if len(rawTime.Sunrise) > 0 {
		sunrise, err := parseDateTimeArray(rawTime.Sunrise, loc)
		if err != nil {
			return nil, err
		}
		daily.Sunrise = sunrise
	}
	if len(rawTime.Sunset) > 0 {
		sunset, err := parseDateTimeArray(rawTime.Sunset, loc)
		if err != nil {
			return nil, err
		}
		daily.Sunset = sunset
	}

	return daily, nil
}
