# API Specification

## Travel & Hospitality Procurement and Booking Operations Platform

**Base URL:** `/api/v1`

---

## Response Envelope

All responses follow a standard envelope format:

### Success Response
```json
{
  "success": true,
  "data": { ... }
}
```

### Error Response
```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable message",
    "details": { "field": "validation message" }
  }
}
```

### Paginated Response
```json
{
  "success": true,
  "data": {
    "items": [ ... ],
    "total": 100,
    "page": 1,
    "pageSize": 20,
    "totalPages": 5
  }
}
```

### Error Codes

| Code | HTTP Status | Description |
|---|---|---|
| `BAD_REQUEST` | 400 | Malformed request body |
| `UNAUTHORIZED` | 401 | Missing or invalid JWT token |
| `FORBIDDEN` | 403 | Insufficient role/permissions |
| `NOT_FOUND` | 404 | Resource not found |
| `VALIDATION_ERROR` | 422 | Field validation failure (details map included) |
| `INTERNAL_ERROR` | 500 | Server error |

---

## Health & Readiness

### `GET /health`
Health check endpoint.

**Auth:** None

**Response:** `200 OK`
```json
{ "status": "ok" }
```

### `GET /ready`
Readiness check — verifies database connectivity.

**Auth:** None

**Response:** `200 OK`
```json
{ "status": "ready" }
```

---

## Authentication

### `POST /api/v1/auth/login`
Authenticate user and receive JWT token.

**Auth:** None

**Request Body:**
```json
{
  "email": "string",
  "password": "string"
}
```

**Response:** `200 OK`
```json
{
  "token": "jwt-token-string",
  "user": {
    "id": "uuid",
    "email": "string",
    "status": "active",
    "displayName": "string",
    "roles": ["traveler", "group_organizer"]
  }
}
```

### `POST /api/v1/auth/logout`
Invalidate the current session.

**Auth:** Bearer token required

**Response:** `200 OK`
```json
{ "message": "logged out" }
```

### `GET /api/v1/auth/me`
Get authenticated user information.

**Auth:** Bearer token required

**Response:** `200 OK`
```json
{
  "id": "uuid",
  "email": "string",
  "status": "active",
  "displayName": "string",
  "roles": ["traveler"]
}
```

---

## Users

### `GET /api/v1/users/:id`
Get user by ID.

**Auth:** Bearer token required

**Response:** `200 OK`
```json
{
  "id": "uuid",
  "email": "string",
  "status": "active",
  "displayName": "string",
  "roles": ["traveler"]
}
```

### `PATCH /api/v1/users/:id/profile`
Update user profile.

**Auth:** Bearer token required (own user)

**Request Body:**
```json
{
  "display_name": "string (optional)",
  "phone_masked": "string (optional)"
}
```

**Response:** `200 OK`

### `PATCH /api/v1/users/:id/preferences`
Update user preferences.

**Auth:** Bearer token required (own user)

**Request Body:**
```json
{
  "preferences": { "key": "value" }
}
```

**Response:** `200 OK`

### `GET /api/v1/admin/users`
List all users (admin only).

**Auth:** Bearer token required  
**Role:** `administrator`

**Query Parameters:**
| Param | Type | Default | Description |
|---|---|---|---|
| `page` | int | 1 | Page number |
| `pageSize` | int | 20 | Items per page |
| `status` | string | — | Filter by status |

**Response:** `200 OK` — Paginated list of user objects

---

## Itineraries

### `POST /api/v1/itineraries`
Create a new itinerary.

**Auth:** Bearer token required  
**Role:** `group_organizer`, `administrator`

**Request Body:**
```json
{
  "title": "string",
  "meetupAt": "2026-07-14T18:30:00Z (optional)",
  "meetupLocationText": "string",
  "notes": "string"
}
```

**Response:** `201 Created`
```json
{
  "id": "uuid",
  "organizerId": "uuid",
  "title": "string",
  "meetupAt": "2026-07-14T18:30:00Z",
  "meetupLocationText": "string",
  "notes": "string",
  "status": "draft",
  "publishedAt": null,
  "checkpointsCount": 0,
  "membersCount": 0,
  "createdAt": "timestamp",
  "updatedAt": "timestamp"
}
```

### `GET /api/v1/itineraries`
List itineraries (paginated).

**Auth:** Bearer token required

**Query Parameters:**
| Param | Type | Default | Description |
|---|---|---|---|
| `page` | int | 1 | Page number |
| `pageSize` | int | 20 | Items per page |
| `status` | string | — | Filter by status |

**Response:** `200 OK` — Paginated list of `ItineraryResponse` objects

### `GET /api/v1/itineraries/:id`
Get itinerary detail including checkpoints, members, and form definitions.

**Auth:** Bearer token required (organizer, member, or admin)

