// MaxCo Appliance Repair Service Invoice Template
// Matches the Chicagoland Appliance Repair Inc. format

#let data = json(sys.inputs.data)

// Color definitions
#let green = rgb("#4CAF50")
#let blue = rgb("#2196F3")

// Helper function for checkbox display
#let checkbox(checked) = {
  if checked { box(width: 10pt, height: 10pt, stroke: 0.5pt, inset: 1pt)[#text(size: 8pt)[X]] }
  else { box(width: 10pt, height: 10pt, stroke: 0.5pt) }
}

// Page setup - Letter size with appropriate margins
#set page(
  paper: "us-letter",
  margin: (top: 1cm, bottom: 1cm, left: 1.5cm, right: 1.5cm),
)

#set text(size: 9pt, font: "Arial")

// ============================================
// HEADER SECTION
// ============================================

#align(center)[
  #text(size: 28pt, weight: "bold", fill: green)[#data.sender.name]
  #v(0.2cm)
  #text(size: 12pt, style: "italic", fill: blue)[Appliance Repair | Heating And Cooling | Dryer Vent Cleaning]
  #v(0.3cm)
  #text(size: 11pt, weight: "bold")[#data.sender.address]
  #v(0.1cm)
  #text(size: 12pt, weight: "bold")[#if "phone" in data.sender [#data.sender.phone]]
]

#v(0.3cm)
#line(length: 100%, stroke: 1pt + green)

// ============================================
// DATE AND TECHNICIAN (right-aligned)
// ============================================

#v(0.2cm)
#align(right)[
  #table(
    columns: (auto, 8cm),
    align: (right, left),
    stroke: none,
    inset: 2pt,
    [*Date:*], [#box(width: 100%, stroke: (bottom: 0.5pt))[#data.meta.date]],
    [*Technician:*], [#box(width: 100%, stroke: (bottom: 0.5pt))[#if "service" in data and "technician" in data.service [#data.service.technician]]],
  )
]

// ============================================
// CUSTOMER INFORMATION SECTION
// ============================================

#v(0.2cm)
#text(weight: "bold")[Customer Information:]

#v(0.2cm)
#grid(
  columns: (1fr, 1fr),
  column-gutter: 1cm,
  [
    #table(
      columns: (auto, 1fr),
      align: (left, left),
      stroke: none,
      inset: 2pt,
      [Name:], [#box(width: 100%, stroke: (bottom: 0.5pt))[#data.customer.name]],
      [Address:], [#box(width: 100%, stroke: (bottom: 0.5pt))[#if "address" in data.customer [#data.customer.address]]],
    )
  ],
  [
    #table(
      columns: (auto, 1fr, auto, 1fr),
      align: (left, left, left, left),
      stroke: none,
      inset: 2pt,
      [Phone: C:], [#box(width: 100%, stroke: (bottom: 0.5pt))[#if "phone_cell" in data.customer [#data.customer.phone_cell] else if "phone" in data.customer [#data.customer.phone]]],
      [H:], [#box(width: 100%, stroke: (bottom: 0.5pt))[#if "phone_home" in data.customer [#data.customer.phone_home]]],
    )
  ]
)

// Appliance Info Row
#v(0.1cm)
#grid(
  columns: (1fr, 1fr, 1fr, 1fr),
  column-gutter: 0.5cm,
  [
    Type: #box(width: 80%, stroke: (bottom: 0.5pt))[#if "service" in data and "type" in data.service [#data.service.type]]
  ],
  [
    Make: #box(width: 80%, stroke: (bottom: 0.5pt))[#if "service" in data and "make" in data.service [#data.service.make]]
  ],
  [
    Model: #box(width: 80%, stroke: (bottom: 0.5pt))[#if "service" in data and "model" in data.service [#data.service.model]]
  ],
  [
    Serial: #box(width: 80%, stroke: (bottom: 0.5pt))[#if "service" in data and "serial" in data.service [#data.service.serial]]
  ]
)

