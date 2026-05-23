const state = {
  token: localStorage.getItem("zolix_token") || "",
  user: null,
  dashboard: null,
  orders: [],
  customers: [],
  services: [],
  currentStatus: "Semua",
  currentView: "dashboard",
};

const rupiah = new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR", maximumFractionDigits: 0 });
const dateFmt = new Intl.DateTimeFormat("id-ID", { day: "2-digit", month: "short", year: "numeric" });
const timeFmt = new Intl.DateTimeFormat("id-ID", { hour: "2-digit", minute: "2-digit" });

const statusMeta = {
  "Diterima": ["info", "⊙ DITERIMA"],
  "Cleaning": ["orange", "✺ PROSES"],
  "Drying": ["orange", "✺ PROSES"],
  "Finishing": ["orange", "✺ PROSES"],
  "Ready Pickup": ["warning", "MENUNGGU DIAMBIL"],
  "Completed": ["success", "✓ SELESAI"],
  "Diambil": ["neutral", "▢ DIAMBIL"],
  "Menunggu Diambil": ["warning", "MENUNGGU DIAMBIL"],
  "Dibatalkan": ["danger", "⊗ DIBATALKAN"],
  "Pending": ["neutral", "PENDING"],
};

const shoeStyles = [
  ["#ffffff", "#d1d5db", "-14deg"],
  ["#050505", "#ffffff", "10deg"],
  ["#eee8dc", "#b8b0a4", "14deg"],
  ["#d6d1c7", "#151515", "-10deg"],
  ["#e6e7e8", "#bfc3c9", "5deg"],
  ["#151515", "#ef4444", "8deg"],
];

async function api(path, options) {
  const headers = { "Content-Type": "application/json", ...(options && options.headers ? options.headers : {}) };
  if (state.token) headers.Authorization = `Bearer ${state.token}`;
  const response = await fetch(path, {
    headers,
    ...options,
  });
  if (response.status === 401 && path !== "/api/v1/auth/login") {
    logout();
    throw new Error("Sesi login berakhir");
  }
  if (!response.ok) {
    throw new Error((await response.json()).error || "Request failed");
  }
  return response.json();
}

async function login(event) {
  event.preventDefault();
  const form = new FormData(event.currentTarget);
  const error = document.querySelector("#login-error");
  error.textContent = "";
  try {
    const result = await api("/api/v1/auth/login", {
      method: "POST",
      body: JSON.stringify({
        email: String(form.get("email")),
        password: String(form.get("password")),
      }),
    });
    state.token = result.token;
    state.user = result.user;
    localStorage.setItem("zolix_token", state.token);
    document.querySelector("#login-screen").classList.add("hidden");
    await load();
  } catch (err) {
    error.textContent = err.message || "Login gagal";
  }
}

function logout() {
  state.token = "";
  state.user = null;
  localStorage.removeItem("zolix_token");
  document.querySelector("#login-screen").classList.remove("hidden");
}

async function load() {
  const [dashboard, orders, customers, services] = await Promise.all([
    api("/api/v1/dashboard"),
    api("/api/v1/orders"),
    api("/api/v1/customers"),
    api("/api/v1/services"),
  ]);
  state.dashboard = dashboard;
  state.orders = orders;
  state.customers = customers;
  state.services = services;
  renderAll();
}

function renderAll() {
  renderStats();
  renderTabs("status-tabs");
  renderTabs("orders-status-tabs");
  renderTable();
  renderOrderCards();
  renderActivity();
  renderCustomers();
  renderServices();
  renderSchedule();
  renderPayments();
  renderServiceOptions();
}

function renderStats() {
  const data = state.dashboard;
  const stats = [
    ["Total order hari ini", data.total_today, "+12% dari kemarin"],
    ["Order aktif", data.active_orders, "Dalam proses"],
    ["Order selesai", data.completed_orders, "Siap invoice"],
    ["Menunggu diambil", data.waiting_pickup, "Reminder pickup"],
    ["Pendapatan bulanan", rupiah.format(data.monthly_revenue), "Mei 2025"],
  ];
  document.querySelector("#stats-grid").innerHTML = stats.map(([label, value, note]) => `
    <article class="stat-card">
      <span>${label}</span>
      <strong>${value}</strong>
      <small>${note}</small>
    </article>
  `).join("");
}

