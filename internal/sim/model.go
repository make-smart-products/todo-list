package sim

import (
	"fmt"
	"math"
	"slices"
)

type Station struct {
	ID                  string      `json:"id"`
	Name                string      `json:"name"`
	Region              string      `json:"region"`
	Role                string      `json:"role"`
	PumpCapacity        float64     `json:"pump_capacity"`
	LoadFactor          float64     `json:"load_factor"`
	MinLoadFactor       float64     `json:"min_load_factor"`
	MaxLoadFactor       float64     `json:"max_load_factor"`
	Health              float64     `json:"health"`
	MaintenanceDebt     float64     `json:"maintenance_debt"`
	GridOnline          bool        `json:"grid_online"`
	ReserveReady        bool        `json:"reserve_ready"`
	ReserveActive       bool        `json:"reserve_active"`
	LightningExposure   float64     `json:"lightning_exposure"`
	EffectiveThroughput float64     `json:"effective_throughput"`
	Status              string      `json:"status"`
	LastEvent           string      `json:"last_event"`
	Equipment           []Equipment `json:"equipment"`
}

type Segment struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	FromStation string  `json:"from_station"`
	ToStation   string  `json:"to_station"`
	Capacity    float64 `json:"capacity"`
	Flow        float64 `json:"flow"`
	Utilization float64 `json:"utilization"`
	StormRisk   float64 `json:"storm_risk"`
	Status      string  `json:"status"`
}

type Equipment struct {
	ID                  string  `json:"id"`
	Name                string  `json:"name"`
	Kind                string  `json:"kind"`
	Health              float64 `json:"health"`
	MaintenanceDebt     float64 `json:"maintenance_debt"`
	LoadShare           float64 `json:"load_share"`
	PowerDraw           float64 `json:"power_draw"`
	Status              string  `json:"status"`
	LastServiceHoursAgo int     `json:"last_service_hours_ago"`
}

type Event struct {
	TickHours int    `json:"tick_hours"`
	Severity  string `json:"severity"`
	Message   string `json:"message"`
}

type ForecastEntry struct {
	OffsetHours    int     `json:"offset_hours"`
	StormFront     string  `json:"storm_front"`
	StormSeverity  float64 `json:"storm_severity"`
	GridOutageRisk float64 `json:"grid_outage_risk"`
	WindowLabel    string  `json:"window_label"`
}

type Economy struct {
	Revenue             float64 `json:"revenue"`
	PowerCost           float64 `json:"power_cost"`
	MaintenanceCost     float64 `json:"maintenance_cost"`
	StormLossCost       float64 `json:"storm_loss_cost"`
	PenaltyCost         float64 `json:"penalty_cost"`
	NetBalance          float64 `json:"net_balance"`
	DispatchBudget      float64 `json:"dispatch_budget"`
	PredictedEndBalance float64 `json:"predicted_end_balance"`
}

type Objective struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Current     float64 `json:"current"`
	Target      float64 `json:"target"`
	Unit        string  `json:"unit"`
	Status      string  `json:"status"`
}

type Scenario struct {
	Name               string      `json:"name"`
	Briefing           string      `json:"briefing"`
	ShiftDurationHours int         `json:"shift_duration_hours"`
	HoursRemaining     int         `json:"hours_remaining"`
	Objectives         []Objective `json:"objectives"`
}

type WeatherState struct {
	Season             string  `json:"season"`
	StormFront         string  `json:"storm_front"`
	StormSeverity      float64 `json:"storm_severity"`
	GridOutageRisk     float64 `json:"grid_outage_risk"`
	OperationsAdvisory string  `json:"operations_advisory"`
}

type Snapshot struct {
	Scenario          Scenario        `json:"scenario"`
	ScenarioName      string          `json:"scenario_name"`
	TickHours         int             `json:"tick_hours"`
	TotalPumped       float64         `json:"total_pumped"`
	DeliveredThisHour float64         `json:"delivered_this_hour"`
	HourlyTarget      float64         `json:"hourly_target"`
	PlanTarget        float64         `json:"plan_target"`
	PlanProgress      float64         `json:"plan_progress"`
	Reliability       float64         `json:"reliability"`
	ServiceQuality    float64         `json:"service_quality"`
	Score             float64         `json:"score"`
	BypassShare       float64         `json:"bypass_share"`
	Economy           Economy         `json:"economy"`
	Weather           WeatherState    `json:"weather"`
	Forecast          []ForecastEntry `json:"forecast"`
	Stations          []Station       `json:"stations"`
	Segments          []Segment       `json:"segments"`
	Alerts            []string        `json:"alerts"`
	EventLog          []Event         `json:"event_log"`
}

