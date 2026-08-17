const chartInstances = {};

const CHART_COLORS = {
  accent: '#3d7a54',
  warm: '#c4784a',
  danger: '#dc4c4c',
  blue: '#2563eb',
  purple: '#7c3aed',
  muted: '#a8a29e',
  palette: ['#3d7a54', '#c4784a', '#2563eb', '#7c3aed', '#dc4c4c', '#0891b2', '#ca8a04', '#db2777'],
};

function destroyChart(id) {
  if (chartInstances[id]) {
    chartInstances[id].destroy();
    delete chartInstances[id];
  }
}

function renderCharts(data, prefix = '') {
  if (typeof Chart === 'undefined' || !data) return;

  const plans = data.plansByStatus || [];
  renderBarChart(`${prefix}chart-plans`, plans.map((p) => statusLabel(p.label)), plans.map((p) => p.value), CHART_COLORS.palette);

  const goals = data.goalsProgress || [];
  renderBarChart(`${prefix}chart-goals`, goals.map((g) => truncate(g.label, 16)), goals.map((g) => g.value), [CHART_COLORS.accent]);

  const monthly = data.financeMonthly || [];
  renderFinanceChart(`${prefix}chart-finance`, monthly);
  renderFinanceChart(`${prefix}chart-finance-detail`, monthly);

  const expenses = data.expenseCategories || [];
  renderDoughnutChart(`${prefix}chart-expenses`, expenses.map((e) => financeCatLabel(e.label)), expenses.map((e) => e.value));
  renderDoughnutChart(`${prefix}chart-expenses-detail`, expenses.map((e) => financeCatLabel(e.label)), expenses.map((e) => e.value));

  const habits = data.habitsActivity || [];
  renderBarChart(`${prefix}chart-habits`, habits.map((h) => truncate(h.label, 12)), habits.map((h) => h.value), [CHART_COLORS.warm]);
}

function renderBarChart(canvasId, labels, values, colors) {
  const canvas = document.getElementById(canvasId);
  if (!canvas) return;
  destroyChart(canvasId);

  chartInstances[canvasId] = new Chart(canvas, {
    type: 'bar',
    data: {
      labels,
      datasets: [{
        data: values,
        backgroundColor: colors.length === 1 ? colors[0] : colors,
        borderRadius: 6,
      }],
    },
    options: chartOptions(false),
  });
}

function renderDoughnutChart(canvasId, labels, values) {
  const canvas = document.getElementById(canvasId);
  if (!canvas || !values.length) return;
  destroyChart(canvasId);

  chartInstances[canvasId] = new Chart(canvas, {
    type: 'doughnut',
    data: {
      labels,
      datasets: [{
        data: values,
        backgroundColor: CHART_COLORS.palette,
        borderWidth: 0,
      }],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: { legend: { position: 'bottom', labels: { boxWidth: 12, font: { size: 11 } } } },
    },
  });
}

function renderFinanceChart(canvasId, monthly) {
  const canvas = document.getElementById(canvasId);
  if (!canvas) return;
  destroyChart(canvasId);

  const labels = monthly.map((m) => formatMonth(m.month));
  chartInstances[canvasId] = new Chart(canvas, {
    type: 'line',
    data: {
      labels,
      datasets: [
        {
          label: 'Prihodki',
          data: monthly.map((m) => m.income),
          borderColor: CHART_COLORS.accent,
          backgroundColor: 'rgba(61,122,84,0.1)',
          fill: true,
          tension: 0.35,
        },
        {
          label: 'Odhodki',
          data: monthly.map((m) => m.expense),
          borderColor: CHART_COLORS.danger,
          backgroundColor: 'rgba(220,76,76,0.08)',
          fill: true,
          tension: 0.35,
        },
      ],
    },
    options: chartOptions(true),
  });
}

function chartOptions(showLegend) {
  return {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: { display: showLegend, position: 'bottom' },
    },
    scales: {
      y: { beginAtZero: true, grid: { color: '#e7e2d9' } },
      x: { grid: { display: false } },
    },
  };
}

function statusLabel(s) {
  return { todo: 'Za narediti', in_progress: 'V teku', done: 'Končano' }[s] || s;
}

function financeCatLabel(c) {
  return {
    salary: 'Plača', freelance: 'Freelance', investment: 'Naložbe', food: 'Hrana',
    transport: 'Prevoz', housing: 'Stanovanje', entertainment: 'Zabava', health: 'Zdravje',
    shopping: 'Nakupi', other: 'Ostalo',
  }[c] || c;
}

function formatMonth(m) {
  if (!m) return '';
  const [y, mo] = m.split('-');
  const d = new Date(Number(y), Number(mo) - 1, 1);
  return d.toLocaleDateString('sl-SI', { month: 'short', year: '2-digit' });
}

function truncate(str, len) {
  if (!str || str.length <= len) return str || '';
  return str.slice(0, len - 1) + '…';
}

function formatEUR(amount) {
  return new Intl.NumberFormat('sl-SI', { style: 'currency', currency: 'EUR' }).format(amount || 0);
}
