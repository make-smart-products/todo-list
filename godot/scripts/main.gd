extends Control

@export var backend_url := "http://127.0.0.1:8080"

@onready var http_request: HTTPRequest = $HTTPRequest
@onready var title_label: Label = $MarginContainer/VBox/TitleLabel
@onready var status_label: Label = $MarginContainer/VBox/StatusLabel
@onready var weather_label: Label = $MarginContainer/VBox/WeatherLabel
@onready var stations_label: Label = $MarginContainer/VBox/StationsLabel
@onready var refresh_button: Button = $MarginContainer/VBox/ButtonsRow/RefreshButton
@onready var tick_button: Button = $MarginContainer/VBox/ButtonsRow/TickButton
@onready var maintenance_button: Button = $MarginContainer/VBox/ButtonsRow/MaintenanceButton

var request_mode := ""

func _ready() -> void:
	var configured_backend_url := OS.get_environment("OIL_WORKER_API_URL")
	if configured_backend_url != "":
		backend_url = configured_backend_url

	title_label.text = "Oil Worker Dispatch Prototype"
	http_request.request_completed.connect(_on_request_completed)
	refresh_button.pressed.connect(_on_refresh_pressed)
	tick_button.pressed.connect(_on_tick_pressed)
	maintenance_button.pressed.connect(_on_maintenance_pressed)
	_fetch_state()

func _on_refresh_pressed() -> void:
	_fetch_state()

func _on_tick_pressed() -> void:
	_post_json("%s/api/v1/tick" % backend_url, {"hours": 1}, "tick")

func _on_maintenance_pressed() -> void:
	_post_json(
		"%s/api/v1/stations/maintenance" % backend_url,
		{"station_id": "nps-east-3"},
		"maintenance"
	)

func _fetch_state() -> void:
	request_mode = "state"
	status_label.text = "Loading simulation state..."
	var error := http_request.request("%s/api/v1/state" % backend_url)
	if error != OK:
		status_label.text = "State request failed: %s" % error

func _post_json(url: String, payload: Dictionary, mode: String) -> void:
	request_mode = mode
	status_label.text = "Sending %s command..." % mode
	var headers := PackedStringArray(["Content-Type: application/json"])
	var body := JSON.stringify(payload)
	var error := http_request.request(url, headers, HTTPClient.METHOD_POST, body)
	if error != OK:
		status_label.text = "Command failed to start: %s" % error

func _on_request_completed(result: int, response_code: int, _headers: PackedStringArray, body: PackedByteArray) -> void:
	if result != HTTPRequest.RESULT_SUCCESS:
		status_label.text = "Backend request failed with engine code %s" % result
		return

	if response_code < 200 or response_code >= 300:
		status_label.text = "Backend returned HTTP %s" % response_code
		return

	var parser := JSON.new()
	var parse_error := parser.parse(body.get_string_from_utf8())
	if parse_error != OK:
		status_label.text = "Could not parse backend JSON."
		return

	if request_mode == "maintenance":
		status_label.text = "Maintenance finished. Refreshing state..."
		_fetch_state()
		return

	_render_state(parser.data)

func _render_state(state: Dictionary) -> void:
	status_label.text = "Tick %s h | Pumped %.2f / %.2f | Reliability %.2f" % [
		state.get("tick_hours", 0),
		state.get("total_pumped", 0.0),
		state.get("plan_target", 0.0),
		state.get("reliability", 0.0),
	]

	var weather := state.get("weather", {})
	weather_label.text = "Storm front: %s | severity: %.2f | advisory: %s" % [
		weather.get("storm_front", "unknown"),
		weather.get("storm_severity", 0.0),
		weather.get("operations_advisory", "n/a"),
	]

	var lines := PackedStringArray()
	for station in state.get("stations", []):
		var power_mode := "grid"
		if station.get("reserve_active", false):
			power_mode = "reserve"
		elif not station.get("grid_online", true):
			power_mode = "offline"

		lines.append(
			"%s [%s] | load %.0f%% | health %.1f | debt %.1f | power %s | %s" % [
				station.get("name", "Station"),
				station.get("region", "region"),
				station.get("load_factor", 0.0) * 100.0,
				station.get("health", 0.0),
				station.get("maintenance_debt", 0.0),
				power_mode,
				station.get("last_event", ""),
			]
		)

	stations_label.text = "\n".join(lines)
