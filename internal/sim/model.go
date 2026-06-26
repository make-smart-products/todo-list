package sim

import (
	"fmt"
	"math"
	"slices"
)

type Station struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	Region            string  `json:"region"`
	PumpCapacity      float64 `json:"pump_capacity"`
	LoadFactor        float64 `json:"load_factor"`
	Health            float64 `json:"health"`
	MaintenanceDebt   float64 `json:"maintenance_debt"`
	GridOnline        bool    `json:"grid_online"`
	ReserveReady      bool    `json:"reserve_ready"`
	ReserveActive     bool    `json:"reserve_active"`
	LightningExposure float64 `json:"lightning_exposure"`
	LastEvent         string  `json:"last_event"`
}

type Segment struct {
	ID          string  `json:"id"`
	FromStation string  `json:"from_station"`
	ToStation   string  `json:"to_station"`
	Capacity    float64 `json:"capacity"`
	Utilization float64 `json:"utilization"`
	StormRisk   float64 `json:"storm_risk"`
}

type WeatherState struct {
	Season             string  `json:"season"`
	StormFront         string  `json:"storm_front"`
	StormSeverity      float64 `json:"storm_severity"`
	GridOutageRisk     float64 `json:"grid_outage_risk"`
	OperationsAdvisory string  `json:"operations_advisory"`
}

type Snapshot struct {
	TickHours        int          `json:"tick_hours"`
	TotalPumped      float64      `json:"total_pumped"`
	HourlyThroughput float64      `json:"hourly_throughput"`
	PlanTarget       float64      `json:"plan_target"`
	Reliability      float64      `json:"reliability"`
	Weather          WeatherState `json:"weather"`
	Stations         []Station    `json:"stations"`
	Segments         []Segment    `json:"segments"`
}

type Simulation struct {
	tickHours   int
	totalPumped float64
	planTarget  float64
	stations    []Station
	segments    []Segment
}

func NewDefaultSimulation() *Simulation {
	simulation := &Simulation{
		planTarget: 3800,
		stations: []Station{
			{
				ID:                "nps-west-1",
				Name:              "West Booster",
				Region:            "west",
				PumpCapacity:      420,
				LoadFactor:        0.78,
				Health:            91,
				MaintenanceDebt:   18,
				GridOnline:        true,
				ReserveReady:      true,
				LightningExposure: 0.55,
			},
			{
				ID:                "nps-central-2",
				Name:              "Central Hub",
				Region:            "central",
				PumpCapacity:      610,
				LoadFactor:        0.82,
				Health:            88,
				MaintenanceDebt:   24,
				GridOnline:        true,
				ReserveReady:      true,
				LightningExposure: 0.72,
			},
			{
				ID:                "nps-east-3",
				Name:              "East Transit",
				Region:            "east",
				PumpCapacity:      500,
				LoadFactor:        0.75,
				Health:            86,
				MaintenanceDebt:   27,
				GridOnline:        true,
				ReserveReady:      false,
				LightningExposure: 0.9,
			},
		},
		segments: []Segment{
			{ID: "seg-west-central", FromStation: "nps-west-1", ToStation: "nps-central-2", Capacity: 380, StormRisk: 0.42},
			{ID: "seg-central-east", FromStation: "nps-central-2", ToStation: "nps-east-3", Capacity: 340, StormRisk: 0.68},
		},
	}

	simulation.updateSegments(simulation.currentWeather())
	return simulation
}

func (s *Simulation) Snapshot() Snapshot {
	stations := cloneStationsForSnapshot(s.stations)
	segments := slices.Clone(s.segments)
	weather := s.currentWeather()

	return Snapshot{
		TickHours:        s.tickHours,
		TotalPumped:      round2(s.totalPumped),
		HourlyThroughput: round2(s.hourlyThroughput(weather)),
		PlanTarget:       s.planTarget,
		Reliability:      round2(s.reliabilityScore()),
		Weather:          weather,
		Stations:         stations,
		Segments:         segments,
	}
}