**Response:** `200 OK`
```json
{
  "id": "uuid",
  "organizerId": "uuid",
  "title": "string",
  "meetupAt": "timestamp",
  "meetupLocationText": "string",
  "notes": "string",
  "status": "draft|published|revised|in_progress|completed|cancelled|archived",
  "publishedAt": "timestamp|null",
  "checkpointsCount": 5,
  "membersCount": 3,
  "createdAt": "timestamp",
  "updatedAt": "timestamp",
  "checkpoints": [
    {
      "id": "uuid",
      "checkpointText": "string",
      "sortOrder": 1,
      "eta": "timestamp|null"
    }
  ],
  "members": [
    {
      "userId": "uuid",
      "role": "string",
      "joinedAt": "timestamp"
    }
  ],
  "formDefinitions": [
    {
      "id": "uuid",
      "fieldKey": "string",
      "fieldLabel": "string",
      "fieldType": "text|select|textarea",
      "required": true,
      "options": {},
      "sortOrder": 1
    }
  ]
}
```

### `PATCH /api/v1/itineraries/:id`
Update an itinerary.

**Auth:** Bearer token required  
**Role:** `group_organizer`, `administrator`

**Request Body:**
```json
{
  "title": "string (optional)",
  "meetupAt": "string (optional)",
  "meetupLocationText": "string (optional)",
  "notes": "string (optional)"
}
```

**Response:** `200 OK`

### `POST /api/v1/itineraries/:id/publish`
Publish a draft itinerary.

**Auth:** Bearer token required  
**Role:** `group_organizer`, `administrator`

**Response:** `200 OK`

### `POST /api/v1/itineraries/:id/checkpoints`
Add a checkpoint to an itinerary.

**Auth:** Bearer token required  
**Role:** `group_organizer`, `administrator`

**Request Body:**
```json
{
  "checkpointText": "string",
  "sortOrder": 1,
  "eta": "2026-07-14T09:00:00Z (optional)"
}
```

**Response:** `201 Created`

### `PATCH /api/v1/itineraries/:id/checkpoints/:checkpointId`
Update a checkpoint.

**Auth:** Bearer token required  
**Role:** `group_organizer`, `administrator`

**Request Body:**
```json
{
  "checkpointText": "string (optional)",
  "sortOrder": 1,
  "eta": "string (optional)"
}
```

**Response:** `200 OK`

### `DELETE /api/v1/itineraries/:id/checkpoints/:checkpointId`
Delete a checkpoint.

**Auth:** Bearer token required  
**Role:** `group_organizer`, `administrator`

**Response:** `200 OK`

### `POST /api/v1/itineraries/:id/members`
Add a member to an itinerary.

**Auth:** Bearer token required  
**Role:** `group_organizer`, `administrator`

**Request Body:**
```json
{
  "userId": "uuid",
  "role": "string"
}
```

**Response:** `201 Created`

### `DELETE /api/v1/itineraries/:id/members/:userId`
Remove a member from an itinerary.

**Auth:** Bearer token required  
**Role:** `group_organizer`, `administrator`

**Response:** `200 OK`

### `POST /api/v1/itineraries/:id/form-definitions`
Create a form field definition for member data collection.

**Auth:** Bearer token required  
**Role:** `group_organizer`, `administrator`

**Request Body:**
```json
{
  "fieldKey": "vehicle_plate",
  "fieldLabel": "Vehicle Plate Number",
  "fieldType": "text|select|textarea",
  "required": true,
  "options": {},
  "validation": {},
  "sortOrder": 1
}
```

**Response:** `201 Created`

### `PATCH /api/v1/itineraries/:id/form-definitions/:defId`
Update a form field definition.

**Auth:** Bearer token required  
**Role:** `group_organizer`, `administrator`

**Request Body:**
```json
{
  "fieldLabel": "string (optional)",
  "fieldType": "string (optional)",
  "required": true,
  "options": {},
  "validation": {},
  "active": true,
  "sortOrder": 1
}
```

**Response:** `200 OK`

### `GET /api/v1/itineraries/:id/form-definitions`
Get all form definitions for an itinerary.

**Auth:** Bearer token required

**Response:** `200 OK` — Array of form definition objects

### `POST /api/v1/itineraries/:id/form-submissions`
Submit member form data (one per member per itinerary).

**Auth:** Bearer token required  
**Role:** `traveler`, `group_organizer`, `administrator`

**Request Body:**
```json
{
  "payload": {
    "vehicle_plate": "ABC-1234",
    "emergency_contact": "555-0123"
  }
}
```

**Response:** `201 Created`

### `GET /api/v1/itineraries/:id/form-submissions`
Get all form submissions for an itinerary.

**Auth:** Bearer token required  
**Role:** `group_organizer`, `administrator`

**Response:** `200 OK` — Array of form submission objects

### `GET /api/v1/itineraries/:id/change-events`
Get change event history for an itinerary.

**Auth:** Bearer token required

**Response:** `200 OK`
```json
[
  {
    "id": "uuid",
    "actorId": "uuid",
    "changeType": "string",
    "summary": "string",
    "diff": {},
    "createdAt": "timestamp"
  }
]
```

---

## Bookings

### `GET /api/v1/bookings`
List bookings (paginated).

**Auth:** Bearer token required

**Query Parameters:**
| Param | Type | Default | Description |
|---|---|---|---|
| `page` | int | 1 | Page number |
| `pageSize` | int | 20 | Items per page |

**Response:** `200 OK` — Paginated list of `BookingResponse`

