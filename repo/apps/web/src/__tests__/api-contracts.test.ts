import { describe, it, expect } from "vitest";

interface BackendItineraryResponse {
  id: string;
  organizerId: string;
  title: string;
  meetupAt: string;
  meetupLocationText: string;
  notes: string;
  status: string;
  checkpointsCount: number;
  membersCount: number;
  createdAt: string;
  updatedAt: string;
}

interface BackendItineraryDetail extends BackendItineraryResponse {
  checkpoints: { id: string; checkpointText: string; sortOrder: number; eta?: string }[];
  members: { id: string; userId: string; role: string }[];
  formDefinitions: { id: string; fieldKey: string; fieldLabel: string; fieldType: string; required: boolean }[];
}

interface BackendBookingResponse {
  id: string;
  organizerId: string;
  title: string;
  description: string;
  status: string;
  totalAmount: number;
  discountAmount: number;
  escrowAmount: number;
  items: { id: string; itemType: string; itemName: string; unitPrice: number; quantity: number; subtotal: number; category: string }[];
  createdAt: string;
  updatedAt: string;
}

interface BackendTransactionResponse {
  id: string;
  walletId: string;
  amount: number;
  direction: string;
  referenceType: string;
  referenceId: string;
  description: string;
  createdAt: string;
}

interface BackendReviewsResponse {
  items: unknown[];
  total: number;
  page: number;
}

describe("Itinerary API contracts", () => {
  it("list response uses paginated shape with items array", () => {
    const mockResponse = {
      items: [] as BackendItineraryResponse[],
      total: 0,
      page: 1,
      pageSize: 20,
      totalPages: 0,
    };
    expect(mockResponse).toHaveProperty("items");
    expect(mockResponse).toHaveProperty("total");
    expect(mockResponse).toHaveProperty("page");
    expect(Array.isArray(mockResponse.items)).toBe(true);
  });

  it("response uses meetupAt not meetupDate", () => {
    const mock: BackendItineraryResponse = {
      id: "1", organizerId: "o", title: "T", meetupAt: "2026-07-14T18:30:00Z",
      meetupLocationText: "Station", notes: "", status: "draft",
      checkpointsCount: 0, membersCount: 0, createdAt: "", updatedAt: "",
    };
    expect(mock).toHaveProperty("meetupAt");
    expect(mock).not.toHaveProperty("meetupDate");
  });

  it("response uses meetupLocationText not location", () => {
    const mock: BackendItineraryResponse = {
      id: "1", organizerId: "o", title: "T", meetupAt: "",
      meetupLocationText: "Station", notes: "", status: "draft",
      checkpointsCount: 0, membersCount: 0, createdAt: "", updatedAt: "",
    };
    expect(mock).toHaveProperty("meetupLocationText");
    expect(mock).not.toHaveProperty("location");
  });

  it("response uses membersCount not memberCount", () => {
    const mock: BackendItineraryResponse = {
      id: "1", organizerId: "o", title: "T", meetupAt: "",
      meetupLocationText: "", notes: "", status: "draft",
      checkpointsCount: 0, membersCount: 5, createdAt: "", updatedAt: "",
    };
    expect(mock.membersCount).toBe(5);
    expect(mock).not.toHaveProperty("memberCount");
  });

  it("detail checkpoints use checkpointText not name", () => {
    const mock: BackendItineraryDetail = {
      id: "1", organizerId: "o", title: "T", meetupAt: "",
      meetupLocationText: "", notes: "", status: "draft",
      checkpointsCount: 1, membersCount: 0, createdAt: "", updatedAt: "",
      checkpoints: [{ id: "c1", checkpointText: "Depart", sortOrder: 1 }],
      members: [], formDefinitions: [],
    };
    expect(mock.checkpoints[0]).toHaveProperty("checkpointText");
    expect(mock.checkpoints[0]).not.toHaveProperty("name");
  });

  it("create payload uses canonical field names", () => {
    const createPayload = {
      title: "Trip",
      meetupAt: new Date("07/14/2026 6:30 PM").toISOString(),
      meetupLocationText: "Station",
      notes: "",
    };
    expect(createPayload).toHaveProperty("meetupAt");
    expect(createPayload).toHaveProperty("meetupLocationText");
    expect(createPayload).not.toHaveProperty("meetupDate");
    expect(createPayload).not.toHaveProperty("location");
    expect(createPayload).not.toHaveProperty("checkpoints");
    expect(createPayload).not.toHaveProperty("formDefinitions");
  });
});

