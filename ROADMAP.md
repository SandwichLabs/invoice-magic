# Invoice Magic Product Roadmap
## For Independent Service Professionals

### Current State Summary

Invoice Magic is a solid foundation: a stateless Go CLI/web app that generates professional PDF invoices from JSON or Google Sheets via Typst templating. It has OAuth2 auth, template management, and a basic web UI for previewing/downloading invoices.

---

## Vision: "Data Lives Anywhere" Invoice Solution

The core insight: independent professionals don't want to learn new systems. They already have data scattered across text messages, notes apps, handwritten tickets, photos of job sites, and maybe a spreadsheet. Invoice Magic should meet them where their data lives.

---

## Phase 1: Field-Ready Mobile Experience

**Problem:** Plumbers and technicians are in the field, not at desks. They need to create invoices from a phone immediately after completing work.

### 1.1 Progressive Web App (PWA)
- Offline-capable web interface that works on any phone
- Install-to-homescreen capability
- Local storage sync when connectivity returns
- Camera integration for capturing job photos

### 1.2 SMS/Text-to-Invoice
- Dedicated phone number (Twilio integration)
- Text: `"Johnson residence, replaced water heater, $450 labor, $890 parts"`
- AI parses natural language → structured invoice
- Reply with PDF link or "Confirm to send to customer?"

### 1.3 Voice-to-Invoice
- Call in and dictate invoice details
- Speech-to-text → structured data
- "Hey, just finished at 123 Main St for the Smiths. Replaced the garbage disposal, took 2 hours at 85 an hour, plus 180 for the disposal unit."

### 1.4 Quick Estimate → Invoice Conversion
- Create estimates in the field
- Customer approval flow (text/email link)
- One-click conversion to invoice when job completes
- Track estimate-to-invoice conversion rate

---

## Phase 2: Universal Data Source Connectors

**Problem:** "Data living anywhere" means connecting to wherever professionals already store information.

### 2.1 Spreadsheet Connectors (Beyond Google Sheets)
- Microsoft Excel Online / OneDrive
- Apple Numbers / iCloud
- Airtable bases
- Notion databases
- Local CSV/Excel file watch (desktop sync)

### 2.2 Business Tool Integrations
- **QuickBooks** - sync customers, push invoices
- **Wave** - free accounting software popular with small businesses
- **Square** - payments and customer data
- **Jobber/ServiceTitan** - field service management platforms
- **Housecall Pro** - scheduling integration

### 2.3 CRM/Contact Integrations
- Google Contacts → auto-populate customer info
- Apple Contacts sync
- Import from phone contacts directly

### 2.4 Local-First SQLite Option
- Embedded SQLite database for offline-first operation
- Sync to cloud storage (Dropbox, iCloud, Google Drive) as backup
- No vendor lock-in, data always exportable

---

## Phase 3: Smart Invoice Features

**Problem:** Professionals waste time on repetitive data entry and calculations.

### 3.1 Part/Service Catalog
- Pre-defined parts with costs (auto-update from suppliers?)
- Labor rate presets (hourly, flat-rate jobs)
- Common job templates: "Standard water heater install", "HVAC tune-up"
- Quick-add from catalog while creating invoice

### 3.2 Photo Documentation
- Attach before/after photos to invoices
- Photo of serial numbers → auto-extract via OCR
- Photo of parts receipt → itemize automatically
- Store photos with invoice record for warranty claims

### 3.3 Smart Pricing
- Material markup calculator (cost + X%)
- Travel time/mileage tracking and billing
- Tiered pricing (emergency rates, weekend rates)
- Tax calculation by jurisdiction (US state sales tax complexity)

### 3.4 Recurring Invoice Automation
- Maintenance contracts (monthly/quarterly/annual)
- Auto-generate and send on schedule
- Payment reminder automation
- Contract renewal tracking

---

## Phase 4: Payment Collection

**Problem:** Creating invoices is only half the job. Getting paid is the other half.

### 4.1 Payment Links
- Stripe integration for card payments
- Square payment links
- PayPal invoicing
- Venmo/Zelle QR codes on invoice PDF
- ACH/bank transfer instructions

### 4.2 Payment Tracking
- Mark invoices paid (partial payments supported)
- Payment method recording
- Automatic reconciliation with Stripe/Square
- Outstanding balance dashboard

### 4.3 Late Payment Automation
- Configurable reminder schedule (3 days, 7 days, 14 days)
- Escalating reminder tone
- Late fee calculation and application
- "Final notice" templates

### 4.4 Deposit/Progress Billing
- Collect deposits before starting work
- Progress invoices for larger jobs
- Deposit credit on final invoice

---

## Phase 5: Customer Experience

**Problem:** Professional appearance builds trust and reduces payment friction.

### 5.1 Customer Portal
- Unique link per customer
- View all their invoices/estimates
- Pay directly from portal
- Download receipts
- Request service (loops back to estimate)

