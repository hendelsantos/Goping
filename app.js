const API_BASE = "/api/v1";

const state = {
  monitors: [],
  activities: [],
  search: "",
  filter: "all",
};

const elements = {
  body: document.body,
  monitorList: document.querySelector("#monitor-list"),
  emptyState: document.querySelector("#empty-state"),
  search: document.querySelector("#monitor-search"),
  filterButton: document.querySelector("#filter-button"),
  dialog: document.querySelector("#monitor-dialog"),
  form: document.querySelector("#monitor-form"),
  activityList: document.querySelector("#activity-list"),
  toast: document.querySelector("#toast"),
};

async function apiRequest(path, options = {}) {
  const response = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: options.body ? { "Content-Type": "application/json", ...options.headers } : options.headers,
  });

  if (response.status === 204) return null;
  const payload = await response.json();
  if (!response.ok) throw new Error(payload.error?.message ?? "A API não conseguiu concluir a operação.");
  return payload.data;
}

function normalizeMonitor(item) {
  return {
    ...item,
    latency: item.latency_ms,
    lastCheck: item.last_check,
  };
}

async function loadDashboard() {
  const [monitors, activities] = await Promise.all([apiRequest("/monitors"), apiRequest("/activities")]);
  state.monitors = monitors.map(normalizeMonitor);
  state.activities = activities;
  renderMonitors();
  renderActivities();
}

function createElement(tag, className, text) {
  const element = document.createElement(tag);
  if (className) element.className = className;
  if (text !== undefined) element.textContent = text;
  return element;
}

function createIcon(name) {
  const icon = createElement("i");
  icon.dataset.lucide = name;
  icon.setAttribute("aria-hidden", "true");
  return icon;
}

function refreshIcons() {
  if (window.lucide) window.lucide.createIcons();
}

function statusLabel(status) {
  return { up: "Online", down: "Offline", unknown: "Desconhecido" }[status] ?? "Desconhecido";
}

function serviceInitials(name) {
  return name
    .split(" ")
    .filter(Boolean)
    .slice(0, 2)
    .map((word) => word[0])
    .join("")
    .toUpperCase();
}

// Cada linha é montada com APIs do DOM para que textos informados pelo usuário não virem HTML executável.
function createMonitorRow(monitor) {
  const row = createElement("div", "monitor-row");
  row.setAttribute("role", "row");

  const service = createElement("div", "service-name");
  service.setAttribute("role", "cell");
  service.append(createElement("span", "service-icon", serviceInitials(monitor.name)));
  const serviceCopy = createElement("div");
  serviceCopy.append(createElement("strong", "", monitor.name), createElement("small", "", monitor.url));
  service.append(serviceCopy);

  const status = createElement("span", `status-badge status-${monitor.status}`, statusLabel(monitor.status));
  status.setAttribute("role", "cell");

  const uptime = createElement("span", "mono-value", `${monitor.uptime.toFixed(2).replace(".", ",")}%`);
  uptime.setAttribute("role", "cell");
  const latency = createElement("span", "mono-value", monitor.latency ? `${monitor.latency} ms` : "Timeout");
  latency.setAttribute("role", "cell");
  const lastCheck = createElement("span", "mono-value", monitor.lastCheck);
  lastCheck.setAttribute("role", "cell");

  const removeButton = createElement("button", "icon-button row-action");
  removeButton.type = "button";
  removeButton.title = `Excluir ${monitor.name}`;
  removeButton.setAttribute("aria-label", `Excluir monitor ${monitor.name}`);
  removeButton.dataset.removeMonitor = monitor.id;
  removeButton.append(createIcon("trash-2"));

  row.append(service, status, uptime, latency, lastCheck, removeButton);
  return row;
}

function renderMonitors() {
  const searchTerm = state.search.trim().toLocaleLowerCase("pt-BR");
  const visibleMonitors = state.monitors.filter((monitor) => {
    const matchesSearch = `${monitor.name} ${monitor.url}`.toLocaleLowerCase("pt-BR").includes(searchTerm);
    const matchesFilter = state.filter === "all" || monitor.status === state.filter;
    return matchesSearch && matchesFilter;
  });

  elements.monitorList.replaceChildren(...visibleMonitors.map(createMonitorRow));
  elements.emptyState.hidden = visibleMonitors.length > 0;
  renderSummary();
  refreshIcons();
}

function renderSummary() {
  const total = state.monitors.length;
  const online = state.monitors.filter((monitor) => monitor.status === "up").length;
  const incidents = state.monitors.filter((monitor) => monitor.status === "down").length;
  const measuredLatencies = state.monitors.map((monitor) => monitor.latency).filter(Number.isFinite);
  const averageLatency = measuredLatencies.length
    ? Math.round(measuredLatencies.reduce((sum, value) => sum + value, 0) / measuredLatencies.length)
    : 0;
  const averageUptime = total
    ? state.monitors.reduce((sum, monitor) => sum + monitor.uptime, 0) / total
    : 0;

  document.querySelector("#online-count").textContent = online;
  document.querySelector("#total-count").textContent = ` / ${total}`;
  document.querySelector("#nav-monitor-count").textContent = total;
  document.querySelector("#latency-value").firstChild.textContent = averageLatency;
  document.querySelector("#uptime-value").firstChild.textContent = averageUptime.toFixed(2).replace(".", ",");
  document.querySelector("#incident-count").firstChild.textContent = incidents;
  document.querySelector("#services-caption").textContent = incidents
    ? `${incidents} ${incidents === 1 ? "serviço requer" : "serviços requerem"} atenção`
    : "Todos os serviços estão online";
}