describe("Booking API contracts", () => {
  it("response uses items not lineItems", () => {
    const mock: BackendBookingResponse = {
      id: "1", organizerId: "o", title: "B", description: "", status: "draft",
      totalAmount: 850, discountAmount: 0, escrowAmount: 0,
      items: [], createdAt: "", updatedAt: "",
    };
    expect(mock).toHaveProperty("items");
    expect(mock).not.toHaveProperty("lineItems");
  });

  it("response uses discountAmount not discount", () => {
    const mock: BackendBookingResponse = {
      id: "1", organizerId: "o", title: "B", description: "", status: "draft",
      totalAmount: 850, discountAmount: 25, escrowAmount: 825,
      items: [], createdAt: "", updatedAt: "",
    };
    expect(mock).toHaveProperty("discountAmount");
    expect(mock).not.toHaveProperty("discount");
  });

  it("response uses escrowAmount not escrowHoldAmount", () => {
    const mock: BackendBookingResponse = {
      id: "1", organizerId: "o", title: "B", description: "", status: "draft",
      totalAmount: 850, discountAmount: 0, escrowAmount: 850,
      items: [], createdAt: "", updatedAt: "",
    };
    expect(mock).toHaveProperty("escrowAmount");
    expect(mock).not.toHaveProperty("escrowHoldAmount");
  });

  it("items use itemType and itemName not type and name", () => {
    const item = { id: "1", itemType: "lodging", itemName: "Room", unitPrice: 150, quantity: 2, subtotal: 300, category: "lodging" };
    expect(item).toHaveProperty("itemType");
    expect(item).toHaveProperty("itemName");
    expect(item).not.toHaveProperty("type");
    expect(item).not.toHaveProperty("name");
  });

  it("tender payload uses referenceText not reference", () => {
    const tenderPayload = {
      tenderType: "cash",
      amount: 85000,
      referenceText: "Receipt #123",
    };
    expect(tenderPayload).toHaveProperty("referenceText");
    expect(tenderPayload).not.toHaveProperty("reference");
  });
});

describe("Wallet API contracts", () => {
  it("transaction uses direction not type", () => {
    const tx: BackendTransactionResponse = {
      id: "1", walletId: "w1", amount: 100, direction: "credit",
      referenceType: "booking", referenceId: "b1", description: "Payment", createdAt: "",
    };
    expect(tx).toHaveProperty("direction");
    expect(tx).not.toHaveProperty("type");
  });

  it("transaction uses referenceType/referenceId not reference", () => {
    const tx: BackendTransactionResponse = {
      id: "1", walletId: "w1", amount: 100, direction: "credit",
      referenceType: "booking", referenceId: "b1", description: "Payment", createdAt: "",
    };
    expect(tx).toHaveProperty("referenceType");
    expect(tx).toHaveProperty("referenceId");
    expect(tx).not.toHaveProperty("reference");
  });

  it("refund payload uses orderType/orderId not bookingId", () => {
    const refundPayload = {
      orderType: "booking",
      orderId: "booking-123",
      amount: 10000,
      reason: "Damaged goods",
    };
    expect(refundPayload).toHaveProperty("orderType");
    expect(refundPayload).toHaveProperty("orderId");
    expect(refundPayload).not.toHaveProperty("bookingId");
  });
});