### `POST /api/v1/bookings`
Create a new booking.

**Auth:** Bearer token required

**Request Body:**
```json
{
  "title": "string",
  "description": "string",
  "itineraryId": "uuid (optional)",
  "items": [
    {
      "itemType": "lodging|transport|activity|other",
      "itemName": "string",
      "description": "string",
      "unitPrice": 150.00,
      "quantity": 2,
      "category": "string"
    }
  ]
}
```

**Response:** `201 Created`
```json
{
  "id": "uuid",
  "organizerId": "uuid",
  "itineraryId": "uuid|null",
  "title": "string",
  "description": "string",
  "status": "draft",
  "totalAmount": 300.00,
  "discountAmount": 0.00,
  "escrowAmount": 0.00,
  "pricingSnapshotId": null,
  "items": [
    {
      "id": "uuid",
      "itemType": "lodging",
      "itemName": "string",
      "description": "string",
      "unitPrice": 150.00,
      "quantity": 2,
      "subtotal": 300.00,
      "category": "string"
    }
  ],
  "createdAt": "timestamp",
  "updatedAt": "timestamp"
}
```

### `GET /api/v1/bookings/:id`
Get booking detail with line items.

**Auth:** Bearer token required

**Response:** `200 OK` — `BookingResponse` with `items` array

### `POST /api/v1/bookings/:id/price-preview`
Preview pricing with coupon application.

**Auth:** Bearer token required

**Request Body:**
```json
{
  "couponCodes": ["SAVE25", "LODGE10"],
  "membershipTier": "gold",
  "isNewUser": false
}
```

**Response:** `200 OK`
```json
{
  "subtotal": 500.00,
  "totalDiscount": 50.00,
  "escrowHoldAmount": 450.00,
  "finalPayable": 450.00,
  "eligibleCoupons": [
    {
      "couponId": "uuid",
      "code": "SAVE25",
      "name": "$25 off $200",
      "discountAmount": 25.00
    }
  ],
  "ineligibleCoupons": [
    {
      "couponId": "uuid",
      "code": "LODGE10",
      "name": "10% off lodging",
      "reasonCode": "REASON_STACKING_NOT_ALLOWED",
      "message": "Cannot stack with threshold discount"
    }
  ],
  "appliedDiscounts": [
    {
      "type": "member_tier",
      "description": "Gold tier: 5% discount",
      "amount": 25.00
    },
    {
      "type": "threshold_fixed",
      "description": "$25 off $200 (SAVE25)",
      "amount": 25.00
    }
  ],
  "snapshotId": "uuid"
}
```

### `POST /api/v1/bookings/:id/checkout`
Process booking checkout with idempotency protection.

**Auth:** Bearer token required

**Headers:**
| Header | Required | Description |
|---|---|---|
| `Idempotency-Key` | Yes | Client-generated unique key for deduplication |

**Request Body:**
```json
{
  "pricingSnapshotId": "uuid",
  "couponCodes": ["SAVE25"],
  "idempotencyKey": "string",
  "membershipTier": "gold",
  "isNewUser": false
}
```

**Response:** `200 OK`
```json
{
  "bookingId": "uuid",
  "status": "paid_held_in_escrow",
  "totalAmount": 500.00,
  "discountAmount": 50.00,
  "escrowAmount": 450.00,
  "snapshotId": "uuid"
}
```

### `POST /api/v1/bookings/:id/record-tender`
Record a manual payment instrument against a booking.

**Auth:** Bearer token required

**Request Body:**
```json
{
  "tenderType": "cash|card_on_file_recorded|bank_transfer_recorded|other_manual",
  "amount": 450.00,
  "referenceText": "string"
}
```

**Response:** `200 OK`

### `POST /api/v1/bookings/:id/cancel`
Cancel a booking.

**Auth:** Bearer token required

**Response:** `200 OK`

### `POST /api/v1/bookings/:id/complete`
Mark a booking as completed (triggers escrow release).

**Auth:** Bearer token required

**Response:** `200 OK`

---

## Coupons & Pricing

### `GET /api/v1/coupons/coupons/available`
List all active coupons.

**Auth:** None

**Response:** `200 OK`
```json
[
  {
    "id": "uuid",
    "code": "SAVE25",
    "name": "$25 off orders over $200",
    "discountType": "threshold_fixed|percentage|new_user_gift",
    "amount": 25.00,
    "minSpend": 200.00,
    "percentOff": 0,
    "validFrom": "timestamp",
    "validTo": "timestamp",
    "eligibilityJson": {},
    "stackGroup": "string",
    "exclusive": false,
    "usageLimitTotal": 1000,
    "usageLimitPerUser": 1,
    "active": true,
    "createdAt": "timestamp",
    "updatedAt": "timestamp"
  }
]
```

### `POST /api/v1/coupons/coupons/evaluate`
Evaluate coupon eligibility for a set of items.

**Auth:** Bearer token required

**Request Body:**
```json
{
  "couponCodes": ["SAVE25", "LODGE10"],
  "items": [
    {
      "id": "uuid",
      "bookingId": "uuid",
      "itemType": "lodging",
      "itemName": "Hotel Room",
      "description": "string",
      "unitPrice": 200.00,
      "quantity": 1,
      "subtotal": 200.00,
      "category": "lodging"
    }
  ],
  "membershipTier": "gold",
  "isNewUser": false
}
```

