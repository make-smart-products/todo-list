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
}

func TestMaintenanceImprovesStationCondition(t *testing.T) {
	simulation := NewDefaultSimulation()
	before := simulation.Snapshot()

	targetBefore := before.Stations[2]
	if _, err := simulation.PerformMaintenance(targetBefore.ID); err != nil {
		t.Fatalf("maintenance failed: %v", err)
	}

	after := simulation.Snapshot()
	targetAfter := after.Stations[2]

	if targetAfter.MaintenanceDebt >= targetBefore.MaintenanceDebt {
		t.Fatalf("expected maintenance debt to decrease, before %.2f after %.2f", targetBefore.MaintenanceDebt, targetAfter.MaintenanceDebt)
	}

	if targetAfter.Health <= targetBefore.Health {
		t.Fatalf("expected health to improve, before %.2f after %.2f", targetBefore.Health, targetAfter.Health)
	}
}

func TestStormOutageWithoutReserveDisablesStation(t *testing.T) {
	simulation := NewDefaultSimulation()
	simulation.tickHours = 15

	after := simulation.Advance(1)

	for _, station := range after.Stations {
		if station.ID != "nps-east-3" {
			continue
		}

		if station.GridOnline {
			t.Fatalf("expected east station grid to be offline during storm")
		}

		if station.ReserveActive {
			t.Fatalf("expected east station reserve to remain inactive")
		}

		return
	}

	t.Fatal("expected to find east station in snapshot")
}
