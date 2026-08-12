package omgo

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseHourlyForecast(t *testing.T) {
	data, err := os.ReadFile("testdata/forecast_hourly.json")
	require.NoError(t, err)

	weather, err := parseWeatherResponse(data, nil)
	require.NoError(t, err)

	// Check metadata
	assert.InDelta(t, 52.52, weather.Latitude, 0.01)
	assert.InDelta(t, 13.42, weather.Longitude, 0.01)
	assert.Equal(t, "Europe/Berlin", weather.Timezone)
	assert.Equal(t, "CET", weather.TimezoneAbbreviation)
	assert.Equal(t, 3600, weather.UTCOffsetSeconds)

	// Check hourly data
	require.NotNil(t, weather.Hourly)
	assert.Len(t, weather.Hourly.Times, 3)
	assert.Len(t, weather.Hourly.Temperature2m, 3)
	assert.Len(t, weather.Hourly.RelativeHumidity2m, 3)
	assert.Len(t, weather.Hourly.Precipitation, 3)
	assert.Len(t, weather.Hourly.WeatherCode, 3)
	assert.Len(t, weather.Hourly.WindSpeed10m, 3)

	// Check values
	assert.Equal(t, 2.5, weather.Hourly.Temperature2m[0])
	assert.Equal(t, WeatherCode(3), weather.Hourly.WeatherCode[0])
	assert.Equal(t, WeatherCode(61), weather.Hourly.WeatherCode[1])

	// Check time parsing with timezone
	loc, _ := time.LoadLocation("Europe/Berlin")
	expectedTime := time.Date(2024, 1, 15, 0, 0, 0, 0, loc)
	assert.Equal(t, expectedTime, weather.Hourly.Times[0])

	// Check units
	require.NotNil(t, weather.HourlyUnits)
	assert.Equal(t, "°C", weather.HourlyUnits.Temperature2m)
	assert.Equal(t, "mm", weather.HourlyUnits.Precipitation)
	assert.Equal(t, "km/h", weather.HourlyUnits.WindSpeed10m)
	assert.Empty(t, weather.PrimaryModel)
	assert.Nil(t, weather.HourlyByModel)
	assert.Nil(t, weather.Minutely15ByModel)
	assert.Nil(t, weather.DailyByModel)
}

func TestParseDailyForecast(t *testing.T) {
	data, err := os.ReadFile("testdata/forecast_daily.json")
	require.NoError(t, err)

	weather, err := parseWeatherResponse(data, nil)
	require.NoError(t, err)

	// Check daily data
	require.NotNil(t, weather.Daily)
	assert.Len(t, weather.Daily.Times, 3)
	assert.Len(t, weather.Daily.Temperature2mMax, 3)
	assert.Len(t, weather.Daily.Temperature2mMin, 3)
	assert.Len(t, weather.Daily.Sunrise, 3)
	assert.Len(t, weather.Daily.Sunset, 3)

	// Check values
	assert.Equal(t, 5.2, weather.Daily.Temperature2mMax[0])
	assert.Equal(t, -1.2, weather.Daily.Temperature2mMin[0])

	// Check time parsing
	loc, _ := time.LoadLocation("Europe/Berlin")
	expectedDate := time.Date(2024, 1, 15, 0, 0, 0, 0, loc)
	assert.Equal(t, expectedDate, weather.Daily.Times[0])

	// Check sunrise/sunset parsing
	expectedSunrise := time.Date(2024, 1, 15, 8, 15, 0, 0, loc)
	assert.Equal(t, expectedSunrise, weather.Daily.Sunrise[0])

	expectedSunset := time.Date(2024, 1, 15, 16, 30, 0, 0, loc)
	assert.Equal(t, expectedSunset, weather.Daily.Sunset[0])

	// Check units
	require.NotNil(t, weather.DailyUnits)
	assert.Equal(t, "°C", weather.DailyUnits.Temperature2mMax)
}