function createActivityItem(activity) {
  const item = createElement("li", "activity-item");
  const iconName = { up: "check", down: "x", info: "gauge" }[activity.type];
  const iconWrap = createElement("span", `activity-icon ${activity.type}`);
  iconWrap.append(createIcon(iconName));

  const content = createElement("div", "activity-content");
  content.append(
    createElement("strong", "", activity.title),
    createElement("p", "", activity.description),
    createElement("time", "", activity.time),
  );
  item.append(iconWrap, content);
  return item;
}

function renderActivities() {
  if (!state.activities.length) {
    elements.activityList.replaceChildren(createElement("li", "activity-empty", "Nenhuma atividade recente."));
    return;
  }

  elements.activityList.replaceChildren(...state.activities.map(createActivityItem));
  refreshIcons();
}

let toastTimer;
function showToast(message) {
  clearTimeout(toastTimer);
  elements.toast.textContent = message;
  elements.toast.classList.add("visible");
  toastTimer = setTimeout(() => elements.toast.classList.remove("visible"), 2800);
}

function openDialog() {
  elements.dialog.showModal();
  setTimeout(() => elements.form.elements.name.focus(), 0);
}

function closeDialog() {
  elements.dialog.close();
  elements.form.reset();
}

async function addMonitor(formData) {
  const name = formData.get("name").trim();
  const url = formData.get("url").trim();
  try {
    await apiRequest("/monitors", {
      method: "POST",
      body: JSON.stringify({
        name,
        url,
        interval_seconds: Number(formData.get("interval")) * 60,
        timeout_seconds: Number(formData.get("timeout")),
      }),
    });
    await loadDashboard();
    closeDialog();
    showToast("Monitor salvo no servidor e aguardando a primeira verificação.");
  } catch (error) {
    showToast(error.message);
  }
}

async function removeMonitor(monitorId) {
  const monitor = state.monitors.find((item) => item.id === monitorId);
  if (!monitor) return;

  try {
    await apiRequest(`/monitors/${encodeURIComponent(monitorId)}`, { method: "DELETE" });
    await loadDashboard();
    showToast(`${monitor.name} foi removido.`);
  } catch (error) {
    showToast(error.message);
  }
}

function updateCurrentDate() {
  const dateLabel = document.querySelector(".section-heading .eyebrow");
  const currentDate = new Intl.DateTimeFormat("pt-BR", {
    weekday: "long",
    day: "2-digit",
    month: "long",
  }).format(new Date());
  dateLabel.textContent = currentDate.charAt(0).toUpperCase() + currentDate.slice(1);
}

// Eventos são registrados em um único ponto para deixar o fluxo da página fácil de acompanhar.
document.querySelectorAll("[data-open-monitor]").forEach((button) => button.addEventListener("click", openDialog));
document.querySelector("#close-dialog").addEventListener("click", closeDialog);
document.querySelector("#cancel-dialog").addEventListener("click", closeDialog);

elements.form.addEventListener("submit", (event) => {
  event.preventDefault();
  if (!elements.form.reportValidity()) return;
  void addMonitor(new FormData(elements.form));
});

elements.search.addEventListener("input", (event) => {
  state.search = event.currentTarget.value;
  renderMonitors();
});

elements.filterButton.addEventListener("click", () => {
  const filters = ["all", "up", "down", "pending"];
  const labels = { all: "Todos", up: "Online", down: "Offline", pending: "Pendentes" };
  state.filter = filters[(filters.indexOf(state.filter) + 1) % filters.length];
  elements.filterButton.title = `Filtro: ${labels[state.filter]}`;
  elements.filterButton.setAttribute("aria-label", `Filtro atual: ${labels[state.filter]}`);
  renderMonitors();
  showToast(`Exibindo: ${labels[state.filter]}.`);
});

elements.monitorList.addEventListener("click", (event) => {
  const button = event.target.closest("[data-remove-monitor]");
  if (button) void removeMonitor(button.dataset.removeMonitor);
});

document.querySelector("#clear-activity").addEventListener("click", async () => {
  try {
    await apiRequest("/activities", { method: "DELETE" });
    state.activities = [];
    renderActivities();
    showToast("Atividade recente limpa no servidor.");
  } catch (error) {
    showToast(error.message);
  }
});

document.querySelector("#menu-button").addEventListener("click", () => {
  elements.body.classList.toggle("menu-open");
});

document.querySelectorAll(".nav-link").forEach((link) => {
  link.addEventListener("click", () => elements.body.classList.remove("menu-open"));
});

elements.dialog.addEventListener("click", (event) => {
  const bounds = elements.dialog.getBoundingClientRect();
  const clickedOutside =
    event.clientX < bounds.left || event.clientX > bounds.right || event.clientY < bounds.top || event.clientY > bounds.bottom;
  if (clickedOutside) closeDialog();
});

document.querySelector("#chart-period").addEventListener("change", (event) => {
  const labels = { 24: "24 horas", 7: "7 dias", 30: "30 dias" };
  showToast(`Gráfico atualizado para ${labels[event.currentTarget.value]}.`);
});

window.addEventListener("DOMContentLoaded", async () => {
  updateCurrentDate();
  refreshIcons();
  try {
    await loadDashboard();
  } catch {
    renderMonitors();
    renderActivities();
    showToast("Não foi possível acessar a API. Inicie o servidor GoPing.");
  }
});