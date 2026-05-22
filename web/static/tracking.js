const trackingStatusMeta = {
  "Diterima": ["info", "⊙ DITERIMA"],
  "Cleaning": ["orange", "✺ PROSES"],
  "Drying": ["orange", "✺ DRYING"],
  "Finishing": ["orange", "✺ FINISHING"],
  "Ready Pickup": ["warning", "MENUNGGU DIAMBIL"],
  "Completed": ["success", "✓ SELESAI"],
  "Diambil": ["neutral", "▢ DIAMBIL"],
  "Menunggu Diambil": ["warning", "MENUNGGU DIAMBIL"],
  "Dibatalkan": ["danger", "⊗ DIBATALKAN"],
  "Pending": ["neutral", "PENDING"],
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

function badge(status) {
  const [name, label] = trackingStatusMeta[status] || ["neutral", status];
  return `<span class="badge ${name}">${label}</span>`;
}

function fmtDate(value) {
  const date = new Date(value);
  if (date.getFullYear() < 2000) return "Menunggu update";
  return `${trackingDate.format(date)} • ${trackingTime.format(date).replace(".", ":")} WIB`;
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
    return;
  }
  if (configResponse.ok) paymentConfig = await configResponse.json();
  const order = await response.json();
  currentOrder = order;
  document.title = `${order.invoice_number} - Zolix`;
  document.querySelector("#invoice-title").textContent = order.invoice_number;
  document.querySelector("#customer-line").textContent = `${order.customer_name} • ${order.customer_phone}`;
  document.querySelector("#tracking-status").innerHTML = badge(order.status);
  document.querySelector("#whatsapp-link").href = `https://wa.me/${order.customer_phone.replace(/\D/g, "")}`;

  document.querySelector("#tracking-timeline").innerHTML = order.timeline.map((step) => `
    <div class="timeline-item ${step.done ? "done" : ""}">
      <span class="dot"></span>
      <div>
        <strong>${step.label}</strong>
        <small>${step.done ? fmtDate(step.time) : "Menunggu update"}</small>
      </div>
    </div>
  `).join("");

  document.querySelector("#invoice-summary").innerHTML = `
    <div class="summary-row"><span>Status pembayaran</span><strong>${order.payment_status}</strong></div>
    <div class="summary-row"><span>Metode</span><strong>${order.payment_method}</strong></div>
    ${order.items.map((item) => `<div class="summary-row"><span>${item.service} × ${item.qty}</span><strong>${trackingRupiah.format(item.price * item.qty)}</strong></div>`).join("")}
    <div class="summary-row total"><span>Total</span><strong>${trackingRupiah.format(order.total_price)}</strong></div>
  `;
  renderSelfPayment(order);

  const media = order.media && order.media.length ? order.media : [];
  document.querySelector("#tracking-media").innerHTML = media.length
    ? media.map((item) => `<figure><img src="${item.url}" alt="${item.type}"><figcaption>${item.type}</figcaption></figure>`).join("")
    : `<p style="padding:0 20px 20px">Foto before/after belum tersedia.</p>`;
}

function renderSelfPayment(order) {
  const transfer = paymentConfig && paymentConfig.transfer ? paymentConfig.transfer : {};
  const cash = paymentConfig && paymentConfig.cash ? paymentConfig.cash : {};
  const qrisDisabled = paymentConfig && paymentConfig.qris_enabled === false;
  const qrisPanel = order.payment_method === "QRIS" && (order.payment_qr_url || order.payment_qr_string) ? `
    <div class="self-payment-result">
      <strong>QRIS siap dibayar</strong>
      <span>Ref: ${order.payment_reference || "-"}</span>
      ${order.payment_qr_url ? `<img src="${order.payment_qr_url}" alt="QRIS pembayaran">` : ""}
      ${order.payment_qr_string ? `<textarea readonly rows="3">${order.payment_qr_string}</textarea>` : ""}
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
        <span>${transfer.bank_name || "Transfer bank"}</span>
      </button>
      <button class="payment-method-card ${order.payment_method === "Cash" ? "active" : ""}" data-self-pay="Cash">
        <strong>Cash</strong>
        <span>Bayar di outlet</span>
      </button>
    </div>
    ${qrisPanel}
    ${order.payment_method === "Transfer" ? `
      <div class="manual-payment-info">
        <strong>${transfer.bank_name || "Transfer Bank"}</strong>
        <span>No. Rekening</span>
        <b>${transfer.account_number || "-"}</b>
        <span>Atas Nama</span>
        <b>${transfer.account_name || "Zolix Shoe Care"}</b>
        <p>${transfer.instructions || "Transfer sesuai total invoice lalu kirim bukti pembayaran melalui WhatsApp."}</p>
      </div>
    ` : ""}
    ${order.payment_method === "Cash" ? `
      <div class="manual-payment-info">
        <strong>Cash</strong>
        <p>${cash.instructions || "Bayar langsung di outlet saat pickup atau saat menyerahkan sepatu."}</p>
      </div>
    ` : ""}
    <p class="self-payment-note">Status saat ini: <strong>${order.payment_status}</strong></p>
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
    renderSelfPayment(currentOrder);
    document.querySelector("#invoice-summary").innerHTML = `
      <div class="summary-row"><span>Status pembayaran</span><strong>${currentOrder.payment_status}</strong></div>
      <div class="summary-row"><span>Metode</span><strong>${currentOrder.payment_method}</strong></div>
      ${currentOrder.items.map((item) => `<div class="summary-row"><span>${item.service} × ${item.qty}</span><strong>${trackingRupiah.format(item.price * item.qty)}</strong></div>`).join("")}
      <div class="summary-row total"><span>Total</span><strong>${trackingRupiah.format(currentOrder.total_price)}</strong></div>
    `;
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
    renderSelfPayment(currentOrder);
    await loadTracking();
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
