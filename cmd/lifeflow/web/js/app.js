const API = '/api';

const LABELS = {
  category: {
    work: 'Delo',
    health: 'Zdravje',
    personal: 'Osebno',
    learning: 'Učenje',
    other: 'Ostalo',
  },
  priority: {
    low: 'Nizka',
    medium: 'Srednja',
    high: 'Visoka',
  },
  status: {
    todo: 'Za narediti',
    in_progress: 'V teku',
    done: 'Končano',
  },
  goalStatus: {
    active: 'Aktiven',
    completed: 'Dokončan',
    paused: 'Pavza',
  },
};

const VIEW_META = {
  overview: { title: 'Pregled', subtitle: 'Tvoj dnevni povzetek produktivnosti', add: false },
  plans: { title: 'Načrti', subtitle: 'Upravljaj svoje naloge in projekte', add: true, addLabel: 'Dodaj načrt' },
  goals: { title: 'Cilji', subtitle: 'Dolgoročni cilji in napredek', add: true, addLabel: 'Dodaj cilj' },
  habits: { title: 'Navade', subtitle: 'Dnevne rutine za boljše življenje', add: true, addLabel: 'Dodaj navado' },
};

let currentView = 'overview';
let plans = [];
let goals = [];
let habits = [];
let overview = null;
let editingId = null;

const $ = (sel) => document.querySelector(sel);
const $$ = (sel) => document.querySelectorAll(sel);

document.addEventListener('DOMContentLoaded', init);

async function init() {
  setTodayDate();
  bindNavigation();
  bindModal();
  bindFilters();
  await refreshAll();
}

function setTodayDate() {
  const el = $('#today-date');
  if (!el) return;
  el.textContent = new Date().toLocaleDateString('sl-SI', {
    weekday: 'long',
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  });
}

function bindNavigation() {
  $$('.nav-item').forEach((btn) => {
    btn.addEventListener('click', () => switchView(btn.dataset.view));
  });

  $('#add-btn').addEventListener('click', () => openModal(currentView));
}

function bindModal() {
  $('#modal-close').addEventListener('click', closeModal);
  $('#modal-cancel').addEventListener('click', closeModal);
  $('#modal-overlay').addEventListener('click', (e) => {
    if (e.target === $('#modal-overlay')) closeModal();
  });

  $('#modal-form').addEventListener('submit', handleFormSubmit);
}

function bindFilters() {
  ['plan-search', 'plan-filter-status', 'plan-filter-category'].forEach((id) => {
    const el = document.getElementById(id);
    if (el) el.addEventListener('input', renderPlans);
    if (el) el.addEventListener('change', renderPlans);
  });
}

function switchView(view) {
  currentView = view;
  $$('.nav-item').forEach((b) => b.classList.toggle('active', b.dataset.view === view));
  $$('.view').forEach((v) => v.classList.toggle('active', v.id === `view-${view}`));

  const meta = VIEW_META[view];
  $('#view-title').textContent = meta.title;
  $('#view-subtitle').textContent = meta.subtitle;

  const addBtn = $('#add-btn');
  if (meta.add) {
    addBtn.hidden = false;
    addBtn.innerHTML = `<span>+</span> ${meta.addLabel}`;
  } else {
    addBtn.hidden = true;
  }
}

async function refreshAll() {
  try {
    [overview, plans, goals, habits] = await Promise.all([
      fetchJSON(`${API}/overview`),
      fetchJSON(`${API}/plans`),
      fetchJSON(`${API}/goals`),
      fetchJSON(`${API}/habits`),
    ]);
    renderOverview();
    renderPlans();
    renderGoals();
    renderHabits();
  } catch (err) {
    toast('Napaka pri nalaganju podatkov', true);
    console.error(err);
  }
}

async function fetchJSON(url, opts = {}) {
  const res = await fetch(url, {
    headers: { 'Content-Type': 'application/json' },
    ...opts,
  });
  if (res.status === 204) return null;
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || 'Request failed');
  return data;
}