func (s *Simulation) Advance(hours int) Snapshot {
	if hours < 1 {
		hours = 1
	}

	for range hours {
		s.tickHours++
		weather := s.currentWeather()

		for index := range s.stations {
			station := &s.stations[index]
			station.ReserveActive = false
			station.GridOnline = true
			station.LastEvent = "stable operations"

			debtGrowth := 0.6 * station.LoadFactor * (1 + weather.StormSeverity*0.6)
			if station.Region == weather.StormFront {
				debtGrowth *= 1.4
			}
			station.MaintenanceDebt = math.Min(100, station.MaintenanceDebt+debtGrowth)

			healthDrop := 0.12 * station.LoadFactor
			if station.MaintenanceDebt > 55 {
				healthDrop += 0.25
			}

			outageTriggered := weather.StormSeverity > 0.72 &&
				station.Region == weather.StormFront &&
				station.LightningExposure > 0.65

			if outageTriggered {
				station.GridOnline = false
				if station.ReserveReady {
					station.ReserveActive = true
					station.LastEvent = "grid outage, reserve power started"
					station.MaintenanceDebt = math.Min(100, station.MaintenanceDebt+1.2)
					healthDrop += 0.2
				} else {
					station.LastEvent = "grid outage, station throughput collapsed"
					healthDrop += 0.6
				}
			}

			station.Health = math.Max(35, station.Health-healthDrop)
		}

		s.updateSegments(weather)
		s.totalPumped += s.hourlyThroughput(weather)
	}

	return s.Snapshot()
}

func (s *Simulation) PerformMaintenance(stationID string) (Station, error) {
	for index := range s.stations {
		station := &s.stations[index]
		if station.ID != stationID {
			continue
		}

		station.MaintenanceDebt = math.Max(0, station.MaintenanceDebt-28)
		station.Health = math.Min(100, station.Health+6)
		station.ReserveReady = true
		station.LastEvent = "planned maintenance completed"

		s.updateSegments(s.currentWeather())
		return *station, nil
	}

	return Station{}, fmt.Errorf("station %q not found", stationID)
}

func (s *Simulation) currentWeather() WeatherState {
	stormCycle := s.tickHours % 24

	severity := 0.2
	switch {
	case stormCycle >= 0 && stormCycle < 6:
		severity = 0.22
	case stormCycle >= 6 && stormCycle < 12:
		severity = 0.46
	case stormCycle >= 12 && stormCycle < 16:
		severity = 0.84
	case stormCycle >= 16 && stormCycle < 20:
		severity = 0.78
	default:
		severity = 0.58
	}

	front := "west"
	switch {
	case stormCycle >= 8 && stormCycle < 16:
		front = "central"
	case stormCycle >= 16:
		front = "east"
	}

	advisory := "Normal pumping window."
	if severity > 0.7 {
		advisory = "Thunderstorm season peak: keep reserve power available."
	} else if severity > 0.4 {
		advisory = "Storm cells are building, review maintenance queues."
	}

	return WeatherState{
		Season:             "thunderstorm",
		StormFront:         front,
		StormSeverity:      round2(severity),
		GridOutageRisk:     round2(severity * 0.9),
		OperationsAdvisory: advisory,
	}
}

func (s *Simulation) hourlyThroughput(weather WeatherState) float64 {
	total := 0.0
	for _, station := range s.stations {
		efficiency := station.Health / 100
		if station.MaintenanceDebt > 45 {
			efficiency *= 0.92
		}

		load := station.PumpCapacity * station.LoadFactor
		switch {
		case !station.GridOnline && station.ReserveActive:
			load *= 0.7
		case !station.GridOnline:
			load = 0
		}

		if station.Region == weather.StormFront {
			load *= 1 - weather.StormSeverity*0.08
		}

		total += load * efficiency
	}

	return round2(total / float64(len(s.stations)))
}

func (s *Simulation) reliabilityScore() float64 {
	total := 0.0
	for _, station := range s.stations {
		score := station.Health - station.MaintenanceDebt*0.35
		if !station.GridOnline && !station.ReserveActive {
			score -= 8
		}
		total += score
	}

	return math.Max(0, total/float64(len(s.stations)))
}

func (s *Simulation) updateSegments(weather WeatherState) {
	throughput := s.hourlyThroughput(weather)
	for index := range s.segments {
		segment := &s.segments[index]
		utilization := throughput / segment.Capacity
		if segment.ToStation == "nps-east-3" && weather.StormFront == "east" {
			utilization *= 0.88
		}
		segment.Utilization = round2(math.Min(1.2, utilization))
	}
}

func round2(value float64) float64 {
	return math.Round(value*100) / 100
}

func cloneStationsForSnapshot(stations []Station) []Station {
	cloned := slices.Clone(stations)
	for index := range cloned {
		cloned[index].Health = round2(cloned[index].Health)
		cloned[index].MaintenanceDebt = round2(cloned[index].MaintenanceDebt)
		cloned[index].PumpCapacity = round2(cloned[index].PumpCapacity)
		cloned[index].LoadFactor = round2(cloned[index].LoadFactor)
		cloned[index].LightningExposure = round2(cloned[index].LightningExposure)
	}

	return cloned
}