func TestParseMultipleModelForecast(t *testing.T) {
	data, err := os.ReadFile("testdata/forecast_multiple_models.json")
	require.NoError(t, err)

	weather, err := parseWeatherResponse(data, []string{"gfs_global", "ecmwf_ifs"})
	require.NoError(t, err)

	assert.Equal(t, "gfs_global", weather.PrimaryModel)
	assert.Equal(t, 3.5, *weather.Current.Temperature2m)
	assert.Equal(t, "°C", weather.CurrentUnits.Temperature2m)

	require.Len(t, weather.HourlyByModel, 2)
	require.Len(t, weather.Minutely15ByModel, 2)
	require.Len(t, weather.DailyByModel, 2)
	assert.Contains(t, weather.HourlyByModel, "gfs_global")
	assert.Contains(t, weather.HourlyByModel, "ecmwf_ifs")
	assert.Contains(t, weather.Minutely15ByModel, "gfs_global")
	assert.Contains(t, weather.Minutely15ByModel, "ecmwf_ifs")
	assert.Contains(t, weather.DailyByModel, "gfs_global")
	assert.Contains(t, weather.DailyByModel, "ecmwf_ifs")

	gfsHourly := weather.HourlyByModel["gfs_global"]
	ecmwfHourly := weather.HourlyByModel["ecmwf_ifs"]
	require.NotNil(t, gfsHourly)
	require.NotNil(t, ecmwfHourly)
	assert.Equal(t, []float64{2.8, 2.6, 2.4}, gfsHourly.Temperature2m)
	assert.Equal(t, []float64{2.5, 2.3, 2.1}, ecmwfHourly.Temperature2m)
	assert.Equal(t, []float64{0.2, 0.0, 0.1}, gfsHourly.Precipitation)
	assert.Nil(t, gfsHourly.RelativeHumidity2m)
	assert.Equal(t, []float64{85, 86, 87}, ecmwfHourly.RelativeHumidity2m)
	assert.Nil(t, ecmwfHourly.Precipitation)
	assert.Len(t, gfsHourly.Times, 3)
	assert.Len(t, ecmwfHourly.Times, 3)
	assert.Same(t, gfsHourly, weather.Hourly)

	assert.Equal(t, []float64{2.8, 2.7, 2.6}, weather.Minutely15ByModel["gfs_global"].Temperature2m)
	assert.Equal(t, []float64{2.5, 2.4, 2.3}, weather.Minutely15ByModel["ecmwf_ifs"].Temperature2m)
	assert.Equal(t, []float64{0.1, 0.0, 0.0}, weather.Minutely15ByModel["gfs_global"].Precipitation)
	assert.Equal(t, []float64{0.0, 0.0, 0.1}, weather.Minutely15ByModel["ecmwf_ifs"].Precipitation)
	assert.Len(t, weather.Minutely15ByModel["gfs_global"].Times, 3)
	assert.Len(t, weather.Minutely15ByModel["ecmwf_ifs"].Times, 3)
	assert.Same(t, weather.Minutely15ByModel["gfs_global"], weather.Minutely15)

	loc, err := time.LoadLocation("Europe/Berlin")
	require.NoError(t, err)
	expectedHourlyTime := time.Date(2024, 1, 15, 0, 0, 0, 0, loc)
	expectedMinutelyTime := time.Date(2024, 1, 15, 0, 15, 0, 0, loc)
	expectedDailyTime := time.Date(2024, 1, 16, 0, 0, 0, 0, loc)
	assert.Equal(t, expectedHourlyTime, weather.HourlyByModel["gfs_global"].Times[0])
	assert.Equal(t, expectedHourlyTime, weather.HourlyByModel["ecmwf_ifs"].Times[0])
	assert.Equal(t, expectedMinutelyTime, weather.Minutely15ByModel["gfs_global"].Times[1])
	assert.Equal(t, expectedMinutelyTime, weather.Minutely15ByModel["ecmwf_ifs"].Times[1])
	assert.Equal(t, expectedDailyTime, weather.DailyByModel["gfs_global"].Times[1])
	assert.Equal(t, expectedDailyTime, weather.DailyByModel["ecmwf_ifs"].Times[1])
	assert.Equal(t, time.Date(2024, 1, 15, 8, 20, 0, 0, loc), weather.DailyByModel["gfs_global"].Sunrise[0])
	assert.Equal(t, time.Date(2024, 1, 15, 8, 15, 0, 0, loc), weather.DailyByModel["ecmwf_ifs"].Sunrise[0])
	assert.Equal(t, time.Date(2024, 1, 17, 16, 39, 0, 0, loc), weather.DailyByModel["gfs_global"].Sunset[2])
	assert.Equal(t, []float64{5.8, 6.4, 5.1}, weather.DailyByModel["gfs_global"].Temperature2mMax)
	assert.Equal(t, []float64{5.2, 6.1, 4.8}, weather.DailyByModel["ecmwf_ifs"].Temperature2mMax)
	assert.Equal(t, []WeatherCode{2, 61, 3}, weather.DailyByModel["gfs_global"].WeatherCode)
	assert.Equal(t, []WeatherCode{3, 61, 45}, weather.DailyByModel["ecmwf_ifs"].WeatherCode)
	assert.Len(t, weather.DailyByModel["gfs_global"].Times, 3)
	assert.Len(t, weather.DailyByModel["ecmwf_ifs"].Times, 3)
	assert.Same(t, weather.DailyByModel["gfs_global"], weather.Daily)

	require.NotNil(t, weather.HourlyUnits)
	assert.Equal(t, "°C", weather.HourlyUnits.Temperature2m)
	assert.Equal(t, "mm", weather.HourlyUnits.Precipitation)
	require.NotNil(t, weather.Minutely15Units)
	assert.Equal(t, "°C", weather.Minutely15Units.Temperature2m)
	assert.Equal(t, "mm", weather.Minutely15Units.Precipitation)
	require.NotNil(t, weather.DailyUnits)
	assert.Equal(t, "°C", weather.DailyUnits.Temperature2mMax)
	assert.Equal(t, "wmo code", weather.DailyUnits.WeatherCode)
	assert.Equal(t, "iso8601", weather.DailyUnits.Sunrise)
}