function renderTabs(targetId) {
  const counts = state.dashboard.status_counts || {};
  const tabs = [
    ["Semua", "Semua", state.orders.length],
    ["Diterima", "Diterima", counts.Diterima || 0],
    ["Proses", "Cleaning", counts.Cleaning || 0],
    ["Selesai", "Completed", counts.Completed || 0],
    ["Diambil", "Diambil", counts.Diambil || 0],
    ["Menunggu Diambil", "Menunggu Diambil", counts["Menunggu Diambil"] || 0],
    ["Dibatalkan", "Dibatalkan", counts.Dibatalkan || 0],
  ];
  document.querySelector(`#${targetId}`).innerHTML = tabs.map(([label, value, count]) => `
    <button class="tab ${state.currentStatus === value ? "active" : ""}" data-status="${value}">
      ${label}<span class="count ${countClass(label)}">${count}</span>
    </button>
  `).join("");
}

function countClass(label) {
  if (label === "Diterima") return "badge info";
  if (label === "Proses") return "badge orange";
  if (label === "Selesai") return "badge success";
  if (label === "Menunggu Diambil") return "badge warning";
  if (label === "Dibatalkan") return "badge danger";
  return "badge neutral";
}

function filteredOrders() {
  const query = document.querySelector("#search-input").value.toLowerCase().trim();
  return state.orders.filter((order) => {
    const statusOk = state.currentStatus === "Semua" || order.status === state.currentStatus;
    const text = `${order.invoice_number} ${order.customer_name} ${order.customer_phone} ${order.items.map((item) => item.service).join(" ")}`.toLowerCase();
    return statusOk && (!query || text.includes(query));
  });
}

function renderTable() {
  const orders = filteredOrders();
  document.querySelector("#orders-table").innerHTML = orders.map((order, index) => {
    const item = order.items[0];
    const shoe = shoeStyles[index % shoeStyles.length];
    return `
      <tr>
        <td>
          <div class="shoe-cell">
            <div>
              <strong>${order.invoice_number}</strong>
              <small>${formatDate(order.created_at)} • ${formatTime(order.created_at)}</small>
            </div>
            <div class="shoe-thumb" style="--shoe:${shoe[0]};--sole:${shoe[1]};--angle:${shoe[2]}"></div>
          </div>
        </td>
        <td><strong>${order.customer_name}</strong><small>☏ ${order.customer_phone}</small></td>
        <td><span>${item.service}</span><br><small>${item.qty} Pasang</small></td>
        <td><span>${formatDate(order.created_at)}</span><br><small>${formatTime(order.created_at)} WIB</small></td>
        <td><span>${formatDate(order.estimated_done_at)}</span><br><small>${formatTime(order.estimated_done_at)} WIB</small></td>
        <td>${statusBadge(order.status)}</td>
        <td><strong>${rupiah.format(order.total_price)}</strong></td>
        <td><div class="actions"><button data-open="${order.id}" title="Detail invoice">⊙</button><button data-wa="${order.id}" title="WhatsApp">⋮</button></div></td>
      </tr>
    `;
  }).join("");
  document.querySelector("#table-count").textContent = `Menampilkan 1 - ${orders.length} dari ${state.orders.length} order`;
}

function renderOrderCards() {
  const orders = filteredOrders();
  document.querySelector("#orders-list").innerHTML = orders.map((order) => `
    <article class="order-card">
      <div><strong>${order.invoice_number}</strong><small>${formatDate(order.created_at)} • ${formatTime(order.created_at)} WIB</small></div>
      <div><strong>${order.customer_name}</strong><small>${order.customer_phone}</small></div>
      <div>${order.items.map((item) => item.service).join(", ")}<small>${order.items.length} item</small></div>
      <div>${statusBadge(order.status)}</div>
      <div class="actions"><button data-open="${order.id}">⊙</button><button data-wa="${order.id}">⋮</button></div>
    </article>
  `).join("");
}

function renderActivity() {
  document.querySelector("#activity-list").innerHTML = state.dashboard.activities.map((item, index) => `
    <div class="activity-item">
      <strong>${item}</strong>
      <span>${index + 2} menit lalu</span>
    </div>
  `).join("");
}

function renderCustomers() {
  document.querySelector("#customer-grid").innerHTML = state.customers.map((customer) => `
    <article class="customer-card">
      <strong>${customer.name}</strong>
      <span>${customer.phone}</span>
    </article>
  `).join("");
}

