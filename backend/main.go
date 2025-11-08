package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"time"
)

type EstimateRequest struct {
	VCpu              float64 `json:"vcpu"`
	MemoryGB          float64 `json:"memory_gb"`
	Concurrency       int     `json:"concurrency"`
	AvgDurationMs     int     `json:"avg_duration_ms"`
	ReqPerMinute      int     `json:"requests_per_min"`
	Region            string  `json:"region"`
	MinInstances      int     `json:"min_instances"`
	MaxInstances      int     `json:"max_instances"`
	IdleUtilizationPc float64 `json:"idle_utilization_pc"`
}

type MonthlyForecast struct {
	Requests   int     `json:"requests"`
	CostUSD    float64 `json:"cost_usd"`
	CO2kg      float64 `json:"co2_kg"`
	EnergyKWh  float64 `json:"energy_kwh"`
	Assumption string  `json:"assumption"`
}

type EstimateResponse struct {
	Per1kRequests struct {
		EnergyKWh float64 `json:"energy_kwh"`
		CO2g      float64 `json:"co2_g"`
		CostUSD   float64 `json:"cost_usd"`
	} `json:"per_1k_requests"`
	PerHour struct {
		EnergyKWh float64 `json:"energy_kwh"`
		CO2g      float64 `json:"co2_g"`
		CostUSD   float64 `json:"cost_usd"`
	} `json:"per_hour"`
	RiskScore       int             `json:"risk_score"`       // 0..100
	SuggestedYAML   string          `json:"suggested_yaml"`   // optimized Cloud Run snippet
	MonthlyForecast MonthlyForecast `json:"monthly_forecast"` // cost + co2
	Assumptions     map[string]any  `json:"assumptions"`
	Advice          []string        `json:"advice"`
}

var gridIntensity = map[string]float64{
	"us-central1":     400,
	"us-east1":        420,
	"us-west1":        300,
	"europe-west1":    230,
	"europe-west4":    180,
	"asia-south1":     700,
	"asia-southeast1": 500,
}

const (
	costPerVCPUSecond   = 0.000024
	costPerGBSecond     = 0.0000025
	requestsPerChunk    = 1000.0
	defaultIdleCPUUsage = 0.10
	wattsPerVCPUAtFull  = 12.0
	wattsPerGBMem       = 0.35
	daysPerMonth        = 30.0
)

func clamp(x, lo, hi float64) float64 { return math.Max(lo, math.Min(hi, x)) }
func regionIntensity(r string) float64 {
	if v, ok := gridIntensity[r]; ok {
		return v
	}
	return 450
}
func estimatePowerWatts(vcpu, memGB, cpuUtil float64) float64 {
	return wattsPerVCPUAtFull*vcpu*clamp(cpuUtil, 0, 1) + wattsPerGBMem*memGB
}
func kwhFromWattsOverSeconds(watts, seconds float64) float64 {
	return (watts * seconds) / 1000.0 / 3600.0
}
func costFromRuntimeSeconds(vcpu, memGB, seconds float64) float64 {
	return vcpu*seconds*costPerVCPUSecond + memGB*seconds*costPerGBSecond
}
func round(v float64, places int) float64 {
	p := math.Pow(10, float64(places))
	return math.Round(v*p) / p
}

func estimateHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w, r)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "use POST", http.StatusMethodNotAllowed)
		return
	}

	var req EstimateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
		return
	}
	normalize(&req)

	// traffic & sizing
	rps := float64(req.ReqPerMinute) / 60.0
	activeInstances := math.Ceil((rps * (float64(req.AvgDurationMs) / 1000.0)) / float64(req.Concurrency))
	if activeInstances < float64(req.MinInstances) {
		activeInstances = float64(req.MinInstances)
	}
	if activeInstances > float64(req.MaxInstances) {
		activeInstances = float64(req.MaxInstances)
	}

	util := 0.0
	if activeInstances > 0 {
		util = (rps * (float64(req.AvgDurationMs) / 1000.0)) / (activeInstances * float64(req.Concurrency))
	}
	util = clamp(util, 0, 1)

	secondsPerRequest := float64(req.AvgDurationMs) / 1000.0
	secPer1k := secondsPerRequest * requestsPerChunk

	// Per-1k
	wattsActive := estimatePowerWatts(req.VCpu, req.MemoryGB, util)
	kwhPer1k := kwhFromWattsOverSeconds(wattsActive, secPer1k)
	costPer1k := costFromRuntimeSeconds(req.VCpu, req.MemoryGB, secPer1k)

	// Hourly incl. min instances
	idleUtil := clamp(req.IdleUtilizationPc/100.0, 0, 1)
	idleInstances := math.Max(float64(req.MinInstances), 0)
	wattsIdleOne := estimatePowerWatts(req.VCpu, req.MemoryGB, idleUtil)
	wattsActiveFleet := wattsActive * activeInstances
	extraIdle := idleInstances - activeInstances
	if extraIdle < 0 {
		extraIdle = 0
	}
	totalWattsNow := wattsActiveFleet + (wattsIdleOne * extraIdle)
	kwhPerHour := kwhFromWattsOverSeconds(totalWattsNow, 3600)

	// Hourly cost (simplified active runtime basis)
	secPerHourActive := 3600.0 * util
	costPerHour := costFromRuntimeSeconds(req.VCpu*activeInstances, req.MemoryGB*activeInstances, secPerHourActive)

	co2Intensity := regionIntensity(req.Region)
	co2Per1k := kwhPer1k * co2Intensity
	co2PerHour := kwhPerHour * co2Intensity

	// Forecast month
	monthlyReq := int(float64(req.ReqPerMinute) * 60 * 24 * daysPerMonth)
	monthlyCost := (float64(monthlyReq) / 1000.0) * costPer1k
	monthlyEnergy := (float64(monthlyReq) / 1000.0) * kwhPer1k
	monthlyCO2kg := (monthlyEnergy * co2Intensity) / 1000.0 // g -> kg

	// Risk score + YAML suggestion
	score := riskScore(req, util, co2Intensity, costPer1k)
	yaml := suggestYAML(req, util, co2Intensity, costPer1k)

	var resp EstimateResponse
	resp.Assumptions = map[string]any{
		"region":                   req.Region,
		"grid_intensity_g_per_kwh": co2Intensity,
		"active_instances":         activeInstances,
		"idle_instances":           idleInstances,
		"cpu_utilization_est":      round(util*100, 1),
		"seconds_per_request":      secondsPerRequest,
	}
	resp.Per1kRequests.EnergyKWh = round(kwhPer1k, 5)
	resp.Per1kRequests.CO2g = round(co2Per1k, 2)
	resp.Per1kRequests.CostUSD = round(costPer1k, 4)
	resp.PerHour.EnergyKWh = round(kwhPerHour, 5)
	resp.PerHour.CO2g = round(co2PerHour, 2)
	resp.PerHour.CostUSD = round(costPerHour, 4)
	resp.RiskScore = score
	resp.SuggestedYAML = yaml
	resp.MonthlyForecast = MonthlyForecast{
		Requests:   monthlyReq,
		CostUSD:    round(monthlyCost, 2),
		CO2kg:      round(monthlyCO2kg, 2),
		EnergyKWh:  round(monthlyEnergy, 2),
		Assumption: fmt.Sprintf("%d days @ %d req/min steady", int(daysPerMonth), req.ReqPerMinute),
	}
	resp.Advice = makeAdvice(req, util, co2Intensity, &resp)

	writeJSON(w, resp)
}

func normalize(req *EstimateRequest) {
	if req.IdleUtilizationPc <= 0 {
		req.IdleUtilizationPc = defaultIdleCPUUsage * 100
	}
	if req.Concurrency <= 0 {
		req.Concurrency = 80
	}
	if req.VCpu <= 0 {
		req.VCpu = 1
	}
	if req.MemoryGB <= 0 {
		req.MemoryGB = 1
	}
	if req.AvgDurationMs <= 0 {
		req.AvgDurationMs = 200
	}
	if req.MinInstances < 0 {
		req.MinInstances = 0
	}
	if req.MaxInstances <= 0 {
		req.MaxInstances = 1
	}
}

