package omgo

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestForecastRequestURL(t *testing.T) {
	req, err := NewForecastRequest(52.52, 13.41)
	require.NoError(t, err)

	req.WithHourly(HourlyTemperature2m, HourlyPrecipitation).
		WithDaily(DailyTemperature2mMax).
		WithTemperatureUnit(Celsius).
		WithTimezone("Europe/Berlin").
		WithForecastDays(7)

	rawURL := req.buildURL("https://api.open-meteo.com/v1/forecast", "")

	// Parse the URL to check parameters
	parsed, err := url.Parse(rawURL)
	require.NoError(t, err)

	params := parsed.Query()
	assert.Equal(t, "52.52", params.Get("latitude"))
	assert.Equal(t, "13.41", params.Get("longitude"))
	assert.Equal(t, "precipitation,temperature_2m", params.Get("hourly")) // sorted
	assert.Equal(t, "temperature_2m_max", params.Get("daily"))
	assert.Equal(t, "celsius", params.Get("temperature_unit"))
	assert.Equal(t, "Europe/Berlin", params.Get("timezone"))
	assert.Equal(t, "7", params.Get("forecast_days"))
	assert.Empty(t, params.Get("apikey"))
}

func TestForecastRequestWithAllOptions(t *testing.T) {
	req, err := NewForecastRequest(40.7128, -74.0060)
	require.NoError(t, err)

	tilt := 45.0
	azimuth := 180.0

	req.WithHourly(HourlyTemperature2m).
		WithDaily(DailySunrise, DailySunset).
		WithCurrent(CurrentTemperature2m, CurrentWeatherCode).
		WithMinutely15(Minutely15Precipitation).
		WithTemperatureUnit(Fahrenheit).
		WithWindSpeedUnit(MilesPerHour).
		WithPrecipitationUnit(Inches).
		WithTimezone("America/New_York").
		WithForecastDays(14).
		WithPastDays(2).
		WithCellSelection(CellSelectionNearest).
		WithTilt(tilt).
		WithAzimuth(azimuth)

	rawURL := req.buildURL("https://api.open-meteo.com/v1/forecast", "")
	parsed, err := url.Parse(rawURL)
	require.NoError(t, err)

	params := parsed.Query()
	assert.Equal(t, "40.7128", params.Get("latitude"))
	assert.Equal(t, "-74.006", params.Get("longitude"))
	assert.Equal(t, "temperature_2m", params.Get("hourly"))
	assert.Equal(t, "sunrise,sunset", params.Get("daily"))
	assert.Equal(t, "temperature_2m,weather_code", params.Get("current"))
	assert.Equal(t, "precipitation", params.Get("minutely_15"))
	assert.Equal(t, "fahrenheit", params.Get("temperature_unit"))
	assert.Equal(t, "mph", params.Get("wind_speed_unit"))
	assert.Equal(t, "inch", params.Get("precipitation_unit"))
	assert.Equal(t, "America/New_York", params.Get("timezone"))
	assert.Equal(t, "14", params.Get("forecast_days"))
	assert.Equal(t, "2", params.Get("past_days"))
	assert.Equal(t, "nearest", params.Get("cell_selection"))
	assert.Equal(t, "45", params.Get("tilt"))
	assert.Equal(t, "180", params.Get("azimuth"))
}

func TestHistoricalRequestURL(t *testing.T) {
	req, err := NewHistoricalRequest(52.52, 13.41, "2023-01-01", "2023-01-31")
	require.NoError(t, err)

	req.WithHourly(HourlyTemperature2m, HourlyPrecipitation).
		WithDaily(DailyTemperature2mMax, DailyTemperature2mMin).
		WithTimezone("Europe/Berlin")

	rawURL := req.buildURL("https://archive-api.open-meteo.com/v1/archive", "")
	parsed, err := url.Parse(rawURL)
	require.NoError(t, err)

	params := parsed.Query()
	assert.Equal(t, "52.52", params.Get("latitude"))
	assert.Equal(t, "13.41", params.Get("longitude"))
	assert.Equal(t, "2023-01-01", params.Get("start_date"))
	assert.Equal(t, "2023-01-31", params.Get("end_date"))
	assert.Equal(t, "precipitation,temperature_2m", params.Get("hourly")) // sorted
	assert.Equal(t, "temperature_2m_max,temperature_2m_min", params.Get("daily"))
	assert.Equal(t, "Europe/Berlin", params.Get("timezone"))
}

