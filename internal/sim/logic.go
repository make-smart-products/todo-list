package sim

import (
	"fmt"
	"math"
)

func (s *Simulation) Advance(hours int) Snapshot {
	if hours < 1 {
		hours = 1
	}

	for range hours {
		s.tickHours++
		weather := s.currentWeather()

		for index := range s.stations {
			station := &s.stations[index]
			station.GridOnline = true
			station.ReserveActive = false
			station.Status = "stable"
			station.LastEvent = "steady pumping window"

			debtGrowth := 0.45 + station.LoadFactor*0.5
			if station.Role == "hub" || station.Role == "terminal" {
				debtGrowth += 0.1
			}
			if station.Region == weather.StormFront {
				debtGrowth *= 1 + weather.StormSeverity*0.9
			}
			station.MaintenanceDebt = clamp(station.MaintenanceDebt+debtGrowth, 0, 100)
			s.applyEquipmentWear(station, weather)

			healthDrop := 0.05 + station.LoadFactor*0.12
			if station.MaintenanceDebt > 55 {
				healthDrop += 0.18
			}
			if station.LoadFactor > 0.96 {
				healthDrop += 0.22
				station.Status = "overloaded"
				station.LastEvent = "high dispatch load is stressing the pumps"
			}

			if s.shouldTriggerOutage(*station, weather) {
				station.GridOnline = false
				if station.ReserveReady {
					station.ReserveActive = true
					station.Status = "reserve"
					station.LastEvent = "storm outage absorbed by reserve power"
					station.MaintenanceDebt = clamp(station.MaintenanceDebt+1.1, 0, 100)
					healthDrop += 0.12
					s.markReserveEquipment(station, "active")
					s.pushEvent("warning", station.Name+" transferred to reserve power during thunderstorm activity.")
				} else {
					station.Status = "outage"
					station.LastEvent = "storm outage cut pumping power"
					healthDrop += 0.45
					s.gridOutageHours++
					s.pushEvent("critical", station.Name+" lost external power with no ready reserve.")
				}
			}

			station.Health = clamp(station.Health-healthDrop, 35, 100)
			s.recalculateStationHealth(station)
		}

		s.refreshDerivedState()
		s.totalPumped += s.deliveredThisHour
		s.revenue += s.deliveredThisHour * 46
		s.powerCost += currentGridLoadPowerCost(s.stations) * 0.35
		if weather.StormSeverity > 0.74 {
			s.stormLossCost += 900 + float64(s.gridOutageHours)*80
		}

		if s.deliveredThisHour < s.hourlyTarget*0.88 {
			s.pushEvent("warning", "Hourly dispatch target missed. Review routing and station loading.")
			s.penaltyCost += 1800
			s.targetMissHours++
		}
		if s.reliability < 76 {
			s.pushEvent("warning", "Reliability dropped below dispatch comfort band.")
			s.penaltyCost += 700
		}
		if weather.StormSeverity > 0.74 {
			s.pushEvent("info", "Thunderstorm peak active over the "+weather.StormFront+" corridor.")
		}
	}

	s.refreshDerivedState()
	return s.Snapshot()
}

func (s *Simulation) PerformMaintenance(stationID string) (Snapshot, error) {
	station, err := s.stationByID(stationID)
	if err != nil {
		return Snapshot{}, err
	}

	station.MaintenanceDebt = clamp(station.MaintenanceDebt-26, 0, 100)
	station.Health = clamp(station.Health+7, 35, 100)
	station.ReserveReady = true
	station.Status = "service"
	station.LastEvent = "planned maintenance improved station readiness"
	for index := range station.Equipment {
		equipment := &station.Equipment[index]
		equipment.Health = clamp(equipment.Health+8, 35, 100)
		equipment.MaintenanceDebt = clamp(equipment.MaintenanceDebt-28, 0, 100)
		equipment.LastServiceHoursAgo = 0
		if equipment.Kind == "reserve" {
			equipment.Status = "ready"
		} else {
			equipment.Status = "stable"
		}
	}
	s.maintenanceCost += 7200
	s.pushEvent("info", station.Name+" completed planned maintenance and reserve checks.")
	s.recalculateStationHealth(station)
	s.refreshDerivedState()

	return s.Snapshot(), nil
}