func riskScore(req EstimateRequest, util float64, intensity float64, costPer1k float64) int {
	score := 100
	// Penalize high carbon grid
	if intensity > 500 {
		score -= 20
	} else if intensity > 350 {
		score -= 10
	}
	// Penalize low concurrency
	if req.Concurrency < 40 {
		score -= 15
	}
	// Penalize high memory with low util
	if req.MemoryGB > 1.5 && util < 0.35 {
		score -= 15
	}
	// Penalize min instances at low traffic
	if req.MinInstances >= 1 && req.ReqPerMinute < 60 {
		score -= 10
	}
	// Penalize high cost per 1k
	if costPer1k > 0.015 {
		score -= 15
	} else if costPer1k > 0.01 {
		score -= 8
	}
	// Bound 0..100
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	return score
}

func suggestYAML(req EstimateRequest, util float64, intensity float64, costPer1k float64) string {
	// Mutate a copy with gentle optimizations
	c := req
	if c.Concurrency < 80 {
		c.Concurrency = 80
	}
	if c.MemoryGB > 1.5 && util < 0.35 {
		c.MemoryGB = 1.5
	}
	if intensity > 400 && c.Region != "europe-west4" && c.Region != "us-west1" {
		c.Region = "europe-west4"
	}
	if c.MinInstances >= 1 && c.ReqPerMinute < 60 {
		c.MinInstances = 0
	}
	return fmt.Sprintf(
		`# Suggested Cloud Run config (opt.)
region: %s
cpu: %.0f
memory: %.1fGi
minInstances: %d
maxInstances: %d
concurrency: %d
ingress: all
`, c.Region, c.VCpu, c.MemoryGB, c.MinInstances, c.MaxInstances, c.Concurrency)
}

func makeAdvice(req EstimateRequest, util float64, intensity float64, resp *EstimateResponse) []string {
	adv := []string{}
	if req.Concurrency < 40 {
		adv = append(adv, fmt.Sprintf("Increase concurrency to ~80 (current %d) to reduce instance count/idle overhead.", req.Concurrency))
	}
	if req.MemoryGB > 1.0 && util < 0.35 {
		adv = append(adv, fmt.Sprintf("Right-size memory: %.1fGi with ~%.0f%% CPU util → try 1–1.5Gi.", req.MemoryGB, util*100))
	}
	if req.MinInstances >= 1 && req.ReqPerMinute < 60 {
		adv = append(adv, "Traffic is low but min_instances>=1; use min_instances=0 with startup probe.")
	}
	if intensity > 400 {
		adv = append(adv, "High CO₂ grid region; prefer europe-west4 or us-west1 for lower emissions.")
	}
	if req.AvgDurationMs > 300 {
		adv = append(adv, "High latency; cache hot data, reuse connections, reduce cold I/O.")
	}
	if resp.Per1kRequests.CostUSD > 0.01 {
		adv = append(adv, fmt.Sprintf("Cost/1k $%.4f; baseline CPU=1, MEM=1Gi, concurrency=80, then re-measure.", resp.Per1kRequests.CostUSD))
	}
	if len(adv) == 0 {
		adv = append(adv, "Configuration looks healthy. Keep automated checks per deploy.")
	}
	return adv
}

func enableCORS(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin == "" {
		origin = "*"
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Vary", "Origin")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
func rootHandler(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("DevOps Intelligence Hub API"))
}
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", rootHandler)
	mux.HandleFunc("/healthz", healthHandler)
	mux.HandleFunc("/estimate", estimateHandler)
	addr := ":" + port()
	s := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	log.Printf("listening on %s", addr)
	log.Fatal(s.ListenAndServe())
}
func port() string {
	if p := os.Getenv("PORT"); p != "" {
		return p
	}
	return "8080"
}
