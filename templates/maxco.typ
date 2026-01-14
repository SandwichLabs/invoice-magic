// MaxCo Appliance Repair Service Invoice Template
// Matches the Chicagoland Appliance Repair Inc. format

#let data = json(sys.inputs.data)

// Color definitions
#let green = rgb("#228B22")
#let blue = rgb("#1E90FF")

// Helper function for checkbox display
#let checkbox(checked) = {
  if checked { box(width: 10pt, height: 10pt, stroke: 0.5pt, inset: 1pt)[#text(size: 7pt)[X]] }
  else { box(width: 10pt, height: 10pt, stroke: 0.5pt) }
}

// Gear/cog icon - using Unicode gear character
#let gear-icon = text(size: 32pt, fill: green)[⚙]

// Page setup - Letter size with appropriate margins
#set page(
  paper: "us-letter",
  margin: (top: 0.6cm, bottom: 0.6cm, left: 1cm, right: 1cm),
)

#set text(size: 9pt, font: "Liberation Sans")

// ============================================
// HEADER SECTION
// ============================================

#align(center)[
  #grid(
    columns: (auto, 1fr, auto),
    align: (center, center, center),
    column-gutter: 0.3cm,
    gear-icon,
    [
      #text(size: 24pt, weight: "bold", fill: green)[Chicagoland Appliance Repair Inc.]
      #v(0.1cm)
      #text(size: 10pt, style: "italic", fill: blue)[Appliance Repair | Heating And Cooling | Dryer Vent Cleaning]
    ],
    gear-icon,
  )
  #v(0.15cm)
  #line(length: 100%, stroke: 1.5pt + green)
  #v(0.1cm)
  #text(size: 9pt, weight: "bold")[#data.sender.address]
  #v(0.05cm)
  #text(size: 10pt, weight: "bold")[#if "phone" in data.sender [#data.sender.phone]]
  #v(0.05cm)
  #text(size: 8pt)[Chicago, IL 60607]
]

#v(0.25cm)

// ============================================
// CUSTOMER INFO AND DATE/TECH SECTION
// ============================================

