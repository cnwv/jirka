const DEMO_DATA = [
  {
    title: "To Do", color: "#58a6ff", sections: null,
    tickets: [
      { key: "DEMO-101", summary: "Add user authentication to API", status: "Open", priority: "High", type: "Story", assignee: "alice", pIcon: "\u2227",
        desc: "Implement JWT-based authentication for all API endpoints.\n\n*Acceptance criteria:*\n- Bearer token validation on every request\n- Token refresh endpoint\n- Rate limiting per user\n\n{code}Authorization: Bearer <token>{code}" },
      { key: "DEMO-102", summary: "Database migration for new schema", status: "Open", priority: "Critical", type: "Task", assignee: "bob", pIcon: "\u23F6",
        desc: "Migration adds `sessions` and `audit_log` tables.\n\n*Steps:*\n1. Create migration file\n2. Add indexes on `user_id` and `created_at`\n3. Run on staging first\n4. Verify rollback works\n\n{code}rails generate migration AddSessionsTable{code}\n\n*Blocked by:* schema review from DBA team." },
      { key: "DEMO-103", summary: "Fix broken pagination on search results", status: "Open", priority: "Major", type: "Bug", assignee: "carol", pIcon: "\u2227",
        desc: "Search results page shows duplicate items when navigating past page 3. Likely an off-by-one error in the offset calculation." },
    ]
  },
  {
    title: "In Progress", color: "#3fb950",
    sections: [
      { name: "BUGS", color: "#f85149", tickets: [
        { key: "DEMO-201", summary: "Memory leak in background worker", status: "In Progress", priority: "Critical", type: "Bug", assignee: "dave", pIcon: "\u23F6",
          desc: "Sidekiq workers are accumulating memory over time. RSS grows from 256MB to 2GB+ within 24 hours. Likely caused by unbounded caching in the stats aggregator." },
        { key: "DEMO-202", summary: "Timeout on large file uploads", status: "In Progress", priority: "Major", type: "Bug", assignee: "eve", pIcon: "\u2227",
          desc: "Files over 50MB consistently timeout at the nginx proxy layer. Need to adjust `client_max_body_size` and `proxy_read_timeout`." },
      ]},
      { name: "FEATURES", color: "#3fb950", tickets: [
        { key: "DEMO-203", summary: "Implement webhook notifications", status: "In Progress", priority: "Medium", type: "Story", assignee: "frank", pIcon: "\u2228",
          desc: "Add webhook delivery system for real-time event notifications. Supports retry with exponential backoff." },
      ]},
    ]
  },
  {
    title: "Code Review", color: "#d29922", sections: null,
    tickets: [
      { key: "DEMO-301", summary: "Refactor payment processing module", status: "In Review", priority: "High", type: "Task", assignee: "grace", pIcon: "\u2227",
        desc: "Extracted payment logic from the monolithic controller into dedicated service objects. Each payment provider now has its own adapter class." },
      { key: "DEMO-302", summary: "Add rate limiting to public API", status: "In Review", priority: "Major", type: "Story", assignee: "henry", pIcon: "\u2227",
        desc: "Implements token bucket rate limiting: 100 req/min for authenticated, 20 req/min for anonymous. Uses Redis for distributed counting." },
    ]
  },
  {
    title: "Blocked", color: "#db6d28", sections: null,
    tickets: [
      { key: "DEMO-401", summary: "Waiting for third-party API credentials", status: "Blocked", priority: "High", type: "Task", assignee: "ivan", pIcon: "\u2227",
        desc: "Payment provider requires production API keys. Submitted request 2 weeks ago, still pending their security review." },
      { key: "DEMO-402", summary: "Infrastructure upgrade pending approval", status: "Blocked", priority: "Critical", type: "Task", assignee: "julia", pIcon: "\u23F6",
        desc: "K8s cluster needs upgrade from 1.27 to 1.29. Change request submitted to infrastructure team, awaiting CAB approval." },
    ]
  },
  {
    title: "QA", color: "#bc8cff",
    sections: [
      { name: "READY", color: "#58a6ff", tickets: [
        { key: "DEMO-501", summary: "Test new search filters", status: "Ready for QA", priority: "Medium", type: "Story", assignee: "kate", pIcon: "\u2228",
          desc: "New faceted search filters added for date range, status, and priority. Needs cross-browser testing and performance validation with 10k+ records." },
      ]},
      { name: "TESTING", color: "#d29922", tickets: [
        { key: "DEMO-502", summary: "Regression test for checkout flow", status: "Testing", priority: "High", type: "Bug", assignee: "leo", pIcon: "\u2227",
          desc: "Checkout flow regression after payment refactor. QA needs to verify: cart total, discount application, tax calculation, and payment confirmation email." },
      ]},
    ]
  },
  {
    title: "Done", color: "#39d353",
    sections: [
      { name: "THIS WEEK", color: "#3fb950", tickets: [
        { key: "DEMO-601", summary: "Fix CSV export encoding", status: "Done", priority: "Medium", type: "Bug", assignee: "mia", pIcon: "\u2228",
          desc: "CSV exports now use UTF-8 BOM for proper Excel compatibility. Fixed issue where special characters were garbled in non-English locales." },
        { key: "DEMO-602", summary: "Update dependencies to latest versions", status: "Done", priority: "Low", type: "Task", assignee: "nick", pIcon: "\u2228",
          desc: "Updated all gems to latest minor/patch versions. No breaking changes. Security advisory CVE-2024-1234 resolved by updating nokogiri." },
      ]},
      { name: "DEPLOYED", color: "#79c0ff", tickets: [] },
    ]
  },
];