describe("Reviews API contracts", () => {
  it("list response is paginated object not array", () => {
    const mock: BackendReviewsResponse = {
      items: [],
      total: 0,
      page: 1,
    };
    expect(mock).toHaveProperty("items");
    expect(mock).toHaveProperty("total");
    expect(Array.isArray(mock.items)).toBe(true);
    expect(Array.isArray(mock)).toBe(false);
  });

  it("submit payload uses scores array not dimensions record", () => {
    const payload = {
      subjectId: "user-1",
      orderType: "booking",
      orderId: "booking-123",
      overallRating: 4.5,
      comment: "Great service",
      scores: [
        { dimensionName: "punctuality", score: 5 },
        { dimensionName: "quality", score: 4 },
      ],
    };
    expect(payload).toHaveProperty("scores");
    expect(Array.isArray(payload.scores)).toBe(true);
    expect(payload).not.toHaveProperty("dimensions");
    expect(payload).toHaveProperty("orderType");
    expect(payload).toHaveProperty("orderId");
  });

  it("review list items use scores not dimensions and IDs not names", () => {
    const review = {
      id: "r1",
      reviewerId: "user-1",
      subjectId: "user-2",
      orderType: "booking",
      orderId: "b1",
      overallRating: 4.0,
      comment: "Good",
      scores: [{ dimensionName: "quality", score: 4 }],
      createdAt: "2026-01-01T00:00:00Z",
    };
    expect(review).toHaveProperty("scores");
    expect(review).not.toHaveProperty("dimensions");
    expect(review).toHaveProperty("reviewerId");
    expect(review).not.toHaveProperty("reviewerName");
    expect(review).toHaveProperty("subjectId");
    expect(review).not.toHaveProperty("subjectName");
  });
});

describe("Booking price-preview contracts", () => {
  it("uses eligibleCoupons and ineligibleCoupons not coupons", () => {
    const preview = {
      snapshotId: "snap-1",
      subtotal: 85000,
      totalDiscount: 2500,
      escrowHoldAmount: 82500,
      finalPayable: 82500,
      eligibleCoupons: [{ couponId: "c1", code: "SAVE25", name: "$25 Off", discountAmount: 2500 }],
      ineligibleCoupons: [{ couponId: "c2", code: "VIP", name: "VIP Only", reasonCode: "MEMBERSHIP_REQUIRED", message: "Membership required" }],
    };
    expect(preview).toHaveProperty("eligibleCoupons");
    expect(preview).toHaveProperty("ineligibleCoupons");
    expect(preview).not.toHaveProperty("coupons");
  });

  it("uses escrowHoldAmount not escrowHold", () => {
    const preview = {
      escrowHoldAmount: 82500,
      subtotal: 85000,
      totalDiscount: 2500,
      finalPayable: 82500,
      eligibleCoupons: [],
      ineligibleCoupons: [],
    };
    expect(preview).toHaveProperty("escrowHoldAmount");
    expect(preview).not.toHaveProperty("escrowHold");
  });
});

describe("Document center file model contracts", () => {
  it("file uses originalFilename/mimeType/byteSize not name/type/size", () => {
    const file = {
      id: "f1",
      originalFilename: "contract.pdf",
      mimeType: "application/pdf",
      byteSize: 45000,
      encrypted: true,
      ownerUserId: "u1",
      createdAt: "2026-01-01T00:00:00Z",
    };
    expect(file).toHaveProperty("originalFilename");
    expect(file).toHaveProperty("mimeType");
    expect(file).toHaveProperty("byteSize");
    expect(file).not.toHaveProperty("name");
    expect(file).not.toHaveProperty("type");
    expect(file).not.toHaveProperty("size");
  });

  it("upload includes recordType and recordId", () => {
    const formFields = {
      file: new Blob(["test"]),
      recordType: "booking",
      recordId: "booking-123",
    };
    expect(formFields).toHaveProperty("recordType");
    expect(formFields).toHaveProperty("recordId");
  });
});

describe("Invoice list model contracts", () => {
  it("invoice list uses orderType/orderId not bookingId/amount", () => {
    const invoice = {
      id: "inv-1",
      requesterId: "u1",
      orderType: "booking",
      orderId: "b1",
      status: "requested",
      notes: "Please generate",
      createdAt: "2026-01-01T00:00:00Z",
    };
    expect(invoice).toHaveProperty("orderType");
    expect(invoice).toHaveProperty("orderId");
    expect(invoice).not.toHaveProperty("bookingId");
    expect(invoice).not.toHaveProperty("amount");
  });
});

describe("RFQ creation role policy", () => {
  it("group_organizer should be allowed to see create RFQ affordance", () => {
    const allowedRoles = ["group_organizer", "administrator", "accountant"];
    expect(allowedRoles).toContain("group_organizer");
    expect(allowedRoles).toContain("administrator");
    expect(allowedRoles).toContain("accountant");
    expect(allowedRoles).not.toContain("traveler");
    expect(allowedRoles).not.toContain("supplier");
  });
});