function renderOverview() {
  if (!overview) return;

  $('#stats-grid').innerHTML = `
    <div class="stat-card">
      <div class="label">Skupaj načrtov</div>
      <div class="value accent">${overview.totalPlans}</div>
    </div>
    <div class="stat-card">
      <div class="label">V teku</div>
      <div class="value">${overview.inProgressPlans}</div>
    </div>
    <div class="stat-card">
      <div class="label">Končano</div>
      <div class="value accent">${overview.completedPlans}</div>
    </div>
    <div class="stat-card">
      <div class="label">Z zamudo</div>
      <div class="value danger">${overview.overduePlans}</div>
    </div>
    <div class="stat-card">
      <div class="label">Aktivni cilji</div>
      <div class="value warm">${overview.activeGoals}</div>
    </div>
    <div class="stat-card">
      <div class="label">Povp. napredek</div>
      <div class="value">${Math.round(overview.avgGoalProgress)}%</div>
    </div>
    <div class="stat-card">
      <div class="label">Navade</div>
      <div class="value">${overview.totalHabits}</div>
    </div>
    <div class="stat-card">
      <div class="label">Skupni streak</div>
      <div class="value warm">${overview.totalStreak}🔥</div>
    </div>
  `;

  renderMiniList('#recent-plans', overview.recentPlans, '📋', 'Ni načrtov še');
  renderMiniList('#upcoming-plans', overview.upcomingPlans, '📅', 'Ni prihajajočih rokov');

  const cats = overview.plansByCategory || {};
  const max = Math.max(...Object.values(cats), 1);
  const catEl = $('#category-bars');

  const entries = Object.entries(cats);
  if (entries.length === 0) {
    catEl.innerHTML = '<div class="empty-state"><span>📊</span>Še ni podatkov</div>';
  } else {
    catEl.innerHTML = entries.map(([key, count]) => `
      <div class="cat-bar-row">
        <span class="cat-bar-label">${LABELS.category[key] || key}</span>
        <div class="cat-bar-track">
          <div class="cat-bar-fill" style="width: ${(count / max) * 100}%"></div>
        </div>
        <span class="cat-bar-count">${count}</span>
      </div>
    `).join('');
  }
}

function renderMiniList(selector, items, icon, emptyText) {
  const el = $(selector);
  if (!items || items.length === 0) {
    el.innerHTML = `<div class="empty-state"><span>${icon}</span>${emptyText}</div>`;
    return;
  }

  el.innerHTML = items.map((p) => {
    const color = statusColor(p.status);
    return `
      <div class="mini-item">
        <span class="dot" style="background:${color}"></span>
        <div class="info">
          <div class="title">${esc(p.title)}</div>
          <div class="meta">${LABELS.category[p.category] || p.category}${p.dueDate ? ' · ' + formatDate(p.dueDate) : ''}</div>
        </div>
      </div>
    `;
  }).join('');
}

function renderPlans() {
  const search = ($('#plan-search')?.value || '').toLowerCase();
  const statusFilter = $('#plan-filter-status')?.value || '';
  const catFilter = $('#plan-filter-category')?.value || '';
  const today = todayStr();

  const filtered = plans.filter((p) => {
    if (search && !p.title.toLowerCase().includes(search) && !(p.description || '').toLowerCase().includes(search)) return false;
    if (statusFilter && p.status !== statusFilter) return false;
    if (catFilter && p.category !== catFilter) return false;
    return true;
  });

  const el = $('#plans-list');
  if (filtered.length === 0) {
    el.innerHTML = '<div class="empty-state" style="grid-column:1/-1"><span>📋</span>Ni načrtov. Dodaj prvega!</div>';
    return;
  }

  el.innerHTML = filtered.map((p) => {
    const overdue = p.dueDate && p.dueDate < today && p.status !== 'done';
    return `
      <div class="plan-card" data-id="${p.id}">
        <div class="plan-card-header">
          <h4>${esc(p.title)}</h4>
          <span class="badge badge-${p.status}">${LABELS.status[p.status]}</span>
        </div>
        ${p.description ? `<p class="plan-desc">${esc(p.description)}</p>` : ''}
        <div class="plan-meta">
          <span class="badge badge-cat-${p.category}">${LABELS.category[p.category]}</span>
          <span class="badge badge-priority-${p.priority}">${LABELS.priority[p.priority]}</span>
          ${p.dueDate ? `<span class="badge ${overdue ? 'badge-overdue' : ''}">${overdue ? '⚠ ' : ''}${formatDate(p.dueDate)}</span>` : ''}
        </div>
        <div class="plan-actions">
          ${p.status !== 'done' ? `<button class="btn btn-sm btn-primary" onclick="cycleStatus('${p.id}')">${nextStatusLabel(p.status)}</button>` : ''}
          <button class="btn btn-sm btn-ghost" onclick="editPlan('${p.id}')">Uredi</button>
          <button class="btn btn-sm btn-danger" onclick="deletePlan('${p.id}')">Izbriši</button>
        </div>
      </div>
    `;
  }).join('');
}