### 5.2 White-Label Branding
- Custom logo placement
- Brand colors in templates
- Custom email domain for sending
- Remove "Invoice Magic" branding

### 5.3 Multi-Channel Delivery
- Email with PDF attachment
- SMS with payment link
- Print-ready formatting for mail
- WhatsApp integration (huge internationally)

### 5.4 Review/Rating Request
- Post-payment satisfaction survey
- "Leave a Google review" prompt with direct link
- Testimonial collection for marketing

---

## Phase 6: Business Intelligence

**Problem:** Professionals need insights without becoming data analysts.

### 6.1 Simple Dashboard
- Revenue this week/month/year
- Outstanding receivables aging
- Top customers by revenue
- Average job value trends

### 6.2 Tax Preparation
- Quarterly revenue summaries
- Expense categorization (if parts tracking enabled)
- Export for accountant (CSV, QBO format)
- 1099 tracking for subcontractors

### 6.3 Job Costing
- Track actual vs estimated time
- Material cost tracking
- Profitability per job type
- "Jobs like this usually take X hours"

---

## Phase 7: Multi-User & Team Features

**Problem:** Growing businesses have multiple technicians.

### 7.1 Technician Accounts
- Each tech can create invoices
- Admin approval workflow (optional)
- Commission/payout tracking
- Individual performance metrics

### 7.2 Dispatcher Integration
- Assign jobs to technicians
- Tech completes job → invoice auto-generated
- Route optimization suggestions
- Real-time job status

### 7.3 Subcontractor Management
- Issue purchase orders
- Track subcontractor invoices received
- Net payment calculation

---

## Technical Architecture Recommendations

### Preserve Core Strengths
- Keep stateless PDF generation (current Typst pipeline is excellent)
- Maintain CLI for power users and automation
- Single-binary distribution is a huge advantage

### Add Selective State
```
┌─────────────────────────────────────────────────────┐
│                  Invoice Magic                       │
├─────────────────────────────────────────────────────┤
│  Data Sources        │  Core Engine    │  Outputs   │
│  ─────────────       │  ───────────    │  ───────   │
│  • Google Sheets ✓   │                 │  • PDF ✓   │
│  • Excel/OneDrive    │  Normalize →    │  • HTML ✓  │
│  • Airtable          │  Validate →     │  • Email   │
│  • SQLite (local)    │  Render (Typst) │  • SMS     │
│  • SMS inbound       │                 │  • Portal  │
│  • Voice inbound     │                 │            │
│  • REST API          │                 │            │
└─────────────────────────────────────────────────────┘
```

### Suggested Tech Additions
- **SQLite** - local-first data persistence (go-sqlite, no CGO with modernc.org/sqlite)
- **Litestream** - SQLite replication to S3/cloud for backup
- **HTMX** - already using, double down for reactive UI without JS complexity
- **Twilio** - SMS/voice ingestion
- **Stripe Connect** - payments with platform fee capability
- **OpenAI/Claude API** - natural language parsing for SMS/voice invoices

### Data Portability Principles
1. **Import from anywhere** - CSV, JSON, API, manual entry all produce same internal format
2. **Export everything** - full data export in open formats (JSON, CSV) always available
3. **No lock-in** - SQLite database file is portable, user owns their data
4. **Sync, don't migrate** - bidirectional sync keeps external sources as source of truth if user prefers

---

## Prioritized Implementation Order

| Priority | Feature | Impact | Effort |
|----------|---------|--------|--------|
| 1 | PWA mobile interface | High | Medium |
| 2 | Payment links (Stripe) | High | Medium |
| 3 | Local SQLite storage option | High | Medium |
| 4 | SMS-to-invoice | Very High | Medium |
| 5 | Part/service catalog | High | Low |
| 6 | Customer portal | Medium | Medium |
| 7 | Payment tracking & reminders | High | Medium |
| 8 | QuickBooks sync | Medium | High |
| 9 | Photo documentation | Medium | Low |
| 10 | Multi-technician support | Medium | High |

---

## Competitive Differentiation

**vs. QuickBooks/FreshBooks:**
- Simpler, focused on field service
- Data portability (not locked into ecosystem)
- Works offline, syncs later

**vs. Jobber/ServiceTitan:**
- Dramatically simpler and cheaper
- No monthly per-user fees
- Self-hostable option

**vs. Wave/Invoice Ninja:**
- Mobile-first, field-ready
- SMS/voice input unique
- Better for non-desk workers

---

## Monetization Options

1. **Open Core** - CLI free forever, hosted service with premium features
2. **Usage-Based** - Free tier (X invoices/month), paid tiers for volume
3. **Payment Processing Cut** - Free software, small % on payments collected
4. **White-Label Licensing** - Sell to field service software vendors

---

This roadmap transforms Invoice Magic from a capable PDF generator into a complete invoicing solution that meets independent professionals wherever they work, with whatever tools they already use.
