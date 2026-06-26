package sim

import (
	"fmt"
	"math"
	"slices"
)

type Station struct {
	ID                  string  `json:"id"`
	Name                string  `json:"name"`
	Region              string  `json:"region"`
	Role                string  `json:"role"`
	PumpCapacity        float64 `json:"pump_capacity"`
	LoadFactor          float64 `json:"load_factor"`
	MinLoadFactor       float64 `json:"min_load_factor"`
	MaxLoadFactor       float64 `json:"max_load_factor"`
	Health              float64 `json:"health"`
	MaintenanceDebt     float64 `json:"maintenance_debt"`
	GridOnline          bool    `json:"grid_online"`
	ReserveReady        bool    `json:"reserve_ready"`
	ReserveActive       bool    `json:"reserve_active"`
	LightningExposure   float64 `json:"lightning_exposure"`
	EffectiveThroughput float64 `json:"effective_throughput"`
	Status              string  `json:"status"`
	LastEvent           string  `json:"last_event"`
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

type Event struct {
	TickHours int    `json:"tick_hours"`
	Severity  string `json:"severity"`
	Message   string `json:"message"`
}

type WeatherState struct {
	Season             string  `json:"season"`
	StormFront         string  `json:"storm_front"`
	StormSeverity      float64 `json:"storm_severity"`
	GridOutageRisk     float64 `json:"grid_outage_risk"`
	OperationsAdvisory string  `json:"operations_advisory"`
}

type Snapshot struct {
	ScenarioName      string       `json:"scenario_name"`
	TickHours         int          `json:"tick_hours"`
	TotalPumped       float64      `json:"total_pumped"`
	DeliveredThisHour float64      `json:"delivered_this_hour"`
	HourlyTarget      float64      `json:"hourly_target"`
	PlanTarget        float64      `json:"plan_target"`
	PlanProgress      float64      `json:"plan_progress"`
	Reliability       float64      `json:"reliability"`
	ServiceQuality    float64      `json:"service_quality"`
	Score             float64      `json:"score"`
	BypassShare       float64      `json:"bypass_share"`
	Weather           WeatherState `json:"weather"`
	Stations          []Station    `json:"stations"`
	Segments          []Segment    `json:"segments"`
	Alerts            []string     `json:"alerts"`
	EventLog          []Event      `json:"event_log"`
}

type Simulation struct {
	scenarioName      string
	tickHours         int
	totalPumped       float64
	deliveredThisHour float64
	planTarget        float64
	hourlyTarget      float64
	bypassShare       float64
	reliability       float64
	serviceQuality    float64
	score             float64
	stations          []Station
	segments          []Segment
	alerts            []string
	eventLog          []Event
}

func NewDefaultSimulation() *Simulation {
	simulation := &Simulation{
		scenarioName: "Thunderstorm Dispatch Week",
		planTarget:   6200,
		hourlyTarget: 255,
		bypassShare:  0.22,
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
		Weather:           s.currentWeather(),
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