function renderGoals() {
  const el = $('#goals-list');
  if (goals.length === 0) {
    el.innerHTML = '<div class="empty-state" style="grid-column:1/-1"><span>🎯</span>Ni ciljev. Postavi si prvega!</div>';
    return;
  }

  el.innerHTML = goals.map((g) => `
    <div class="goal-card" data-id="${g.id}">
      <h4>${esc(g.title)}</h4>
      ${g.description ? `<p class="goal-desc">${esc(g.description)}</p>` : ''}
      <div class="progress-label">
        <span>Napredek</span>
        <span>${g.progress}%</span>
      </div>
      <div class="progress-bar">
        <div class="progress-fill" style="width:${g.progress}%"></div>
      </div>
      <div class="plan-meta">
        <span class="badge badge-${g.status === 'completed' ? 'done' : 'in_progress'}">${LABELS.goalStatus[g.status]}</span>
        ${g.targetDate ? `<span class="badge">${formatDate(g.targetDate)}</span>` : ''}
      </div>
      <div class="goal-actions">
        <button class="btn btn-sm btn-primary" onclick="editGoal('${g.id}')">Uredi</button>
        <button class="btn btn-sm btn-danger" onclick="deleteGoal('${g.id}')">Izbriši</button>
      </div>
    </div>
  `).join('');
}

function renderHabits() {
  const el = $('#habits-list');
  const today = todayStr();

  if (habits.length === 0) {
    el.innerHTML = '<div class="empty-state" style="grid-column:1/-1"><span>🔥</span>Ni navad. Začni z novo rutino!</div>';
    return;
  }

  el.innerHTML = habits.map((h) => {
    const doneToday = h.lastDone === today;
    return `
      <div class="habit-card" data-id="${h.id}">
        <div class="habit-icon">${h.icon || '✨'}</div>
        <h4>${esc(h.name)}</h4>
        <div class="habit-streak">${h.streak}</div>
        <div class="habit-streak-label">${h.streak === 1 ? 'dan zapored' : 'dni zapored'}</div>
        <button class="btn btn-primary habit-done-btn ${doneToday ? 'done-today' : ''}"
          onclick="markHabitDone('${h.id}')" ${doneToday ? 'disabled' : ''}>
          ${doneToday ? '✓ Danes opravljeno' : 'Označi kot narejeno'}
        </button>
        <div class="habit-actions">
          <button class="btn btn-sm btn-danger" onclick="deleteHabit('${h.id}')">Izbriši</button>
        </div>
      </div>
    `;
  }).join('');
}

function openModal(type, id = null) {
  editingId = id;
  const form = $('#modal-form');
  form.innerHTML = '';

  if (type === 'plans') {
    $('#modal-title').textContent = id ? 'Uredi načrt' : 'Nov načrt';
    const plan = id ? plans.find((p) => p.id === id) : null;
    form.innerHTML = planFormHTML(plan);
  } else if (type === 'goals') {
    $('#modal-title').textContent = id ? 'Uredi cilj' : 'Nov cilj';
    const goal = id ? goals.find((g) => g.id === id) : null;
    form.innerHTML = goalFormHTML(goal);
  } else if (type === 'habits') {
    $('#modal-title').textContent = 'Nova navada';
    form.innerHTML = habitFormHTML();
  }

  $('#modal-overlay').hidden = false;
}

