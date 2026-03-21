# PraestOS UX Specification
_Source: Stitch designs provided by Matthias, 2026-03-20_

## Design Language

**Palette:**
- Background: `#F7F8FA` (main), `#FAFBFC` (sidebar)
- Cards: `#FFFFFF` with `1px #E5E7EB` border, `border-radius: 8px`
- Primary CTA: `#1B3A4B` (dark teal-navy), white text
- Text primary: `#1A1D23` / secondary: `#6B7280` / labels: `#7C8DB0`
- Status: green `#22C55E` / amber `#F59E0B` / red `#EF4444` / blue `#3B82F6`

**Typography:**
- Font: Inter (or DM Sans)
- Page titles: 36-44px bold, sometimes with trailing period
- Section headings: 20-24px bold
- Table headers: 11-12px uppercase, letter-spaced, semi-bold
- Body: 14-15px regular
- Labels/categories: 11px uppercase, letter-spaced

## Layout Shell

```
┌─────────────────────────────────────────────┐
│  Sidebar (210px) │  Top Bar (56px)           │
│  ───────────────  ──────────────────────── │
│  Logo + name     │  Nav tabs  Search  Bell  │
│  ───────────────  ──────────────────────── │
│  MAIN HUB        │                          │
│  Dashboard       │  Breadcrumb              │
│  ───────────────  Page Title.               │
│  SALES           │  Description             │
│  Pipeline        │                          │
│  Contacts        │  [Content]               │
│  Accounts        │                          │
│  ───────────────  ──────────────────────── │
│  Settings        │  Pagination / footer     │
│  Sign Out        │                          │
└─────────────────────────────────────────────┘
```

## Pages to Build

### 1. Navigation Sidebar
- Logo + "PraestOS" + "CRM ENTERPRISE" subtitle
- Groups: MAIN HUB, SALES OPERATIONS, SERVICE, INSIGHTS
- Items: Dashboard, Pipeline, Contacts, Accounts, Leads, Tickets, Reports
- Active state: `#E8EDF2` bg, bolder text
- Bottom: Settings, Sign Out
- Footer: "+ New Entry" dark CTA button

### 2. Dashboard Overview
- KPI cards row: Total Revenue, Pipeline Value, Active Deals, Open Tickets
- Bar chart (muted blue-gray bars)
- Quick Actions 2×2 grid: Add Lead, Schedule, Email, Invoice
- Task checklist card
- Activity feed

### 3. Ticket List
- Filter bar: Status, Priority, Assigned To dropdowns
- Table: Subject, Contact (avatar+name), Status badge, Priority dot+text, Assigned To, Created At
- Status badges: OPEN (teal), PENDING (amber), RESOLVED (green), CLOSED (gray)
- Priority dots: Critical=red, High=orange, Medium=blue, Low=gray
- Summary metrics: Avg Response Time, Resolution Rate, First Touch Resolution

### 4. Ticket Detail (3-column)
- Left: Ticket thread + reply composer
  - Messages differentiated: client (gray), agent (light blue), internal (yellow + lock icon)
  - Rich text toolbar
  - "INTERNAL NOTE" checkbox
  - "Submit Reply" primary + "Save as Draft" secondary
- Right panel:
  - Contact info card (avatar, name, email, phone, location)
  - Ticket properties: Assigned To, Source, Category
  - SLA status with overdue bar (red when overdue)
  - Internal tags with "+ ADD TAG"

### 5. Contact Detail
- Header: photo/avatar, name, title, company (link), Edit + Send Email actions
- Interaction timeline (icon-differentiated: call, email, meeting)
- Open opportunities mini-cards with pipeline progress bars
- Private notes (lock icon, auto-save)
- Record footer: ID, Created, Modified, Export PDF, Audit Log, Delete (red)

### 6. Accounts List
- Table with icon avatars, industry badges (AEROSPACE, FINTECH, etc.)
- Activity status dots: green=Active, amber=Pending, gray=Archived
- Export + New Account actions

### 7. Leads List
- Similar to Contacts/Accounts table pattern
- Lead score indicator
- Conversion status

## Component Patterns

**KPI Card:**
```
┌─────────────────┐
│ LABEL (caps)    │
│ $1,284,000      │  ← 36-48px bold
│ ↑ +12% trend    │
└─────────────────┘
```

**Status Badge:** filled pill, 6px v-padding, 12px h-padding, white text

**Table Row:** 60px height, avatar + name, colored status badge, colored priority dot

**Right Detail Panel:** stacked cards per section, icon-labeled fields, chevron dropdowns

## Naming Convention
The designs use "Rational Archive" branding — for PraestOS use "PraestOS" in all UI strings.
