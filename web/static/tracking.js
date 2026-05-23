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
  if (Number.isNaN(date.getTime()) || date.getFullYear() < 2000) return "Pending";
  return `${trackingDate.format(date)}<br>${trackingTime.format(date).replace(".", ":")}`;
}

function plainDate(value) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime()) || date.getFullYear() < 2000) return "Menunggu update";
  return `${trackingDate.format(date)}, ${trackingTime.format(date).replace(".", ":")}`;
}

function paymentClass(status) {
  const normalized = String(status || "").toLowerCase();
  if (normalized.includes("lunas") || normalized.includes("paid")) return "success";
  if (normalized.includes("batal") || normalized.includes("gagal")) return "danger";
  return "warning";
}

function whatsappUrl(order, message) {
  const phone = String(order.customer_phone || "").replace(/\D/g, "");
  const text = encodeURIComponent(message || `Halo ZOLIX, saya ingin menanyakan nota ${order.invoice_number}.`);
  return phone ? `https://wa.me/${phone}?text=${text}` : "#";
}

function timelineIcon(label) {
  const normalized = String(label || "").toLowerCase();
  if (normalized.includes("terima")) return "✓";
  if (normalized.includes("cuci") || normalized.includes("clean")) return "◉";
  if (normalized.includes("dry")) return "✽";
  if (normalized.includes("finish")) return "✦";
  if (normalized.includes("ambil") || normalized.includes("pickup")) return "□";
  return "•";
}

function statusMessage(order) {
  const status = String(order.status || "").toLowerCase();
  if (status.includes("completed") || status.includes("selesai") || status.includes("ready")) {
    return ["Pesanan Anda sudah siap!", "Silakan lakukan pengambilan sesuai instruksi dari ZOLIX Shoe Care."];
  }
  if (status.includes("batal")) {
    return ["Pesanan dibatalkan.", "Silakan hubungi tim ZOLIX jika membutuhkan bantuan lanjutan."];
  }
  return ["Pesanan Anda sedang diproses!", "Kami sedang mengerjakan sepatu Anda dengan maksimal. Notifikasi akan dikirim jika pesanan sudah selesai."];
}

function renderTimeline(order) {
  const timeline = order.timeline && order.timeline.length ? order.timeline : [
    { label: "Diterima", done: true, time: order.created_at },
    { label: "Dicuci", done: false },
    { label: "Drying", done: false },
    { label: "Finishing", done: false },
    { label: "Diambil", done: false },
  ];
  const doneCount = timeline.filter((step) => step.done).length;
  const progress = Math.max(0, Math.min(100, timeline.length > 1 ? ((doneCount - 1) / (timeline.length - 1)) * 100 : 0));

  document.querySelector("#tracking-timeline").innerHTML = `
    <div class="nota-progress-line"><i style="width:${progress}%"></i></div>
    ${timeline.map((step) => `
      <div class="nota-progress-step ${step.done ? "done" : ""}">
        <div class="nota-step-dot">${timelineIcon(step.label)}</div>
        <strong>${safeText(step.label)}</strong>
        <span>${step.done ? fmtDate(step.time) : "Pending"}</span>
      </div>
    `).join("")}
  `;
}

function renderOrderMeta(order) {
  document.querySelector("#order-meta").innerHTML = `
    <div><span>▣</span><div><p>Tanggal Order</p><strong>${plainDate(order.created_at)}</strong></div></div>
    <div><span>◷</span><div><p>Estimasi Selesai</p><strong>${plainDate(order.estimated_done_at || order.estimated_at)}</strong></div></div>
    <div><span>◎</span><div><p>Lokasi</p><strong>${safeText(order.location || "Balikpapan")}</strong></div></div>
    <div><span>▤</span><div><p>Metode Pembayaran</p><strong class="nota-payment-state ${paymentClass(order.payment_status)}">${safeText(order.payment_status || "-")}</strong></div></div>
  `;
}