func TestForecastRequestWithAPIKey(t *testing.T) {
	req, err := NewForecastRequest(52.52, 13.41)
	require.NoError(t, err)

	req.WithHourly(HourlyTemperature2m)

	rawURL := req.buildURL("https://api.open-meteo.com/v1/forecast", "test-api-key")
	parsed, err := url.Parse(rawURL)
	require.NoError(t, err)

	params := parsed.Query()
	assert.Equal(t, "test-api-key", params.Get("apikey"))
}

func TestHistoricalRequestWithAPIKey(t *testing.T) {
	req, err := NewHistoricalRequest(52.52, 13.41, "2023-01-01", "2023-01-31")
	require.NoError(t, err)

	req.WithHourly(HourlyTemperature2m)

	rawURL := req.buildURL("https://archive-api.open-meteo.com/v1/archive", "test-api-key")
	parsed, err := url.Parse(rawURL)
	require.NoError(t, err)

	params := parsed.Query()
	assert.Equal(t, "test-api-key", params.Get("apikey"))
}

func TestHistoricalRequestValidation(t *testing.T) {
	// Missing start date
	_, err := NewHistoricalRequest(52.52, 13.41, "", "2023-01-31")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "startDate is required")

	// Missing end date
	_, err = NewHistoricalRequest(52.52, 13.41, "2023-01-01", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "endDate is required")
}

func TestLocationValidation(t *testing.T) {
	// Valid location
	loc, err := NewLocation(52.52, 13.41)
	require.NoError(t, err)
	assert.Equal(t, 52.52, loc.Latitude)
	assert.Equal(t, 13.41, loc.Longitude)

	// Invalid latitude (too high)
	_, err = NewLocation(91, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "latitude")

	// Invalid latitude (too low)
	_, err = NewLocation(-91, 0)
	assert.Error(t, err)

	// Invalid longitude (too high)
	_, err = NewLocation(0, 181)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "longitude")

	// Invalid longitude (too low)
	_, err = NewLocation(0, -181)
	assert.Error(t, err)

	// Edge cases (valid)
	_, err = NewLocation(90, 180)
	require.NoError(t, err)

	_, err = NewLocation(-90, -180)
	require.NoError(t, err)
}

func TestLocationWithElevation(t *testing.T) {
	loc, err := NewLocation(52.52, 13.41)
	require.NoError(t, err)

	loc = loc.WithElevation(100.5)
	require.NotNil(t, loc.Elevation)
	assert.Equal(t, 100.5, *loc.Elevation)
}

func TestMetricsDeduplication(t *testing.T) {
	req, err := NewForecastRequest(52.52, 13.41)
	require.NoError(t, err)

	// Add duplicate metrics via multiple calls
	req.WithHourly(HourlyTemperature2m, HourlyPrecipitation).
		WithHourly(HourlyTemperature2m).                    // duplicate
		WithHourly(HourlyWindSpeed10m, HourlyPrecipitation) // another duplicate

	rawURL := req.buildURL("https://api.open-meteo.com/v1/forecast", "")
	parsed, err := url.Parse(rawURL)
	require.NoError(t, err)

	params := parsed.Query()
	// Should be deduplicated and sorted
	assert.Equal(t, "precipitation,temperature_2m,wind_speed_10m", params.Get("hourly"))
}

