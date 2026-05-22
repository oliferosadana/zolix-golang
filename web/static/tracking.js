const trackingStatusMeta = {
  Diterima: ["info", "DITERIMA"],
  Cleaning: ["orange", "PROSES"],
  Drying: ["orange", "DRYING"],
  Finishing: ["orange", "FINISHING"],
  "Ready Pickup": ["warning", "MENUNGGU DIAMBIL"],
  Completed: ["success", "SELESAI"],
  Diambil: ["neutral", "DIAMBIL"],
  "Menunggu Diambil": ["warning", "MENUNGGU DIAMBIL"],
  Dibatalkan: ["danger", "DIBATALKAN"],
  Pending: ["neutral", "PENDING"],
};

const trackingRupiah = new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR", maximumFractionDigits: 0 });
const trackingDate = new Intl.DateTimeFormat("id-ID", { day: "2-digit", month: "long", year: "numeric" });
const trackingTime = new Intl.DateTimeFormat("id-ID", { hour: "2-digit", minute: "2-digit" });
let currentOrder = null;
let paymentConfig = null;

function invoiceFromPath() {
  const parts = window.location.pathname.split("/");
  return decodeURIComponent(parts[parts.length - 1] || "");
}

function safeText(value) {
  return String(value ?? "").replace(/[&<>"']/g, (char) => ({
    "&": "&amp;",
    "<": "&lt;",
    ">": "&gt;",
    '"': "&quot;",
    "'": "&#39;",
  }[char]));
}

function badge(status) {
  const [name, label] = trackingStatusMeta[status] || ["neutral", status || "PENDING"];
  return `<span class="badge ${name}">${safeText(label)}</span>`;
}

function fmtDate(value) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime()) || date.getFullYear() < 2000) return "Menunggu update";
  return `${trackingDate.format(date)} - ${trackingTime.format(date).replace(".", ":")} WIB`;
}

function paymentClass(status) {
  const normalized = String(status || "").toLowerCase();
  if (normalized.includes("lunas") || normalized.includes("paid")) return "success";
  if (normalized.includes("batal") || normalized.includes("gagal")) return "danger";
  return "warning";
}

function renderOverview(order) {
  document.querySelector("#tracking-overview").innerHTML = `
    <div class="tracking-overview-item">
      <span>Total Tagihan</span>
      <strong>${trackingRupiah.format(order.total_price || 0)}</strong>
    </div>
    <div class="tracking-overview-item">
      <span>Pembayaran</span>
      <strong class="payment-state ${paymentClass(order.payment_status)}">${safeText(order.payment_status || "-")}</strong>
    </div>
    <div class="tracking-overview-item">
      <span>Metode</span>
      <strong>${safeText(order.payment_method || "-")}</strong>
    </div>
  `;
}

function renderInvoiceSummary(order) {
  const itemRows = (order.items || []).map((item) => `
    <div class="invoice-item">
      <div>
        <strong>${safeText(item.service || "Layanan")}</strong>
        <span>${safeText(item.shoe_name || "Sepatu")} - Qty ${safeText(item.qty || 0)}</span>
      </div>
      <b>${trackingRupiah.format((item.price || 0) * (item.qty || 0))}</b>
    </div>
  `).join("");

  document.querySelector("#invoice-summary").innerHTML = `
    <div class="invoice-total-card">
      <span>Total Invoice</span>
      <strong>${trackingRupiah.format(order.total_price || 0)}</strong>
    </div>
    <div class="invoice-detail-list">
      <div class="summary-row"><span>Invoice</span><strong>${safeText(order.invoice_number)}</strong></div>
      <div class="summary-row"><span>Pelanggan</span><strong>${safeText(order.customer_name)}</strong></div>
      <div class="summary-row"><span>Status pembayaran</span><strong>${safeText(order.payment_status || "-")}</strong></div>
      <div class="summary-row"><span>Metode</span><strong>${safeText(order.payment_method || "-")}</strong></div>
    </div>
    <div class="invoice-items">
      ${itemRows || `<p class="empty-state">Item layanan belum tersedia.</p>`}
    </div>
  `;
}

function renderTimeline(order) {
  document.querySelector("#tracking-timeline").innerHTML = (order.timeline || []).map((step) => `
    <div class="timeline-item ${step.done ? "done" : ""}">
      <span class="dot"></span>
      <div>
        <strong>${safeText(step.label)}</strong>
        <small>${step.done ? fmtDate(step.time) : "Menunggu update"}</small>
      </div>
    </div>
  `).join("");
}

function renderMedia(order) {
  const media = order.media && order.media.length ? order.media : [];
  document.querySelector("#tracking-media").innerHTML = media.length
    ? media.map((item) => `<figure><img src="${safeText(item.url)}" alt="${safeText(item.type)}"><figcaption>${safeText(item.type)}</figcaption></figure>`).join("")
    : `<div class="empty-state">Foto before/after belum tersedia.</div>`;
}

function refreshOrderView(order) {
  currentOrder = order;
  document.title = `${order.invoice_number} - Zolix`;
  document.querySelector("#invoice-title").textContent = order.invoice_number;
  document.querySelector("#customer-line").textContent = `${order.customer_name} - ${order.customer_phone}`;
  document.querySelector("#tracking-status").innerHTML = badge(order.status);
  document.querySelector("#whatsapp-link").href = `https://wa.me/${String(order.customer_phone || "").replace(/\D/g, "")}`;
  renderOverview(order);
  renderTimeline(order);
  renderInvoiceSummary(order);
  renderSelfPayment(order);
  renderMedia(order);
}