function renderCustomer(order) {
  document.querySelector("#customer-details").innerHTML = `
    <span>Nama</span><strong>${safeText(order.customer_name || "-")}</strong>
    <span>No. WhatsApp</span><strong>${safeText(order.customer_phone || "-")}</strong>
    <span>Alamat</span><strong>${safeText(order.customer_address || order.address || "-")}</strong>
  `;
}

function renderItems(order) {
  const items = order.items || [];
  document.querySelector("#invoice-summary").innerHTML = `
    <div class="nota-item-head">
      <span>Item</span><span>Layanan</span><span>Qty</span><span>Harga Satuan</span><span>Subtotal</span>
    </div>
    ${items.length ? items.map((item, index) => {
      const qty = item.qty || 0;
      const price = item.price || 0;
      return `
        <div class="nota-item-row">
          <div class="nota-item-product">
            <div class="nota-shoe-thumb ${index % 2 ? "dark" : ""}"></div>
            <div>
              <strong>${safeText(item.shoe_name || "Sepatu")}</strong>
              <small>${safeText(item.note || item.size || "")}</small>
            </div>
          </div>
          <span class="nota-service-pill">${safeText(item.service || "Layanan")}</span>
          <span>${safeText(qty)} Pasang</span>
          <span>${trackingRupiah.format(price)}</span>
          <strong>${trackingRupiah.format(price * qty)}</strong>
        </div>
      `;
    }).join("") : `<div class="empty-state">Item layanan belum tersedia.</div>`}
  `;
}

function renderPaymentSummary(order) {
  const subtotal = (order.items || []).reduce((sum, item) => sum + ((item.price || 0) * (item.qty || 0)), 0);
  const total = order.total_price || subtotal;
  const discount = Math.max(0, subtotal - total);

  document.querySelector("#payment-summary").innerHTML = `
    <div class="payment-line"><span>Subtotal</span><strong>${trackingRupiah.format(subtotal || total)}</strong></div>
    <div class="payment-line"><span>Diskon</span><strong class="danger">${discount ? `- ${trackingRupiah.format(discount)}` : trackingRupiah.format(0)}</strong></div>
    <div class="payment-line total"><span>Total</span><strong>${trackingRupiah.format(total)}</strong></div>
    <div class="payment-grand"><span>Total yang harus dibayar</span><strong>${trackingRupiah.format(total)}</strong></div>
  `;
}

function renderMedia(order) {
  const media = order.media && order.media.length ? order.media : [];
  if (!media.length) {
    document.querySelector("#tracking-media").innerHTML = `<div class="empty-state">Foto before/after belum tersedia.</div>`;
    return;
  }

  const before = media.filter((item) => String(item.type || "").toLowerCase().includes("before"));
  const after = media.filter((item) => String(item.type || "").toLowerCase().includes("after"));
  const fallbackBefore = before.length ? before : media.slice(0, Math.ceil(media.length / 2));
  const fallbackAfter = after.length ? after : media.slice(Math.ceil(media.length / 2));
  const group = (title, list) => `
    <div class="nota-media-group">
      <strong>${title}</strong>
      <div>
        ${list.slice(0, 5).map((item) => `<figure><img src="${safeText(item.url)}" alt="${safeText(item.type || title)}"><figcaption>${safeText(item.label || item.type || title)}</figcaption></figure>`).join("")}
      </div>
    </div>
  `;
  document.querySelector("#tracking-media").innerHTML = `${group("Before", fallbackBefore)}${group("After", fallbackAfter)}`;
}

function refreshOrderView(order) {
  currentOrder = order;
  const [title, body] = statusMessage(order);
  document.title = `${order.invoice_number} - Zolix`;
  document.querySelector("#invoice-title").textContent = order.invoice_number;
  document.querySelector("#invoice-subtitle").textContent = order.invoice_number;
  document.querySelector("#tracking-status").innerHTML = badge(order.status);
  document.querySelector("#whatsapp-link").href = whatsappUrl(order);
  document.querySelector("#confirm-payment-link").href = whatsappUrl(order, `Halo ZOLIX, saya sudah melakukan pembayaran untuk nota ${order.invoice_number}.`);
  document.querySelector("#status-message-title").textContent = title;
  document.querySelector("#status-message-body").textContent = body;
  renderTimeline(order);
  renderOrderMeta(order);
  renderCustomer(order);
  renderItems(order);
  renderPaymentSummary(order);
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
    document.querySelector("#invoice-subtitle").textContent = "Periksa kembali nomor invoice Anda.";
    document.querySelector("#tracking-status").innerHTML = badge("Dibatalkan");
    return;
  }
  if (configResponse.ok) paymentConfig = await configResponse.json();
  refreshOrderView(await response.json());
}

