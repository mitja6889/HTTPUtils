const API = '/api';

const LABELS = {
  category: { work: 'Delo', health: 'Zdravje', personal: 'Osebno', learning: 'Učenje', other: 'Ostalo' },
  priority: { low: 'Nizka', medium: 'Srednja', high: 'Visoka' },
  status: { todo: 'Za narediti', in_progress: 'V teku', done: 'Končano' },
  goalStatus: { active: 'Aktiven', completed: 'Dokončan', paused: 'Pavza' },
  goalKind: { manual: 'Ročni', financial: 'Finančni' },
  txType: { income: 'Prihodek', expense: 'Odhodek' },
  financeCat: {
    salary: 'Plača', freelance: 'Freelance', investment: 'Naložbe', food: 'Hrana',
    transport: 'Prevoz', housing: 'Stanovanje', entertainment: 'Zabava', health: 'Zdravje',
    shopping: 'Nakupi', other: 'Ostalo',
  },
};

const VIEW_META = {
  overview: { title: 'Pregled', subtitle: 'Tvoj dnevni povzetek produktivnosti', add: false },
  plans: { title: 'Načrti', subtitle: 'Upravljaj svoje naloge in projekte', add: true, addLabel: 'Dodaj načrt' },
  finance: { title: 'Finance', subtitle: 'Stroški, zaslužki in stanje', add: true, addLabel: 'Dodaj transakcijo' },
  goals: { title: 'Cilji', subtitle: 'Dolgoročni cilji z dinamičnim napredkom', add: true, addLabel: 'Dodaj cilj' },
  habits: { title: 'Navade', subtitle: 'Dnevne rutine za boljše življenje', add: true, addLabel: 'Dodaj navado' },
};

const HABIT_ICONS = ['✨', '🏃', '📚', '💧', '🧘', '💪', '🥗', '😴', '📝', '🎨', '🌿', '🔥', '⭐', '🎯', '☀️', '🌙', '🍎', '🚶', '🧠', '❤️'];

let currentView = 'overview';
let plans = [];
let goals = [];
let habits = [];
let transactions = [];
let overview = null;
let editingId = null;
let selectedWeekDay = localDateStr();
let planViewMode = 'day';
let modalTasks = [];
let modalType = null;
let loadError = null;

const $ = (sel) => document.querySelector(sel);
const $$ = (sel) => document.querySelectorAll(sel);

document.addEventListener('DOMContentLoaded', init);

async function init() {
  setTodayDate();
  bindNavigation();
  bindModal();
  bindFilters();
  bindDelegatedActions();
  await refreshAll();
}

function bindNavigation() {
  $$('.nav-item, .bottom-nav-item').forEach((btn) => {
    btn.addEventListener('click', () => {
      if (btn.dataset.view === 'more') {
        openMoreSheet();
        return;
      }
      closeMoreSheet();
      switchView(btn.dataset.view);
    });
  });

  $$('.more-sheet-item').forEach((btn) => {
    btn.addEventListener('click', () => {
      closeMoreSheet();
      switchView(btn.dataset.view);
    });
  });

  document.querySelector('[data-action="close-more"]')?.addEventListener('click', closeMoreSheet);

  $('#add-btn').addEventListener('click', () => openModal(currentView));
  $('#fab-add').addEventListener('click', () => openModal(currentView));
  $('#mobile-add-btn').addEventListener('click', () => openModal(currentView));

  $('#charts-toggle')?.addEventListener('click', () => {
    const section = $('#charts-section');
    const btn = $('#charts-toggle');
    const open = section.hidden;
    section.hidden = !open;
    btn.setAttribute('aria-expanded', open ? 'true' : 'false');
    if (open && overview?.charts) {
      requestAnimationFrame(() => renderCharts(overview.charts));
    }
  });
}

function openMoreSheet() {
  $('#more-sheet').hidden = false;
  $$('.bottom-nav-item').forEach((b) => b.classList.remove('active'));
  $('#more-nav-btn')?.classList.add('active');
}

function closeMoreSheet() {
  const sheet = $('#more-sheet');
  if (sheet) sheet.hidden = true;
}

function bindModal() {
  $('#modal-close').addEventListener('click', closeModal);
  $('#modal-cancel').addEventListener('click', closeModal);
  $('#modal-overlay').addEventListener('click', (e) => {
    if (e.target === $('#modal-overlay')) closeModal();
  });
  $('#modal-form').addEventListener('submit', handleFormSubmit);
  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape' && !$('#modal-overlay').hidden) closeModal();
  });
}

function bindFilters() {
  ['plan-search', 'plan-filter-status', 'plan-filter-category', 'plan-filter-view'].forEach((id) => {
    const el = document.getElementById(id);
    if (!el) return;
    el.addEventListener('input', onPlanFiltersChanged);
    el.addEventListener('change', onPlanFiltersChanged);
  });
  const financeFilter = document.getElementById('finance-filter-type');
  if (financeFilter) financeFilter.addEventListener('change', renderTransactions);
}

function onPlanFiltersChanged() {
  planViewMode = $('#plan-filter-view')?.value || 'day';
  const weekNav = $('#week-nav');
  if (weekNav) weekNav.hidden = planViewMode !== 'day';
  renderPlans();
}