function renderServices() {
  document.querySelector("#service-grid").innerHTML = state.services.map((service) => `
    <article class="service-card">
      <strong>${service.name}</strong>
      <span>${rupiah.format(service.price)} • ${service.duration}</span>
    </article>
  `).join("");
}

function renderSchedule() {
  const waiting = state.orders.filter((order) => ["Menunggu Diambil", "Completed", "Ready Pickup"].includes(order.status));
  document.querySelector("#schedule-list").innerHTML = waiting.map((order) => `
    <article class="order-card">
      <div><strong>${order.invoice_number}</strong><small>${order.customer_name}</small></div>
      <div><strong>Estimasi pickup</strong><small>${formatDate(order.estimated_done_at)} • ${formatTime(order.estimated_done_at)} WIB</small></div>
      <div>${statusBadge(order.status)}</div>
      <div><strong>${rupiah.format(order.total_price)}</strong></div>
      <div class="actions"><button data-wa="${order.id}">⋮</button></div>
    </article>
  `).join("");
}

function pendingPaymentOrders() {
  return state.orders.filter((order) => ["Belum Bayar", "DP"].includes(order.payment_status));
}

function paymentAge(order) {
  const createdAt = new Date(order.created_at);
  if (Number.isNaN(createdAt.getTime())) return "Tanggal tidak valid";
  const diffHours = Math.max(0, Math.floor((Date.now() - createdAt.getTime()) / 36e5));
  if (diffHours < 1) return "Baru dibuat";
  if (diffHours < 24) return `${diffHours} jam pending`;
  const days = Math.floor(diffHours / 24);
  return `${days} hari pending`;
}

function paymentMethodLabel(order) {
  const method = order.payment_method || "Belum dipilih";
  const provider = order.payment_provider ? ` via ${order.payment_provider}` : "";
  return `${method}${provider}`;
}

function renderPayments() {
  const pending = pendingPaymentOrders();
  const unpaid = pending.filter((order) => order.payment_status === "Belum Bayar");
  const partial = pending.filter((order) => order.payment_status === "DP");
  const pendingAmount = pending.reduce((sum, order) => sum + (order.total_price || 0), 0);
  const qrisPending = pending.filter((order) => order.payment_method === "QRIS").length;

  document.querySelector("#payment-summary-grid").innerHTML = [
    ["Total Pending", pending.length, `${unpaid.length} belum bayar, ${partial.length} DP`],
    ["Nominal Pending", rupiah.format(pendingAmount), "Perlu follow up pembayaran"],
    ["QRIS Pending", qrisPending, "Bisa dicek status AutoGoPay"],
  ].map(([label, value, note]) => `
    <article class="payment-stat-card">
      <span>${label}</span>
      <strong>${value}</strong>
      <small>${note}</small>
    </article>
  `).join("");

  document.querySelector("#payment-list").innerHTML = pending.length ? pending.map((order) => `
    <article class="payment-card">
      <div class="payment-card-main">
        <div>
          <strong>${order.invoice_number}</strong>
          <small>${formatDate(order.created_at)} • ${formatTime(order.created_at)} WIB</small>
        </div>
        <div>
          <span>${order.customer_name}</span>
          <small>${order.customer_phone}</small>
        </div>
      </div>
      <div class="payment-card-info">
        <div><span>Status Bayar</span><strong>${paymentBadge(order.payment_status)}</strong></div>
        <div><span>Metode</span><strong>${paymentMethodLabel(order)}</strong></div>
        <div><span>Total</span><strong>${rupiah.format(order.total_price)}</strong></div>
        <div><span>Umur Pending</span><strong>${paymentAge(order)}</strong></div>
      </div>
      <div class="payment-card-ref">
        <span>Referensi pembayaran</span>
        <strong>${order.payment_reference || "Belum ada reference"}</strong>
        <small>${order.payment_qr_url || order.payment_qr_string ? "QRIS tersedia untuk pelanggan." : "Belum ada QRIS aktif."}</small>
      </div>
      <div class="payment-card-actions">
        <button class="secondary-button" data-open="${order.id}">Detail</button>
        ${order.payment_method === "QRIS" && !order.payment_reference ? `<button class="secondary-button" data-generate-qris="${order.id}">Generate QRIS</button>` : ""}
        ${order.payment_reference ? `<button class="secondary-button" data-check-qris="${order.id}">Cek Status</button>` : ""}
        <a class="secondary-button" href="/order/${encodeURIComponent(order.invoice_number)}" target="_blank" rel="noopener">Tracking</a>
      </div>
    </article>
  `).join("") : `
    <div class="payment-empty">
      <strong>Tidak ada pembayaran pending.</strong>
      <span>Semua invoice sudah lunas atau belum ada order aktif.</span>
    </div>
  `;
}

