# oil-worker-game

Hybrid prototype for an oil pipeline operations game where:

- **Go** runs the simulation of trunk pipelines, pumping stations, maintenance, and thunderstorm-season risks.
- **Godot** renders the dispatch UI, scenario flow, and player actions.

## Current repository layout

```text
.
├── cmd/simserver/          # Go API entry point
├── docs/                   # Architecture notes
├── godot/                  # Godot 4 client project
├── internal/api/           # HTTP API for the simulation
├── internal/sim/           # Domain model and simulation rules
└── go.mod
```

## Simulation focus

The first slice models:

- pumping throughput across a simple national pipeline chain,
- station health and maintenance debt,
- reserve power readiness,
- thunderstorm season and storm-front migration,
- grid outages that reduce or stop throughput.

## Run the Go backend

```bash
go run ./cmd/simserver
```

Optional environment variable:

```bash
OIL_WORKER_API_ADDR=:8080
```

Available endpoints:

- `GET /healthz`
- `GET /api/v1/state`
- `POST /api/v1/tick` with `{"hours":1}`
- `POST /api/v1/stations/maintenance` with `{"station_id":"nps-east-3"}`

## Run the Godot client

1. Open the `godot/` directory in Godot 4.
2. Start the project.
3. The main scene will call the backend at `http://127.0.0.1:8080`.

Optional environment variable for the Godot client:

```bash
OIL_WORKER_API_URL=http://127.0.0.1:8080
```

## Tests

```bash
go test ./...
```

## Next development steps

1. Expand the network model from a linear route to a branching country map.
2. Add maintenance planning windows and station-level staffing constraints.
3. Introduce mission/scenario definitions that Godot can load from data files.
4. Replace the text UI with an interactive network map and station drill-down screens.