let selectedTicket = null;
let activePanel = 0;
let focusDetail = false;
let detailScroll = 0;

// Per-panel cursor index
const cursorByPanel = DEMO_DATA.map(() => 0);

function getAllTickets(panel) {
  if (panel.sections) {
    return panel.sections.flatMap(s => s.tickets);
  }
  return panel.tickets;
}

function currentTickets() {
  return getAllTickets(DEMO_DATA[activePanel]);
}

function syncSelection() {
  const tickets = currentTickets();
  const idx = cursorByPanel[activePanel];
  if (tickets.length > 0 && idx < tickets.length) {
    selectedTicket = tickets[idx];
  }
  renderDetail();
}

function renderPanels() {
  const grid = document.getElementById('panels');
  grid.innerHTML = '';

  DEMO_DATA.forEach((panel, pi) => {
    const div = document.createElement('div');
    div.className = 'panel' + (pi === activePanel && !focusDetail ? ' active' : '');
    div.style.setProperty('--panel-color', panel.color);

    const allTickets = getAllTickets(panel);
    const header = document.createElement('div');
    header.className = 'panel-header';
    header.innerHTML = `<span>${pi + 1}  ${panel.title}</span><span class="count">${allTickets.length}</span>`;

    const body = document.createElement('div');
    body.className = 'panel-body';

    let ticketIdx = 0;
    if (panel.sections) {
      panel.sections.forEach(sec => {
        const label = document.createElement('div');
        label.className = 'section-label';
        label.style.color = sec.color;
        label.textContent = `\u2500\u2500 ${sec.name} \u2500\u2500`;
        body.appendChild(label);

        if (sec.tickets.length === 0) {
          const empty = document.createElement('div');
          empty.className = 'ticket-row';
          empty.style.color = 'var(--text-dim)';
          empty.style.fontStyle = 'italic';
          empty.textContent = '  EMPTY';
          body.appendChild(empty);
        }

        sec.tickets.forEach(t => {
          body.appendChild(makeTicketRow(t, pi, ticketIdx));
          ticketIdx++;
        });
      });
    } else {
      panel.tickets.forEach(t => {
        body.appendChild(makeTicketRow(t, pi, ticketIdx));
        ticketIdx++;
      });
    }

    div.appendChild(header);
    div.appendChild(body);
    div.addEventListener('click', () => {
      activePanel = pi;
      focusDetail = false;
      const cursor = cursorByPanel[pi];
      if (allTickets[cursor]) selectTicket(allTickets[cursor]);
      renderPanels();
    });
    grid.appendChild(div);
  });

  // Update detail pane border to show focus
  const detailPane = document.querySelector('.detail-pane');
  if (detailPane) {
    detailPane.style.borderColor = focusDetail ? 'var(--accent)' : '';
  }
}

function makeTicketRow(ticket, panelIdx, ticketIdx) {
  const row = document.createElement('div');
  const isSelected = panelIdx === activePanel && cursorByPanel[panelIdx] === ticketIdx;
  row.className = 'ticket-row' + (isSelected ? ' selected' : '');
  row.innerHTML = `<span class="ticket-priority">${ticket.pIcon}</span><span class="ticket-key">${ticket.key}</span><span class="ticket-summary">${ticket.summary}</span>`;
  row.addEventListener('click', (e) => {
    e.stopPropagation();
    activePanel = panelIdx;
    focusDetail = false;
    cursorByPanel[panelIdx] = ticketIdx;
    selectTicket(ticket);
    renderPanels();
  });
  return row;
}

function selectTicket(ticket) {
  selectedTicket = ticket;
  detailScroll = 0;
  renderDetail();
}

function renderDetail() {
  const header = document.getElementById('detail-header');
  const body = document.getElementById('detail-body');

  if (!selectedTicket) {
    header.textContent = 'Select a ticket';
    body.innerHTML = '<p style="color: var(--text-dim)">Click on a ticket or press j/k to navigate</p>';
    return;
  }

  const t = selectedTicket;
  header.textContent = `${t.key} \u2014 ${t.summary}`;

  let desc = t.desc || '';
  desc = desc.replace(/\{code\}([\s\S]*?)\{code\}/g, '<div class="code-block">$1</div>');
  desc = desc.replace(/\*(.*?)\*/g, '<strong>$1</strong>');
  desc = desc.replace(/\n/g, '<br>');

  body.innerHTML = `
    <div class="detail-field"><span class="detail-label">Status</span><span class="detail-value">${t.status}</span></div>
    <div class="detail-field"><span class="detail-label">Priority</span><span class="detail-value">${t.pIcon} ${t.priority}</span></div>
    <div class="detail-field"><span class="detail-label">Type</span><span class="detail-value">${t.type}</span></div>
    <div class="detail-field"><span class="detail-label">Assignee</span><span class="detail-value">${t.assignee}</span></div>
    <div class="detail-desc">
      <h3>Description</h3>
      <p>${desc}</p>
    </div>
  `;
}