#grid(
  columns: (55%, 45%),
  column-gutter: 0.3cm,
  [
    // Customer Information header bar
    #block(width: 100%, fill: blue, inset: 3pt)[
      #text(fill: white, weight: "bold", size: 9pt)[Customer Information:]
    ]
    #v(0.1cm)
    #grid(
      columns: (auto, 1fr),
      column-gutter: 0.2cm,
      row-gutter: 0.15cm,
      [Name:], [#box(width: 100%, stroke: (bottom: 0.5pt), inset: (bottom: 2pt))[#data.customer.name]],
      [Address:], [#box(width: 100%, stroke: (bottom: 0.5pt), inset: (bottom: 2pt))[#if "address" in data.customer [#data.customer.address]]],
    )
  ],
  [
    #v(0.15cm)
    #grid(
      columns: (auto, 1fr),
      column-gutter: 0.2cm,
      row-gutter: 0.15cm,
      [*Date:*], [#box(width: 100%, stroke: (bottom: 0.5pt), inset: (bottom: 2pt))[#data.meta.date]],
      [*Technician:*], [#box(width: 100%, stroke: (bottom: 0.5pt), inset: (bottom: 2pt))[#if "service" in data and "technician" in data.service [#data.service.technician]]],
    )
    #v(0.1cm)
    #grid(
      columns: (auto, 1fr, auto, 1fr),
      column-gutter: 0.1cm,
      [Phone: C:], [#box(width: 100%, stroke: (bottom: 0.5pt), inset: (bottom: 2pt))[#if "phone" in data.customer [#data.customer.phone]]],
      [H:], [#box(width: 100%, stroke: (bottom: 0.5pt), inset: (bottom: 2pt))[#if "phone_home" in data.customer [#data.customer.phone_home]]],
    )
  ]
)

// Appliance Info Row
#v(0.15cm)
#grid(
  columns: (1fr, 1fr, 1fr, 1fr),
  column-gutter: 0.2cm,
  [Type: #box(width: 65%, stroke: (bottom: 0.5pt), inset: (bottom: 2pt))[#if "service" in data and "type" in data.service [#data.service.type]]],
  [Make: #box(width: 65%, stroke: (bottom: 0.5pt), inset: (bottom: 2pt))[#if "service" in data and "make" in data.service [#data.service.make]]],
  [Model: #box(width: 60%, stroke: (bottom: 0.5pt), inset: (bottom: 2pt))[#if "service" in data and "model" in data.service [#data.service.model]]],
  [Serial: #box(width: 60%, stroke: (bottom: 0.5pt), inset: (bottom: 2pt))[#if "service" in data and "serial" in data.service [#data.service.serial]]],
)

// Email Address
#v(0.1cm)
Email Address: #box(width: 30%, stroke: (bottom: 0.5pt), inset: (bottom: 2pt))[#if "email" in data.customer [#data.customer.email]]

// ============================================
// WARRANTY AND SERVICE TYPE
// ============================================

#v(0.15cm)
#let warranty_type = if "service" in data and "warranty_type" in data.service { data.service.warranty_type } else { "" }
#let service_type = if "service" in data and "service_type" in data.service { data.service.service_type } else { "" }

#text(size: 8pt)[
  *Warranty:*
  #h(0.15cm)
  #checkbox(warranty_type == "none") None
  #h(0.15cm)
  #checkbox(warranty_type == "90") 90
  #h(0.15cm)
  #checkbox(warranty_type == "parts_labor") Parts & Labor
  #h(0.15cm)
  #checkbox(warranty_type == "parts_only") Parts Only
  #h(0.15cm)
  #checkbox(warranty_type == "labor_only") Labor Only
  #h(0.4cm)
  *Nature of Service:*
  #h(0.15cm)
  #checkbox(service_type == "repair") Repair
  #h(0.15cm)
  #checkbox(service_type == "install") Install
]

// ============================================
// TERMS AND CONDITIONS
// ============================================

#v(0.2cm)
#set text(size: 6pt)
#block(width: 100%, inset: 3pt, stroke: 0.5pt)[
*Warranty:* Unless specified above, there is no warranty. All warranties specified apply under normal use of unit only and same location (address). Company is not responsible for food loss, medicine or other perishables, or damage to carpet, tile, floor, counter, wall or any other personal property that may occur while company's agents or employees repair, service or move unit as stated per this agreement. There is no warranty on refrigerant leaks unless the specified leak is repaired by company.

*Payment:* Payment is due upon delivery of part(s) and/or completion of service, unless payment terms are specified set forth in the handwritten portion of this agreement. In case of returned checks or credit / debit card charge back (NSF, stop payment, closed account, etc.), customer is responsible for reasonable attorney fees and collection costs incurred as a result of the dishonored check or credit / debit card. Deposit/payments for parts specially ordered for the customer are non-refundable once ordered by company. If a refund is requested for any reason, the customer must notify the company in writing. Refunds are limited to amount paid, less a charge for the service call plus a 15% handling/restocking fee. Refunds or cancellations for any reason will void any warranty provided by the company.

*Parts:* Certain parts are recycled by company or sent to manufacturers. Customer must indicate in written portion of this contract if customer intends to keep parts that were replaced. For all other parts, the customer is responsible for their disposal.

I, the undersigned, contend to be of legal age and/or completely responsible and/or fully authorized to order and accept diagnostic and repair services according to the terms and conditions set forth and the prices quoted herein. I further contend that this is the complete, only and final agreement between company and myself. If married (or a person other than the owner/user of the appliance being serviced), my signature represents that I am acting pursuant to authorization from my spouse (or the owner/user).

*Credit Card Surcharge:* A 3.5% credit card processing fee is added to the total cost of service for using a credit card or a debit card to cover the merchant's processing fee.
]

#set text(size: 9pt)

// ============================================
// LINE ITEMS TABLE
// ============================================

#v(0.2cm)
#table(
  columns: (1fr, 1.8cm),
  align: (left, right),
  stroke: 0.5pt,
  inset: 5pt,

  // Header row with blue background
  table.cell(fill: blue)[#text(fill: white, weight: "bold")[Description]],
  table.cell(fill: blue)[#text(fill: white, weight: "bold")[Amount]],

  // Item rows
  ..data.items.map((item) => (
    item.description,
    if "amount" in item { str(item.amount) } else { str(calc.round(item.qty * item.unit_price, digits: 2)) },
  )).flatten(),

  // Empty rows for handwritten entries
  [], [],
  [], [],
)

// ============================================
// TOTALS SECTION
// ============================================

#v(0.1cm)
#align(right)[
  #grid(
    columns: (auto, 4cm),
    column-gutter: 0.3cm,
    row-gutter: 0.1cm,
    align: (right, right),
    [*Total*], [#box(width: 100%, stroke: (bottom: 0.5pt), inset: (bottom: 2pt))[#str(data.totals.gross)]],
    [*Deposit*], [#box(width: 100%, stroke: (bottom: 0.5pt), inset: (bottom: 2pt))[#if "deposit" in data.totals [#str(data.totals.deposit)]]],
    [*Balance*], [#box(width: 100%, stroke: (bottom: 0.5pt), inset: (bottom: 2pt))[#if "balance" in data.totals [#str(data.totals.balance)]]],
  )
]

// ============================================
// PAYMENT SECTION
// ============================================

#v(0.2cm)
*Payment* #h(0.2cm) #checkbox(false) Check \# #box(width: 4cm, stroke: (bottom: 0.5pt))[]
#v(0.08cm)
#h(1.4cm) #checkbox(false) Credit Card \# Last 4 Digits \# #box(width: 3cm, stroke: (bottom: 0.5pt))[]
#v(0.08cm)
#h(1.4cm) #checkbox(false) Confirmation (Approval) \# #box(width: 3.5cm, stroke: (bottom: 0.5pt))[]

// ============================================
// NOTES SECTION (if present)
// ============================================

#if "notes" in data and data.notes != "" [
  #v(0.15cm)
  #block(width: 100%, inset: 3pt, stroke: 0.5pt + blue)[
    *Notes:* #data.notes
  ]
]

// ============================================
// SIGNATURE SECTION
// ============================================

#v(0.25cm)
#line(length: 100%, stroke: 0.5pt)
#v(0.2cm)

#grid(
  columns: (1fr, 1fr),
  column-gutter: 1cm,
  [
    *Customer Signature:* (Service call or estimate approved)
    #v(0.5cm)
    #line(length: 85%, stroke: 0.5pt)
  ],
  [
    *Customer Signature:* (upon completion)
    #v(0.5cm)
    #line(length: 85%, stroke: 0.5pt)
  ]
)