function renderServiceOptions() {
  const select = document.querySelector("#service-select");
  select.innerHTML = state.services.map((service) => `<option value="${service.name}" data-price="${service.price}">${service.name}</option>`).join("");
}

function statusBadge(status) {
  const [className, label] = statusMeta[status] || ["neutral", status];
  return `<span class="badge ${className}">${label}</span>`;
}

function paymentBadge(status) {
  if (status === "Lunas") return `<span class="badge success">LUNAS</span>`;
  if (status === "DP") return `<span class="badge warning">DP</span>`;
  return `<span class="badge danger">BELUM BAYAR</span>`;
}

function formatDate(value) {
  return dateFmt.format(new Date(value));
}

function formatTime(value) {
  return timeFmt.format(new Date(value)).replace(".", ":");
}

function setView(view) {
  state.currentView = view;
  document.querySelectorAll(".view").forEach((node) => node.classList.remove("active-view"));
  document.querySelector(`#${view}-view`).classList.add("active-view");
  document.querySelectorAll(".nav-item").forEach((node) => node.classList.toggle("active", node.dataset.view === view));
  closeSidebar();
  const titles = {
    dashboard: ["Dashboard", "Ringkasan operasional ZOLIX Shoe Care hari ini."],
    orders: ["List Order", "Kelola semua order dengan mudah dan cepat."],
    create: ["Tambah Order", "Buat order baru, invoice, dan tracking customer."],
    payments: ["Pembayaran", "Pantau pembayaran pending, QRIS, transfer, dan cash."],
    schedule: ["Jadwal", "Pantau estimasi selesai dan pickup pelanggan."],
    customers: ["Pelanggan", "Tracking customer dan histori order."],
    services: ["Layanan", "Kelola layanan, harga, dan estimasi pengerjaan."],
    settings: ["Pengaturan", "Integrasi sistem dan konfigurasi operasional."],
  };
  document.querySelector("#page-title").textContent = titles[view][0];
  document.querySelector("#page-subtitle").textContent = titles[view][1];
}

function openSidebar() {
  document.body.classList.add("sidebar-open");
  document.querySelector("#menu-button").setAttribute("aria-expanded", "true");
}

function closeSidebar() {
  document.body.classList.remove("sidebar-open");
  document.querySelector("#menu-button").setAttribute("aria-expanded", "false");
}