function scrollDetail(delta) {
  const body = document.getElementById('detail-body');
  if (body) body.scrollBy({ top: delta, behavior: 'smooth' });
}

function showKeyHint(key) {
  let hint = document.getElementById('key-hint');
  if (!hint) {
    hint = document.createElement('div');
    hint.id = 'key-hint';
    hint.style.cssText = 'position:fixed;bottom:20px;left:50%;transform:translateX(-50%);background:#1c2128;border:1px solid #30363d;color:#58a6ff;padding:6px 16px;border-radius:6px;font-size:0.85rem;z-index:999;opacity:0;transition:opacity 0.15s;pointer-events:none;font-family:inherit;';
    document.body.appendChild(hint);
  }
  hint.textContent = key;
  hint.style.opacity = '1';
  clearTimeout(hint._timer);
  hint._timer = setTimeout(() => { hint.style.opacity = '0'; }, 600);
}

// Keyboard handler
document.addEventListener('keydown', (e) => {
  // Don't capture when typing in inputs
  if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA') return;

  const key = e.key;

  // 1-6: focus panel
  if (key >= '1' && key <= '6') {
    const idx = parseInt(key) - 1;
    if (idx < DEMO_DATA.length) {
      e.preventDefault();
      activePanel = idx;
      focusDetail = false;
      syncSelection();
      renderPanels();
      showKeyHint(`${key}  ${DEMO_DATA[idx].title}`);
    }
    return;
  }

  // 7: focus detail (N+1 for 6 panels)
  if (key === '7') {
    e.preventDefault();
    focusDetail = true;
    renderPanels();
    showKeyHint('7  Detail');
    return;
  }

  // Tab / Shift+Tab: cycle focus
  if (key === 'Tab') {
    e.preventDefault();
    if (e.shiftKey) {
      if (focusDetail) {
        focusDetail = false;
      } else {
        activePanel = (activePanel - 1 + DEMO_DATA.length) % DEMO_DATA.length;
      }
    } else {
      if (!focusDetail && activePanel === DEMO_DATA.length - 1) {
        focusDetail = true;
      } else if (focusDetail) {
        focusDetail = false;
        activePanel = 0;
      } else {
        activePanel++;
      }
    }
    if (!focusDetail) syncSelection();
    renderPanels();
    showKeyHint(focusDetail ? 'Detail' : `${activePanel + 1}  ${DEMO_DATA[activePanel].title}`);
    return;
  }

  // j/Down: move cursor down or scroll detail
  if (key === 'j' || key === 'ArrowDown') {
    e.preventDefault();
    if (focusDetail) {
      scrollDetail(60);
      showKeyHint('\u2193 scroll');
    } else {
      const tickets = currentTickets();
      if (cursorByPanel[activePanel] < tickets.length - 1) {
        cursorByPanel[activePanel]++;
        syncSelection();
        renderPanels();
        showKeyHint(`\u2193 ${selectedTicket ? selectedTicket.key : ''}`);
      }
    }
    return;
  }

  // k/Up: move cursor up or scroll detail
  if (key === 'k' || key === 'ArrowUp') {
    e.preventDefault();
    if (focusDetail) {
      scrollDetail(-60);
      showKeyHint('\u2191 scroll');
    } else {
      if (cursorByPanel[activePanel] > 0) {
        cursorByPanel[activePanel]--;
        syncSelection();
        renderPanels();
        showKeyHint(`\u2191 ${selectedTicket ? selectedTicket.key : ''}`);
      }
    }
    return;
  }

  // r: simulate refresh
  if (key === 'r') {
    e.preventDefault();
    const statusSpan = document.querySelector('.status-bar span');
    if (statusSpan) {
      statusSpan.textContent = 'Refreshing...';
      setTimeout(() => {
        const now = new Date();
        statusSpan.textContent = `Last update: ${now.toLocaleTimeString()}`;
      }, 500);
    }
    showKeyHint('r  refresh');
    return;
  }
});

function copyInstall(el) {
  navigator.clipboard.writeText('brew install cnwv/apps/jirka').then(() => {
    const hint = el.querySelector('.copy-hint');
    hint.textContent = 'copied!';
    hint.style.opacity = '1';
    setTimeout(() => {
      hint.textContent = 'click to copy';
      hint.style.opacity = '';
    }, 1500);
  });
}

// Init
cursorByPanel[0] = 1; // start on DEMO-102
selectedTicket = DEMO_DATA[0].tickets[1];
renderPanels();
renderDetail();