**Response:** `200 OK` — `PricingResult` object (same structure as price-preview response)

### `POST /api/v1/coupons/coupons/redeem-preview`
Evaluate and create a pricing snapshot for redemption.

**Auth:** Bearer token required

**Request Body:** Same as evaluate

**Response:** `200 OK` — `PricingResult` with `snapshotId` populated

---

## Finance

### `GET /api/v1/wallets/:ownerId`
Get wallet for a user.

**Auth:** Bearer token required

**Query Parameters:**
| Param | Type | Description |
|---|---|---|
| `type` | string | Wallet type filter |

**Response:** `200 OK`
```json
{
  "id": "uuid",
  "ownerId": "uuid",
  "walletType": "customer|supplier|courier|platform_clearing|escrow_control|refund_clearing|fee_revenue",
  "balance": 1500.00,
  "currency": "USD",
  "createdAt": "timestamp",
  "updatedAt": "timestamp"
}
```

### `GET /api/v1/wallets/:ownerId/transactions`
Get wallet transaction history (paginated).

**Auth:** Bearer token required

**Query Parameters:**
| Param | Type | Default | Description |
|---|---|---|---|
| `page` | int | 1 | Page number |
| `pageSize` | int | 20 | Items per page |
| `type` | string | — | Filter by transaction type |

**Response:** `200 OK`
```json
{
  "items": [
    {
      "id": "uuid",
      "walletId": "uuid",
      "amount": 100.00,
      "direction": "credit|debit",
      "referenceType": "booking|procurement|withdrawal",
      "referenceId": "uuid",
      "description": "string",
      "createdAt": "timestamp"
    }
  ],
  "total": 50,
  "page": 1,
  "pageSize": 20,
  "totalPages": 3
}
```

### `GET /api/v1/escrows/:ownerId`
Get active escrow accounts for an owner.

**Auth:** Bearer token required

**Response:** `200 OK`
```json
[
  {
    "id": "uuid",
    "orderType": "booking",
    "orderId": "uuid",
    "amountHeld": 450.00,
    "amountReleased": 0.00,
    "status": "held|partially_released|released|refunded",
    "createdAt": "timestamp"
  }
]
```

### `POST /api/v1/payments/record-tender`
Record a manual tender payment (accountant/admin).

**Auth:** Bearer token required  
**Role:** `accountant`, `administrator`

**Request Body:**
```json
{
  "orderType": "booking|procurement",
  "orderId": "uuid",
  "tenderType": "cash|card_on_file_recorded|bank_transfer_recorded|other_manual",
  "amount": 500.00,
  "currency": "USD",
  "referenceText": "string"
}
```

**Response:** `200 OK`
```json
{
  "id": "uuid",
  "recordedAt": "timestamp"
}
```

### `POST /api/v1/settlements/:orderId/release`
Release escrowed funds.

**Auth:** Bearer token required  
**Role:** `accountant`, `administrator`

**Request Body:**
```json
{
  "orderType": "booking",
  "amount": 450.00
}
```

**Response:** `200 OK`

### `POST /api/v1/refunds`
Process a refund (minimum $1.00).

**Auth:** Bearer token required  
**Role:** `accountant`, `administrator`

**Request Body:**
```json
{
  "orderType": "booking",
  "orderId": "uuid",
  "amount": 50.00,
  "reason": "string",
  "items": [
    {
      "itemId": "uuid",
      "itemType": "string",
      "amount": 50.00
    }
  ]
}
```

**Response:** `200 OK`

### `POST /api/v1/withdrawals`
Request a courier withdrawal (daily cap: $2,500.00).

**Auth:** Bearer token required  
**Role:** `courier_runner`

**Request Body:**
```json
{
  "amount": 500.00
}
```

**Response:** `201 Created`
```json
{
  "id": "uuid",
  "courierId": "uuid",
  "requestAmount": 500.00,
  "status": "requested",
  "requestedAt": "timestamp",
  "createdAt": "timestamp"
}
```

### `POST /api/v1/withdrawals/:id/approve`
Approve a withdrawal request.

**Auth:** Bearer token required  
**Role:** `accountant`, `administrator`

**Response:** `200 OK`

### `POST /api/v1/withdrawals/:id/reject`
Reject a withdrawal request.

**Auth:** Bearer token required  
**Role:** `accountant`, `administrator`

**Request Body:**
```json
{
  "reason": "string"
}
```

**Response:** `200 OK`

### `GET /api/v1/reconciliation`
Get reconciliation report data.

**Auth:** Bearer token required  
**Role:** `accountant`, `administrator`

**Response:** `200 OK`
```json
{
  "openingBalance": 10000.00,
  "inflows": 5000.00,
  "outflows": 2000.00,
  "heldInEscrow": 3000.00,
  "released": 1500.00,
  "refunded": 200.00,
  "netPayable": 1300.00,
  "unreconciledItems": []
}
```

---

## Contracts & Invoices

