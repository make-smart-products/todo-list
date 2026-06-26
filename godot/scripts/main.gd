extends Control

@export var backend_url := "http://127.0.0.1:8080"

@onready var http_request: HTTPRequest = $HTTPRequest
@onready var title_label: Label = $MarginContainer/VBox/TitleLabel
@onready var summary_label: Label = $MarginContainer/VBox/SummaryLabel
@onready var weather_label: Label = $MarginContainer/VBox/WeatherLabel
@onready var refresh_button: Button = $MarginContainer/VBox/ButtonsRow/RefreshButton
@onready var tick_button: Button = $MarginContainer/VBox/ButtonsRow/TickButton
@onready var tick_six_button: Button = $MarginContainer/VBox/ButtonsRow/TickSixButton
@onready var map_canvas = $MarginContainer/VBox/MainSplit/MapPanel/MapVBox/MapCanvas
@onready var segment_label: Label = $MarginContainer/VBox/MainSplit/MapPanel/MapVBox/SegmentLabel
@onready var selected_station_label: Label = $MarginContainer/VBox/MainSplit/Sidebar/SidebarVBox/SelectedStationLabel
@onready var station_details_label: Label = $MarginContainer/VBox/MainSplit/Sidebar/SidebarVBox/StationDetailsLabel
@onready var load_value_label: Label = $MarginContainer/VBox/MainSplit/Sidebar/SidebarVBox/LoadValueLabel
@onready var load_slider: HSlider = $MarginContainer/VBox/MainSplit/Sidebar/SidebarVBox/LoadSlider
@onready var apply_load_button: Button = $MarginContainer/VBox/MainSplit/Sidebar/SidebarVBox/ActionButtonsRow/ApplyLoadButton
@onready var maintenance_button: Button = $MarginContainer/VBox/MainSplit/Sidebar/SidebarVBox/ActionButtonsRow/MaintenanceButton
@onready var reserve_button: Button = $MarginContainer/VBox/MainSplit/Sidebar/SidebarVBox/ActionButtonsRow/ReserveButton
@onready var bypass_value_label: Label = $MarginContainer/VBox/MainSplit/Sidebar/SidebarVBox/BypassValueLabel
@onready var bypass_slider: HSlider = $MarginContainer/VBox/MainSplit/Sidebar/SidebarVBox/BypassSlider
@onready var apply_bypass_button: Button = $MarginContainer/VBox/MainSplit/Sidebar/SidebarVBox/ApplyBypassButton
@onready var alerts_label: Label = $MarginContainer/VBox/MainSplit/Sidebar/SidebarVBox/AlertsLabel
@onready var event_log_label: Label = $MarginContainer/VBox/MainSplit/Sidebar/SidebarVBox/EventLogLabel

var request_mode := ""
var snapshot: Dictionary = {}
var selected_station_id := ""

func _ready() -> void:
	var configured_backend_url := OS.get_environment("OIL_WORKER_API_URL")
	if configured_backend_url != "":
		backend_url = configured_backend_url

	title_label.text = "Oil Worker Dispatch Prototype"
	http_request.request_completed.connect(_on_request_completed)
	refresh_button.pressed.connect(_on_refresh_pressed)
	tick_button.pressed.connect(_on_tick_pressed)
	tick_six_button.pressed.connect(_on_tick_six_pressed)
	maintenance_button.pressed.connect(_on_maintenance_pressed)
	apply_load_button.pressed.connect(_on_apply_load_pressed)
	reserve_button.pressed.connect(_on_prepare_reserve_pressed)
	apply_bypass_button.pressed.connect(_on_apply_bypass_pressed)
	load_slider.value_changed.connect(_on_load_slider_changed)
	bypass_slider.value_changed.connect(_on_bypass_slider_changed)
	map_canvas.station_selected.connect(_on_station_selected)

	_on_load_slider_changed(load_slider.value)
	_on_bypass_slider_changed(bypass_slider.value)
	_fetch_state()

func _on_refresh_pressed() -> void:
	_fetch_state()

func _on_tick_pressed() -> void:
	_post_json("%s/api/v1/tick" % backend_url, {"hours": 1}, "tick-1h")

func _on_tick_six_pressed() -> void:
	_post_json("%s/api/v1/tick" % backend_url, {"hours": 6}, "tick-6h")

func _on_maintenance_pressed() -> void:
	if selected_station_id == "":
		return

	_post_json(
		"%s/api/v1/stations/maintenance" % backend_url,
		{"station_id": selected_station_id},
		"maintenance"
	)

func _on_apply_load_pressed() -> void:
	if selected_station_id == "":
		return

	_post_json(
		"%s/api/v1/stations/load" % backend_url,
		{"station_id": selected_station_id, "load_factor": load_slider.value / 100.0},
		"load"
	)

func _on_prepare_reserve_pressed() -> void:
	if selected_station_id == "":
		return

	_post_json(
		"%s/api/v1/stations/prepare_reserve" % backend_url,
		{"station_id": selected_station_id},
		"prepare-reserve"
	)

func _on_apply_bypass_pressed() -> void:
	_post_json(
		"%s/api/v1/network/bypass_share" % backend_url,
		{"share": bypass_slider.value / 100.0},
		"bypass"
	)

func _fetch_state() -> void:
	request_mode = "state"
	summary_label.text = "Loading simulation state..."
	var error := http_request.request("%s/api/v1/state" % backend_url)
	if error != OK:
		summary_label.text = "State request failed: %s" % error