function bindDelegatedActions() {
  document.body.addEventListener('click', async (e) => {
    const btn = e.target.closest('[data-action]');
    if (!btn) return;

    const { action, id } = btn.dataset;
    switch (action) {
      case 'cycle-status': await cycleStatus(id); break;
      case 'edit-plan': openModal('plans', id); break;
      case 'delete-plan': await deletePlan(id); break;
      case 'edit-goal': openModal('goals', id); break;
      case 'delete-goal': await deleteGoal(id); break;
      case 'edit-habit': openModal('habits', id); break;
      case 'delete-habit': await deleteHabit(id); break;
      case 'habit-done': await markHabitDone(id); break;
      case 'habit-undo': await undoHabitDone(id); break;
      case 'delete-transaction': await deleteTransaction(id); break;
      case 'edit-transaction': openModal('finance', id); break;
      case 'toggle-task': await toggleTask(id, btn.dataset.taskId); break;
      case 'select-day': selectWeekDay(btn.dataset.date); break;
      case 'retry-load': await refreshAll(); break;
      case 'add-task-row': addTaskRow(); break;
      case 'remove-task-row': btn.closest('.task-editor-row')?.remove(); break;
    }
  });
}

function switchView(view) {
  if (!VIEW_META[view]) return;

  currentView = view;
  closeMoreSheet();

  $$('.nav-item').forEach((b) => b.classList.toggle('active', b.dataset.view === view));
  $$('.bottom-nav-item').forEach((b) => {
    if (b.id === 'more-nav-btn') {
      b.classList.remove('active');
    } else {
      b.classList.toggle('active', b.dataset.view === view);
    }
  });

  $$('.view').forEach((v) => v.classList.toggle('active', v.id === `view-${view}`));

  const meta = VIEW_META[view];
  $('#view-title').textContent = meta.title;
  $('#view-subtitle').textContent = meta.subtitle;

  const mobileTitle = $('#mobile-title');
  if (mobileTitle) mobileTitle.textContent = meta.title;

  const showAdd = meta.add;
  $('#add-btn').hidden = !showAdd;
  $('#fab-add').hidden = !showAdd;
  $('#mobile-add-btn').hidden = !showAdd;

  if (showAdd) {
    $('#add-btn').innerHTML = `<span>+</span> ${meta.addLabel}`;
  }

  window.scrollTo({ top: 0, behavior: 'smooth' });
}

async function refreshAll() {
  setLoading(true);
  loadError = null;
  try {
    const today = localDateStr();
    [overview, plans, goals, habits, transactions] = await Promise.all([
      fetchJSON(`${API}/overview?today=${today}`),
      fetchJSON(`${API}/plans`),
      fetchJSON(`${API}/goals`),
      fetchJSON(`${API}/habits?today=${today}`),
      fetchJSON(`${API}/transactions`),
    ]);
    renderAll();
  } catch (err) {
    loadError = err.message || 'Napaka pri nalaganju';
    toast(loadError, true);
    renderErrorBanner();
  } finally {
    setLoading(false);
  }
}

function renderAll() {
  renderErrorBanner();
  renderOverview();
  renderWeekNav();
  renderPlans();
  renderGoals();
  renderHabits();
  renderFinance();
}

function renderErrorBanner() {
  const main = $('.main');
  let banner = $('#error-banner');
  if (!loadError) {
    banner?.remove();
    return;
  }
  if (!banner) {
    banner = document.createElement('div');
    banner.id = 'error-banner';
    banner.className = 'error-banner';
    main.insertBefore(banner, main.firstChild.nextSibling);
  }
  banner.innerHTML = `
    <span>${esc(loadError)}</span>
    <button class="btn btn-sm btn-primary" data-action="retry-load">Poskusi znova</button>
  `;
}

function setLoading(on) {
  const el = $('#loading');
  if (el) el.hidden = !on;
}

async function fetchJSON(url, opts = {}) {
  const res = await fetch(url, {
    headers: { 'Content-Type': 'application/json' },
    ...opts,
  });
  if (res.status === 204) return null;

  const contentType = res.headers.get('content-type') || '';
  let data = null;
  if (contentType.includes('application/json')) {
    data = await res.json();
  } else {
    const text = await res.text();
    throw new Error(text || `HTTP ${res.status}`);
  }

  if (!res.ok) throw new Error(data?.error || `HTTP ${res.status}`);
  return data;
}

function setTodayDate() {
  const now = new Date();
  const formatted = now.toLocaleDateString('sl-SI', {
    weekday: 'long', day: 'numeric', month: 'long', year: 'numeric',
  });
  const el = $('#today-date');
  if (el) el.textContent = formatted;
  const mobile = $('#mobile-date');
  if (mobile) {
    mobile.textContent = now.toLocaleDateString('sl-SI', { weekday: 'short', day: 'numeric', month: 'short' });
  }
}

