# PRD.md

# Zolix Shoe Care Management System

---

# Product Requirement Document (PRD)

## Product Name

Zolix

## Product Type

Digital Laundry & Shoe Care Management Platform

## Platform

* Web Application
* Mobile Responsive
* Desktop Dashboard

---

# 1. Product Overview

Zolix adalah platform manajemen laundry sepatu modern yang dirancang untuk membantu operasional bisnis shoe care secara realtime, ringan, modern, dan scalable.

Aplikasi ini mendukung:

* Dashboard realtime
* Manajemen order
* Tracking pelanggan
* WhatsApp automation
* Invoice digital
* Upload before & after
* Multi item order
* Reporting bisnis

---

# 2. Product Goals

Tujuan utama aplikasi:

* Mempermudah operasional laundry sepatu
* Mengurangi proses manual
* Memberikan tracking realtime kepada pelanggan
* Mengintegrasikan WhatsApp automation
* Menyediakan invoice digital modern
* Mendukung pertumbuhan bisnis laundry

---

# 3. Target Users

## Super Admin

Mengelola seluruh sistem.

## Staff / Kasir

Mengelola order dan pembayaran.

## Cleaner

Mengupdate progres cleaning.

## Customer

Melihat status order dan invoice.

---

# 4. Technology Stack

## Backend

* Go (Golang)
* Fiber Framework
* GORM
* JWT Authentication
* WebSocket

## Frontend

* Next.js
* TypeScript
* TailwindCSS
* Zustand

## Database

* PostgreSQL

## Infrastructure

* Docker
* Docker Compose
* Redis
* Nginx

## Storage

* MinIO
  atau
* Supabase Storage

## WhatsApp Integration

* WAHA API

---

# 5. Core Features

# 5.1 Authentication

## Features

* Login email & password
* JWT Authentication
* Session management
* Role based access

## Roles

* Super Admin
* Staff
* Cleaner

---

# 5.2 Dashboard

## Dashboard Widgets

* Total order hari ini
* Pendapatan harian
* Pendapatan bulanan
* Order aktif
* Order selesai
* Statistik layanan
* Recent orders
* Realtime activity

## Realtime System

Menggunakan WebSocket.

---

# 5.3 Order Management

## Create Order

### Input Fields

* Nama pelanggan
* Nomor WhatsApp
* Jenis layanan
* Jumlah item
* Catatan
* Estimasi selesai
* Harga
* Metode pembayaran

---

## Multi Item Support

Satu order dapat memiliki:

* beberapa pasang sepatu
* beberapa layanan berbeda

---

## Order Status Flow

```text
Pending
↓
Diterima
↓
Cleaning
↓
Drying
↓
Finishing
↓
Ready Pickup
↓
Completed
```

---

# 5.4 Upload Before & After

## Features

* Multi upload
* Drag & drop
* Image compression
* Image optimization
* Before & after comparison

## Supported Format

* JPG
* PNG
* WEBP

---

# 5.5 Customer Tracking Page

## Features

Customer dapat:

* melihat status order
* melihat timeline progress
* melihat before/after
* melihat invoice
* download invoice PDF

## URL Example

```text
/order/TRX-2026-0001
```

---

# 5.6 Invoice System

## Features

* Invoice digital
* QR Code
* Download PDF
* Share WhatsApp
* Before & after preview

## Invoice Data

* Nama pelanggan
* Nomor invoice
* Detail layanan
* Total pembayaran
* Status pembayaran

---

# 5.7 WhatsApp Integration

Menggunakan:

* WAHA API

## Automation

* Order dibuat
* Order diproses
* Order selesai
* Reminder pickup
* Invoice terkirim

## Message Template

```text
Halo {{nama}}

Order sepatu Anda sedang diproses.

Status:
{{status}}

Track order:
{{link}}
```

---

# 5.8 Service Management

## Default Services

* Fast Clean
* Deep Clean
* Unyellowing
* Repaint
* Repair

## Features

* Edit harga
* Edit estimasi
* Aktif/nonaktif layanan

---

# 5.9 Payment System

## Payment Methods

* Cash
* Transfer
* QRIS

## Features

* Upload bukti transfer
* Status pembayaran
* Auto invoice

---

# 5.10 Reporting

## Reports

* Daily revenue
* Monthly revenue
* Service analytics
* Customer analytics
* Order analytics

## Export

* PDF
* Excel

---

# 6. Admin Modules

## Modules

* Dashboard
* Orders
* Customers
* Services
* Payments
* Gallery
* Reports
* Settings

---

# 7. Technical Architecture

# 7.1 Backend Architecture

## Main Stack

* Go
* Fiber
* PostgreSQL
* Redis
* WebSocket
* JWT

## Backend Responsibilities

* REST API
* Authentication
* Order Processing
* Media Management
* WhatsApp Integration
* Realtime Notification

---

# 7.2 Frontend Architecture

## Main Stack

* Next.js
* TailwindCSS
* Zustand

## Frontend Responsibilities

* Dashboard UI
* Order Management
* Customer Tracking
* Invoice Rendering

---

# 7.3 Infrastructure

## Services

* Docker
* Docker Compose
* Nginx
* Redis
* PostgreSQL
* MinIO

---

# 8. Suggested Folder Structure

```text
/backend
├── cmd
├── internal
│   ├── auth
│   ├── order
│   ├── customer
│   ├── invoice
│   ├── media
│   ├── whatsapp
│   ├── payment
│   └── dashboard
├── pkg
├── configs
├── migrations
├── storage
└── docker
```

---

# 9. API Design

# Authentication

```http
POST /api/v1/auth/login
POST /api/v1/auth/logout
POST /api/v1/auth/refresh
```

# Orders

```http
GET /api/v1/orders
POST /api/v1/orders
GET /api/v1/orders/:id
PUT /api/v1/orders/:id
DELETE /api/v1/orders/:id
```

# Upload

```http
POST /api/v1/upload
```

# Dashboard

```http
GET /api/v1/dashboard
```

---

# 10. Database Design

# users

```text
id
name
email
password
role
created_at
updated_at
```

# customers

```text
id
name
phone
created_at
```

# orders

```text
id
invoice_number
customer_id
status
total_price
payment_status
created_at
updated_at
```

# order_items

```text
id
order_id
service_id
shoe_name
qty
price
```

# services

```text
id
name
price
duration
```

# media

```text
id
order_id
type
url
```

---

# 11. Security

## Security Features

* JWT Authentication
* Rate Limiter
* Role Permission
* Upload Validation
* SQL Injection Protection
* Image Sanitization

---

# 12. Performance Goals

## Targets

* API response < 200ms
* Support 1000+ orders/day
* Optimized for ARM64
* Low RAM usage

---

# 13. Deployment Target

## Supported Platforms

* VPS
* HG680P
* Armbian ARM64
* Ubuntu Server

---

# 14. Future Features

## Planned Features

* AI stain detection
* Loyalty system
* Mobile app
* Pickup scheduling
* Courier integration
* Thermal printer support
* Barcode tracking
* QR tracking system

---

# 15. Success Metrics

## KPI

* Faster order processing
* Reduced admin workload
* Improved customer retention
* Better realtime visibility
* Reduced manual WhatsApp handling

---

# 16. Final Architecture

```text
Client (Mobile/Desktop)
        │
        ▼
Next.js Frontend
        │
 REST API + WebSocket
        ▼
Fiber Backend (Go)
        │
 ├── PostgreSQL
 ├── Redis
 ├── WAHA
 ├── MinIO/Supabase
 └── Worker Queue
```