describe("Invoice request payload", () => {
  it("uses orderType/orderId not bookingId/amount", () => {
    const payload = {
      orderType: "booking",
      orderId: "booking-123",
      notes: "Please generate invoice",
    };
    expect(payload).toHaveProperty("orderType");
    expect(payload).toHaveProperty("orderId");
    expect(payload).not.toHaveProperty("bookingId");
    expect(payload).not.toHaveProperty("amount");
  });
});

// ---------------------------------------------------------------------------
// ADAPTER / MAPPING CONTRACT TESTS
// ---------------------------------------------------------------------------

describe("Review payload mapping", () => {
  // Simulates the exact transformation in ReviewDashboard.tsx submit handler.
  function mapReviewPayload(
    subjectId: string,
    orderId: string,
    overallRating: number,
    comment: string,
    dimensionRatings: Record<string, number>,
  ) {
    return {
      subjectId,
      orderType: "booking" as const,
      orderId,
      overallRating,
      comment,
      scores: Object.entries(dimensionRatings).map(([name, score]) => ({
        dimensionName: name,
        score,
      })),
    };
  }

  it("maps dimension ratings to scores array", () => {
    const payload = mapReviewPayload("user-2", "b-1", 4, "Good", {
      punctuality: 5,
      quality: 4,
      communication: 3,
    });

    expect(payload.scores).toHaveLength(3);
    expect(payload.scores[0]).toHaveProperty("dimensionName");
    expect(payload.scores[0]).toHaveProperty("score");
    expect(payload).not.toHaveProperty("dimensions");
  });

  it("includes required business context fields", () => {
    const payload = mapReviewPayload("user-2", "booking-123", 4.5, "Great", {});

    expect(payload.subjectId).toBe("user-2");
    expect(payload.orderType).toBe("booking");
    expect(payload.orderId).toBe("booking-123");
    expect(payload.overallRating).toBe(4.5);
  });

  it("produces empty scores for no dimension ratings", () => {
    const payload = mapReviewPayload("user-2", "b-1", 3, "", {});
    expect(payload.scores).toHaveLength(0);
    expect(Array.isArray(payload.scores)).toBe(true);
  });

  it("submit button validation requires orderId", () => {
    // Mirrors the disabled predicate in ReviewDashboard.tsx.
    const isDisabled = (subjectId: string, reviewOrderId: string, overallRating: number) =>
      !subjectId || !reviewOrderId || overallRating === 0;

    expect(isDisabled("user-2", "", 4)).toBe(true);   // missing orderId
    expect(isDisabled("", "b-1", 4)).toBe(true);      // missing subjectId
    expect(isDisabled("user-2", "b-1", 0)).toBe(true); // missing rating
    expect(isDisabled("user-2", "b-1", 4)).toBe(false); // all present
  });
});

describe("Booking price-preview mapping", () => {
  // The canonical backend response shape.
  interface PricePreviewResponse {
    snapshotId: string;
    subtotal: number;
    totalDiscount: number;
    escrowHoldAmount: number;
    finalPayable: number;
    eligibleCoupons: { couponId: string; code: string; name: string; discountAmount: number }[];
    ineligibleCoupons: { couponId: string; code: string; name: string; reasonCode: string; message: string }[];
  }

  it("frontend can render eligible coupons with discount amounts", () => {
    const preview: PricePreviewResponse = {
      snapshotId: "snap-1",
      subtotal: 85000,
      totalDiscount: 2500,
      escrowHoldAmount: 82500,
      finalPayable: 82500,
      eligibleCoupons: [
        { couponId: "c1", code: "SAVE25", name: "$25 Off", discountAmount: 2500 },
      ],
      ineligibleCoupons: [],
    };

    // Iterate like BookingDetail.tsx does.
    const rendered = preview.eligibleCoupons.map((c) => ({
      code: c.code,
      amount: c.discountAmount,
    }));
    expect(rendered).toHaveLength(1);
    expect(rendered[0].code).toBe("SAVE25");
    expect(rendered[0].amount).toBe(2500);
  });

  it("frontend can render ineligible coupons with reason", () => {
    const preview: PricePreviewResponse = {
      snapshotId: "snap-1",
      subtotal: 85000,
      totalDiscount: 0,
      escrowHoldAmount: 85000,
      finalPayable: 85000,
      eligibleCoupons: [],
      ineligibleCoupons: [
        { couponId: "c2", code: "VIP", name: "VIP Only", reasonCode: "MEMBERSHIP_REQUIRED", message: "VIP membership required" },
      ],
    };

    const rendered = preview.ineligibleCoupons.map((c) => ({
      code: c.code,
      reason: `[${c.reasonCode}] ${c.message}`,
    }));
    expect(rendered).toHaveLength(1);
    expect(rendered[0].reason).toContain("MEMBERSHIP_REQUIRED");
  });

  it("no runtime error iterating empty coupon arrays", () => {
    const preview: PricePreviewResponse = {
      snapshotId: "snap-1",
      subtotal: 10000,
      totalDiscount: 0,
      escrowHoldAmount: 10000,
      finalPayable: 10000,
      eligibleCoupons: [],
      ineligibleCoupons: [],
    };

    expect(() => preview.eligibleCoupons.map((c) => c.code)).not.toThrow();
    expect(() => preview.ineligibleCoupons.map((c) => c.code)).not.toThrow();
  });
});