func TestSplitModelFields(t *testing.T) {
	data := json.RawMessage(`{"time":["2024-01-15T00:00"],"shared":"value","temperature_2m_global":[1],"temperature_2m_gfs_global":[2],"sunrise_gfs_global":["2024-01-15T08:00"]}`)
	shared, byModel, err := splitModelFields(data, []string{"global", "gfs_global"})
	require.NoError(t, err)

	var sharedFields map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(shared, &sharedFields))
	assert.Contains(t, sharedFields, "time")
	assert.Contains(t, sharedFields, "shared")
	assert.NotContains(t, sharedFields, "temperature_2m_global")
	assert.Len(t, byModel, 2)
	assert.Contains(t, byModel, "global")
	assert.Contains(t, byModel, "gfs_global")

	var gfsFields map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(byModel["gfs_global"], &gfsFields))
	assert.Contains(t, gfsFields, "time")
	assert.Contains(t, gfsFields, "shared")
	assert.Contains(t, gfsFields, "sunrise")
	assert.JSONEq(t, `[2]`, string(gfsFields["temperature_2m"]))

	var globalFields map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(byModel["global"], &globalFields))
	assert.JSONEq(t, `[1]`, string(globalFields["temperature_2m"]))
}

func TestSplitModelFieldsModelOverrideAndPartialAvailability(t *testing.T) {
	data := json.RawMessage(`{"time":["2024-01-15T00:00"],"temperature_2m":[0],"temperature_2m_gfs_global":[2],"rain_gfs_global":[1]}`)
	_, byModel, err := splitModelFields(data, []string{"gfs_global", "ecmwf_ifs"})
	require.NoError(t, err)
	assert.Contains(t, byModel, "gfs_global")
	assert.NotContains(t, byModel, "ecmwf_ifs")

	var gfsFields map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(byModel["gfs_global"], &gfsFields))
	assert.JSONEq(t, `[2]`, string(gfsFields["temperature_2m"]))
}