function openDrawer(orderID) {
  const order = state.orders.find((item) => item.id === Number(orderID));
  if (!order) return;
  const media = order.media && order.media.length ? order.media : [
    { type: "before", url: "/assets/image1.png" },
    { type: "after", url: "/assets/nota_order.png" },
  ];
  document.querySelector("#drawer-content").innerHTML = `
    <div class="invoice-head">
      <h2>${order.invoice_number}</h2>
      <p>${order.customer_name} • ${order.customer_phone}</p>
    </div>
    <div class="settings-grid" style="grid-template-columns:1fr 1fr;padding:0;margin-bottom:18px">
      <div><strong>Status</strong>${statusBadge(order.status)}</div>
      <div><strong>Pembayaran</strong><span>${order.payment_status} via ${order.payment_method}</span></div>
      <div><strong>Total</strong><span>${rupiah.format(order.total_price)}</span></div>
      <div><strong>Estimasi</strong><span>${formatDate(order.estimated_done_at)}</span></div>
    </div>
    <div class="payment-box">
      <div>
        <strong>QRIS AutoGoPay</strong>
        <span>${order.payment_reference ? `Ref: ${order.payment_reference}` : "Belum ada transaksi QRIS"}</span>
      </div>
      ${order.payment_qr_url ? `<img src="${order.payment_qr_url}" alt="QRIS pembayaran ${order.invoice_number}">` : ""}
      ${order.payment_qr_string ? `<textarea readonly rows="3">${order.payment_qr_string}</textarea>` : ""}
      <div class="payment-actions">
        <button class="secondary-button" data-generate-qris="${order.id}">Generate QRIS</button>
        ${order.payment_reference ? `<button class="secondary-button" data-check-qris="${order.id}">Cek Status</button>` : ""}
      </div>
    </div>
    <h3>Timeline Progress</h3>
    <div class="timeline">
      ${order.timeline.map((step) => `
        <div class="timeline-item ${step.done ? "done" : ""}">
          <span class="dot"></span>
          <div><strong>${step.label}</strong><small>${step.done ? `${formatDate(step.time)} • ${formatTime(step.time)} WIB` : "Menunggu update cleaner"}</small></div>
        </div>
      `).join("")}
    </div>
    <h3>Before / After</h3>
    <div class="media-grid">
      ${media.map((item) => `<figure><img src="${item.url}" alt="${item.type} cleaning"><figcaption>${item.type}</figcaption></figure>`).join("")}
    </div>
    <form class="upload-form" data-upload-order="${order.id}">
      <label>Jenis foto
        <select name="type"><option value="before">Before</option><option value="after">After</option></select>
      </label>
      <label>File foto
        <input name="file" type="file" accept="image/jpeg,image/png,image/webp" required>
      </label>
      <button class="secondary-button" type="submit">Upload Foto</button>
    </form>
    <p style="margin:18px 0">${order.notes || "Tidak ada catatan tambahan."}</p>
    <label style="margin-bottom:12px">Update status
      <select id="drawer-status">
        ${Object.keys(statusMeta).map((status) => `<option value="${status}" ${status === order.status ? "selected" : ""}>${status}</option>`).join("")}
      </select>
    </label>
    <div style="display:flex;gap:10px;flex-wrap:wrap">
      <button class="primary-button" data-save-status="${order.id}">Simpan Status</button>
      <a class="secondary-button" style="display:flex;align-items:center;text-decoration:none" href="/order/${encodeURIComponent(order.invoice_number)}" target="_blank">Tracking</a>
      <button class="whatsapp" data-wa="${order.id}">Share WhatsApp</button>
      <button class="secondary-button" data-delete-order="${order.id}">Hapus</button>
    </div>
  `;
  document.querySelector("#drawer").classList.add("open");
}

function closeDrawer() {
  document.querySelector("#drawer").classList.remove("open");
}

async function shareWhatsApp(orderID) {
  try {
    await api("/api/v1/whatsapp/send", {
      method: "POST",
      body: JSON.stringify({ order_id: Number(orderID) }),
    });
    alert("Pesan WhatsApp berhasil dikirim.");
  } catch (error) {
    alert(error.message || "Pesan WhatsApp gagal dikirim.");
  }
}

async function generateQRIS(orderID) {
  try {
    await api("/api/v1/payments/qris/generate", {
      method: "POST",
      body: JSON.stringify({ order_id: Number(orderID) }),
    });
    await load();
    openDrawer(orderID);
  } catch (error) {
    alert(error.message || "Gagal membuat QRIS");
  }
}

async function checkQRIS(orderID) {
  try {
    await api("/api/v1/payments/qris/status", {
      method: "POST",
      body: JSON.stringify({ order_id: Number(orderID) }),
    });
    await load();
    openDrawer(orderID);
  } catch (error) {
    alert(error.message || "Gagal mengecek status QRIS");
  }
}

async function createOrder(event) {
  event.preventDefault();
  const form = new FormData(event.currentTarget);
  const service = String(form.get("service"));
  const payload = {
    customer_name: String(form.get("customer_name")),
    customer_phone: String(form.get("customer_phone")),
    payment_method: String(form.get("payment_method")),
    payment_status: String(form.get("payment_status")),
    notes: String(form.get("notes")),
    items: [{
      service,
      shoe_name: String(form.get("shoe_name") || "Sneakers"),
      qty: Number(form.get("qty") || 1),
      price: Number(form.get("price") || 0),
    }],
  };
  await api("/api/v1/orders", { method: "POST", body: JSON.stringify(payload) });
  event.currentTarget.reset();
  await load();
  setView("orders");
}