function closeModal() {
  $('#modal-overlay').hidden = true;
  editingId = null;
}

function planFormHTML(plan) {
  return `
    <div class="form-group">
      <label for="f-title">Naslov *</label>
      <input class="form-input" id="f-title" required value="${esc(plan?.title || '')}">
    </div>
    <div class="form-group">
      <label for="f-desc">Opis</label>
      <textarea class="form-textarea" id="f-desc">${esc(plan?.description || '')}</textarea>
    </div>
    <div class="form-row">
      <div class="form-group">
        <label for="f-category">Kategorija</label>
        <select class="form-select" id="f-category">
          ${selectOptions(LABELS.category, plan?.category || 'personal')}
        </select>
      </div>
      <div class="form-group">
        <label for="f-priority">Prioriteta</label>
        <select class="form-select" id="f-priority">
          ${selectOptions(LABELS.priority, plan?.priority || 'medium')}
        </select>
      </div>
    </div>
    <div class="form-row">
      <div class="form-group">
        <label for="f-status">Status</label>
        <select class="form-select" id="f-status">
          ${selectOptions(LABELS.status, plan?.status || 'todo')}
        </select>
      </div>
      <div class="form-group">
        <label for="f-due">Rok</label>
        <input class="form-input" type="date" id="f-due" value="${plan?.dueDate || ''}">
      </div>
    </div>
  `;
}

function goalFormHTML(goal) {
  return `
    <div class="form-group">
      <label for="f-title">Naslov *</label>
      <input class="form-input" id="f-title" required value="${esc(goal?.title || '')}">
    </div>
    <div class="form-group">
      <label for="f-desc">Opis</label>
      <textarea class="form-textarea" id="f-desc">${esc(goal?.description || '')}</textarea>
    </div>
    <div class="form-row">
      <div class="form-group">
        <label for="f-progress">Napredek (%)</label>
        <input class="form-input" type="number" id="f-progress" min="0" max="100" value="${goal?.progress ?? 0}">
      </div>
      <div class="form-group">
        <label for="f-target">Ciljni datum</label>
        <input class="form-input" type="date" id="f-target" value="${goal?.targetDate || ''}">
      </div>
    </div>
    <div class="form-group">
      <label for="f-gstatus">Status</label>
      <select class="form-select" id="f-gstatus">
        ${selectOptions(LABELS.goalStatus, goal?.status || 'active')}
      </select>
    </div>
  `;
}

function habitFormHTML() {
  const icons = ['✨', '🏃', '📚', '💧', '🧘', '💪', '🥗', '😴', '📝', '🎨'];
  return `
    <div class="form-group">
      <label for="f-name">Ime navade *</label>
      <input class="form-input" id="f-name" required placeholder="npr. Jutranja meditacija">
    </div>
    <div class="form-group">
      <label for="f-icon">Ikona</label>
      <select class="form-select" id="f-icon">
        ${icons.map((i) => `<option value="${i}">${i}</option>`).join('')}
      </select>
    </div>
  `;
}

function selectOptions(labels, selected) {
  return Object.entries(labels).map(([k, v]) =>
    `<option value="${k}" ${k === selected ? 'selected' : ''}>${v}</option>`
  ).join('');
}

