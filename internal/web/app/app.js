const state = {
  snapshot: null,
  selectedStationId: "nps-central-hub",
};

const els = {
  scenarioBriefing: document.getElementById("scenario-briefing"),
  metricTotalPumped: document.getElementById("metric-total-pumped"),
  metricPlanProgress: document.getElementById("metric-plan-progress"),
  metricHourlyThroughput: document.getElementById("metric-hourly-throughput"),
  metricHourlyTarget: document.getElementById("metric-hourly-target"),
  metricReliability: document.getElementById("metric-reliability"),
  metricServiceQuality: document.getElementById("metric-service-quality"),
  metricBalance: document.getElementById("metric-balance"),
  metricScore: document.getElementById("metric-score"),
  weatherBanner: document.getElementById("weather-banner"),
  segmentsList: document.getElementById("segments-list"),
  forecastList: document.getElementById("forecast-list"),
  stationTitle: document.getElementById("station-title"),
  stationStatusBadge: document.getElementById("station-status-badge"),
  stationSummary: document.getElementById("station-summary"),
  equipmentList: document.getElementById("equipment-list"),
  alertsList: document.getElementById("alerts-list"),
  objectivesList: document.getElementById("objectives-list"),
  economyList: document.getElementById("economy-list"),
  eventLogList: document.getElementById("event-log-list"),
  loadSlider: document.getElementById("load-slider"),
  loadValue: document.getElementById("load-value"),
  bypassSlider: document.getElementById("bypass-slider"),
  bypassValue: document.getElementById("bypass-value"),
  refreshButton: document.getElementById("refresh-button"),
  tick1Button: document.getElementById("tick-1-button"),
  tick6Button: document.getElementById("tick-6-button"),
  resetButton: document.getElementById("reset-button"),
  applyLoadButton: document.getElementById("apply-load-button"),
  applyBypassButton: document.getElementById("apply-bypass-button"),
  maintenanceButton: document.getElementById("maintenance-button"),
  reserveButton: document.getElementById("reserve-button"),
};

async function requestJson(url, options = {}) {
  const response = await fetch(url, {
    headers: { "Content-Type": "application/json" },
    ...options,
  });

  if (!response.ok) {
    const body = await response.text();
    throw new Error(body || `HTTP ${response.status}`);
  }

  return response.json();
}

function formatNumber(value) {
  return new Intl.NumberFormat("ru-RU", { maximumFractionDigits: 2 }).format(value);
}

function setSelectedStation(stationId) {
  state.selectedStationId = stationId;
  render();
}

function getStation(stationId = state.selectedStationId) {
  return state.snapshot?.stations?.find((station) => station.id === stationId) ?? null;
}

function renderMetrics(snapshot) {
  const scenario = snapshot.scenario ?? {};
  const economy = snapshot.economy ?? {};

  els.scenarioBriefing.textContent =
    `${scenario.briefing ?? ""} Осталось часов: ${scenario.hours_remaining ?? 0}.`;
  els.metricTotalPumped.textContent = `${formatNumber(snapshot.total_pumped)} т`;
  els.metricPlanProgress.textContent =
    `План: ${formatNumber(snapshot.plan_target)} т | ${Math.round((snapshot.plan_progress ?? 0) * 100)}%`;
  els.metricHourlyThroughput.textContent = `${formatNumber(snapshot.delivered_this_hour)} т/ч`;
  els.metricHourlyTarget.textContent =
    `Цель часа: ${formatNumber(snapshot.hourly_target)} т/ч`;
  els.metricReliability.textContent = formatNumber(snapshot.reliability);
  els.metricServiceQuality.textContent =
    `Качество обслуживания: ${formatNumber(snapshot.service_quality)}`;
  els.metricBalance.textContent = `${formatNumber(economy.net_balance ?? 0)} у.е.`;
  els.metricScore.textContent = `Очки сценария: ${formatNumber(snapshot.score)}`;
  els.weatherBanner.textContent =
    `Фронт: ${snapshot.weather.storm_front} | гроза ${Math.round(snapshot.weather.storm_severity * 100)}%`;
}

