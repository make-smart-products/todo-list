# Go + Godot Architecture

## Why this split fits Oil Worker

The game is primarily a **systems simulation**:

- national-scale pumping targets,
- equipment wear,
- electrical reliability,
- thunderstorm-season disruptions,
- station maintenance quality.

Those parts benefit from a backend-style domain layer with deterministic rules and straightforward automated tests. Go is a good fit for that simulation core.

Godot is used for:

- the country map,
- dispatch and station UI,
- player input,
- scenario progression,
- alerts, audio, and presentation.

## Responsibility split

### Go layer

The Go layer owns:

- simulation state,
- time advancement,
- station condition changes,
- outage effects,
- throughput calculations,
- maintenance actions,
- API contracts for the client.

### Godot layer

The Godot layer owns:

- the main scene tree,
- visual rendering of the network,
- action buttons and later map interactions,
- event log and warnings,
- camera and UX flow,
- mission briefings and results screens.

## Integration contract

The first integration step uses a local HTTP API.

### Initial endpoints

- `GET /api/v1/state`
  - returns the current snapshot of the network, stations, and weather
- `POST /api/v1/tick`
  - advances the simulation by a given number of hours
- `POST /api/v1/stations/maintenance`
  - applies maintenance to a single station

This API-first split keeps the domain logic independent from the Godot scene graph and makes the simulation easy to test without the game client.

## Why not embed Go directly into Godot

Direct embedding is possible only with extra native integration layers and toolchain complexity. For the first versions, an HTTP boundary is simpler because it:

- keeps iteration fast,
- allows headless simulation tests,
- supports future web dashboards or debug tooling,
- makes data contracts explicit.

If performance or packaging later becomes a concern, the architecture can evolve toward:

- a native extension,
- a shared library bridge,
- or an in-process messaging layer.

## Suggested near-term roadmap

### Milestone 1

- stabilize the simulation model,
- define scenario data,
- build a clickable map UI in Godot,
- show real station and segment states instead of text-only summaries.

### Milestone 2

- add branching routes and bottlenecks,
- model reserve power tests and failures,
- add maintenance queues and scheduled service windows,
- introduce scoring by throughput, reliability, and outages.

### Milestone 3

- add campaign progression,
- expose weather forecasts,
- add alarms, overlays, and station detail panels,
- package the Go backend and Godot client into one runnable distribution.