func TestSplitModelFieldsCompatibilityCases(t *testing.T) {
	data := json.RawMessage(`{ "time": [] }`)
	shared, byModel, err := splitModelFields(data, []string{"", ""})
	require.NoError(t, err)
	assert.Nil(t, byModel)
	assert.Equal(t, data, shared)

	shared, byModel, err = splitModelFields(data, []string{"gfs_global"})
	require.NoError(t, err)
	assert.Nil(t, byModel)
	assert.Equal(t, data, shared)

	for _, input := range []json.RawMessage{json.RawMessage(`[`), json.RawMessage(`[]`), json.RawMessage(`"value"`), json.RawMessage(`1`), json.RawMessage(`true`)} {
		_, _, err = splitModelFields(input, []string{"gfs_global"})
		assert.Error(t, err)
	}
	shared, byModel, err = splitModelFields(json.RawMessage(`null`), []string{"gfs_global"})
	require.NoError(t, err)
	assert.Equal(t, "null", string(shared))
	assert.Nil(t, byModel)
}

func TestSplitModelFieldsOnlyMatchesRequestedModelsAndPreservesInput(t *testing.T) {
	data := json.RawMessage(`{"time":[],"temperature_2m_ecmwf_ifs":[1],"temperature_2m_gfs_global":[2]}`)
	models := []string{"", "ecmwf_ifs", "ecmwf_ifs"}
	wantModels := append([]string(nil), models...)

	_, byModel, err := splitModelFields(data, models)
	require.NoError(t, err)
	assert.Equal(t, wantModels, models)
	require.Len(t, byModel, 1)
	assert.Contains(t, byModel, "ecmwf_ifs")
	assert.NotContains(t, byModel, "gfs_global")

	var fields map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(byModel["ecmwf_ifs"], &fields))
	assert.JSONEq(t, `[1]`, string(fields["temperature_2m"]))
	assert.JSONEq(t, `[2]`, string(fields["temperature_2m_gfs_global"]))
}

func TestParseEnsembleResponse(t *testing.T) {
	data, err := os.ReadFile("testdata/ensemble.json")
	require.NoError(t, err)
	weather, err := parseEnsembleResponse(data, "icon_seamless")
	require.NoError(t, err)
	assert.InDelta(t, 52.52, weather.Latitude, 0.01)
	assert.InDelta(t, 13.41, weather.Longitude, 0.01)
	assert.Equal(t, 38.0, weather.Elevation)
	assert.Equal(t, "icon_seamless", weather.Model)
	assert.Equal(t, "Europe/Berlin", weather.Timezone)
	assert.Equal(t, "CET", weather.TimezoneAbbreviation)
	assert.Equal(t, 3600, weather.UTCOffsetSeconds)
	assert.Equal(t, 1.5, weather.GenerationTimeMs)
	require.Len(t, weather.HourlyByMember, 3)
	assert.Equal(t, []float64{2, 2.1, 2.2}, weather.HourlyByMember["member00"].Temperature2m)
	assert.Equal(t, []float64{3, 3.1, 3.2}, weather.HourlyByMember["member01"].Temperature2m)
	assert.Nil(t, weather.HourlyByMember["member02"].Precipitation)
	for _, member := range weather.HourlyByMember {
		assert.Equal(t, []int{0, 0, 0}, member.IsDay)
		assert.Len(t, member.Times, 3)
	}
	require.Len(t, weather.DailyByMember, 3)
	assert.Nil(t, weather.DailyByMember["member01"].Temperature2mMax)
	assert.Equal(t, []float64{1, 2}, weather.DailyByMember["member01"].PrecipitationSum)
	assert.Nil(t, weather.DailyByMember["member00"].PrecipitationSum)
	assert.Equal(t, []float64{7, 8}, weather.DailyByMember["member02"].Temperature2mMax)
	for _, member := range weather.DailyByMember {
		assert.Len(t, member.Times, 2)
		assert.Len(t, member.Sunrise, 2)
		assert.Len(t, member.Sunset, 2)
		assert.Equal(t, []float64{29700, 29880}, member.DaylightDuration)
	}
	require.NotNil(t, weather.HourlyUnits)
	assert.Equal(t, "°C", weather.HourlyUnits.Temperature2m)
	require.NotNil(t, weather.DailyUnits)
	assert.Equal(t, "°C", weather.DailyUnits.Temperature2mMax)
	loc, err := time.LoadLocation("Europe/Berlin")
	require.NoError(t, err)
	assert.Equal(t, time.Date(2024, 1, 15, 0, 0, 0, 0, loc), weather.HourlyByMember["member00"].Times[0])
	assert.Equal(t, time.Date(2024, 1, 15, 0, 0, 0, 0, loc), weather.DailyByMember["member02"].Times[0])
	assert.Equal(t, time.Date(2024, 1, 15, 8, 15, 0, 0, loc), weather.DailyByMember["member01"].Sunrise[0])
}