function renderSegments(snapshot) {
  els.segmentsList.innerHTML = "";
  snapshot.segments.forEach((segment) => {
    const card = document.createElement("article");
    card.className = "stack-card";
    card.innerHTML = `
      <h4>${segment.name}</h4>
      <p>${formatNumber(segment.flow)} / ${formatNumber(segment.capacity)} т/ч</p>
      <p class="muted">Загрузка: ${Math.round(segment.utilization * 100)}% | Статус: ${segment.status}</p>
    `;
    els.segmentsList.appendChild(card);

    const line = document.getElementById(`${segment.id}-line`);
    if (line) {
      line.style.stroke =
        segment.status === "congested"
          ? "#ff7b54"
          : segment.status === "storm-watch"
            ? "#f5d547"
            : segment.status === "idle"
              ? "#52606d"
              : "#58a6ff";
    }
  });
}

function renderForecast(snapshot) {
  els.forecastList.innerHTML = "";
  snapshot.forecast.forEach((entry) => {
    const card = document.createElement("article");
    card.className = "stack-card";
    card.innerHTML = `
      <h4>${entry.window_label}</h4>
      <p>Фронт: ${entry.storm_front}</p>
      <p class="muted">Сила грозы: ${Math.round(entry.storm_severity * 100)}% | Риск потери питания: ${Math.round(entry.grid_outage_risk * 100)}%</p>
    `;
    els.forecastList.appendChild(card);
  });
}

function renderStation(snapshot) {
  const station = getStation();
  if (!station) {
    els.stationTitle.textContent = "Станция";
    els.stationSummary.textContent = "Станция не выбрана.";
    return;
  }

  els.stationTitle.textContent = `${station.name} (${station.region})`;
  els.stationStatusBadge.textContent = station.status;
  els.stationStatusBadge.className = "status-badge";
  els.stationSummary.textContent =
    `Роль: ${station.role}. Поток: ${formatNumber(station.effective_throughput)} т/ч. Нагрузка: ${Math.round(station.load_factor * 100)}%. ` +
    `Health: ${formatNumber(station.health)}. Debt: ${formatNumber(station.maintenance_debt)}. ${station.last_event}`;

  els.loadSlider.value = Math.round(station.load_factor * 100);
  els.loadValue.textContent = `${els.loadSlider.value}%`;
  els.bypassSlider.value = Math.round((snapshot.bypass_share ?? 0.22) * 100);
  els.bypassValue.textContent = `${els.bypassSlider.value}%`;

  els.equipmentList.innerHTML = "";
  station.equipment.forEach((equipment) => {
    const card = document.createElement("article");
    card.className = "stack-card";
    card.innerHTML = `
      <h4>${equipment.name}</h4>
      <p>${equipment.kind} | status: ${equipment.status}</p>
      <p class="muted">Health: ${formatNumber(equipment.health)} | Debt: ${formatNumber(equipment.maintenance_debt)} | Draw: ${formatNumber(equipment.power_draw)} МВт</p>
      <p class="muted">Последнее ТО: ${equipment.last_service_hours_ago} ч назад</p>
    `;
    els.equipmentList.appendChild(card);
  });

  document.querySelectorAll(".station-node").forEach((node) => {
    const nodeStation = snapshot.stations.find(
      (item) => `station-${item.id}` === node.id,
    );
    node.classList.remove("selected", "reserve", "outage", "overloaded", "service", "reserve-ready");
    if (!nodeStation) return;
    node.classList.add(nodeStation.status);
    if (nodeStation.id === state.selectedStationId) {
      node.classList.add("selected");
    }
  });
}

function renderAlerts(snapshot) {
  els.alertsList.innerHTML = "";
  if (!snapshot.alerts.length) {
    els.alertsList.innerHTML = `<p class="empty">Активных предупреждений нет.</p>`;
    return;
  }

  snapshot.alerts.forEach((alert) => {
    const card = document.createElement("article");
    card.className = "stack-card";
    card.innerHTML = `<p>${alert}</p>`;
    els.alertsList.appendChild(card);
  });
}