// Email Address
#v(0.1cm)
Email Address: #box(width: 40%, stroke: (bottom: 0.5pt))[#if "email" in data.customer [#data.customer.email]]

// ============================================
// WARRANTY AND SERVICE TYPE
// ============================================

#v(0.2cm)
#let warranty_type = if "service" in data and "warranty_type" in data.service { data.service.warranty_type } else { "" }
#let service_type = if "service" in data and "service_type" in data.service { data.service.service_type } else { "" }

#grid(
  columns: (auto, 1fr),
  column-gutter: 1cm,
  [
    *Warranty:*
    #checkbox(warranty_type == "none") None
    #h(0.3cm)
    #checkbox(warranty_type == "90") 90
    #h(0.3cm)
    #checkbox(warranty_type == "parts_labor") Parts & Labor
    #h(0.3cm)
    #checkbox(warranty_type == "parts_only") Parts Only
    #h(0.3cm)
    #checkbox(warranty_type == "labor_only") Labor Only
  ],
  [
    *Nature of Service:*
    #checkbox(service_type == "repair") Repair
    #h(0.3cm)
    #checkbox(service_type == "install") Install
  ]
)

// ============================================
// TERMS AND CONDITIONS
// ============================================

#v(0.3cm)
#set text(size: 7pt)
#block(width: 100%, inset: 4pt, stroke: 0.5pt)[
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

#v(0.3cm)
#table(
  columns: (1fr, auto),
  align: (left, right),
  stroke: 0.5pt,
  inset: 6pt,

  // Header row
  table.header(
    [*Description*], [*Amount*],
  ),

  // Item rows - use amount directly if available, otherwise calculate
  ..data.items.map((item) => (
    item.description,
    if "amount" in item { str(item.amount) } else { str(calc.round(item.qty * item.unit_price, digits: 2)) },
  )).flatten(),

  // Empty rows for handwritten entries
  [], [],
  [], [],
  [], [],
)

// ============================================
// TOTALS SECTION
// ============================================

#v(0.1cm)
#align(right)[
  #table(
    columns: (auto, 6cm),
    align: (right, right),
    stroke: none,
    inset: 3pt,
    [*Total*], [#box(width: 100%, stroke: (bottom: 0.5pt))[#str(data.totals.gross)]],
    [*Deposit*], [#box(width: 100%, stroke: (bottom: 0.5pt))[#if "deposit" in data.totals [#str(data.totals.deposit)]]],
    [*Balance*], [#box(width: 100%, stroke: (bottom: 0.5pt))[#if "balance" in data.totals [#str(data.totals.balance)]]],
  )
]

// ============================================
// PAYMENT SECTION
// ============================================

#v(0.3cm)
*Payment* #checkbox(false) Check  \#  #box(width: 6cm, stroke: (bottom: 0.5pt))[]
#v(0.1cm)
#h(1.35cm) #checkbox(false) Credit Card \# Last 4 Digits \# #box(width: 4cm, stroke: (bottom: 0.5pt))[]
#v(0.1cm)
#h(1.35cm) Confirmation (Approval) \# #box(width: 5cm, stroke: (bottom: 0.5pt))[]

// ============================================
// NOTES SECTION (if present)
// ============================================

#if "notes" in data and data.notes != "" [
  #v(0.3cm)
  #block(width: 100%, inset: 4pt, stroke: 0.5pt + blue)[
    *Notes:* #data.notes
  ]
]

// ============================================
// SIGNATURE SECTION
// ============================================

#v(0.5cm)
#line(length: 100%, stroke: 0.5pt)
#v(0.3cm)

#grid(
  columns: (1fr, 1fr),
  column-gutter: 2cm,
  [
    *Customer Signature:* (Service call or estimate approved)
    #v(0.8cm)
    #line(length: 90%, stroke: 0.5pt)
  ],
  [
    *Customer Signature:* (upon completion)
    #v(0.8cm)
    #line(length: 90%, stroke: 0.5pt)
  ]
)