func TestEnsembleRequestURLAndValidation(t *testing.T) {
	_, err := NewEnsembleRequest(52.52, 13.41, "")
	require.ErrorContains(t, err, "model is required")
	_, err = NewEnsembleRequest(52.52, 13.41, "icon_seamless,gfs_seamless")
	require.ErrorContains(t, err, "exactly one model is supported")
	_, err = NewEnsembleRequest(91, 13.41, "icon_seamless")
	require.ErrorContains(t, err, "latitude")

	req, err := NewEnsembleRequest(52.52, 13.41, " custom model ")
	require.NoError(t, err)
	hourlyMetrics := []HourlyMetric{HourlyTemperature2m, HourlyPrecipitation, HourlyTemperature2m}
	dailyMetrics := []DailyMetric{DailySunset, DailyTemperature2mMax, DailySunset}
	wantHourlyMetrics := append([]HourlyMetric(nil), hourlyMetrics...)
	wantDailyMetrics := append([]DailyMetric(nil), dailyMetrics...)
	req.WithLocation(req.location.WithElevation(44)).
		WithHourly(hourlyMetrics...).
		WithDaily(dailyMetrics...).
		WithTemperatureUnit(Fahrenheit).WithWindSpeedUnit(MilesPerHour).
		WithPrecipitationUnit(Inches).WithTimezone("Europe/Berlin").
		WithForecastDays(0).WithPastDays(-1).WithForecastHours(0).WithPastHours(-2).
		WithDateRange("2024-01-01", "2024-01-02").WithHourRange("2024-01-01T00:00", "2024-01-02T00:00").
		WithCellSelection(CellSelectionNearest).WithTilt(45).WithAzimuth(180)
	parsed, err := url.Parse(req.buildURL("https://example.test/ensemble", "key"))
	require.NoError(t, err)
	assert.Equal(t, wantHourlyMetrics, hourlyMetrics)
	assert.Equal(t, wantDailyMetrics, dailyMetrics)
	assert.Equal(t, wantHourlyMetrics, req.hourlyMetrics)
	assert.Equal(t, wantDailyMetrics, req.dailyMetrics)
	q := parsed.Query()
	assert.Equal(t, "52.52", q.Get("latitude"))
	assert.Equal(t, "13.41", q.Get("longitude"))
	assert.Equal(t, "44", q.Get("elevation"))
	require.Len(t, q["models"], 1)
	assert.Equal(t, " custom model ", q.Get("models"))
	assert.Equal(t, "precipitation,temperature_2m", q.Get("hourly"))
	assert.Equal(t, "sunset,temperature_2m_max", q.Get("daily"))
	assert.Equal(t, "fahrenheit", q.Get("temperature_unit"))
	assert.Equal(t, "mph", q.Get("wind_speed_unit"))
	assert.Equal(t, "inch", q.Get("precipitation_unit"))
	assert.Equal(t, "Europe/Berlin", q.Get("timezone"))
	assert.Equal(t, "0", q.Get("forecast_days"))
	assert.Equal(t, "-1", q.Get("past_days"))
	assert.Equal(t, "0", q.Get("forecast_hours"))
	assert.Equal(t, "-2", q.Get("past_hours"))
	assert.Equal(t, "2024-01-01", q.Get("start_date"))
	assert.Equal(t, "2024-01-02", q.Get("end_date"))
	assert.Equal(t, "2024-01-01T00:00", q.Get("start_hour"))
	assert.Equal(t, "2024-01-02T00:00", q.Get("end_hour"))
	assert.Equal(t, "nearest", q.Get("cell_selection"))
	assert.Equal(t, "45", q.Get("tilt"))
	assert.Equal(t, "180", q.Get("azimuth"))
	assert.Equal(t, "key", q.Get("apikey"))
	assert.Equal(t, "https://example.test/ensemble", parsed.Scheme+"://"+parsed.Host+parsed.Path)

	bare, err := NewEnsembleRequest(1, 2, "icon_seamless")
	require.NoError(t, err)
	bareQ, err := url.Parse(bare.buildURL("https://example.test", ""))
	require.NoError(t, err)
	assert.False(t, bareQ.Query().Has("forecast_days"))
	assert.False(t, bareQ.Query().Has("forecast_hours"))
	assert.False(t, bareQ.Query().Has("past_days"))
	assert.False(t, bareQ.Query().Has("past_hours"))
	assert.False(t, bareQ.Query().Has("apikey"))
}
