// Default Invoice Template for invoice-generator-pro
// This template creates a professional invoice from JSON data

// Parse invoice data from input
#let data = json(sys.inputs.data)

// Page setup
#set page(
  paper: "a4",
  margin: (top: 2cm, bottom: 2cm, left: 2cm, right: 2cm),
)

#set text(font: "Linux Libertine", size: 11pt)

// Header
#align(center)[
  #text(size: 24pt, weight: "bold")[INVOICE]
]

#v(1cm)

// Invoice metadata
#grid(
  columns: (1fr, 1fr),
  align: (left, right),
  [
    *Invoice Number:* #data.meta.invoice_number \
    *Date:* #data.meta.date \
    #if "due_date" in data.meta [*Due Date:* #data.meta.due_date]
  ],
  []
)

#v(1cm)

// Sender and Customer
#grid(
  columns: (1fr, 1fr),
  column-gutter: 2cm,
  [
    *From:* \
    #text(weight: "bold")[#data.sender.name] \
    #data.sender.address \
    #if "email" in data.sender [#data.sender.email] \
    #if "phone" in data.sender [#data.sender.phone]
  ],
  [
    *To:* \
    #text(weight: "bold")[#data.customer.name] \
    #if "company" in data.customer [#data.customer.company \]
    #if "address" in data.customer [#data.customer.address] \
    #if "email" in data.customer [#data.customer.email]
  ]
)

#v(1cm)

// Line items table
#table(
  columns: (auto, 1fr, auto, auto, auto),
  align: (center, left, right, right, right),
  stroke: 0.5pt,
  inset: 8pt,

  // Header
  [*#*], [*Description*], [*Qty*], [*Unit Price*], [*Amount*],

  // Items
  ..data.items.enumerate().map(((i, item)) => (
    [#(i + 1)],
    [#item.description],
    [#item.qty],
    [#item.unit_price],
    [#calc.round(item.qty * item.unit_price, digits: 2)],
  )).flatten()
)

#v(0.5cm)

// Totals
#align(right)[
  #table(
    columns: (auto, auto),
    align: (left, right),
    stroke: none,
    inset: 4pt,

    [*Subtotal:*], [#data.totals.net],
    [*Tax:*], [#data.totals.tax],
    table.hline(stroke: 1pt),
    [*Total:*], [*#data.totals.gross*],
  )
]

#v(1cm)

// Notes
#if "notes" in data [
  #line(length: 100%, stroke: 0.5pt)
  #v(0.3cm)
  *Notes:* #data.notes
]