func (s *Simulation) SetStationLoad(stationID string, loadFactor float64) (Snapshot, error) {
	station, err := s.stationByID(stationID)
	if err != nil {
		return Snapshot{}, err
	}

	station.LoadFactor = clamp(loadFactor, station.MinLoadFactor, station.MaxLoadFactor)
	station.LastEvent = "dispatch setpoint updated"
	if station.LoadFactor > 0.96 {
		station.Status = "overloaded"
	} else {
		station.Status = "stable"
	}

	s.pushEvent("info", station.Name+" dispatch load updated to "+formatPercent(station.LoadFactor)+".")
	s.recalculateStationHealth(station)
	s.refreshDerivedState()

	return s.Snapshot(), nil
}

func (s *Simulation) PrepareReserve(stationID string) (Snapshot, error) {
	station, err := s.stationByID(stationID)
	if err != nil {
		return Snapshot{}, err
	}

	station.ReserveReady = true
	station.MaintenanceDebt = clamp(station.MaintenanceDebt+1.8, 0, 100)
	station.LoadFactor = clamp(station.LoadFactor-0.03, station.MinLoadFactor, station.MaxLoadFactor)
	station.LastEvent = "reserve power tested and marked ready"
	station.Status = "reserve-ready"
	s.markReserveEquipment(station, "ready")
	s.maintenanceCost += 1800
	s.pushEvent("info", station.Name+" reserve power train has been prepared for storm conditions.")
	s.recalculateStationHealth(station)
	s.refreshDerivedState()

	return s.Snapshot(), nil
}

func (s *Simulation) SetBypassShare(share float64) Snapshot {
	s.bypassShare = clamp(share, 0.05, 0.55)
	s.pushEvent("info", "Central dispatch rerouted "+formatPercent(s.bypassShare)+" of flow toward the southern bypass.")
	s.refreshDerivedState()
	return s.Snapshot()
}

func (s *Simulation) currentWeather() WeatherState {
	return s.weatherAtHour(s.tickHours)
}

func (s *Simulation) weatherAtHour(hour int) WeatherState {
	stormCycle := hour % 24
	severity := 0.24
	switch {
	case stormCycle >= 0 && stormCycle < 6:
		severity = 0.24
	case stormCycle >= 6 && stormCycle < 12:
		severity = 0.46
	case stormCycle >= 12 && stormCycle < 18:
		severity = 0.82
	default:
		severity = 0.76
	}

	front := "west"
	switch {
	case stormCycle >= 4 && stormCycle < 8:
		front = "central"
	case stormCycle >= 8 && stormCycle < 12:
		front = "north"
	case stormCycle >= 12 && stormCycle < 18:
		front = "south"
	case stormCycle >= 18:
		front = "east"
	}

	advisory := "Normal dispatch window. Build throughput carefully."
	if severity > 0.74 {
		advisory = "Storm peak active: protect reserve power and keep bypass flexibility available."
	} else if severity > 0.42 {
		advisory = "Storm cells building: clear maintenance backlog before the next peak."
	}

	return WeatherState{
		Season:             "thunderstorm",
		StormFront:         front,
		StormSeverity:      round2(severity),
		GridOutageRisk:     round2(severity * 0.92),
		OperationsAdvisory: advisory,
	}
}

