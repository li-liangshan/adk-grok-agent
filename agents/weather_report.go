package agents

import (
	"adk-weatherreport/main/models"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
	"google.golang.org/genai"
)

type getWeatherReportArgs struct {
	City string `json:"city" jsonschema:"The city for which to get the weather report."`
}

type getWeatherReportResult struct {
	Status string `json:"status"`
	Report string `json:"report,omitempty"`
}

type getWeatherForecastArgs struct {
	City string `json:"city" jsonschema:"The city for which to get the weather forecast."`
	Days int    `json:"days,omitempty" jsonschema:"Number of forecast days, from 1 to 5. Defaults to 3."`
}

type getWeatherForecastResult struct {
	Status   string `json:"status"`
	Forecast string `json:"forecast,omitempty"`
}

// WeatherResponse struct maps the JSON response from OpenWeather API
// Contains city name and main temperature details
type WeatherResponse struct {
	CityName string `json:"name"`
	Weather  []struct {
		Main        string `json:"main"`
		Description string `json:"description"`
	} `json:"weather"`
	Main struct {
		Temp float64 `json:"temp"`
	} `json:"main"`
}

// ForecastItem maps one 3-hour entry from the OpenWeather forecast API.
type ForecastItem struct {
	DtTxt   string `json:"dt_txt"`
	Weather []struct {
		Description string `json:"description"`
	} `json:"weather"`
	Main struct {
		TempMin float64 `json:"temp_min"`
		TempMax float64 `json:"temp_max"`
	} `json:"main"`
}

