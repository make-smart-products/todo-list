# oil-worker-game

Hybrid prototype for an oil pipeline operations game where:

- **Go** runs the simulation of trunk pipelines, pumping stations, maintenance, and thunderstorm-season risks.
- **Godot** renders the dispatch UI, scenario flow, and player actions.
- **Browser client** offers a playable dispatch console directly from the Go server.

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

The current slice models:

- a branching national pipeline network with a mainline and southern bypass,
- station health, maintenance debt, and reserve power readiness,
- dispatch load setpoints that can be changed per station,
- thunderstorm season and storm-front migration by region,
- grid outages that reduce or stop throughput,
- alerts, event log entries, and scoring for the current dispatch scenario,
- scenario objectives, forecast windows, equipment-level status, and dispatch economics.

## Run the Go backend

```bash
go run ./cmd/simserver
```

Optional environment variable:

```bash
OIL_WORKER_API_ADDR=:8080
```

Available endpoints:

- `GET /` for the playable browser client
- `GET /healthz`
- `GET /api/v1/state`
- `POST /api/v1/reset`
- `POST /api/v1/tick` with `{"hours":1}`
- `POST /api/v1/stations/maintenance` with `{"station_id":"nps-south-loop"}`
- `POST /api/v1/stations/load` with `{"station_id":"nps-west-booster","load_factor":0.92}`
- `POST /api/v1/stations/prepare_reserve` with `{"station_id":"nps-south-loop"}`
- `POST /api/v1/network/bypass_share` with `{"share":0.35}`

## Play in a web browser

1. Start the backend:

```bash
go run ./cmd/simserver
```

2. Open:

```text
http://127.0.0.1:8080/
```

The browser client includes:

- a clickable map of the country-wide pipeline network,
- station equipment panels,
- economic indicators,
- storm forecast,
- scenario objectives,
- buttons for load changes, maintenance, reserve preparation, routing, and shift reset.

## Run the Godot client

1. Open the `godot/` directory in Godot 4.
2. Start the project.
3. The main scene will call the backend at `http://127.0.0.1:8080`.
4. Click stations on the map to change load, prepare reserve, or launch maintenance.

Optional environment variable for the Godot client:

```bash
OIL_WORKER_API_URL=http://127.0.0.1:8080
```

## Tests

```bash
go test ./...
```

## Next development steps

1. Add maintenance planning windows and station-level staffing constraints.
2. Introduce mission/scenario definitions that Godot can load from data files.
3. Add more weather patterns, forecast horizons, and region-specific risk modifiers.
4. Replace the prototype labels with richer widgets, charts, and station drill-down panels.
5. Explore packaging the browser client and Godot client behind a shared save/progression model.