### `GET /api/v1/contract-templates`
List available contract templates.

**Auth:** Bearer token required

**Response:** `200 OK`
```json
[
  {
    "id": "uuid",
    "name": "Supplier Agreement",
    "active": true,
    "version": 1
  }
]
```

### `POST /api/v1/contracts/generate`
Generate a contract from a template.

**Auth:** Bearer token required

**Request Body:**
```json
{
  "templateId": "uuid",
  "variables": {
    "supplier_name": "Acme Hotels",
    "contract_date": "2026-07-14"
  }
}
```

**Response:** `201 Created`
```json
{
  "id": "uuid",
  "templateId": "uuid",
  "fileId": "uuid",
  "generatedBy": "uuid",
  "generatedAt": "timestamp",
  "version": 1
}
```

### `POST /api/v1/invoice-requests`
Request an invoice for an order.

**Auth:** Bearer token required

**Request Body:**
```json
{
  "orderType": "booking|procurement",
  "orderId": "uuid",
  "notes": "string"
}
```

**Response:** `201 Created`
```json
{
  "id": "uuid",
  "requesterId": "uuid",
  "orderType": "booking",
  "orderId": "uuid",
  "status": "requested",
  "notes": "string",
  "createdAt": "timestamp",
  "updatedAt": "timestamp"
}
```

### `GET /api/v1/invoice-requests`
List invoice requests.

**Auth:** Bearer token required

**Response:** `200 OK` — Array of `InvoiceRequestResponse`

### `POST /api/v1/invoice-requests/:id/approve`
Approve an invoice request.

**Auth:** Bearer token required  
**Role:** `administrator`, `accountant`

**Response:** `200 OK`

### `POST /api/v1/invoices/:id/generate`
Generate an invoice PDF from an approved request.

**Auth:** Bearer token required  
**Role:** `administrator`, `accountant`

**Response:** `200 OK`
```json
{
  "id": "uuid",
  "requestId": "uuid",
  "invoiceNumber": "INV-2026-0001",
  "orderType": "booking",
  "orderId": "uuid",
  "amount": 450.00,
  "fileId": "uuid",
  "generatedAt": "timestamp"
}
```

---

## Files

### `POST /api/v1/files/upload`
Upload a file (optionally encrypted).

**Auth:** Bearer token required

**Content-Type:** `multipart/form-data`

**Form Fields:**
| Field | Type | Required | Description |
|---|---|---|---|
| `file` | file | Yes | The file to upload |
| `recordType` | string | Yes | Record type to link (e.g., "booking", "procurement") |
| `recordId` | string | Yes | Record ID to link |
| `encrypt` | boolean | No | Whether to encrypt with AES-256-GCM |

**Response:** `201 Created`
```json
{
  "id": "uuid",
  "originalFilename": "receipt.pdf",
  "mimeType": "application/pdf",
  "byteSize": 204800,
  "encrypted": true,
  "sha256": "hex-string",
  "createdAt": "timestamp"
}
```

### `POST /api/v1/files/:id/download-token`
Generate a time-limited download token.

**Auth:** Bearer token required

**Request Body:**
```json
{
  "ttlSeconds": 3600,
  "singleUse": true
}
```

**Response:** `200 OK`
```json
{
  "token": "uuid-token",
  "expiresAt": "timestamp",
  "singleUse": true
}
```

### `GET /api/v1/files/download/:token`
Download a file using a token.

**Auth:** None (token-based)

**Response:** `200 OK` — Binary file stream with appropriate `Content-Type` and `Content-Disposition` headers

### `GET /api/v1/files/record/:recordType/:recordId`
Get files linked to a record.

**Auth:** Bearer token required

**Response:** `200 OK` — Array of file metadata objects

---

## Notifications

### `GET /api/v1/notifications`
List notifications for the current user (paginated).

**Auth:** Bearer token required

**Query Parameters:**
| Param | Type | Default | Description |
|---|---|---|---|
| `page` | int | 1 | Page number |
| `pageSize` | int | 20 | Items per page |
| `unreadOnly` | boolean | false | Filter to unread only |

**Response:** `200 OK`
```json
{
  "items": [
    {
      "id": "uuid",
      "eventType": "itinerary_updated",
      "sourceType": "itinerary",
      "sourceId": "uuid",
      "channel": "in_app",
      "status": "delivered",
      "payload": {},
      "deliveredAt": "timestamp",
      "readAt": null,
      "createdAt": "timestamp"
    }
  ],
  "total": 15,
  "page": 1,
  "pageSize": 20,
  "totalPages": 1
}
```

### `POST /api/v1/notifications/:id/read`
Mark a notification as read.

**Auth:** Bearer token required

**Response:** `200 OK`

### `GET /api/v1/messages`
List messages for the current user (paginated).

**Auth:** Bearer token required

**Query Parameters:**
| Param | Type | Default | Description |
|---|---|---|---|
| `page` | int | 1 | Page number |
| `pageSize` | int | 20 | Items per page |

**Response:** `200 OK` — Paginated list of `MessageDTO`

### `POST /api/v1/messages/callback-queue/export`
Export callback queue entries as JSON.

**Auth:** Bearer token required  
**Role:** `administrator`, `accountant`