async function handleFormSubmit(e) {
  e.preventDefault();

  try {
    if (currentView === 'plans') {
      const body = {
        title: $('#f-title').value,
        description: $('#f-desc').value,
        category: $('#f-category').value,
        priority: $('#f-priority').value,
        status: $('#f-status').value,
        dueDate: $('#f-due').value,
      };

      if (editingId) {
        await fetchJSON(`${API}/plans/${editingId}`, { method: 'PATCH', body: JSON.stringify(body) });
        toast('Načrt posodobljen');
      } else {
        await fetchJSON(`${API}/plans`, { method: 'POST', body: JSON.stringify(body) });
        toast('Načrt dodan');
      }
    } else if (currentView === 'goals') {
      const body = {
        title: $('#f-title').value,
        description: $('#f-desc').value,
        progress: parseInt($('#f-progress').value, 10) || 0,
        targetDate: $('#f-target').value,
        status: $('#f-gstatus').value,
      };

      if (editingId) {
        await fetchJSON(`${API}/goals/${editingId}`, { method: 'PATCH', body: JSON.stringify(body) });
        toast('Cilj posodobljen');
      } else {
        await fetchJSON(`${API}/goals`, { method: 'POST', body: JSON.stringify(body) });
        toast('Cilj dodan');
      }
    } else if (currentView === 'habits') {
      const body = {
        name: $('#f-name').value,
        icon: $('#f-icon').value,
      };
      await fetchJSON(`${API}/habits`, { method: 'POST', body: JSON.stringify(body) });
      toast('Navada dodana');
    }

    closeModal();
    await refreshAll();
  } catch (err) {
    toast(err.message || 'Napaka', true);
  }
}

window.editPlan = (id) => openModal('plans', id);
window.editGoal = (id) => openModal('goals', id);

window.cycleStatus = async (id) => {
  const plan = plans.find((p) => p.id === id);
  if (!plan) return;
  const next = plan.status === 'todo' ? 'in_progress' : plan.status === 'in_progress' ? 'done' : 'done';
  try {
    await fetchJSON(`${API}/plans/${id}`, { method: 'PATCH', body: JSON.stringify({ status: next }) });
    toast('Status posodobljen');
    await refreshAll();
  } catch (err) {
    toast(err.message, true);
  }
};

window.deletePlan = async (id) => {
  if (!confirm('Res želiš izbrisati ta načrt?')) return;
  try {
    await fetchJSON(`${API}/plans/${id}`, { method: 'DELETE' });
    toast('Načrt izbrisan');
    await refreshAll();
  } catch (err) {
    toast(err.message, true);
  }
};

window.deleteGoal = async (id) => {
  if (!confirm('Res želiš izbrisati ta cilj?')) return;
  try {
    await fetchJSON(`${API}/goals/${id}`, { method: 'DELETE' });
    toast('Cilj izbrisan');
    await refreshAll();
  } catch (err) {
    toast(err.message, true);
  }
};

window.deleteHabit = async (id) => {
  if (!confirm('Res želiš izbrisati to navado?')) return;
  try {
    await fetchJSON(`${API}/habits/${id}`, { method: 'DELETE' });
    toast('Navada izbrisana');
    await refreshAll();
  } catch (err) {
    toast(err.message, true);
  }
};

window.markHabitDone = async (id) => {
  try {
    await fetchJSON(`${API}/habits/${id}/done`, { method: 'POST' });
    toast('Odlično! 🔥');
    await refreshAll();
  } catch (err) {
    toast(err.message, true);
  }
};

function nextStatusLabel(status) {
  if (status === 'todo') return 'Začni';
  if (status === 'in_progress') return 'Končaj';
  return '';
}

function statusColor(status) {
  if (status === 'done') return '#3d7a54';
  if (status === 'in_progress') return '#2563eb';
  return '#a8a29e';
}

function formatDate(d) {
  if (!d) return '';
  return new Date(d + 'T00:00:00').toLocaleDateString('sl-SI', { day: 'numeric', month: 'short' });
}

function todayStr() {
  return new Date().toISOString().slice(0, 10);
}

function esc(str) {
  const d = document.createElement('div');
  d.textContent = str || '';
  return d.innerHTML;
}

function toast(msg, isError = false) {
  const container = $('#toast-container');
  const el = document.createElement('div');
  el.className = 'toast';
  el.textContent = msg;
  if (isError) el.style.background = '#dc4c4c';
  container.appendChild(el);
  setTimeout(() => el.remove(), 3000);
}
