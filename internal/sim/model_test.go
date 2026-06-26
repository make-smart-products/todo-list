package sim

import "testing"

func TestAdvanceIncreasesThroughputAndTracksStormSeason(t *testing.T) {
	simulation := NewDefaultSimulation()

	before := simulation.Snapshot()
	after := simulation.Advance(6)

	if after.TickHours != 6 {
		t.Fatalf("expected tick hours to advance to 6, got %d", after.TickHours)
	}

	if after.TotalPumped <= before.TotalPumped {
		t.Fatalf("expected total pumped to increase, before %.2f after %.2f", before.TotalPumped, after.TotalPumped)
	}

	if after.Weather.Season != "thunderstorm" {
		t.Fatalf("expected thunderstorm season, got %q", after.Weather.Season)
	}

	if len(after.EventLog) == 0 {
		t.Fatal("expected event log entries after advancing the simulation")
	}
}

func TestMaintenanceImprovesStationCondition(t *testing.T) {
	simulation := NewDefaultSimulation()
	before := simulation.Snapshot()

	targetBefore := before.Stations[3]
	if _, err := simulation.PerformMaintenance(targetBefore.ID); err != nil {
		t.Fatalf("maintenance failed: %v", err)
	}

	after := simulation.Snapshot()
	targetAfter := findStation(t, after, targetBefore.ID)

	if targetAfter.MaintenanceDebt >= targetBefore.MaintenanceDebt {
		t.Fatalf("expected maintenance debt to decrease, before %.2f after %.2f", targetBefore.MaintenanceDebt, targetAfter.MaintenanceDebt)
	}

	if targetAfter.Health <= targetBefore.Health {
		t.Fatalf("expected health to improve, before %.2f after %.2f", targetBefore.Health, targetAfter.Health)
	}

	if !targetAfter.ReserveReady {
		t.Fatal("expected maintenance to make reserve power ready")
	}
}

func TestSetStationLoadUpdatesDispatchSetpoint(t *testing.T) {
	simulation := NewDefaultSimulation()

	before := findStation(t, simulation.Snapshot(), "nps-west-booster")
	after, err := simulation.SetStationLoad("nps-west-booster", 1.2)
	if err != nil {
		t.Fatalf("set station load failed: %v", err)
	}

	station := findStation(t, after, "nps-west-booster")
	if station.LoadFactor <= before.LoadFactor {
		t.Fatalf("expected load factor to increase, before %.2f after %.2f", before.LoadFactor, station.LoadFactor)
	}

	if station.LoadFactor != station.MaxLoadFactor {
		t.Fatalf("expected load factor to clamp to max %.2f, got %.2f", station.MaxLoadFactor, station.LoadFactor)
	}
}

func TestBypassShareRedirectsFlow(t *testing.T) {
	simulation := NewDefaultSimulation()

	before := simulation.Snapshot()
	beforeFlow := findSegment(t, before, "seg-central-south").Flow

	after := simulation.SetBypassShare(0.5)
	afterFlow := findSegment(t, after, "seg-central-south").Flow

	if after.BypassShare <= before.BypassShare {
		t.Fatalf("expected bypass share to increase, before %.2f after %.2f", before.BypassShare, after.BypassShare)
	}

	if afterFlow <= beforeFlow {
		t.Fatalf("expected southern bypass flow to grow, before %.2f after %.2f", beforeFlow, afterFlow)
	}
}

func TestPrepareReserveProtectsSouthLoopDuringStorm(t *testing.T) {
	simulation := NewDefaultSimulation()
	if _, err := simulation.PrepareReserve("nps-south-loop"); err != nil {
		t.Fatalf("prepare reserve failed: %v", err)
	}

	simulation.tickHours = 11
	after := simulation.Advance(1)
	south := findStation(t, after, "nps-south-loop")

	if !south.ReserveActive {
		t.Fatal("expected south loop to switch to reserve during southern storm peak")
	}
}

func findStation(t *testing.T, snapshot Snapshot, stationID string) Station {
	t.Helper()

	for _, station := range snapshot.Stations {
		if station.ID == stationID {
			return station
		}
	}

	t.Fatalf("station %s not found", stationID)
	return Station{}
}

func findSegment(t *testing.T, snapshot Snapshot, segmentID string) Segment {
	t.Helper()

	for _, segment := range snapshot.Segments {
		if segment.ID == segmentID {
			return segment
		}
	}

	t.Fatalf("segment %s not found", segmentID)
	return Segment{}
}