func (s *Simulation) refreshDerivedState() {
	weather := s.currentWeather()

	north := s.stationByIDMust("nps-north-intake")
	west := s.stationByIDMust("nps-west-booster")
	central := s.stationByIDMust("nps-central-hub")
	south := s.stationByIDMust("nps-south-loop")
	east := s.stationByIDMust("nps-east-export")

	northCentral := s.segmentByIDMust("seg-north-central")
	westCentral := s.segmentByIDMust("seg-west-central")
	centralEast := s.segmentByIDMust("seg-central-east")
	centralSouth := s.segmentByIDMust("seg-central-south")
	southEast := s.segmentByIDMust("seg-south-east")

	northFlow := math.Min(stationPotential(*north, weather), northCentral.Capacity)
	westFlow := math.Min(stationPotential(*west, weather), westCentral.Capacity)
	incomingCentral := northFlow + westFlow

	centralCapacity := stationPotential(*central, weather)
	centralHandled := math.Min(incomingCentral, centralCapacity)

	mainDesired := centralHandled * (1 - s.bypassShare)
	bypassDesired := centralHandled * s.bypassShare

	mainFlow := math.Min(mainDesired, centralEast.Capacity)
	southIngress := math.Min(bypassDesired, math.Min(centralSouth.Capacity, stationPotential(*south, weather)))
	southFlow := math.Min(southIngress, southEast.Capacity)

	eastCapacity := stationPotential(*east, weather)
	delivered := mainFlow + southFlow
	if delivered > eastCapacity && delivered > 0 {
		scale := eastCapacity / delivered
		mainFlow *= scale
		southFlow *= scale
		southIngress = southFlow
		delivered = eastCapacity
	}

	north.EffectiveThroughput = northFlow
	west.EffectiveThroughput = westFlow
	central.EffectiveThroughput = mainFlow + southIngress
	south.EffectiveThroughput = southFlow
	east.EffectiveThroughput = delivered

	northCentral.Flow = northFlow
	westCentral.Flow = westFlow
	centralEast.Flow = mainFlow
	centralSouth.Flow = southIngress
	southEast.Flow = southFlow

	s.updateSegmentStatus(northCentral, weather)
	s.updateSegmentStatus(westCentral, weather)
	s.updateSegmentStatus(centralEast, weather)
	s.updateSegmentStatus(centralSouth, weather)
	s.updateSegmentStatus(southEast, weather)

	s.deliveredThisHour = round2(delivered)
	s.reliability = round2(s.calculateReliability())
	s.serviceQuality = round2(s.calculateServiceQuality())
	netBalance := s.dispatchBudget + s.revenue - s.powerCost - s.maintenanceCost - s.stormLossCost - s.penaltyCost
	s.score = round2(s.totalPumped*0.42 + s.reliability*9 + s.serviceQuality*6 + netBalance*0.002)
	s.alerts = s.currentAlerts(weather)
}

func (s *Simulation) calculateReliability() float64 {
	total := 0.0
	for _, station := range s.stations {
		score := station.Health - station.MaintenanceDebt*0.34
		if station.Status == "outage" {
			score -= 11
		}
		if station.Status == "reserve" {
			score -= 4
		}
		if station.Status == "overloaded" {
			score -= 3
		}
		total += score
	}

	return clamp(total/float64(len(s.stations)), 0, 100)
}

func (s *Simulation) calculateServiceQuality() float64 {
	total := 0.0
	for _, station := range s.stations {
		score := 100 - station.MaintenanceDebt*0.58
		score += (station.Health - 80) * 0.45
		if station.ReserveReady {
			score += 3
		}
		if station.Status == "outage" {
			score -= 12
		}
		total += score
	}

	return clamp(total/float64(len(s.stations)), 0, 100)
}

func (s *Simulation) currentAlerts(weather WeatherState) []string {
	alerts := make([]string, 0, 12)
	if weather.StormSeverity > 0.74 {
		alerts = append(alerts, "Severe thunderstorm activity over the "+weather.StormFront+" corridor.")
	}
	if s.deliveredThisHour < s.hourlyTarget {
		alerts = append(alerts, "Dispatch target is under pressure. Increase healthy throughput or reroute flow.")
	}
	if s.bypassShare > 0.4 {
		alerts = append(alerts, "Southern bypass is carrying an elevated share of national flow.")
	}
	if s.currentEconomy().NetBalance < s.dispatchBudget*0.6 {
		alerts = append(alerts, "Operating margin is shrinking. Maintenance and penalty costs are rising.")
	}

	for _, station := range s.stations {
		switch station.Status {
		case "outage":
			alerts = append(alerts, station.Name+" is offline after a power loss.")
		case "reserve":
			alerts = append(alerts, station.Name+" is running on reserve power.")
		case "overloaded":
			alerts = append(alerts, station.Name+" is above the preferred dispatch envelope.")
		}
		if station.MaintenanceDebt > 55 {
			alerts = append(alerts, station.Name+" has a critical maintenance backlog.")
		}
		for _, equipment := range station.Equipment {
			if equipment.Health < 70 || equipment.MaintenanceDebt > 60 {
				alerts = append(alerts, station.Name+": "+equipment.Name+" requires intervention.")
			}
		}
	}

	return alerts
}

func (s *Simulation) updateSegmentStatus(segment *Segment, weather WeatherState) {
	segment.Utilization = round2(clamp(segment.Flow/segment.Capacity, 0, 1.35))
	segment.Status = "stable"
	if segment.Utilization > 0.95 {
		segment.Status = "congested"
	} else if segment.Utilization > 0.75 {
		segment.Status = "busy"
	}

	if weather.StormSeverity > 0.72 && segment.StormRisk > 0.6 {
		segment.Status = "storm-watch"
	}
	if segment.Flow == 0 {
		segment.Status = "idle"
	}
}