type ForecastResponse struct {
	City struct {
		Name string `json:"name"`
	} `json:"city"`
	List []ForecastItem `json:"list"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

const weatherAgentInstruction = "You are a helpful assistant that provides weather information and analyzes the sentiment of user feedback. " +
	"**If the user asks for the current weather in a specific city, use the 'get_weather_report' tool to retrieve the weather details.** " +
	"**If the user asks for a weather forecast in a specific city, use the 'get_weather_forecast' tool; its 'days' argument supports 1 to 5 days (default 3).** " +
	"**If a tool returns a 'success' status, provide the result to the user.** " +
	"**If a tool returns an 'error' status, inform the user that the weather information for the specified city is not available and ask if they have another city in mind.** "

func newWeatherTools() ([]tool.Tool, error) {
	weatherTool, err := functiontool.New(
		functiontool.Config{
			Name:        "get_weather_report",
			Description: "Retrieves the current weather report for a specified city.",
		},
		getWeatherReport,
	)
	if err != nil {
		return nil, err
	}

	forecastTool, err := functiontool.New(
		functiontool.Config{
			Name:        "get_weather_forecast",
			Description: "Retrieves a daily weather forecast for a specified city, for up to 5 days.",
		},
		getWeatherForecast,
	)
	if err != nil {
		return nil, err
	}

	return []tool.Tool{weatherTool, forecastTool}, nil
}

func NewWeatherReportAgent(ctx context.Context) (agent.Agent, error) {
	model, err := models.NewDeepSeekModel(ctx, "deepseek-flash", &genai.ClientConfig{
		APIKey: os.Getenv("DEEPSEEK_API_KEY"),
	})
	if err != nil {
		log.Fatalf("Failed to create model: %v", err)
		return nil, err
	}

	tools, err := newWeatherTools()
	if err != nil {
		return nil, err
	}

	weatherReportAgent, err := llmagent.New(llmagent.Config{
		Name:        "weather_report_agent",
		Model:       model,
		Instruction: weatherAgentInstruction,
		Tools:       tools,
	})
	if err != nil {
		log.Fatal(err)
	}

	return weatherReportAgent, nil
}

// fetchOWM calls the OpenWeatherMap API and returns the response body.
func fetchOWM(apiURL string) ([]byte, error) {
	res, err := (&http.Client{Timeout: 10 * time.Second}).Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		var errorResponse ErrorResponse
		if err := json.Unmarshal(body, &errorResponse); err == nil && errorResponse.Message != "" {
			return nil, fmt.Errorf("weather API error: %s", errorResponse.Message)
		}
		return nil, fmt.Errorf("weather API returned status %d", res.StatusCode)
	}

	return body, nil
}

func getWeatherReport(ctx tool.Context, args getWeatherReportArgs) (getWeatherReportResult, error) {
	apiKey := os.Getenv("OWM_API_KEY")
	if apiKey == "" {
		return getWeatherReportResult{Status: "error", Report: "OWM_API_KEY is not configured"}, nil
	}

	apiURL := fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?appid=%s&q=%s&units=metric",
		url.QueryEscape(apiKey), url.QueryEscape(args.City))
	body, err := fetchOWM(apiURL)
	if err != nil {
		return getWeatherReportResult{Status: "error", Report: err.Error()}, nil
	}

	var weatherData WeatherResponse
	if err := json.Unmarshal(body, &weatherData); err != nil {
		return getWeatherReportResult{Status: "error", Report: err.Error()}, nil
	}
	if len(weatherData.Weather) == 0 {
		return getWeatherReportResult{Status: "error", Report: "no weather data returned for the specified city"}, nil
	}

	report := fmt.Sprintf("The weather in %s is %s with a temperature of %.1f degrees Celsius.",
		weatherData.CityName, weatherData.Weather[0].Description, weatherData.Main.Temp)
	return getWeatherReportResult{Status: "success", Report: report}, nil
}

func getWeatherForecast(ctx tool.Context, args getWeatherForecastArgs) (getWeatherForecastResult, error) {
	days := args.Days
	if days <= 0 {
		days = 3
	}
	if days > 5 {
		days = 5
	}

	apiKey := os.Getenv("OWM_API_KEY")
	if apiKey == "" {
		return getWeatherForecastResult{Status: "error", Forecast: "OWM_API_KEY is not configured"}, nil
	}

	// The free 5-day/3-hour forecast endpoint returns 8 entries per day.
	apiURL := fmt.Sprintf("http://api.openweathermap.org/data/2.5/forecast?appid=%s&q=%s&units=metric&cnt=%d",
		url.QueryEscape(apiKey), url.QueryEscape(args.City), days*8)
	body, err := fetchOWM(apiURL)
	if err != nil {
		return getWeatherForecastResult{Status: "error", Forecast: err.Error()}, nil
	}

	var forecastData ForecastResponse
	if err := json.Unmarshal(body, &forecastData); err != nil {
		return getWeatherForecastResult{Status: "error", Forecast: err.Error()}, nil
	}
	if len(forecastData.List) == 0 {
		return getWeatherForecastResult{Status: "error", Forecast: "no forecast data returned for the specified city"}, nil
	}

	daily := aggregateDailyForecast(forecastData.List)
	if len(daily) > days {
		daily = daily[:days]
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Forecast for %s:\n", forecastData.City.Name)
	for _, day := range daily {
		fmt.Fprintf(&b, "%s: %s, %.1f~%.1f degrees Celsius\n",
			day.Date, day.Description, day.Min, day.Max)
	}

	return getWeatherForecastResult{Status: "success", Forecast: strings.TrimRight(b.String(), "\n")}, nil
}

type dailyForecast struct {
	Date        string
	Description string
	Min         float64
	Max         float64
}

// aggregateDailyForecast groups 3-hourly entries into one summary per day,
// using the entry closest to noon for the weather description.
func aggregateDailyForecast(items []ForecastItem) []dailyForecast {
	var daily []dailyForecast
	var noonDistance time.Duration

	for _, item := range items {
		if len(item.DtTxt) < 10 || len(item.Weather) == 0 {
			continue
		}

		date := item.DtTxt[:10]
		if len(daily) == 0 || daily[len(daily)-1].Date != date {
			daily = append(daily, dailyForecast{
				Date:        date,
				Description: item.Weather[0].Description,
				Min:         item.Main.TempMin,
				Max:         item.Main.TempMax,
			})
			noonDistance = distanceFromNoon(item.DtTxt)
			continue
		}

		day := &daily[len(daily)-1]
		if item.Main.TempMin < day.Min {
			day.Min = item.Main.TempMin
		}
		if item.Main.TempMax > day.Max {
			day.Max = item.Main.TempMax
		}
		if distance := distanceFromNoon(item.DtTxt); distance < noonDistance {
			day.Description = item.Weather[0].Description
			noonDistance = distance
		}
	}

	return daily
}

func distanceFromNoon(dtTxt string) time.Duration {
	t, err := time.Parse("2006-01-02 15:04:05", dtTxt)
	if err != nil {
		return 24 * time.Hour
	}
	noon := time.Date(t.Year(), t.Month(), t.Day(), 12, 0, 0, 0, t.Location())
	distance := t.Sub(noon)
	if distance < 0 {
		return -distance
	}
	return distance
}

func NewWeatherReportAgentWithGemini(ctx context.Context) (agent.Agent, error) {
	geminiModel, err := gemini.NewModel(ctx, "gemini-3-pro-preview", &genai.ClientConfig{
		APIKey: os.Getenv("GEMINI_API_KEY"),
	})
	if err != nil {
		log.Fatalf("Failed to create model: %v", err)
		return nil, err
	}

	tools, err := newWeatherTools()
	if err != nil {
		return nil, err
	}

	weatherReportAgent, err := llmagent.New(llmagent.Config{
		Name:        "weather_report_agent",
		Model:       geminiModel,
		Instruction: weatherAgentInstruction,
		Tools:       tools,
	})
	if err != nil {
		log.Fatal(err)
	}

	return weatherReportAgent, nil
}