type Simulation struct {
	scenarioName       string
	scenarioBriefing   string
	shiftDurationHours int
	tickHours          int
	totalPumped        float64
	deliveredThisHour  float64
	planTarget         float64
	hourlyTarget       float64
	bypassShare        float64
	reliability        float64
	serviceQuality     float64
	score              float64
	revenue            float64
	powerCost          float64
	maintenanceCost    float64
	stormLossCost      float64
	penaltyCost        float64
	dispatchBudget     float64
	gridOutageHours    int
	targetMissHours    int
	stations           []Station
	segments           []Segment
	alerts             []string
	eventLog           []Event
}

func NewDefaultSimulation() *Simulation {
	simulation := &Simulation{
		scenarioName:       "Thunderstorm Dispatch Week",
		scenarioBriefing:   "Balance national pumping volume against equipment health, reserve readiness, and storm-season outages.",
		shiftDurationHours: 24,
		planTarget:         6200,
		hourlyTarget:       255,
		bypassShare:        0.22,
		dispatchBudget:     210000,
		stations: []Station{
			{
				ID:                "nps-north-intake",
				Name:              "North Intake",
				Region:            "north",
				Role:              "source",
				PumpCapacity:      320,
				LoadFactor:        0.82,
				MinLoadFactor:     0.55,
				MaxLoadFactor:     1.05,
				Health:            93,
				MaintenanceDebt:   12,
				GridOnline:        true,
				ReserveReady:      true,
				LightningExposure: 0.45,
				Status:            "stable",
				Equipment: []Equipment{
					{ID: "north-pump-a", Name: "Pump Train A", Kind: "process", Health: 94, MaintenanceDebt: 10, LoadShare: 0.52, PowerDraw: 41, Status: "stable", LastServiceHoursAgo: 18},
					{ID: "north-switchgear", Name: "110kV Switchgear", Kind: "electrical", Health: 92, MaintenanceDebt: 13, LoadShare: 0.28, PowerDraw: 18, Status: "stable", LastServiceHoursAgo: 22},
					{ID: "north-generator", Name: "Reserve Generator", Kind: "reserve", Health: 93, MaintenanceDebt: 12, LoadShare: 0.20, PowerDraw: 9, Status: "ready", LastServiceHoursAgo: 12},
				},
			},
			{
				ID:                "nps-west-booster",
				Name:              "West Booster",
				Region:            "west",
				Role:              "source",
				PumpCapacity:      430,
				LoadFactor:        0.78,
				MinLoadFactor:     0.55,
				MaxLoadFactor:     1.05,
				Health:            90,
				MaintenanceDebt:   20,
				GridOnline:        true,
				ReserveReady:      true,
				LightningExposure: 0.58,
				Status:            "stable",
				Equipment: []Equipment{
					{ID: "west-pump-a", Name: "Pump Train B", Kind: "process", Health: 91, MaintenanceDebt: 18, LoadShare: 0.55, PowerDraw: 56, Status: "stable", LastServiceHoursAgo: 26},
					{ID: "west-rectifier", Name: "Power Rectifier", Kind: "electrical", Health: 89, MaintenanceDebt: 22, LoadShare: 0.25, PowerDraw: 24, Status: "stable", LastServiceHoursAgo: 34},
					{ID: "west-generator", Name: "Diesel Reserve", Kind: "reserve", Health: 90, MaintenanceDebt: 20, LoadShare: 0.20, PowerDraw: 12, Status: "ready", LastServiceHoursAgo: 20},
				},
			},
			{
				ID:                "nps-central-hub",
				Name:              "Central Hub",
				Region:            "central",
				Role:              "hub",
				PumpCapacity:      620,
				LoadFactor:        0.84,
				MinLoadFactor:     0.60,
				MaxLoadFactor:     1.05,
				Health:            89,
				MaintenanceDebt:   26,
				GridOnline:        true,
				ReserveReady:      true,
				LightningExposure: 0.70,
				Status:            "stable",
				Equipment: []Equipment{
					{ID: "central-pump-a", Name: "Main Pump Train", Kind: "process", Health: 90, MaintenanceDebt: 24, LoadShare: 0.48, PowerDraw: 68, Status: "stable", LastServiceHoursAgo: 28},
					{ID: "central-transformer", Name: "Traction Transformer", Kind: "electrical", Health: 88, MaintenanceDebt: 28, LoadShare: 0.30, PowerDraw: 31, Status: "stable", LastServiceHoursAgo: 42},
					{ID: "central-generator", Name: "Reserve Turbine", Kind: "reserve", Health: 89, MaintenanceDebt: 26, LoadShare: 0.22, PowerDraw: 16, Status: "ready", LastServiceHoursAgo: 24},
				},
			},
			{
				ID:                "nps-south-loop",
				Name:              "South Loop",
				Region:            "south",
				Role:              "bypass",
				PumpCapacity:      280,
				LoadFactor:        0.70,
				MinLoadFactor:     0.55,
				MaxLoadFactor:     0.98,
				Health:            87,
				MaintenanceDebt:   22,
				GridOnline:        true,
				ReserveReady:      false,
				LightningExposure: 0.82,
				Status:            "stable",
				Equipment: []Equipment{
					{ID: "south-pump-a", Name: "Bypass Pump Train", Kind: "process", Health: 88, MaintenanceDebt: 20, LoadShare: 0.50, PowerDraw: 34, Status: "stable", LastServiceHoursAgo: 30},
					{ID: "south-substation", Name: "Substation Relay Group", Kind: "electrical", Health: 85, MaintenanceDebt: 25, LoadShare: 0.28, PowerDraw: 15, Status: "stable", LastServiceHoursAgo: 44},
					{ID: "south-generator", Name: "Portable Generator", Kind: "reserve", Health: 84, MaintenanceDebt: 27, LoadShare: 0.22, PowerDraw: 8, Status: "standby", LastServiceHoursAgo: 54},
				},
			},
			{
				ID:                "nps-east-export",
				Name:              "East Export",
				Region:            "east",
				Role:              "terminal",
				PumpCapacity:      540,
				LoadFactor:        0.88,
				MinLoadFactor:     0.60,
				MaxLoadFactor:     1.02,
				Health:            92,
				MaintenanceDebt:   18,
				GridOnline:        true,
				ReserveReady:      true,
				LightningExposure: 0.65,
				Status:            "stable",
				Equipment: []Equipment{
					{ID: "east-pump-a", Name: "Export Pump Train", Kind: "process", Health: 93, MaintenanceDebt: 16, LoadShare: 0.50, PowerDraw: 52, Status: "stable", LastServiceHoursAgo: 16},
					{ID: "east-switchyard", Name: "Switchyard Protection", Kind: "electrical", Health: 91, MaintenanceDebt: 18, LoadShare: 0.25, PowerDraw: 21, Status: "stable", LastServiceHoursAgo: 18},
					{ID: "east-generator", Name: "Gas Turbine Reserve", Kind: "reserve", Health: 92, MaintenanceDebt: 17, LoadShare: 0.25, PowerDraw: 14, Status: "ready", LastServiceHoursAgo: 15},
				},
			},
		},
		segments: []Segment{
			{ID: "seg-north-central", Name: "North-Central Mainline", FromStation: "nps-north-intake", ToStation: "nps-central-hub", Capacity: 300, StormRisk: 0.35},
			{ID: "seg-west-central", Name: "West-Central Mainline", FromStation: "nps-west-booster", ToStation: "nps-central-hub", Capacity: 360, StormRisk: 0.42},
			{ID: "seg-central-east", Name: "Central-East Mainline", FromStation: "nps-central-hub", ToStation: "nps-east-export", Capacity: 420, StormRisk: 0.58},
			{ID: "seg-central-south", Name: "Central-South Bypass", FromStation: "nps-central-hub", ToStation: "nps-south-loop", Capacity: 240, StormRisk: 0.64},
			{ID: "seg-south-east", Name: "South-East Export Loop", FromStation: "nps-south-loop", ToStation: "nps-east-export", Capacity: 230, StormRisk: 0.71},
		},
	}

	simulation.pushEvent("info", "Dispatch shift started. Country-wide pumping plan loaded.")
	simulation.refreshDerivedState()
	return simulation
}