function renderOverview() {
  if (!overview) return;

  $('#stats-grid').innerHTML = `
    <div class="stat-card"><div class="label">Načrti</div><div class="value accent">${overview.totalPlans}</div></div>
    <div class="stat-card"><div class="label">V teku</div><div class="value">${overview.inProgressPlans}</div></div>
    <div class="stat-card stat-extra"><div class="label">Končano</div><div class="value accent">${overview.completedPlans}</div></div>
    <div class="stat-card"><div class="label">Z zamudo</div><div class="value danger">${overview.overduePlans}</div></div>
    <div class="stat-card stat-extra"><div class="label">Aktivni cilji</div><div class="value warm">${overview.activeGoals}</div></div>
    <div class="stat-card stat-extra"><div class="label">Povp. napredek</div><div class="value">${Math.round(overview.avgGoalProgress)}%</div></div>
    <div class="stat-card stat-extra"><div class="label">Navade</div><div class="value">${overview.totalHabits}</div></div>
    <div class="stat-card"><div class="label">Stanje</div><div class="value ${overview.finance?.balance >= 0 ? 'accent' : 'danger'}">${formatEUR(overview.finance?.balance || 0)}</div></div>
    <div class="stat-card stat-extra"><div class="label">Streak</div><div class="value warm">${overview.totalStreak}🔥</div></div>
  `;

  renderFinanceSummary('#finance-summary', overview.finance);

  if (overview.charts && !$('#charts-section')?.hidden) {
    renderCharts(overview.charts);
  }

  renderMiniList('#recent-plans', overview.recentPlans, '📋', 'Ni načrtov še');
  renderMiniList('#upcoming-plans', overview.upcomingPlans, '📅', 'Ni prihajajočih rokov');
  renderMiniList('#overdue-plans', overview.overduePlansList, '⚠️', 'Ni zamujenih načrtov');

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
        <div class="cat-bar-track"><div class="cat-bar-fill" style="width:${(count / max) * 100}%"></div></div>
        <span class="cat-bar-count">${count}</span>
      </div>
    `).join('');
  }
}

function renderMiniList(selector, items, icon, emptyText) {
  const el = $(selector);
  if (!items?.length) {
    el.innerHTML = `<div class="empty-state"><span>${icon}</span>${emptyText}</div>`;
    return;
  }
  el.innerHTML = items.map((p) => `
    <div class="mini-item">
      <span class="dot" style="background:${statusColor(p.status)}"></span>
      <div class="info">
        <div class="title">${esc(p.title)}</div>
        <div class="meta">${LABELS.category[p.category] || p.category}${p.dueDate ? ' · ' + formatDate(p.dueDate) : ''}</div>
      </div>
    </div>
  `).join('');
}

function renderWeekNav() {
  const el = $('#week-nav');
  if (!el) return;

  const today = localDateStr();
  const overdueCount = plans.filter((p) => isOverdue(p, today)).length;

  let html = '';
  if (overdueCount > 0) {
    html += `
      <button type="button" class="week-day overdue-tab ${selectedWeekDay === 'overdue' ? 'active' : ''}"
        data-action="select-day" data-date="overdue">
        <div class="week-day-name">Zamuda</div>
        <div class="week-day-num">⚠</div>
        <div class="week-day-count">${overdueCount}</div>
      </button>`;
  }

  for (let i = 0; i < 7; i++) {
    const dateStr = addDays(today, i);
    const d = parseLocalDate(dateStr);
    const count = plans.filter((p) => p.dueDate === dateStr).length;
    html += `
      <button type="button" class="week-day ${selectedWeekDay === dateStr ? 'active' : ''} ${dateStr === today ? 'today' : ''}"
        data-action="select-day" data-date="${dateStr}">
        <div class="week-day-name">${d.toLocaleDateString('sl-SI', { weekday: 'short' })}</div>
        <div class="week-day-num">${d.getDate()}</div>
        <div class="week-day-count">${count} ${count === 1 ? 'načrt' : 'načrti'}</div>
      </button>`;
  }

  el.innerHTML = html;
  el.hidden = planViewMode !== 'day';
}

function selectWeekDay(dateStr) {
  selectedWeekDay = dateStr;
  planViewMode = dateStr === 'overdue' ? 'overdue' : 'day';
  const viewFilter = $('#plan-filter-view');
  if (viewFilter) viewFilter.value = planViewMode === 'overdue' ? 'overdue' : 'day';
  renderWeekNav();
  renderPlans();
}

function renderPlans() {
  const search = ($('#plan-search')?.value || '').toLowerCase();
  const statusFilter = $('#plan-filter-status')?.value || '';
  const catFilter = $('#plan-filter-category')?.value || '';
  const viewMode = $('#plan-filter-view')?.value || planViewMode;
  const today = localDateStr();

  const filtered = plans.filter((p) => {
    if (search && !matchesSearch(p, search)) return false;
    if (statusFilter && p.status !== statusFilter) return false;
    if (catFilter && p.category !== catFilter) return false;

    if (viewMode === 'all') return true;
    if (viewMode === 'overdue') return isOverdue(p, today);
    if (viewMode === 'day') {
      if (selectedWeekDay === 'overdue') return isOverdue(p, today);
      if (p.dueDate) return p.dueDate === selectedWeekDay;
      return selectedWeekDay === today;
    }
    return true;
  });

  const el = $('#plans-list');
  if (!filtered.length) {
    el.innerHTML = `<div class="empty-state"><span>📋</span>${emptyPlansMessage(viewMode)}</div>`;
    return;
  }

  el.innerHTML = groupPlansByDay(filtered).map(({ label, items }) => `
    <div class="day-group">
      <div class="day-group-header">
        <h3>${label}</h3>
        <span class="day-label">${items.length} ${items.length === 1 ? 'načrt' : 'načrti'}</span>
      </div>
      <div class="day-group-plans">${items.map((p) => planCardHTML(p, today)).join('')}</div>
    </div>
  `).join('');
}

function emptyPlansMessage(viewMode) {
  if (viewMode === 'overdue') return 'Ni zamujenih načrtov.';
  if (viewMode === 'all') return 'Ni načrtov. Dodaj prvega!';
  return 'Ni načrtov za ta dan.';
}

function matchesSearch(p, search) {
  return p.title.toLowerCase().includes(search) || (p.description || '').toLowerCase().includes(search);
}

function groupPlansByDay(items) {
  const map = new Map();
  for (const p of items) {
    const key = p.dueDate || 'no-date';
    if (!map.has(key)) map.set(key, []);
    map.get(key).push(p);
  }
  return [...map.entries()]
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([date, plansForDay]) => ({
      label: date === 'no-date' ? 'Brez roka' : formatDayLabel(date),
      items: plansForDay,
    }));
}

function planCardHTML(p, today) {
  const overdue = isOverdue(p, today);
  const tasks = p.tasks || [];
  const doneTasks = tasks.filter((t) => t.done).length;

  return `
    <div class="plan-card" data-id="${p.id}">
      <div class="plan-card-header">
        <h4>${esc(p.title)}</h4>
        <span class="badge badge-${p.status}">${LABELS.status[p.status]}</span>
      </div>
      ${p.description ? `<p class="plan-desc">${esc(p.description)}</p>` : ''}
      ${tasks.length ? `<div class="task-progress">${doneTasks}/${tasks.length} podnalog</div>` : ''}
      ${tasks.length ? `<div class="task-list">${tasks.map((t) => `
        <label class="task-item ${t.done ? 'done' : ''}">
          <input type="checkbox" ${t.done ? 'checked' : ''} data-action="toggle-task" data-id="${p.id}" data-task-id="${t.id}">
          <span>${esc(t.title)}</span>
        </label>`).join('')}</div>` : ''}
      <div class="plan-meta">
        <span class="badge badge-cat-${p.category}">${LABELS.category[p.category]}</span>
        <span class="badge badge-priority-${p.priority}">${LABELS.priority[p.priority]}</span>
        ${p.dueDate ? `<span class="badge ${overdue ? 'badge-overdue' : ''}">${overdue ? '⚠ ' : ''}${formatDate(p.dueDate)}</span>` : ''}
      </div>
      <div class="plan-actions">
        ${p.status !== 'done' ? `<button class="btn btn-sm btn-primary" data-action="cycle-status" data-id="${p.id}">${nextStatusLabel(p.status)}</button>` : ''}
        ${p.status === 'done' ? `<button class="btn btn-sm btn-ghost" data-action="cycle-status" data-id="${p.id}">Ponovno odpri</button>` : ''}
        <button class="btn btn-sm btn-ghost" data-action="edit-plan" data-id="${p.id}">Uredi</button>
        <button class="btn btn-sm btn-danger" data-action="delete-plan" data-id="${p.id}">Izbriši</button>
      </div>
    </div>`;
}

function renderGoals() {
  const el = $('#goals-list');
  if (!goals.length) {
    el.innerHTML = '<div class="empty-state"><span>🎯</span>Ni ciljev. Postavi si prvega!</div>';
    return;
  }

  el.innerHTML = goals.map((g) => {
    const badgeClass = g.status === 'completed' ? 'done' : g.status === 'paused' ? 'paused' : 'in_progress';
    const isFinancial = g.kind === 'financial';
    const progress = isFinancial && g.targetAmount > 0
      ? Math.min(100, Math.round((g.currentAmount / g.targetAmount) * 100))
      : g.progress;
    return `
      <div class="goal-card" data-id="${g.id}">
        <div class="plan-card-header">
          <h4>${esc(g.title)}</h4>
          ${isFinancial ? '<span class="goal-financial-tag">💰 Finančni</span>' : ''}
        </div>
        ${g.description ? `<p class="goal-desc">${esc(g.description)}</p>` : ''}
        ${isFinancial ? `<div class="progress-label"><span>Privarčevano</span><span>${formatEUR(g.currentAmount)} / ${formatEUR(g.targetAmount)}</span></div>` : ''}
        <div class="progress-label"><span>Napredek</span><span>${progress}%</span></div>
        <div class="progress-bar"><div class="progress-fill" style="width:${Math.min(100, progress)}%"></div></div>
        <div class="plan-meta">
          <span class="badge badge-${badgeClass}">${LABELS.goalStatus[g.status]}</span>
          ${g.targetDate ? `<span class="badge">${formatDate(g.targetDate)}</span>` : ''}
        </div>
        <div class="goal-actions">
          <button class="btn btn-sm btn-primary" data-action="edit-goal" data-id="${g.id}">Uredi</button>
          <button class="btn btn-sm btn-danger" data-action="delete-goal" data-id="${g.id}">Izbriši</button>
        </div>
      </div>`;
  }).join('');
}

function renderHabits() {
  const el = $('#habits-list');
  const today = localDateStr();

  if (!habits.length) {
    el.innerHTML = '<div class="empty-state"><span>🔥</span>Ni navad. Začni z novo rutino!</div>';
    return;
  }

  el.innerHTML = habits.map((h) => {
    const doneToday = h.lastDone === today;
    return `
      <div class="habit-card" data-id="${h.id}">
        <div class="habit-icon">${esc(h.icon || '✨')}</div>
        <h4>${esc(h.name)}</h4>
        <div class="habit-streak">${h.streak}</div>
        <div class="habit-streak-label">${h.streak === 1 ? 'dan zapored' : 'dni zapored'}</div>
        <button class="btn btn-primary habit-done-btn ${doneToday ? 'done-today' : ''}"
          data-action="${doneToday ? 'habit-undo' : 'habit-done'}" data-id="${h.id}" ${!doneToday && false ? 'disabled' : ''}>
          ${doneToday ? '↩ Razveljavi danes' : 'Označi kot narejeno'}
        </button>
        <div class="habit-actions">
          <button class="btn btn-sm btn-ghost" data-action="edit-habit" data-id="${h.id}">Uredi</button>
          <button class="btn btn-sm btn-danger" data-action="delete-habit" data-id="${h.id}">Izbriši</button>
        </div>
      </div>`;
  }).join('');
}

function openModal(type, id = null) {
  closeMoreSheet();
  editingId = id;
  modalType = type;
  const fields = $('#modal-fields');
  fields.innerHTML = '';

  if (type === 'plans') {
    $('#modal-title').textContent = id ? 'Uredi načrt' : 'Nov načrt';
    const plan = id ? plans.find((p) => p.id === id) : null;
    modalTasks = plan?.tasks ? plan.tasks.map((t) => ({ ...t })) : [];
    fields.innerHTML = planFormHTML(plan);
  } else if (type === 'goals') {
    $('#modal-title').textContent = id ? 'Uredi cilj' : 'Nov cilj';
    fields.innerHTML = goalFormHTML(id ? goals.find((g) => g.id === id) : null);
  } else if (type === 'habits') {
    $('#modal-title').textContent = id ? 'Uredi navado' : 'Nova navada';
    fields.innerHTML = habitFormHTML(id ? habits.find((h) => h.id === id) : null);
  } else if (type === 'finance') {
    $('#modal-title').textContent = id ? 'Uredi transakcijo' : 'Nova transakcija';
    fields.innerHTML = transactionFormHTML(id ? transactions.find((t) => t.id === id) : null);
    updateTxCategories();
    if (id) {
      const tx = transactions.find((t) => t.id === id);
      if (tx) $('#f-tx-category').value = tx.category;
    }
  }

  $('#modal-overlay').hidden = false;
  document.body.classList.add('modal-open');
  requestAnimationFrame(() => {
    const first = fields.querySelector('input, textarea, select');
    if (first && window.matchMedia('(max-width: 900px)').matches) first.focus({ preventScroll: true });
  });
}

function closeModal() {
  $('#modal-overlay').hidden = true;
  document.body.classList.remove('modal-open');
  editingId = null;
  modalType = null;
  modalTasks = [];
}

function planFormHTML(plan) {
  return `
    <div class="form-group"><label for="f-title">Naslov *</label><input class="form-input" id="f-title" required value="${esc(plan?.title || '')}"></div>
    <div class="form-group"><label for="f-desc">Opis</label><textarea class="form-textarea" id="f-desc">${esc(plan?.description || '')}</textarea></div>
    <div class="form-row">
      <div class="form-group"><label for="f-category">Kategorija</label><select class="form-select" id="f-category">${selectOptions(LABELS.category, plan?.category || 'personal')}</select></div>
      <div class="form-group"><label for="f-priority">Prioriteta</label><select class="form-select" id="f-priority">${selectOptions(LABELS.priority, plan?.priority || 'medium')}</select></div>
    </div>
    <div class="form-row">
      <div class="form-group"><label for="f-status">Status</label><select class="form-select" id="f-status">${selectOptions(LABELS.status, plan?.status || 'todo')}</select></div>
      <div class="form-group"><label for="f-due">Rok</label><input class="form-input" type="date" id="f-due" value="${plan?.dueDate || (selectedWeekDay === 'overdue' ? localDateStr() : selectedWeekDay)}"></div>
    </div>
    <div class="form-group">
      <label>Podnaloge</label>
      <div class="tasks-editor" id="tasks-editor">${renderTaskEditorRows()}</div>
      <button type="button" class="btn btn-sm btn-ghost" data-action="add-task-row" style="margin-top:8px">+ Dodaj podnalogo</button>
    </div>`;
}

function renderTaskEditorRows() {
  if (!modalTasks.length) return '';
  return modalTasks.map((t, i) => `
    <div class="task-editor-row">
      <input class="form-input" data-task-index="${i}" value="${esc(t.title)}" placeholder="Podnaloga">
      <button type="button" class="btn btn-sm btn-danger" data-action="remove-task-row">×</button>
    </div>`).join('');
}

function addTaskRow() {
  modalTasks.push({ id: '', title: '', done: false, createdAt: '' });
  $('#tasks-editor').innerHTML = renderTaskEditorRows();
}

function collectTasksFromEditor() {
  const inputs = $$('#tasks-editor input[data-task-index]');
  const result = [];
  inputs.forEach((input, i) => {
    const title = input.value.trim();
    if (!title) return;
    const existing = modalTasks[i] || {};
    result.push({
      id: existing.id || '',
      title,
      done: !!existing.done,
      createdAt: existing.createdAt || '',
    });
  });
  return result;
}

function goalFormHTML(goal) {
  const isFinancial = goal?.kind === 'financial';
  return `
    <div class="form-group"><label for="f-title">Naslov *</label><input class="form-input" id="f-title" required value="${esc(goal?.title || '')}"></div>
    <div class="form-group"><label for="f-desc">Opis</label><textarea class="form-textarea" id="f-desc">${esc(goal?.description || '')}</textarea></div>
    <div class="form-group">
      <label for="f-kind">Tip cilja</label>
      <select class="form-select" id="f-kind" onchange="toggleGoalKindFields()">
        <option value="manual" ${!isFinancial ? 'selected' : ''}>Ročni napredek</option>
        <option value="financial" ${isFinancial ? 'selected' : ''}>Finančni (dinamičen)</option>
      </select>
    </div>
    <div id="goal-manual-fields" ${isFinancial ? 'hidden' : ''}>
      <div class="form-group"><label for="f-progress">Napredek (%)</label><input class="form-input" type="number" id="f-progress" min="0" max="100" value="${goal?.progress ?? 0}"></div>
    </div>
    <div id="goal-financial-fields" ${isFinancial ? '' : 'hidden'}>
      <div class="form-group"><label for="f-target-amount">Ciljni znesek (€)</label><input class="form-input" type="number" id="f-target-amount" min="0" step="0.01" value="${goal?.targetAmount || ''}"></div>
      <p style="font-size:0.82rem;color:var(--text-muted)">Napredek se avtomatsko posodablja iz transakcij, povezanih s tem ciljem.</p>
    </div>
    <div class="form-group"><label for="f-target">Ciljni datum</label><input class="form-input" type="date" id="f-target" value="${goal?.targetDate || ''}"></div>
    <div class="form-group"><label for="f-gstatus">Status</label><select class="form-select" id="f-gstatus">${selectOptions(LABELS.goalStatus, goal?.status || 'active')}</select></div>`;
}

window.toggleGoalKindFields = () => {
  const kind = $('#f-kind').value;
  $('#goal-manual-fields').hidden = kind === 'financial';
  $('#goal-financial-fields').hidden = kind !== 'financial';
};

function habitFormHTML(habit) {
  return `
    <div class="form-group"><label for="f-name">Ime navade *</label><input class="form-input" id="f-name" required value="${esc(habit?.name || '')}" placeholder="npr. Jutranja meditacija"></div>
    <div class="form-group"><label for="f-icon">Ikona</label><select class="form-select" id="f-icon">${HABIT_ICONS.map((i) => `<option value="${i}" ${habit?.icon === i ? 'selected' : ''}>${i}</option>`).join('')}</select></div>`;
}

function selectOptions(labels, selected) {
  return Object.entries(labels).map(([k, v]) => `<option value="${k}" ${k === selected ? 'selected' : ''}>${v}</option>`).join('');
}

async function handleFormSubmit(e) {
  e.preventDefault();
  setLoading(true);
  try {
    const type = modalType || currentView;
    if (type === 'plans') {
      const body = {
        title: $('#f-title').value,
        description: $('#f-desc').value,
        category: $('#f-category').value,
        priority: $('#f-priority').value,
        status: $('#f-status').value,
        dueDate: $('#f-due').value,
        tasks: collectTasksFromEditor(),
      };
      if (editingId) {
        await fetchJSON(`${API}/plans/${editingId}`, { method: 'PATCH', body: JSON.stringify(body) });
        toast('Načrt posodobljen');
      } else {
        await fetchJSON(`${API}/plans`, { method: 'POST', body: JSON.stringify(body) });
        toast('Načrt dodan');
      }
    } else if (type === 'goals') {
      const kind = $('#f-kind').value;
      const body = {
        title: $('#f-title').value,
        description: $('#f-desc').value,
        kind,
        targetDate: $('#f-target').value,
        status: $('#f-gstatus').value,
      };
      if (kind === 'financial') {
        body.targetAmount = parseFloat($('#f-target-amount').value) || 0;
      } else {
        body.progress = parseInt($('#f-progress').value, 10) || 0;
      }
      if (editingId) {
        await fetchJSON(`${API}/goals/${editingId}`, { method: 'PATCH', body: JSON.stringify(body) });
        toast('Cilj posodobljen');
      } else {
        await fetchJSON(`${API}/goals`, { method: 'POST', body: JSON.stringify(body) });
        toast('Cilj dodan');
      }
    } else if (type === 'habits') {
      const body = { name: $('#f-name').value, icon: $('#f-icon').value };
      if (editingId) {
        await fetchJSON(`${API}/habits/${editingId}`, { method: 'PATCH', body: JSON.stringify(body) });
        toast('Navada posodobljena');
      } else {
        await fetchJSON(`${API}/habits`, { method: 'POST', body: JSON.stringify(body) });
        toast('Navada dodana');
      }
    } else if (type === 'finance') {
      const body = {
        type: $('#f-tx-type').value,
        amount: parseFloat($('#f-amount').value),
        category: $('#f-tx-category').value,
        description: $('#f-tx-desc').value,
        date: $('#f-tx-date').value,
        goalId: $('#f-tx-goal').value,
      };
      if (editingId) {
        await fetchJSON(`${API}/transactions/${editingId}`, { method: 'PATCH', body: JSON.stringify(body) });
        toast('Transakcija posodobljena');
      } else {
        await fetchJSON(`${API}/transactions`, { method: 'POST', body: JSON.stringify(body) });
        toast('Transakcija dodana');
      }
    }
    closeModal();
    await refreshAll();
  } catch (err) {
    toast(err.message || 'Napaka', true);
  } finally {
    setLoading(false);
  }
}

async function cycleStatus(id) {
  const plan = plans.find((p) => p.id === id);
  if (!plan) return;
  let next = plan.status;
  if (plan.status === 'todo') next = 'in_progress';
  else if (plan.status === 'in_progress') next = 'done';
  else next = 'todo';

  try {
    await fetchJSON(`${API}/plans/${id}`, { method: 'PATCH', body: JSON.stringify({ status: next }) });
    toast('Status posodobljen');
    await refreshAll();
  } catch (err) {
    toast(err.message, true);
  }
}

async function toggleTask(planId, taskId) {
  const plan = plans.find((p) => p.id === planId);
  if (!plan) return;
  const tasks = (plan.tasks || []).map((t) => t.id === taskId ? { ...t, done: !t.done } : t);
  try {
    await fetchJSON(`${API}/plans/${planId}`, { method: 'PATCH', body: JSON.stringify({ tasks }) });
    await refreshAll();
  } catch (err) {
    toast(err.message, true);
  }
}

async function deletePlan(id) {
  if (!confirm('Res želiš izbrisati ta načrt?')) return;
  try {
    await fetchJSON(`${API}/plans/${id}`, { method: 'DELETE' });
    toast('Načrt izbrisan');
    await refreshAll();
  } catch (err) {
    toast(err.message, true);
  }
}

async function deleteGoal(id) {
  if (!confirm('Res želiš izbrisati ta cilj?')) return;
  try {
    await fetchJSON(`${API}/goals/${id}`, { method: 'DELETE' });
    toast('Cilj izbrisan');
    await refreshAll();
  } catch (err) {
    toast(err.message, true);
  }
}

async function deleteHabit(id) {
  if (!confirm('Res želiš izbrisati to navado?')) return;
  try {
    await fetchJSON(`${API}/habits/${id}`, { method: 'DELETE' });
    toast('Navada izbrisana');
    await refreshAll();
  } catch (err) {
    toast(err.message, true);
  }
}

async function markHabitDone(id) {
  try {
    await fetchJSON(`${API}/habits/${id}/done`, {
      method: 'POST',
      body: JSON.stringify({ date: localDateStr() }),
    });
    toast('Odlično! 🔥');
    await refreshAll();
  } catch (err) {
    toast(err.message, true);
  }
}

async function undoHabitDone(id) {
  try {
    await fetchJSON(`${API}/habits/${id}/undo`, {
      method: 'POST',
      body: JSON.stringify({ date: localDateStr() }),
    });
    toast('Check-in razveljavljen');
    await refreshAll();
  } catch (err) {
    toast(err.message, true);
  }
}

function renderFinanceSummary(selector, finance) {
  const el = $(selector);
  if (!el || !finance) return;
  el.innerHTML = `
    <div class="finance-card"><div class="label">Stanje</div><div class="amount ${finance.balance >= 0 ? 'positive' : 'negative'}">${formatEUR(finance.balance)}</div></div>
    <div class="finance-card"><div class="label">Prihodki (mesec)</div><div class="amount positive">${formatEUR(finance.monthIncome)}</div></div>
    <div class="finance-card"><div class="label">Odhodki (mesec)</div><div class="amount negative">${formatEUR(finance.monthExpense)}</div></div>
    <div class="finance-card"><div class="label">Bilanca (mesec)</div><div class="amount ${finance.monthBalance >= 0 ? 'positive' : 'negative'}">${formatEUR(finance.monthBalance)}</div></div>
  `;
}

function renderFinance() {
  const finance = overview?.finance;
  renderFinanceSummary('#finance-hero', finance);

  if (overview?.charts) {
    renderCharts(overview.charts, '');
  }

  renderTransactions();
}

function renderTransactions() {
  const el = $('#transactions-list');
  if (!el) return;

  const typeFilter = $('#finance-filter-type')?.value || '';
  const filtered = transactions.filter((t) => !typeFilter || t.type === typeFilter);

  if (!filtered.length) {
    el.innerHTML = '<div class="empty-state"><span>💰</span>Ni transakcij. Dodaj prvo!</div>';
    return;
  }

  el.innerHTML = filtered.map((t) => `
    <div class="transaction-item">
      <div class="transaction-icon ${t.type}">${t.type === 'income' ? '↑' : '↓'}</div>
      <div class="transaction-info">
        <div class="title">${esc(t.description || LABELS.financeCat[t.category])}</div>
        <div class="meta">${LABELS.financeCat[t.category]} · ${formatDate(t.date)}${t.goalId ? ' · 🎯 Cilj' : ''}</div>
      </div>
      <div class="transaction-amount ${t.type}">${t.type === 'income' ? '+' : '-'}${formatEUR(t.amount)}</div>
      <div class="transaction-actions">
        <button class="btn btn-sm btn-ghost" data-action="edit-transaction" data-id="${t.id}">Uredi</button>
        <button class="btn btn-sm btn-danger" data-action="delete-transaction" data-id="${t.id}">Izbriši</button>
      </div>
    </div>
  `).join('');
}

function transactionFormHTML(tx) {
  const financialGoals = goals.filter((g) => g.kind === 'financial');
  const type = tx?.type || 'expense';
  return `
    <div class="form-row">
      <div class="form-group">
        <label for="f-tx-type">Tip</label>
        <select class="form-select" id="f-tx-type" onchange="updateTxCategories()">
          <option value="income" ${type === 'income' ? 'selected' : ''}>Prihodek</option>
          <option value="expense" ${type === 'expense' ? 'selected' : ''}>Odhodek</option>
        </select>
      </div>
      <div class="form-group">
        <label for="f-amount">Znesek (€) *</label>
        <input class="form-input" type="number" id="f-amount" min="0.01" step="0.01" required value="${tx?.amount || ''}">
      </div>
    </div>
    <div class="form-group">
      <label for="f-tx-category">Kategorija</label>
      <select class="form-select" id="f-tx-category"></select>
    </div>
    <div class="form-group"><label for="f-tx-desc">Opis</label><input class="form-input" id="f-tx-desc" value="${esc(tx?.description || '')}"></div>
    <div class="form-group"><label for="f-tx-date">Datum</label><input class="form-input" type="date" id="f-tx-date" value="${tx?.date || localDateStr()}"></div>
    <div class="form-group">
      <label for="f-tx-goal">Povezan finančni cilj (opcijsko)</label>
      <select class="form-select" id="f-tx-goal">
        <option value="">— Brez —</option>
        ${financialGoals.map((g) => `<option value="${g.id}" ${tx?.goalId === g.id ? 'selected' : ''}>${esc(g.title)}</option>`).join('')}
      </select>
    </div>`;
}

window.updateTxCategories = () => {
  const type = $('#f-tx-type').value;
  const cats = type === 'income'
    ? { salary: 'Plača', freelance: 'Freelance', investment: 'Naložbe', other: 'Ostalo' }
    : { food: 'Hrana', transport: 'Prevoz', housing: 'Stanovanje', entertainment: 'Zabava', health: 'Zdravje', shopping: 'Nakupi', other: 'Ostalo' };
  $('#f-tx-category').innerHTML = Object.entries(cats).map(([k, v]) => `<option value="${k}">${v}</option>`).join('');
};

async function deleteTransaction(id) {
  if (!confirm('Res želiš izbrisati to transakcijo?')) return;
  try {
    await fetchJSON(`${API}/transactions/${id}`, { method: 'DELETE' });
    toast('Transakcija izbrisana');
    await refreshAll();
  } catch (err) {
    toast(err.message, true);
  }
}

function formatEUR(amount) {
  return new Intl.NumberFormat('sl-SI', { style: 'currency', currency: 'EUR' }).format(amount || 0);
}

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

function isOverdue(plan, today) {
  return plan.dueDate && plan.dueDate < today && plan.status !== 'done';
}

function formatDate(d) {
  if (!d) return '';
  return parseLocalDate(d).toLocaleDateString('sl-SI', { day: 'numeric', month: 'short' });
}

function formatDayLabel(dateStr) {
  const today = localDateStr();
  if (dateStr === today) return 'Danes';
  if (dateStr === addDays(today, 1)) return 'Jutri';
  return parseLocalDate(dateStr).toLocaleDateString('sl-SI', { weekday: 'long', day: 'numeric', month: 'long' });
}

function localDateStr(date = new Date()) {
  const y = date.getFullYear();
  const m = String(date.getMonth() + 1).padStart(2, '0');
  const d = String(date.getDate()).padStart(2, '0');
  return `${y}-${m}-${d}`;
}

function parseLocalDate(dateStr) {
  const [y, m, d] = dateStr.split('-').map(Number);
  return new Date(y, m - 1, d);
}

function addDays(dateStr, days) {
  const d = parseLocalDate(dateStr);
  d.setDate(d.getDate() + days);
  return localDateStr(d);
}

function esc(str) {
  const el = document.createElement('div');
  el.textContent = str || '';
  return el.innerHTML;
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