func TestSplitMemberFields(t *testing.T) {
	data := json.RawMessage(`{"time":["2024-01-01T00:00"],"is_day":[0],"temperature_2m":[1],"temperature_2m_member01":[2],"rain_member03":[3]}`)
	wantData := append(json.RawMessage(nil), data...)
	members, err := splitMemberFields(data, map[string]bool{"is_day": true})
	require.NoError(t, err)
	assert.Equal(t, wantData, data)
	require.Len(t, members, 3)
	var member1 map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(members["member01"], &member1))
	assert.JSONEq(t, `[0]`, string(member1["is_day"]))
	assert.JSONEq(t, `[2]`, string(member1["temperature_2m"]))
	var member3 map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(members["member03"], &member3))
	assert.NotContains(t, member3, "temperature_2m")

	sharedOnly, err := splitMemberFields(json.RawMessage(`{"time":[],"is_day":[]}`), map[string]bool{"is_day": true})
	require.NoError(t, err)
	assert.Contains(t, sharedOnly, "member00")
	invariantOnly, err := splitMemberFields(json.RawMessage(`{"is_day":[]}`), map[string]bool{"is_day": true})
	require.NoError(t, err)
	assert.Contains(t, invariantOnly, "member00")
	timeOnly, err := splitMemberFields(json.RawMessage(`{"time":[]}`), map[string]bool{"is_day": true})
	require.NoError(t, err)
	assert.Nil(t, timeOnly)
	empty, err := splitMemberFields(json.RawMessage(`{}`), map[string]bool{"is_day": true})
	require.NoError(t, err)
	assert.Nil(t, empty)

	for _, key := range []string{"metric_member", "metric_memberx", "metric_member01_extra", "_member01", "metric_member١"} {
		members, err := splitMemberFields(json.RawMessage(`{"`+key+`":[]}`), nil)
		require.NoError(t, err)
		assert.Contains(t, members, "member00", key)
	}
	for _, input := range []json.RawMessage{json.RawMessage(`[]`), json.RawMessage(`"value"`), json.RawMessage(`{`)} {
		_, err := splitMemberFields(input, nil)
		assert.Error(t, err)
	}
}

func TestParseEnsembleResponseRejectsAmbiguityAndNulls(t *testing.T) {
	for _, test := range []struct {
		name     string
		body     string
		contains []string
	}{
		{"normalized collision", `{"hourly":{"time":[],"temperature_2m":[],"temperature_2m_member00":[]}}`, []string{"hourly", "member00", "temperature_2m"}},
		{"shared collision", `{"hourly":{"time":[],"time_member01":[]}}`, []string{"hourly", "member01", "time"}},
		{"null series", `{"hourly":{"time":[],"temperature_2m":null}}`, []string{"hourly", "member00", "temperature_2m", "null"}},
		{"null element", `{"hourly":{"time":[],"temperature_2m":[1,null]}}`, []string{"hourly", "member00", "temperature_2m", "element 1"}},
		{"top-level null", `null`, []string{"ensemble response", "JSON object"}},
		{"top-level array", `[]`, []string{"ensemble response"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := parseEnsembleResponse([]byte(test.body), "icon")
			require.Error(t, err)
			for _, part := range test.contains {
				assert.Contains(t, err.Error(), part)
			}
		})
	}

	weather, err := parseEnsembleResponse([]byte(`{"hourly":null,"hourly_units":null,"daily":null,"daily_units":null}`), "icon")
	require.NoError(t, err)
	assert.Nil(t, weather.HourlyByMember)
	assert.Nil(t, weather.HourlyUnits)
	assert.Nil(t, weather.DailyByMember)
	assert.Nil(t, weather.DailyUnits)

	_, err = parseEnsembleResponse([]byte(`{"hourly":{"time":[],"temperature_2m_member07":[1,null]}}`), "icon")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "hourly member member07")
	assert.Contains(t, err.Error(), "temperature_2m")
	assert.Contains(t, err.Error(), "element 1")
	_, err = parseEnsembleResponse([]byte(`{"hourly":{"time":["not-a-time"],"temperature_2m_member07":[1]}}`), "icon")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing hourly member member07")
}