func (s *Simulation) Snapshot() Snapshot {
	return Snapshot{
		Scenario:          s.currentScenario(),
		ScenarioName:      s.scenarioName,
		TickHours:         s.tickHours,
		TotalPumped:       round2(s.totalPumped),
		DeliveredThisHour: round2(s.deliveredThisHour),
		HourlyTarget:      round2(s.hourlyTarget),
		PlanTarget:        round2(s.planTarget),
		PlanProgress:      round2(clamp(s.totalPumped/s.planTarget, 0, 2)),
		Reliability:       round2(s.reliability),
		ServiceQuality:    round2(s.serviceQuality),
		Score:             round2(s.score),
		BypassShare:       round2(s.bypassShare),
		Economy:           s.currentEconomy(),
		Weather:           s.currentWeather(),
		Forecast:          s.currentForecast(6),
		Stations:          cloneStationsForSnapshot(s.stations),
		Segments:          cloneSegmentsForSnapshot(s.segments),
		Alerts:            slices.Clone(s.alerts),
		EventLog:          slices.Clone(s.eventLog),
	}
}

func (s *Simulation) pushEvent(severity, message string) {
	s.eventLog = append(s.eventLog, Event{
		TickHours: s.tickHours,
		Severity:  severity,
		Message:   message,
	})
	if len(s.eventLog) > 12 {
		s.eventLog = slices.Clone(s.eventLog[len(s.eventLog)-12:])
	}
}