function renderObjectives(snapshot) {
  els.objectivesList.innerHTML = "";
  snapshot.scenario.objectives.forEach((objective) => {
    const card = document.createElement("article");
    card.className = "stack-card";
    card.innerHTML = `
      <h4>${objective.title}</h4>
      <p>${objective.description}</p>
      <p class="status-${objective.status}">${formatNumber(objective.current)} / ${formatNumber(objective.target)} ${objective.unit} — ${objective.status}</p>
    `;
    els.objectivesList.appendChild(card);
  });
}

function renderEconomy(snapshot) {
  const economy = snapshot.economy;
  els.economyList.innerHTML = "";
  [
    ["Выручка", economy.revenue],
    ["Энергозатраты", economy.power_cost],
    ["Обслуживание", economy.maintenance_cost],
    ["Потери от гроз", economy.storm_loss_cost],
    ["Штрафы", economy.penalty_cost],
    ["Ожидаемый финал смены", economy.predicted_end_balance],
  ].forEach(([label, value]) => {
    const card = document.createElement("article");
    card.className = "stack-card";
    card.innerHTML = `<h4>${label}</h4><p>${formatNumber(value)} у.е.</p>`;
    els.economyList.appendChild(card);
  });
}

function renderEventLog(snapshot) {
  els.eventLogList.innerHTML = "";
  [...snapshot.event_log].reverse().forEach((event) => {
    const card = document.createElement("article");
    card.className = "stack-card";
    card.innerHTML = `
      <h4>[${event.tick_hours}ч] ${event.severity.toUpperCase()}</h4>
      <p>${event.message}</p>
    `;
    els.eventLogList.appendChild(card);
  });
}

function render() {
  const snapshot = state.snapshot;
  if (!snapshot) return;

  renderMetrics(snapshot);
  renderSegments(snapshot);
  renderForecast(snapshot);
  renderStation(snapshot);
  renderAlerts(snapshot);
  renderObjectives(snapshot);
  renderEconomy(snapshot);
  renderEventLog(snapshot);
}

async function refreshState() {
  state.snapshot = await requestJson("/api/v1/state");
  if (!getStation()) {
    state.selectedStationId = state.snapshot.stations?.[0]?.id ?? "";
  }
  render();
}

async function postAction(path, payload = {}) {
  state.snapshot = await requestJson(path, {
    method: "POST",
    body: JSON.stringify(payload),
  });
  render();
}

function wireEvents() {
  els.refreshButton.addEventListener("click", () => refreshState());
  els.tick1Button.addEventListener("click", () => postAction("/api/v1/tick", { hours: 1 }));
  els.tick6Button.addEventListener("click", () => postAction("/api/v1/tick", { hours: 6 }));
  els.resetButton.addEventListener("click", () => postAction("/api/v1/reset"));
  els.applyLoadButton.addEventListener("click", () =>
    postAction("/api/v1/stations/load", {
      station_id: state.selectedStationId,
      load_factor: Number(els.loadSlider.value) / 100,
    }),
  );
  els.maintenanceButton.addEventListener("click", () =>
    postAction("/api/v1/stations/maintenance", { station_id: state.selectedStationId }),
  );
  els.reserveButton.addEventListener("click", () =>
    postAction("/api/v1/stations/prepare_reserve", { station_id: state.selectedStationId }),
  );
  els.applyBypassButton.addEventListener("click", () =>
    postAction("/api/v1/network/bypass_share", { share: Number(els.bypassSlider.value) / 100 }),
  );

  els.loadSlider.addEventListener("input", () => {
    els.loadValue.textContent = `${els.loadSlider.value}%`;
  });
  els.bypassSlider.addEventListener("input", () => {
    els.bypassValue.textContent = `${els.bypassSlider.value}%`;
  });

  document.querySelectorAll(".station-node").forEach((node) => {
    node.addEventListener("click", () => {
      const stationId = node.id.replace("station-", "");
      setSelectedStation(stationId);
    });
  });
}

wireEvents();
refreshState().catch((error) => {
  els.scenarioBriefing.textContent = `Не удалось загрузить игру: ${error.message}`;
});
