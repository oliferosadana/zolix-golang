ZOLIX SHOE CARE — UI/UX DESIGN SPECIFICATION

1. Brand Identity

Brand Name

ZOLIX Shoe Care

Brand Personality

Modern

Sporty

Premium

Clean

Fast

Professional

Design Direction

Minimal clean UI with:

White background

Neon lime accent

Black typography

Rounded modern components

Spacious layout

Soft shadows

2. Color System

Primary Colors

NameHexUsageNeon Lime#CBFF00Primary actionBlack#0A0A0AMain textWhite#FFFFFFBackground 

Secondary Colors

NameHexLight Background#F8F9FABorder Gray#E5E7EBText Gray#6B7280Card Gray#FAFAFA 

Status Colors

Success

background: #DCFCE7; color: #16A34A; 

Warning

background: #FEF3C7; color: #F59E0B; 

Danger

background: #FEE2E2; color: #EF4444; 

Info

background: #DBEAFE; color: #2563EB; 

3. Typography

Primary Font

font-family: "Inter", sans-serif; 

Secondary Font

font-family: "Montserrat", sans-serif; 

4. Font Scale

--text-xs: 12px; --text-sm: 14px; --text-base: 16px; --text-lg: 18px; --text-xl: 22px; --text-2xl: 28px; --text-3xl: 36px; 

5. Spacing System

4px 8px 12px 16px 20px 24px 32px 40px 48px 64px 

6. Border Radius

--radius-sm: 8px; --radius-md: 12px; --radius-lg: 16px; --radius-xl: 24px; 

7. Shadow System

Soft Shadow

box-shadow: 0 4px 12px rgba(0,0,0,0.04); 

Card Shadow

box-shadow: 0 8px 24px rgba(0,0,0,0.06); 

8. Layout System

Desktop Layout

Sidebar Width

280px 

Content Padding

32px 

Max Content Width

1600px 

9. Mobile Layout

Max Width

430px 

Bottom Navigation Height

72px 

Safe Padding

padding-bottom: 96px; 

10. Buttons

Primary Button

background: #CBFF00; color: #0A0A0A; height: 52px; border-radius: 14px; font-weight: 700; 

Secondary Button

background: #FFFFFF; border: 1px solid #E5E7EB; color: #111111; 

WhatsApp Button

background: #16A34A; color: #FFFFFF; 

11. Input Fields

height: 52px; border-radius: 14px; border: 1px solid #E5E7EB; padding: 0 16px; font-size: 14px; 

Focus State

border-color: #CBFF00; box-shadow: 0 0 0 3px rgba(203,255,0,0.2); 

12. Card Component

background: #FFFFFF; border-radius: 20px; padding: 24px; border: 1px solid #E5E7EB; box-shadow: 0 8px 24px rgba(0,0,0,0.04); 

13. Badge Component

Base Badge

padding: 6px 12px; border-radius: 999px; font-size: 12px; font-weight: 700; 

14. Service Badge

FastClean

background: #ECFCCB; color: #65A30D; 

Deep Clean

background: #E0F2FE; color: #0284C7; 

Unyellowing

background: #FEF3C7; color: #D97706; 

Repair

background: #F3E8FF; color: #9333EA; 

15. Navigation

Desktop Sidebar Menu

Dashboard

List Order

Tambah Order

Jadwal

Pelanggan

Layanan

Pengaturan

Mobile Bottom Navigation

Dashboard

Order

Tambah

Pelanggan

Akun

16. Dashboard Components

Summary Cards

Display:

Total Order

Order Proses

Order Selesai

Menunggu Diambil

Total Pendapatan

Charts

Recommended:

Line chart

Donut chart

Service popularity chart

17. Order Status

StatusColorDiterimaBlueProsesOrangeSelesaiGreenDiambilGrayMenunggu DiambilYellowDibatalkanRed 

18. Order Progress Timeline

Steps:

Diterima

Dicuci

Drying

Finishing

Selesai

Diambil

19. Upload Multi Photo

Required Photo Types

Before

Depan

Belakang

Samping kiri

Samping kanan

Sol bawah

Detail noda

After

Depan

Belakang

Samping kiri

Samping kanan

Sol bawah

20. Customer Tracking Page

Features:

Order status

Timeline progress

Before-after photos

Payment summary

Download PDF

WhatsApp support button

21. WhatsApp Integration

Features

Send invoice

Send status update

Pickup reminder

Share tracking page

Example Template

Halo Kak 👋 Sepatu Anda dengan nomor nota INV-250518-0012 sudah selesai dicuci dan siap diambil. Terima kasih telah menggunakan ZOLIX Shoe Care 🙌 

22. Recommended Tech Stack

Frontend

Next.js

React

TailwindCSS

Framer Motion

Icons

lucide-react

Backend

Supabase or

Node.js + PostgreSQL

23. Recommended Libraries

Table

TanStack Table

Form

React Hook Form

Zod

Modal

Radix UI

Upload

UploadThing or

Supabase Storage

Charts

Recharts

24. Animation Style

Use:

Smooth fade

Soft hover

Scale hover

Slide transition

Avoid:

Heavy animation

Flashing effects

Too many gradients

25. Responsive Breakpoints

Mobile: 0 - 640px Tablet: 641px - 1024px Desktop: 1025px+ 

26. Accessibility

Minimum text size: 14px

Contrast ratio AA

Large touch target

Keyboard accessible

Visible focus state

27. UI Design Rules

Use White Background

Primary UI should remain bright and clean.

Use Neon Lime Carefully

Only for:

CTA buttons

Active menu

Highlight numbers

Important status

Progress active state

Avoid Overusing Black

Use black mainly for:

Titles

Sidebar

Important text

28. Design Feeling

Target feeling:

Premium sneaker care app

Modern SaaS dashboard

Fast and clean workflow

Minimal but sporty

Inspired by:

Linear

Stripe

Notion

Nike

Sneaker marketplace UI

29. Suggested Folder Structure

/app /dashboard /orders /customers /services /settings /components /ui /cards /tables /forms /charts /lib /styles 

30. Final Notes

The UI should prioritize:

Speed

Readability

Mobile usability

Clean visual hierarchy

Easy scanning for admin

Premium customer experience

Brand focus: WHITE + BLACK + NEON LIME Minimal clean modern sneaker-care aesthetic.

