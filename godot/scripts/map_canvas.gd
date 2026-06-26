extends Control

signal station_selected(station_id: String)

const STATION_POSITIONS := {
	"nps-north-intake": Vector2(140, 90),
	"nps-west-booster": Vector2(120, 250),
	"nps-central-hub": Vector2(360, 180),
	"nps-south-loop": Vector2(520, 320),
	"nps-east-export": Vector2(640, 150),
}

var snapshot: Dictionary = {}
var selected_station_id := ""

func _ready() -> void:
	mouse_filter = Control.MOUSE_FILTER_STOP
	queue_redraw()

func set_snapshot(next_snapshot: Dictionary, next_selected_station_id: String) -> void:
	snapshot = next_snapshot
	selected_station_id = next_selected_station_id
	queue_redraw()

func _gui_input(event: InputEvent) -> void:
	if event is InputEventMouseButton and event.pressed and event.button_index == MOUSE_BUTTON_LEFT:
		for station_id in STATION_POSITIONS.keys():
			if event.position.distance_to(STATION_POSITIONS[station_id]) <= 26.0:
				station_selected.emit(station_id)
				return

func _draw() -> void:
	draw_rect(Rect2(Vector2.ZERO, size), Color("0f151b"))
	_draw_grid()
	_draw_segments()
	_draw_stations()

func _draw_grid() -> void:
	for x in range(0, int(size.x), 48):
		draw_line(Vector2(x, 0), Vector2(x, size.y), Color("17212b"), 1.0)
	for y in range(0, int(size.y), 48):
		draw_line(Vector2(0, y), Vector2(size.x, y), Color("17212b"), 1.0)

func _draw_segments() -> void:
	var font := get_theme_default_font()
	var font_size := get_theme_default_font_size()

	for segment in snapshot.get("segments", []):
		var from_position := STATION_POSITIONS.get(segment.get("from_station", ""), Vector2.ZERO)
		var to_position := STATION_POSITIONS.get(segment.get("to_station", ""), Vector2.ZERO)
		var color := _segment_color(segment.get("status", "stable"))
		draw_line(from_position, to_position, color, 6.0)

		var center := from_position.lerp(to_position, 0.5)
		var label := "%.0f / %.0f" % [segment.get("flow", 0.0), segment.get("capacity", 0.0)]
		draw_string(font, center + Vector2(-32, -8), label, HORIZONTAL_ALIGNMENT_LEFT, -1, font_size, Color.WHITE)

func _draw_stations() -> void:
	var font := get_theme_default_font()
	var font_size := get_theme_default_font_size()

	for station in snapshot.get("stations", []):
		var station_id := station.get("id", "")
		var position := STATION_POSITIONS.get(station_id, Vector2.ZERO)
		var color := _station_color(station.get("status", "stable"))
		var outline_color := Color.WHITE if station_id == selected_station_id else Color("203040")

		draw_circle(position, 24.0, color)
		draw_arc(position, 25.0, 0.0, TAU, 32, outline_color, 3.0)
		draw_string(font, position + Vector2(-48, -32), station.get("name", station_id), HORIZONTAL_ALIGNMENT_LEFT, -1, font_size, Color.WHITE)
		draw_string(
			font,
			position + Vector2(-44, 44),
			"%.0f t/h" % station.get("effective_throughput", 0.0),
			HORIZONTAL_ALIGNMENT_LEFT,
			-1,
			font_size,
			Color("c9d4df")
		)

func _segment_color(status: String) -> Color:
	match status:
		"congested":
			return Color("ff7b54")
		"storm-watch":
			return Color("f5d547")
		"idle":
			return Color("405261")
		_:
			return Color("3aa3ff")

func _station_color(status: String) -> Color:
	match status:
		"outage":
			return Color("d1495b")
		"reserve":
			return Color("f7b538")
		"overloaded":
			return Color("ff7b54")
		"service":
			return Color("7bd389")
		"reserve-ready":
			return Color("6ec6ff")
		_:
			return Color("00a878")