func round2(value float64) float64 {
	return math.Round(value*100) / 100
}

func clamp(value, minValue, maxValue float64) float64 {
	return math.Max(minValue, math.Min(maxValue, value))
}

func cloneStationsForSnapshot(stations []Station) []Station {
	cloned := slices.Clone(stations)
	for index := range cloned {
		cloned[index].PumpCapacity = round2(cloned[index].PumpCapacity)
		cloned[index].LoadFactor = round2(cloned[index].LoadFactor)
		cloned[index].MinLoadFactor = round2(cloned[index].MinLoadFactor)
		cloned[index].MaxLoadFactor = round2(cloned[index].MaxLoadFactor)
		cloned[index].Health = round2(cloned[index].Health)
		cloned[index].MaintenanceDebt = round2(cloned[index].MaintenanceDebt)
		cloned[index].LightningExposure = round2(cloned[index].LightningExposure)
		cloned[index].EffectiveThroughput = round2(cloned[index].EffectiveThroughput)
		cloned[index].Equipment = cloneEquipment(cloned[index].Equipment)
	}

	return cloned
}

func cloneEquipment(equipment []Equipment) []Equipment {
	cloned := slices.Clone(equipment)
	for index := range cloned {
		cloned[index].Health = round2(cloned[index].Health)
		cloned[index].MaintenanceDebt = round2(cloned[index].MaintenanceDebt)
		cloned[index].LoadShare = round2(cloned[index].LoadShare)
		cloned[index].PowerDraw = round2(cloned[index].PowerDraw)
	}

	return cloned
}

func cloneSegmentsForSnapshot(segments []Segment) []Segment {
	cloned := slices.Clone(segments)
	for index := range cloned {
		cloned[index].Capacity = round2(cloned[index].Capacity)
		cloned[index].Flow = round2(cloned[index].Flow)
		cloned[index].Utilization = round2(cloned[index].Utilization)
		cloned[index].StormRisk = round2(cloned[index].StormRisk)
	}

	return cloned
}

func (s *Simulation) stationByID(stationID string) (*Station, error) {
	for index := range s.stations {
		if s.stations[index].ID == stationID {
			return &s.stations[index], nil
		}
	}

	return nil, fmt.Errorf("station %q not found", stationID)
}

