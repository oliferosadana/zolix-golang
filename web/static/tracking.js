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
  const response = await fetch(`/api/v1/public/orders/${encodeURIComponent(invoice)}`);
  if (!response.ok) {
    document.querySelector("#invoice-title").textContent = "Order tidak ditemukan";
    document.querySelector("#customer-line").textContent = "Periksa kembali nomor invoice Anda.";
    return;
  }
  const order = await response.json();
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

  const media = order.media && order.media.length ? order.media : [];
  document.querySelector("#tracking-media").innerHTML = media.length
    ? media.map((item) => `<figure><img src="${item.url}" alt="${item.type}"><figcaption>${item.type}</figcaption></figure>`).join("")
    : `<p style="padding:0 20px 20px">Foto before/after belum tersedia.</p>`;
}

loadTracking();