func _post_json(url: String, payload: Dictionary, mode: String) -> void:
	request_mode = mode
	summary_label.text = "Sending %s command..." % mode
	var headers := PackedStringArray(["Content-Type: application/json"])
	var body := JSON.stringify(payload)
	var error := http_request.request(url, headers, HTTPClient.METHOD_POST, body)
	if error != OK:
		summary_label.text = "Command failed to start: %s" % error

func _on_request_completed(result: int, response_code: int, _headers: PackedStringArray, body: PackedByteArray) -> void:
	if result != HTTPRequest.RESULT_SUCCESS:
		summary_label.text = "Backend request failed with engine code %s" % result
		return

	if response_code < 200 or response_code >= 300:
		summary_label.text = "Backend returned HTTP %s" % response_code
		return

	var parser := JSON.new()
	var parse_error := parser.parse(body.get_string_from_utf8())
	if parse_error != OK:
		summary_label.text = "Could not parse backend JSON."
		return

	_render_state(parser.data)

func _render_state(state: Dictionary) -> void:
	snapshot = state
	if selected_station_id == "" or _find_station(selected_station_id).is_empty():
		var stations: Array = snapshot.get("stations", [])
		if not stations.is_empty():
			selected_station_id = stations[0].get("id", "")

	summary_label.text = "Scenario: %s | Tick %s h | Pumped %.2f / %.2f | Hour %.2f / %.2f | Score %.2f | Reliability %.2f" % [
		state.get("scenario_name", "Oil Worker"),
		state.get("tick_hours", 0),
		state.get("total_pumped", 0.0),
		state.get("plan_target", 0.0),
		state.get("delivered_this_hour", 0.0),
		state.get("hourly_target", 0.0),
		state.get("score", 0.0),
		state.get("reliability", 0.0),
	]

	var weather := state.get("weather", {})
	weather_label.text = "Storm front: %s | severity: %.2f | bypass share: %.0f%% | advisory: %s" % [
		weather.get("storm_front", "unknown"),
		weather.get("storm_severity", 0.0),
		state.get("bypass_share", 0.0) * 100.0,
		weather.get("operations_advisory", "n/a"),
	]

	_render_selected_station()
	_render_segments()
	_render_alerts()
	_render_event_log()
	map_canvas.set_snapshot(snapshot, selected_station_id)

func _render_selected_station() -> void:
	var station := _find_station(selected_station_id)
	if station.is_empty():
		selected_station_label.text = "No station selected"
		station_details_label.text = ""
		return

	selected_station_label.text = "%s (%s / %s)" % [
		station.get("name", "Station"),
		station.get("region", "region"),
		station.get("role", "role"),
	]

	var power_mode := "grid"
	if station.get("reserve_active", false):
		power_mode = "reserve"
	elif not station.get("grid_online", true):
		power_mode = "offline"

	station_details_label.text = "Load %.0f%% | Health %.1f | Debt %.1f | Throughput %.0f t/h | Power %s | Status %s\n%s" % [
		station.get("load_factor", 0.0) * 100.0,
		station.get("health", 0.0),
		station.get("maintenance_debt", 0.0),
		station.get("effective_throughput", 0.0),
		power_mode,
		station.get("status", "stable"),
		station.get("last_event", ""),
	]

	load_slider.set_value_no_signal(station.get("load_factor", 0.7) * 100.0)
	_on_load_slider_changed(load_slider.value)
	bypass_slider.set_value_no_signal(snapshot.get("bypass_share", 0.2) * 100.0)
	_on_bypass_slider_changed(bypass_slider.value)

func _render_segments() -> void:
	var lines := PackedStringArray()
	for segment in snapshot.get("segments", []):
		lines.append(
			"%s | %.0f / %.0f t/h | util %.0f%% | %s" % [
				segment.get("name", "Segment"),
				segment.get("flow", 0.0),
				segment.get("capacity", 0.0),
				segment.get("utilization", 0.0) * 100.0,
				segment.get("status", "stable"),
			]
		)

	segment_label.text = "\n".join(lines)

func _render_alerts() -> void:
	var alerts := snapshot.get("alerts", [])
	if alerts.is_empty():
		alerts_label.text = "No active alerts."
		return

	var lines := PackedStringArray()
	for alert in alerts:
		lines.append("• %s" % alert)
	alerts_label.text = "\n".join(lines)

func _render_event_log() -> void:
	var events := snapshot.get("event_log", [])
	if events.is_empty():
		event_log_label.text = "No events logged yet."
		return

	var lines := PackedStringArray()
	for index in range(events.size() - 1, -1, -1):
		var entry := events[index]
		lines.append("[%sh][%s] %s" % [
			entry.get("tick_hours", 0),
			str(entry.get("severity", "info")).to_upper(),
			entry.get("message", ""),
		])

	event_log_label.text = "\n".join(lines)

func _find_station(station_id: String) -> Dictionary:
	for station in snapshot.get("stations", []):
		if station.get("id", "") == station_id:
			return station
	return {}

func _on_station_selected(station_id: String) -> void:
	selected_station_id = station_id
	_render_selected_station()
	map_canvas.set_snapshot(snapshot, selected_station_id)

func _on_load_slider_changed(value: float) -> void:
	load_value_label.text = "Station load setpoint: %.0f%%" % value

func _on_bypass_slider_changed(value: float) -> void:
	bypass_value_label.text = "Southern bypass share: %.0f%%" % value