func (s *Simulation) stationByIDMust(stationID string) *Station {
	station, err := s.stationByID(stationID)
	if err != nil {
		panic(err)
	}

	return station
}

func (s *Simulation) segmentByIDMust(segmentID string) *Segment {
	for index := range s.segments {
		if s.segments[index].ID == segmentID {
			return &s.segments[index]
		}
	}

	panic(fmt.Sprintf("segment %q not found", segmentID))
}

func (s *Simulation) currentForecast(hoursAhead int) []ForecastEntry {
	forecast := make([]ForecastEntry, 0, hoursAhead)
	for offset := 1; offset <= hoursAhead; offset++ {
		weather := s.weatherAtHour(s.tickHours + offset)
		forecast = append(forecast, ForecastEntry{
			OffsetHours:    offset,
			StormFront:     weather.StormFront,
			StormSeverity:  round2(weather.StormSeverity),
			GridOutageRisk: round2(weather.GridOutageRisk),
			WindowLabel:    fmt.Sprintf("+%dh", offset),
		})
	}

	return forecast
}

func (s *Simulation) currentEconomy() Economy {
	hoursRemaining := float64(maxInt(s.shiftDurationHours-s.tickHours, 0))
	projectedRevenue := s.revenue + hoursRemaining*s.deliveredThisHour*46
	projectedPowerCost := s.powerCost + hoursRemaining*currentGridLoadPowerCost(s.stations)*0.35
	projectedPenalty := s.penaltyCost + float64(s.targetMissHours)*1800

	return Economy{
		Revenue:             round2(s.revenue),
		PowerCost:           round2(s.powerCost),
		MaintenanceCost:     round2(s.maintenanceCost),
		StormLossCost:       round2(s.stormLossCost),
		PenaltyCost:         round2(s.penaltyCost),
		NetBalance:          round2(s.dispatchBudget + s.revenue - s.powerCost - s.maintenanceCost - s.stormLossCost - s.penaltyCost),
		DispatchBudget:      round2(s.dispatchBudget),
		PredictedEndBalance: round2(s.dispatchBudget + projectedRevenue - projectedPowerCost - s.maintenanceCost - s.stormLossCost - projectedPenalty),
	}
}

func (s *Simulation) currentScenario() Scenario {
	return Scenario{
		Name:               s.scenarioName,
		Briefing:           s.scenarioBriefing,
		ShiftDurationHours: s.shiftDurationHours,
		HoursRemaining:     maxInt(s.shiftDurationHours-s.tickHours, 0),
		Objectives: []Objective{
			{
				ID:          "objective-throughput",
				Title:       "Pump national plan",
				Description: "Reach the dispatch target before the shift ends.",
				Current:     round2(s.totalPumped),
				Target:      round2(s.planTarget),
				Unit:        "tons",
				Status:      objectiveStatus(s.totalPumped/s.planTarget, 1.0),
			},
			{
				ID:          "objective-reliability",
				Title:       "Keep reliability high",
				Description: "Hold reliability at or above the operating threshold.",
				Current:     round2(s.reliability),
				Target:      78,
				Unit:        "score",
				Status:      thresholdStatus(s.reliability, 78),
			},
			{
				ID:          "objective-outages",
				Title:       "Minimize outage impact",
				Description: "Limit total grid-outage hours across stations.",
				Current:     float64(s.gridOutageHours),
				Target:      4,
				Unit:        "hours",
				Status:      inverseThresholdStatus(float64(s.gridOutageHours), 4),
			},
		},
	}
}

func objectiveStatus(progress, target float64) string {
	switch {
	case progress >= target:
		return "completed"
	case progress >= target*0.75:
		return "on-track"
	default:
		return "at-risk"
	}
}

func thresholdStatus(current, target float64) string {
	if current >= target {
		return "completed"
	}
	if current >= target*0.92 {
		return "on-track"
	}

	return "at-risk"
}

func inverseThresholdStatus(current, target float64) string {
	if current <= target {
		return "completed"
	}
	if current <= target*1.5 {
		return "on-track"
	}

	return "at-risk"
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}

	return b
}