**Response:** `200 OK`
```json
{
  "entries": [ ... ],
  "count": 10,
  "exportedAt": "timestamp"
}
```

### `GET /api/v1/send-logs`
List notification send logs.

**Auth:** Bearer token required

**Response:** `200 OK` — Array of `SendLogDTO`

### `GET /api/v1/users/:id/dnd`
Get Do-Not-Disturb settings.

**Auth:** Bearer token required

**Response:** `200 OK`
```json
{
  "dndStart": "21:00",
  "dndEnd": "08:00",
  "enabled": true
}
```

### `PATCH /api/v1/users/:id/dnd`
Update Do-Not-Disturb settings.

**Auth:** Bearer token required

**Request Body:**
```json
{
  "dndStart": "21:00",
  "dndEnd": "08:00",
  "enabled": true
}
```

**Response:** `200 OK`

### `GET /api/v1/users/:id/subscriptions`
Get subscription preferences.

**Auth:** Bearer token required

**Response:** `200 OK` — Array of subscription preference objects

### `PATCH /api/v1/users/:id/subscriptions`
Update subscription preferences.

**Auth:** Bearer token required

**Request Body:**
```json
[
  {
    "channelType": "in_app",
    "eventType": "itinerary_updated",
    "enabled": true
  }
]
```

**Response:** `200 OK`

---

## Procurement

### `GET /api/v1/rfqs`
List RFQs.

**Auth:** Bearer token required

**Response:** `200 OK` — Array of `RFQResponse`

### `POST /api/v1/rfqs`
Create a new RFQ.

**Auth:** Bearer token required  
**Role:** `accountant`, `administrator`, `group_organizer`

**Request Body:**
```json
{
  "title": "Hotel supplies for Q3",
  "description": "string",
  "deadline": "2026-08-01T00:00:00Z"
}
```

**Response:** `201 Created`
```json
{
  "id": "uuid",
  "createdBy": "uuid",
  "title": "string",
  "description": "string",
  "deadline": "timestamp",
  "status": "draft",
  "items": [],
  "createdAt": "timestamp",
  "updatedAt": "timestamp"
}
```

### `GET /api/v1/rfqs/:id`
Get RFQ detail.

**Auth:** Bearer token required (scope-limited: admin/accountant see all; suppliers see only invited; organizers see only own)

**Response:** `200 OK` — `RFQResponse` with items

### `POST /api/v1/rfqs/:id/issue`
Issue RFQ to suppliers with finalized items.

**Auth:** Bearer token required  
**Role:** `accountant`, `administrator`

**Request Body:**
```json
{
  "supplierIds": ["uuid", "uuid"],
  "items": [
    {
      "itemName": "Bed linens",
      "specifications": "King size, 300 thread count",
      "quantity": 50,
      "unit": "set",
      "sortOrder": 1
    }
  ]
}
```

**Response:** `200 OK`

### `POST /api/v1/rfqs/:id/quotes`
Submit a quote for an RFQ.

**Auth:** Bearer token required  
**Role:** `supplier`, `administrator`

**Request Body:**
```json
{
  "totalAmount": 2500.00,
  "leadTimeDays": 14,
  "notes": "string",
  "items": [
    {
      "rfqItemId": "uuid",
      "unitPrice": 50.00,
      "quantity": 50,
      "notes": "string"
    }
  ]
}
```

**Response:** `201 Created` — `QuoteResponse`

### `GET /api/v1/rfqs/:id/comparison`
Get supplier quote comparison matrix.

**Auth:** Bearer token required  
**Role:** `accountant`, `administrator`

**Response:** `200 OK`
```json
{
  "rfqId": "uuid",
  "items": [ ... ],
  "quotes": [
    {
      "supplierId": "uuid",
      "totalAmount": 2500.00,
      "leadTimeDays": 14,
      "items": [ ... ]
    }
  ]
}
```

### `POST /api/v1/rfqs/:id/select`
Select a supplier quote.

**Auth:** Bearer token required  
**Role:** `accountant`, `administrator`

**Request Body:**
```json
{
  "quoteId": "uuid"
}
```

**Response:** `200 OK`

### `GET /api/v1/purchase-orders`
List purchase orders.

**Auth:** Bearer token required

**Response:** `200 OK` — Array of `POResponse`

### `POST /api/v1/purchase-orders`
Create a purchase order.

**Auth:** Bearer token required  
**Role:** `accountant`, `administrator`

**Request Body:**
```json
{
  "rfqId": "uuid (optional)",
  "quoteId": "uuid (optional)",
  "supplierId": "uuid",
  "promisedDate": "2026-08-15 (optional)",
  "items": [
    {
      "itemName": "Bed linens",
      "specifications": "King size",
      "unitPrice": 50.00,
      "quantity": 50
    }
  ]
}
```

**Response:** `201 Created` — `POResponse` with auto-generated `poNumber` (format: `PO-YYYY-XXXX`)

### `GET /api/v1/purchase-orders/:id`
Get purchase order detail.

**Auth:** Bearer token required (scope-limited by role)

**Response:** `200 OK` — `POResponse` with items