func (s *Simulation) shouldTriggerOutage(station Station, weather WeatherState) bool {
	if weather.StormSeverity < 0.72 || station.Region != weather.StormFront {
		return false
	}

	risk := station.LightningExposure + station.MaintenanceDebt/100*0.4
	if station.ReserveReady {
		risk -= 0.08
	}

	return risk > 0.82
}

func stationPotential(station Station, weather WeatherState) float64 {
	efficiency := station.Health / 100
	if station.MaintenanceDebt > 45 {
		efficiency *= 0.9
	}

	flow := station.PumpCapacity * station.LoadFactor * efficiency
	if station.Region == weather.StormFront {
		flow *= 1 - weather.StormSeverity*(0.03+station.LightningExposure*0.04)
	}

	switch {
	case !station.GridOnline && station.ReserveActive:
		flow *= 0.68
	case !station.GridOnline:
		flow = 0
	}

	if station.Status == "overloaded" {
		flow *= 0.98
	}

	return round2(math.Max(0, flow))
}

func formatPercent(value float64) string {
	return fmt.Sprintf("%.0f%%", round2(value*100))
}

func (s *Simulation) applyEquipmentWear(station *Station, weather WeatherState) {
	for index := range station.Equipment {
		equipment := &station.Equipment[index]
		equipment.LastServiceHoursAgo++
		if equipment.Kind == "reserve" {
			if station.ReserveReady {
				equipment.Status = "ready"
			} else {
				equipment.Status = "standby"
			}
		} else {
			equipment.Status = "stable"
		}

		debtGrowth := 0.35 + station.LoadFactor*equipment.LoadShare*0.8
		if equipment.Kind == "electrical" && station.Region == weather.StormFront {
			debtGrowth += weather.StormSeverity * 0.9
		}
		if equipment.Kind == "reserve" && station.ReserveReady {
			debtGrowth += 0.08
		}
		equipment.MaintenanceDebt = clamp(equipment.MaintenanceDebt+debtGrowth, 0, 100)

		healthDrop := 0.03 + station.LoadFactor*equipment.LoadShare*0.18
		if equipment.Kind == "electrical" && weather.StormSeverity > 0.7 && station.Region == weather.StormFront {
			healthDrop += 0.12
			equipment.Status = "storm-watch"
		}
		if equipment.MaintenanceDebt > 58 {
			healthDrop += 0.14
			equipment.Status = "degraded"
		}

		equipment.Health = clamp(equipment.Health-healthDrop, 35, 100)
		equipment.PowerDraw = round2(equipment.PowerDraw * (1 + station.LoadFactor*0.001))
	}
}

func (s *Simulation) recalculateStationHealth(station *Station) {
	if len(station.Equipment) == 0 {
		return
	}

	totalHealth := 0.0
	totalDebt := 0.0
	reserveReady := station.ReserveReady
	for _, equipment := range station.Equipment {
		totalHealth += equipment.Health
		totalDebt += equipment.MaintenanceDebt
		if equipment.Kind == "reserve" && equipment.Status != "ready" && equipment.Status != "active" {
			reserveReady = false
		}
	}

	station.Health = clamp(totalHealth/float64(len(station.Equipment)), 35, 100)
	station.MaintenanceDebt = clamp(totalDebt/float64(len(station.Equipment)), 0, 100)
	station.ReserveReady = reserveReady
}

func (s *Simulation) markReserveEquipment(station *Station, status string) {
	for index := range station.Equipment {
		equipment := &station.Equipment[index]
		if equipment.Kind != "reserve" {
			continue
		}

		equipment.Status = status
		if status == "active" {
			equipment.Health = clamp(equipment.Health-0.2, 35, 100)
			equipment.MaintenanceDebt = clamp(equipment.MaintenanceDebt+0.8, 0, 100)
		}
		if status == "ready" {
			equipment.LastServiceHoursAgo = 0
		}
	}
}

func currentGridLoadPowerCost(stations []Station) float64 {
	total := 0.0
	for _, station := range stations {
		for _, equipment := range station.Equipment {
			draw := equipment.PowerDraw
			if equipment.Kind == "reserve" && station.ReserveActive {
				draw *= 1.3
			}
			total += draw * station.LoadFactor
		}
	}

	return round2(total)
}