func TestParseEnsembleResponseSharedOnlyUnitsAndTimezoneFallback(t *testing.T) {
	body := []byte(`{
		"timezone":"Not/A_Timezone",
		"hourly_units":{"temperature_2m":"base","temperature_2m_member01":"ignored"},
		"daily_units":{"temperature_2m_max":"base","temperature_2m_max_member01":"ignored"},
		"hourly":{"is_day":[1]},
		"daily":{"sunrise":["2024-01-15T08:15"]}
	}`)
	weather, err := parseEnsembleResponse(body, " exact model ")
	require.NoError(t, err)
	assert.Equal(t, " exact model ", weather.Model)
	require.NotNil(t, weather.HourlyUnits)
	assert.Equal(t, "base", weather.HourlyUnits.Temperature2m)
	require.NotNil(t, weather.DailyUnits)
	assert.Equal(t, "base", weather.DailyUnits.Temperature2mMax)
	require.Contains(t, weather.HourlyByMember, "member00")
	assert.Equal(t, []int{1}, weather.HourlyByMember["member00"].IsDay)
	require.Contains(t, weather.DailyByMember, "member00")
	assert.Equal(t, time.Date(2024, 1, 15, 8, 15, 0, 0, time.UTC), weather.DailyByMember["member00"].Sunrise[0])
}

func TestSplitMemberFieldsReportsCollisionsDeterministically(t *testing.T) {
	data := json.RawMessage(`{"time":[],"time_member01":[],"is_day":[],"is_day_member01":[]}`)
	for i := 0; i < 20; i++ {
		_, err := splitMemberFields(data, map[string]bool{"is_day": true})
		require.EqualError(t, err, `member member01 field "is_day" collides with a shared field`)
	}
}

func TestFirstModelUsesFirstExactlyNonEmptyIdentifier(t *testing.T) {
	models := []string{"", " ", "gfs_global", "gfs_global"}
	want := append([]string(nil), models...)

	assert.Equal(t, " ", firstModel(models))
	assert.Equal(t, want, models)
}

func TestParseSingleExplicitModelUnsuffixedForecast(t *testing.T) {
	data, err := os.ReadFile("testdata/forecast_hourly.json")
	require.NoError(t, err)
	weather, err := parseWeatherResponse(data, []string{"", "ecmwf_ifs", "ecmwf_ifs"})
	require.NoError(t, err)
	assert.Equal(t, "ecmwf_ifs", weather.PrimaryModel)
	assert.Nil(t, weather.HourlyByModel)
	assert.Equal(t, []float64{2.5, 2.3, 2.1}, weather.Hourly.Temperature2m)
}

func TestParseSingleExplicitModelSuffixedForecast(t *testing.T) {
	data := json.RawMessage(`{"timezone":"UTC","hourly":{"time":["2024-01-15T00:00"],"temperature_2m_ecmwf_ifs":[1]},"hourly_units":{"time":"iso8601","temperature_2m_ecmwf_ifs":"primary-unit"}}`)
	weather, err := parseWeatherResponse(data, []string{"ecmwf_ifs"})
	require.NoError(t, err)

	assert.Equal(t, "ecmwf_ifs", weather.PrimaryModel)
	require.Contains(t, weather.HourlyByModel, "ecmwf_ifs")
	assert.Same(t, weather.HourlyByModel["ecmwf_ifs"], weather.Hourly)
	assert.Equal(t, []float64{1}, weather.Hourly.Temperature2m)
	require.NotNil(t, weather.HourlyUnits)
	assert.Equal(t, "primary-unit", weather.HourlyUnits.Temperature2m)
}

func TestParseSecondaryOnlyModelLeavesCompatibilitySectionShared(t *testing.T) {
	data := json.RawMessage(`{"timezone":"Europe/Berlin","hourly":{"time":["2024-01-15T00:00"],"temperature_2m_secondary":[2]}}`)
	weather, err := parseWeatherResponse(data, []string{"primary", "secondary"})
	require.NoError(t, err)
	assert.Equal(t, "primary", weather.PrimaryModel)
	require.Contains(t, weather.HourlyByModel, "secondary")
	assert.Nil(t, weather.Hourly.Temperature2m)
	assert.Len(t, weather.Hourly.Times, 1)
	assert.NotSame(t, weather.HourlyByModel["secondary"], weather.Hourly)
}