### `POST /api/v1/purchase-orders/:id/accept`
Supplier accepts a purchase order.

**Auth:** Bearer token required  
**Role:** `supplier`, `administrator`

**Response:** `200 OK`

### `POST /api/v1/purchase-orders/:id/deliveries`
Record a delivery against a PO.

**Auth:** Bearer token required  
**Role:** `accountant`, `administrator`

**Request Body:**
```json
{
  "courierId": "uuid (optional)",
  "deliveryDate": "2026-08-14",
  "notes": "string",
  "items": [
    {
      "poItemId": "uuid",
      "quantityDelivered": 50,
      "quantityAccepted": 48,
      "quantityRejected": 2
    }
  ]
}
```

**Response:** `201 Created` — `DeliveryResponse`

### `GET /api/v1/deliveries`
List all deliveries.

**Auth:** Bearer token required

**Response:** `200 OK` — Array of `DeliveryResponse`

### `POST /api/v1/deliveries/:id/inspect`
Perform quality inspection on a delivery.

**Auth:** Bearer token required  
**Role:** `accountant`, `administrator`

**Request Body:**
```json
{
  "status": "passed|failed",
  "notes": "Required notes describing inspection results"
}
```

**Response:** `200 OK`
```json
{
  "id": "uuid",
  "deliveryId": "uuid",
  "poId": "uuid",
  "inspectorId": "uuid",
  "status": "passed|failed",
  "notes": "string",
  "inspectedAt": "timestamp"
}
```

**Note:** A `failed` inspection automatically creates a discrepancy ticket.

### `POST /api/v1/discrepancies`
Create a discrepancy ticket manually.

**Auth:** Bearer token required  
**Role:** `accountant`, `administrator`

**Request Body:**
```json
{
  "poId": "uuid",
  "deliveryId": "uuid (optional)",
  "inspectionId": "uuid (optional)",
  "discrepancyType": "shortage|damage|wrong_item|late_delivery|service_deviation|other",
  "description": "string",
  "notes": "string"
}
```

**Response:** `201 Created`

### `GET /api/v1/exceptions`
List exception cases.

**Auth:** Bearer token required

**Response:** `200 OK`
```json
[
  {
    "id": "uuid",
    "referenceType": "purchase_order|delivery",
    "referenceId": "uuid",
    "status": "open|pending_financial_resolution|pending_waiver|ready_to_close|closed",
    "openedReason": "string",
    "openedAt": "timestamp",
    "closedAt": null,
    "waivers": [],
    "adjustments": [],
    "createdAt": "timestamp"
  }
]
```

### `POST /api/v1/exceptions/:id/waivers`
Submit a waiver for an exception.

**Auth:** Bearer token required  
**Role:** `accountant`, `administrator`

**Request Body:**
```json
{
  "waiverReason": "string"
}
```

**Response:** `200 OK`

### `POST /api/v1/exceptions/:id/settlement-adjustments`
Submit a settlement adjustment (posts journal entry to ledger).

**Auth:** Bearer token required  
**Role:** `accountant`, `administrator`

**Request Body:**
```json
{
  "amount": 100.00,
  "direction": "debit|credit",
  "reason": "string"
}
```

**Response:** `200 OK`

### `POST /api/v1/exceptions/:id/close`
Close an exception case.

**Auth:** Bearer token required  
**Role:** `accountant`, `administrator`

**Prerequisite:** At least one waiver OR settlement adjustment must exist.

**Response:** `200 OK`

### `GET /api/v1/supplier-quotes`
List supplier quotes for the current user (supplier role).

**Auth:** Bearer token required

**Response:** `200 OK` — Array of quote objects

---

## Reviews & Ratings

### `POST /api/v1/reviews`
Submit a review (unique per reviewer + order type + order ID).

**Auth:** Bearer token required

**Request Body:**
```json
{
  "subjectId": "uuid",
  "orderType": "booking|procurement",
  "orderId": "uuid",
  "overallRating": 4.5,
  "comment": "string",
  "scores": [
    { "dimensionName": "punctuality", "score": 5.0 },
    { "dimensionName": "communication", "score": 4.0 },
    { "dimensionName": "quality", "score": 4.5 },
    { "dimensionName": "compliance", "score": 4.0 },
    { "dimensionName": "professionalism", "score": 5.0 }
  ]
}
```

**Available Dimensions:** `punctuality`, `communication`, `quality`, `compliance`, `professionalism`, `accuracy`, `cleanliness`, `route_adherence`, `delivery_integrity`

**Response:** `201 Created`

### `GET /api/v1/reviews/subject/:userId`
Get reviews for a user (paginated).

**Auth:** None

**Query Parameters:**
| Param | Type | Default | Description |
|---|---|---|---|
| `page` | int | 1 | Page number |
| `pageSize` | int | 20 | Items per page |

**Response:** `200 OK`
```json
{
  "items": [
    {
      "id": "uuid",
      "reviewerId": "uuid",
      "subjectId": "uuid",
      "orderType": "booking",
      "orderId": "uuid",
      "overallRating": 4.5,
      "comment": "string",
      "scores": [
        { "dimensionName": "punctuality", "score": 5.0 }
      ],
      "editableUntil": "timestamp",
      "createdAt": "timestamp"
    }
  ],
  "total": 10,
  "page": 1,
  "pageSize": 20,
  "totalPages": 1
}
```