async function loadTracking() {
  const invoice = invoiceFromPath();
  const [response, configResponse] = await Promise.all([
    fetch(`/api/v1/public/orders/${encodeURIComponent(invoice)}`),
    fetch("/api/v1/public/payment-config"),
  ]);
  if (!response.ok) {
    document.querySelector("#invoice-title").textContent = "Order tidak ditemukan";
    document.querySelector("#customer-line").textContent = "Periksa kembali nomor invoice Anda.";
    document.querySelector("#tracking-status").innerHTML = badge("Dibatalkan");
    document.querySelector("#tracking-overview").innerHTML = "";
    return;
  }
  if (configResponse.ok) paymentConfig = await configResponse.json();
  refreshOrderView(await response.json());
}

function renderSelfPayment(order) {
  const transfer = paymentConfig && paymentConfig.transfer ? paymentConfig.transfer : {};
  const cash = paymentConfig && paymentConfig.cash ? paymentConfig.cash : {};
  const qrisDisabled = paymentConfig && paymentConfig.qris_enabled === false;
  const qrisPanel = order.payment_method === "QRIS" && (order.payment_qr_url || order.payment_qr_string) ? `
    <div class="self-payment-result">
      <strong>QRIS siap dibayar</strong>
      <span>Ref: ${safeText(order.payment_reference || "-")}</span>
      ${order.payment_qr_url ? `<img src="${safeText(order.payment_qr_url)}" alt="QRIS pembayaran">` : ""}
      ${order.payment_qr_string ? `<textarea readonly rows="3">${safeText(order.payment_qr_string)}</textarea>` : ""}
      <button class="secondary-button" data-self-check-qris>Cek Status Pembayaran</button>
    </div>
  ` : "";

  document.querySelector("#self-payment").innerHTML = `
    <div class="payment-method-grid">
      <button class="payment-method-card ${order.payment_method === "QRIS" ? "active" : ""}" data-self-pay="QRIS" ${qrisDisabled ? "disabled" : ""}>
        <strong>QRIS</strong>
        <span>${qrisDisabled ? "Belum dikonfigurasi" : "Bayar instan dengan QRIS"}</span>
      </button>
      <button class="payment-method-card ${order.payment_method === "Transfer" ? "active" : ""}" data-self-pay="Transfer">
        <strong>Transfer</strong>
        <span>${safeText(transfer.bank_name || "Transfer bank")}</span>
      </button>
      <button class="payment-method-card ${order.payment_method === "Cash" ? "active" : ""}" data-self-pay="Cash">
        <strong>Cash</strong>
        <span>Bayar di outlet</span>
      </button>
    </div>
    ${qrisPanel}
    ${order.payment_method === "Transfer" ? `
      <div class="manual-payment-info">
        <strong>${safeText(transfer.bank_name || "Transfer Bank")}</strong>
        <span>No. Rekening</span>
        <b>${safeText(transfer.account_number || "-")}</b>
        <span>Atas Nama</span>
        <b>${safeText(transfer.account_name || "Zolix Shoe Care")}</b>
        <p>${safeText(transfer.instructions || "Transfer sesuai total invoice lalu kirim bukti pembayaran melalui WhatsApp.")}</p>
      </div>
    ` : ""}
    ${order.payment_method === "Cash" ? `
      <div class="manual-payment-info">
        <strong>Cash</strong>
        <p>${safeText(cash.instructions || "Bayar langsung di outlet saat pickup atau saat menyerahkan sepatu.")}</p>
      </div>
    ` : ""}
    <p class="self-payment-note">Status saat ini: <strong>${safeText(order.payment_status || "-")}</strong></p>
  `;
}

async function trackingPaymentRequest(path, payload) {
  const response = await fetch(path, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  const body = await response.json().catch(() => ({}));
  if (!response.ok) throw new Error(body.error || "Request pembayaran gagal");
  return body;
}

async function selectSelfPayment(method) {
  if (!currentOrder) return;
  try {
    if (method === "QRIS") {
      const result = await trackingPaymentRequest("/api/v1/public/payments/qris/generate", {
        invoice_number: currentOrder.invoice_number,
      });
      currentOrder = result.order;
    } else {
      const result = await trackingPaymentRequest("/api/v1/public/payments/method", {
        invoice_number: currentOrder.invoice_number,
        method,
      });
      currentOrder = result.order;
    }
    renderOverview(currentOrder);
    renderInvoiceSummary(currentOrder);
    renderSelfPayment(currentOrder);
  } catch (error) {
    alert(error.message || "Gagal memilih pembayaran");
  }
}

async function checkSelfQRIS() {
  if (!currentOrder) return;
  try {
    const result = await trackingPaymentRequest("/api/v1/public/payments/qris/status", {
      invoice_number: currentOrder.invoice_number,
    });
    currentOrder = result.order;
    renderOverview(currentOrder);
    renderInvoiceSummary(currentOrder);
    renderSelfPayment(currentOrder);
  } catch (error) {
    alert(error.message || "Gagal mengecek pembayaran");
  }
}

document.addEventListener("click", (event) => {
  const paymentButton = event.target.closest("[data-self-pay]");
  if (paymentButton && !paymentButton.disabled) selectSelfPayment(paymentButton.dataset.selfPay);

  const checkButton = event.target.closest("[data-self-check-qris]");
  if (checkButton) checkSelfQRIS();
});

loadTracking();