func TestParseUnitsSelectPrimaryModel(t *testing.T) {
	data := json.RawMessage(`{
		"timezone":"UTC",
		"hourly_units":{"time":"iso8601","temperature_2m_secondary":"secondary-unit","temperature_2m_primary":"primary-unit"},
		"minutely_15_units":{"time":"iso8601","precipitation_secondary":"secondary-unit","precipitation_primary":"primary-unit"},
		"daily_units":{"time":"iso8601","temperature_2m_max_secondary":"secondary-unit","temperature_2m_max_primary":"primary-unit"}
	}`)
	weather, err := parseWeatherResponse(data, []string{"primary", "secondary"})
	require.NoError(t, err)

	require.NotNil(t, weather.HourlyUnits)
	assert.Equal(t, "primary-unit", weather.HourlyUnits.Temperature2m)
	require.NotNil(t, weather.Minutely15Units)
	assert.Equal(t, "primary-unit", weather.Minutely15Units.Precipitation)
	require.NotNil(t, weather.DailyUnits)
	assert.Equal(t, "primary-unit", weather.DailyUnits.Temperature2mMax)
}

func TestParseNullableAndAbsentUnitSectionsRemainNil(t *testing.T) {
	data := json.RawMessage(`{"hourly_units":null,"daily_units":null}`)
	weather, err := parseWeatherResponse(data, []string{"gfs_global"})
	require.NoError(t, err)

	assert.Nil(t, weather.HourlyUnits)
	assert.Nil(t, weather.Minutely15Units)
	assert.Nil(t, weather.DailyUnits)
}

func TestParseCurrentWeather(t *testing.T) {
	data, err := os.ReadFile("testdata/forecast_current.json")
	require.NoError(t, err)

	weather, err := parseWeatherResponse(data, nil)
	require.NoError(t, err)

	// Check current data
	require.NotNil(t, weather.Current)
	assert.Equal(t, 900, weather.Current.Interval)

	require.NotNil(t, weather.Current.Temperature2m)
	assert.Equal(t, 3.5, *weather.Current.Temperature2m)

	require.NotNil(t, weather.Current.IsDay)
	assert.Equal(t, 1, *weather.Current.IsDay)
	assert.True(t, weather.Current.IsDaytime())

	require.NotNil(t, weather.Current.WeatherCode)
	assert.Equal(t, PartlyCloudy, *weather.Current.WeatherCode)

	// Check time
	loc, _ := time.LoadLocation("Europe/Berlin")
	expectedTime := time.Date(2024, 1, 15, 14, 0, 0, 0, loc)
	assert.Equal(t, expectedTime, weather.Current.Time)

	// Check units
	require.NotNil(t, weather.CurrentUnits)
	assert.Equal(t, "°C", weather.CurrentUnits.Temperature2m)
}

func TestParseHistorical(t *testing.T) {
	data, err := os.ReadFile("testdata/historical.json")
	require.NoError(t, err)

	weather, err := parseWeatherResponse(data, nil)
	require.NoError(t, err)

	require.NotNil(t, weather.Hourly)
	assert.Len(t, weather.Hourly.Times, 3)
	assert.Equal(t, 18.5, weather.Hourly.Temperature2m[0])
	assert.Empty(t, weather.PrimaryModel)
	assert.Nil(t, weather.HourlyByModel)
	assert.Nil(t, weather.Minutely15ByModel)
	assert.Nil(t, weather.DailyByModel)
}

func TestWeatherCodeString(t *testing.T) {
	tests := []struct {
		code     WeatherCode
		expected string
	}{
		{ClearSky, "Clear sky"},
		{PartlyCloudy, "Partly cloudy"},
		{Fog, "Fog"},
		{RainSlight, "Slight rain"},
		{ThunderstormSlight, "Thunderstorm"},
		{WeatherCode(999), "Unknown (999)"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.code.String())
		})
	}
}