async function updateOrderStatus(orderID) {
  const order = state.orders.find((item) => item.id === Number(orderID));
  const status = document.querySelector("#drawer-status").value;
  if (!order || !status) return;
  await api(`/api/v1/orders/${order.id}`, {
    method: "PUT",
    body: JSON.stringify({
      customer_name: order.customer_name,
      customer_phone: order.customer_phone,
      status,
      payment_status: order.payment_status,
      payment_method: order.payment_method,
      notes: order.notes,
    }),
  });
  await load();
  openDrawer(order.id);
}

async function deleteOrder(orderID) {
  if (!confirm("Hapus order ini?")) return;
  await api(`/api/v1/orders/${orderID}`, { method: "DELETE" });
  closeDrawer();
  await load();
}

async function uploadMedia(event, uploadForm) {
  event.preventDefault();
  const formElement = uploadForm || event.target.closest("[data-upload-order]");
  if (!formElement) return;
  const orderID = formElement.dataset.uploadOrder;
  const submitButton = formElement.querySelector("button[type='submit']");
  const originalLabel = submitButton ? submitButton.textContent : "";
  if (submitButton) {
    submitButton.disabled = true;
    submitButton.textContent = "Mengupload...";
  }
  const form = new FormData(formElement);
  form.append("order_id", orderID);
  try {
    const response = await fetch("/api/v1/upload", {
      method: "POST",
      headers: { Authorization: `Bearer ${state.token}` },
      body: form,
    });
    if (response.status === 401) {
      logout();
      return;
    }
    if (!response.ok) {
      const error = await response.json();
      alert(error.error || "Upload gagal");
      return;
    }
    formElement.reset();
    await load();
    openDrawer(orderID);
  } catch (error) {
    alert(error.message || "Upload gagal");
  } finally {
    if (submitButton) {
      submitButton.disabled = false;
      submitButton.textContent = originalLabel;
    }
  }
}

document.addEventListener("click", (event) => {
  const menuButton = event.target.closest("#menu-button");
  if (menuButton) openSidebar();

  if (event.target.closest("#sidebar-backdrop")) closeSidebar();

  const viewButton = event.target.closest("[data-view]");
  if (viewButton) setView(viewButton.dataset.view);

  const tab = event.target.closest("[data-status]");
  if (tab) {
    state.currentStatus = tab.dataset.status;
    renderTabs("status-tabs");
    renderTabs("orders-status-tabs");
    renderTable();
    renderOrderCards();
  }

  const open = event.target.closest("[data-open]");
  if (open) openDrawer(open.dataset.open);

  const wa = event.target.closest("[data-wa]");
  if (wa) shareWhatsApp(wa.dataset.wa);

  const generateQRISButton = event.target.closest("[data-generate-qris]");
  if (generateQRISButton) generateQRIS(generateQRISButton.dataset.generateQris);

  const checkQRISButton = event.target.closest("[data-check-qris]");
  if (checkQRISButton) checkQRIS(checkQRISButton.dataset.checkQris);

  const saveStatus = event.target.closest("[data-save-status]");
  if (saveStatus) updateOrderStatus(saveStatus.dataset.saveStatus);

  const deleteButton = event.target.closest("[data-delete-order]");
  if (deleteButton) deleteOrder(deleteButton.dataset.deleteOrder);

  const refreshPayments = event.target.closest("[data-refresh-payments]");
  if (refreshPayments) load();
});

document.querySelector("#search-input").addEventListener("input", () => {
  renderTable();
  renderOrderCards();
});

document.querySelector("#order-form").addEventListener("submit", createOrder);
document.addEventListener("submit", (event) => {
  const uploadForm = event.target.closest("[data-upload-order]");
  if (uploadForm) uploadMedia(event, uploadForm);
});
document.querySelector("#service-select").addEventListener("change", (event) => {
  const option = event.target.selectedOptions[0];
  document.querySelector("input[name='price']").value = option.dataset.price || 0;
});
document.querySelector("#drawer-close").addEventListener("click", closeDrawer);
document.querySelector("#drawer-x").addEventListener("click", closeDrawer);
document.addEventListener("keydown", (event) => {
  if (event.key === "Escape") closeSidebar();
});

document.querySelector("#login-form").addEventListener("submit", login);
document.querySelector("#logout-button").addEventListener("click", logout);

if (state.token) {
  document.querySelector("#login-screen").classList.add("hidden");
  load().catch((error) => {
    document.querySelector("#login-error").textContent = error.message;
  });
}