### `GET /api/v1/credit-tiers/:userId`
Get credit tier snapshot for a user.

**Auth:** None

**Response:** `200 OK`
```json
{
  "userId": "uuid",
  "tier": "bronze|silver|gold|platinum|restricted",
  "totalTransactions": 25,
  "avgRating": 4.2,
  "violationCount": 0,
  "computedAt": "timestamp"
}
```

### `POST /api/v1/violations`
Record a violation against a user.

**Auth:** Bearer token required  
**Role:** `administrator`, `accountant`

**Request Body:**
```json
{
  "userId": "uuid",
  "violationType": "string",
  "description": "string",
  "severity": "low|medium|high|critical"
}
```

**Response:** `201 Created`

### `POST /api/v1/no-shows`
Record a no-show.

**Auth:** Bearer token required  
**Role:** `administrator`, `accountant`

**Request Body:**
```json
{
  "userId": "uuid",
  "orderType": "booking|procurement",
  "orderId": "uuid"
}
```

**Response:** `201 Created`

### `POST /api/v1/harassment-flags`
Flag harassment.

**Auth:** Bearer token required

**Request Body:**
```json
{
  "subjectId": "uuid",
  "description": "string",
  "evidenceFileId": "uuid (optional)"
}
```

**Response:** `201 Created`

---

## Risk Management

### `GET /api/v1/risk/:userId`
Get risk summary for a user.

**Auth:** Bearer token required  
**Role:** `administrator`

**Response:** `200 OK`
```json
{
  "userId": "uuid",
  "score": 15.0,
  "isBlacklisted": false,
  "activeThrottles": [
    {
      "id": "uuid",
      "actionType": "create_rfq",
      "reason": "RFQ rate limit exceeded",
      "expiresAt": "timestamp",
      "active": true,
      "createdAt": "timestamp"
    }
  ],
  "recentEvents": [
    {
      "id": "uuid",
      "eventType": "rfq_creation",
      "description": "string",
      "severity": "low",
      "createdAt": "timestamp"
    }
  ]
}
```

### `GET /api/v1/admin/approvals`
Get pending admin approvals.

**Auth:** Bearer token required  
**Role:** `administrator`

**Response:** `200 OK`
```json
[
  {
    "id": "uuid",
    "userId": "uuid",
    "actionType": "create_booking",
    "referenceType": "booking",
    "referenceId": "uuid",
    "status": "pending",
    "requestedBy": "uuid",
    "resolvedBy": null,
    "resolutionNotes": "",
    "createdAt": "timestamp",
    "resolvedAt": null
  }
]
```

### `POST /api/v1/admin/approvals/:id/resolve`
Resolve a pending approval.

**Auth:** Bearer token required  
**Role:** `administrator`

**Request Body:**
```json
{
  "status": "approved|denied",
  "notes": "string"
}
```

**Response:** `200 OK`

### `POST /api/v1/admin/users/:id/blacklist`
Blacklist a user.

**Auth:** Bearer token required  
**Role:** `administrator`

**Request Body:**
```json
{
  "reason": "string"
}
```

**Response:** `200 OK`

### `POST /api/v1/admin/users/:id/unblacklist`
Remove a user from the blacklist.

**Auth:** Bearer token required  
**Role:** `administrator`

**Response:** `200 OK`

---

## Admin

### `GET /api/v1/admin/audit-logs`
Get audit logs (paginated).

**Auth:** Bearer token required  
**Role:** `administrator`

**Query Parameters:**
| Param | Type | Default | Description |
|---|---|---|---|
| `page` | int | 1 | Page number |
| `pageSize` | int | 20 | Items per page |
| `actorId` | string | — | Filter by actor |
| `entityType` | string | — | Filter by entity type |
| `action` | string | — | Filter by action |

**Response:** `200 OK`
```json
{
  "items": [
    {
      "id": "uuid",
      "actor_id": "uuid",
      "action": "POST /api/v1/bookings",
      "entity_type": "api_request",
      "entity_id": null,
      "request_id": "uuid",
      "created_at": "timestamp"
    }
  ],
  "total": 200,
  "page": 1,
  "pageSize": 20,
  "totalPages": 10
}
```

### `GET /api/v1/admin/send-logs`
Get notification send logs (paginated).

**Auth:** Bearer token required  
**Role:** `administrator`

**Query Parameters:**
| Param | Type | Default | Description |
|---|---|---|---|
| `page` | int | 1 | Page number |

**Response:** `200 OK` — Paginated list of `SendLogDTO`

### `GET /api/v1/admin/config`
Get system configuration.

**Auth:** Bearer token required  
**Role:** `administrator`

**Response:** `200 OK`
```json
{
  "dnd_start": "21:00",
  "dnd_end": "08:00",
  "courier_daily_cap": 2500.00,
  "refund_minimum": 1.00,
  "max_cancellations_24h": 8,
  "max_rfqs_10min": 20,
  "download_token_ttl_seconds": 3600
}
```