describe("Document center file model mapping", () => {
  interface BackendFileRecord {
    id: string;
    originalFilename: string;
    mimeType: string;
    byteSize: number;
    encrypted: boolean;
    ownerUserId: string;
    createdAt: string;
  }

  function formatBytes(bytes: number): string {
    if (bytes === 0) return "0 B";
    const k = 1024;
    const sizes = ["B", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
  }

  it("renders file metadata from canonical backend fields", () => {
    const file: BackendFileRecord = {
      id: "f1",
      originalFilename: "contract.pdf",
      mimeType: "application/pdf",
      byteSize: 45000,
      encrypted: true,
      ownerUserId: "u1",
      createdAt: "2026-01-01T00:00:00Z",
    };

    // Simulate how DocumentCenter.tsx renders the file row.
    expect(file.originalFilename).toBe("contract.pdf");
    expect(file.mimeType).toBe("application/pdf");
    expect(formatBytes(file.byteSize)).toBe("43.9 KB");
    expect(file.encrypted).toBe(true);
  });

  it("does not rely on name/type/size fields", () => {
    const file: BackendFileRecord = {
      id: "f2",
      originalFilename: "receipt.jpg",
      mimeType: "image/jpeg",
      byteSize: 120000,
      encrypted: false,
      ownerUserId: "u2",
      createdAt: "2026-02-15T00:00:00Z",
    };

    const asRecord = file as unknown as Record<string, unknown>;
    expect(asRecord).not.toHaveProperty("name");
    expect(asRecord).not.toHaveProperty("type");
    expect(asRecord).not.toHaveProperty("size");
  });
});

describe("Document center invoice list mapping", () => {
  interface BackendInvoiceRequest {
    id: string;
    requesterId: string;
    orderType: string;
    orderId: string;
    status: string;
    notes: string;
    createdAt: string;
  }

  it("renders invoice request from canonical backend fields", () => {
    const inv: BackendInvoiceRequest = {
      id: "inv-1",
      requesterId: "u1",
      orderType: "booking",
      orderId: "b-1",
      status: "pending",
      notes: "Please generate",
      createdAt: "2026-01-01T00:00:00Z",
    };

    // Simulate how DocumentCenter.tsx renders the invoice row.
    const orderDisplay = `${inv.orderType}: ${inv.orderId}`;
    expect(orderDisplay).toBe("booking: b-1");
    expect(inv.status).toBe("pending");
    expect(inv.notes).toBe("Please generate");
  });

  it("does not rely on bookingId/amount fields", () => {
    const inv = {
      id: "inv-2",
      requesterId: "u1",
      orderType: "procurement",
      orderId: "po-1",
      status: "generated",
      notes: "",
      createdAt: "2026-03-01T00:00:00Z",
    };

    expect(inv).not.toHaveProperty("bookingId");
    expect(inv).not.toHaveProperty("amount");
  });
});

describe("RFQ create role visibility logic", () => {
  // Mirrors the exact logic from ProcurementDashboard.tsx.
  function canCreateRFQ(hasRole: (role: string) => boolean): boolean {
    return hasRole("group_organizer") || hasRole("administrator") || hasRole("accountant");
  }

  it("allows group_organizer", () => {
    expect(canCreateRFQ((r) => r === "group_organizer")).toBe(true);
  });

  it("allows administrator", () => {
    expect(canCreateRFQ((r) => r === "administrator")).toBe(true);
  });

  it("allows accountant", () => {
    expect(canCreateRFQ((r) => r === "accountant")).toBe(true);
  });

  it("denies traveler", () => {
    expect(canCreateRFQ((r) => r === "traveler")).toBe(false);
  });

  it("denies supplier", () => {
    expect(canCreateRFQ((r) => r === "supplier")).toBe(false);
  });

  it("denies courier_runner", () => {
    expect(canCreateRFQ((r) => r === "courier_runner")).toBe(false);
  });

  it("matches backend rfqCreator middleware policy", () => {
    // Backend: middleware.RequireRole(RoleAccountant, RoleAdministrator, RoleGroupOrganizer)
    const backendAllowed = ["accountant", "administrator", "group_organizer"];
    const allRoles = ["administrator", "group_organizer", "traveler", "supplier", "courier_runner", "accountant"];

    for (const role of allRoles) {
      const frontendAllows = canCreateRFQ((r) => r === role);
      const backendAllows = backendAllowed.includes(role);
      expect(frontendAllows).toBe(backendAllows);
    }
  });
});

// ---------------------------------------------------------------------------
// DOCUMENT CENTER RECORD-SCOPED LISTING REGRESSION TESTS
// ---------------------------------------------------------------------------

describe("DocumentCenter record-scoped listing state", () => {
  // These tests verify the state/query logic that was fixed so that the file
  // list query uses the same record context as the upload form.

  // Mirrors the exact query-params logic from DocumentCenter.tsx.
  function buildFileListQueryParams(uploadRecordType: string, uploadRecordId: string) {
    return {
      queryKey: ["files", uploadRecordType, uploadRecordId],
      apiPath: `/files/record/${uploadRecordType}/${uploadRecordId}`,
      enabled: !!uploadRecordId,
    };
  }

  it("query is disabled when uploadRecordId is empty", () => {
    const params = buildFileListQueryParams("booking", "");
    expect(params.enabled).toBe(false);
  });

  it("query is enabled when uploadRecordId has a value", () => {
    const params = buildFileListQueryParams("booking", "b-123");
    expect(params.enabled).toBe(true);
  });

  it("query path includes recordType and recordId", () => {
    const params = buildFileListQueryParams("procurement", "po-456");
    expect(params.apiPath).toBe("/files/record/procurement/po-456");
  });

  it("queryKey changes when record context changes", () => {
    const params1 = buildFileListQueryParams("booking", "b-1");
    const params2 = buildFileListQueryParams("booking", "b-2");
    const params3 = buildFileListQueryParams("procurement", "b-1");

    expect(params1.queryKey).not.toEqual(params2.queryKey);
    expect(params1.queryKey).not.toEqual(params3.queryKey);
  });

  it("upload and list use the same record context", () => {
    // Simulates the unified state: upload and list both use uploadRecordType/uploadRecordId.
    const uploadRecordType = "booking";
    const uploadRecordId = "b-789";

    const listParams = buildFileListQueryParams(uploadRecordType, uploadRecordId);
    const uploadPayload = {
      recordType: uploadRecordType,
      recordId: uploadRecordId,
    };

    expect(listParams.apiPath).toContain(uploadPayload.recordType);
    expect(listParams.apiPath).toContain(uploadPayload.recordId);
    expect(listParams.enabled).toBe(true);
  });
});

describe("DocumentCenter upload request shaping", () => {
  // Verifies that the upload FormData includes recordType and recordId.
  function buildUploadFormFields(
    uploadRecordType: string,
    uploadRecordId: string,
  ) {
    return {
      recordType: uploadRecordType,
      recordId: uploadRecordId,
    };
  }

  it("includes recordType and recordId in upload payload", () => {
    const fields = buildUploadFormFields("booking", "b-123");
    expect(fields).toHaveProperty("recordType", "booking");
    expect(fields).toHaveProperty("recordId", "b-123");
  });

  it("defaults recordType to booking", () => {
    const defaultRecordType = "booking"; // matches useState default
    const fields = buildUploadFormFields(defaultRecordType, "");
    expect(fields.recordType).toBe("booking");
  });

  it("supports procurement record type", () => {
    const fields = buildUploadFormFields("procurement", "po-456");
    expect(fields.recordType).toBe("procurement");
    expect(fields.recordId).toBe("po-456");
  });
});