function renderSelfPayment(order) {
  const transfer = paymentConfig && paymentConfig.transfer ? paymentConfig.transfer : {};
  const cash = paymentConfig && paymentConfig.cash ? paymentConfig.cash : {};
  const qrisDisabled = paymentConfig && paymentConfig.qris_enabled === false;
  const qrisDetail = order.payment_method === "QRIS" && (order.payment_qr_url || order.payment_qr_string) ? `
    <div class="nota-payment-detail">
      <strong>QRIS siap dibayar</strong>
      <span>Ref: ${safeText(order.payment_reference || "-")}</span>
      ${order.payment_qr_url ? `<img src="${safeText(order.payment_qr_url)}" alt="QRIS pembayaran">` : ""}
      ${order.payment_qr_string ? `<textarea readonly rows="3">${safeText(order.payment_qr_string)}</textarea>` : ""}
      <button type="button" data-self-check-qris>Cek Status Pembayaran</button>
    </div>
  ` : "";

  document.querySelector("#self-payment").innerHTML = `
    <button class="nota-method ${order.payment_method === "QRIS" ? "active" : ""}" data-self-pay="QRIS" ${qrisDisabled ? "disabled" : ""}>
      <span class="method-icon">▦</span>
      <span class="method-copy">
        <strong>QRIS</strong>
        <small>${qrisDisabled ? "Belum dikonfigurasi" : "Scan QR Code menggunakan aplikasi e-wallet atau mobile banking."}</small>
        <em>Mudah & Instan</em>
      </span>
      ${order.payment_qr_url ? `<img class="method-qr" src="${safeText(order.payment_qr_url)}" alt="QRIS pembayaran">` : `<i class="method-qr-placeholder"></i>`}
      <b>›</b>
    </button>
    <button class="nota-method ${order.payment_method === "Transfer" ? "active" : ""}" data-self-pay="Transfer">
      <span class="method-icon">⌂</span>
      <span class="method-copy">
        <strong>Transfer Bank</strong>
        <small>${safeText(transfer.bank_name || "Transfer ke rekening resmi ZOLIX Shoe Care")}<br>${safeText(transfer.account_number || "1234 5678 9012")}<br>${safeText(transfer.account_name || "ZOLIX SHOE CARE")}</small>
      </span>
      <b>›</b>
    </button>
    <button class="nota-method ${order.payment_method === "Cash" ? "active" : ""}" data-self-pay="Cash">
      <span class="method-icon">◎</span>
      <span class="method-copy">
        <strong>Cash (Tunai)</strong>
        <small>${safeText(cash.instructions || "Bayar tunai saat pengambilan sepatu di store ZOLIX Shoe Care.")}</small>
        <em>Bayar saat ambil</em>
      </span>
      <b>›</b>
    </button>
    ${qrisDetail}
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
    renderPaymentSummary(currentOrder);
    renderSelfPayment(currentOrder);
    renderOrderMeta(currentOrder);
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
    renderPaymentSummary(currentOrder);
    renderSelfPayment(currentOrder);
    renderOrderMeta(currentOrder);
  } catch (error) {
    alert(error.message || "Gagal mengecek pembayaran");
  }
}

document.addEventListener("click", (event) => {
  const paymentButton = event.target.closest("[data-self-pay]");
  if (paymentButton && !paymentButton.disabled) selectSelfPayment(paymentButton.dataset.selfPay);

  const checkButton = event.target.closest("[data-self-check-qris]");
  if (checkButton) checkSelfQRIS();

  if (event.target.closest("#download-nota")) window.print();
});

loadTracking();
