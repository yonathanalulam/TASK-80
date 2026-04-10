
import {
  BookingStatus,
  CouponIneligibilityReason,
  CreditTier,
  DiscountType,
  EscrowStatus,
  ExceptionStatus,
  InspectionStatus,
  InvoiceRequestStatus,
  ItineraryStatus,
  POStatus,
  RFQStatus,
  Role,
  TenderType,
  UserStatus,
  WithdrawalStatus,
} from './enums';

export interface ApiResponse<T = unknown> {
  success: boolean;
  data?: T;
  error?: ApiError;
}

export interface ApiError {
  code: string;
  message: string;
  details?: Record<string, string[]>;
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  user: UserDTO;
}

export interface UserDTO {
  id: string;
  email: string;
  status: UserStatus;
  roles: Role[];
  profile?: UserProfileDTO;
}

export interface UserProfileDTO {
  displayName: string;
  phoneMasked?: string;
}

export interface CreateItineraryRequest {
  title: string;
  meetupAt: string;
  meetupLocationText: string;
  notes?: string;
  checkpoints?: CreateCheckpointRequest[];
  formDefinitions?: CreateFormDefinitionRequest[];
}

export interface ItineraryDTO {
  id: string;
  organizerId: string;
  organizerName?: string;
  title: string;
  meetupAt: string;
  meetupLocationText: string;
  notes?: string;
  status: ItineraryStatus;
  publishedAt?: string;
  checkpoints: CheckpointDTO[];
  memberCount: number;
  createdAt: string;
  updatedAt: string;
}

export interface CheckpointDTO {
  id: string;
  sortOrder: number;
  checkpointText: string;
  eta?: string;
}

export interface CreateCheckpointRequest {
  checkpointText: string;
  sortOrder: number;
  eta?: string;
}

export interface FormDefinitionDTO {
  id: string;
  fieldKey: string;
  fieldLabel: string;
  fieldType: string;
  required: boolean;
  optionsJson?: unknown;
  validationJson?: unknown;
  sortOrder: number;
  active: boolean;
}

export interface CreateFormDefinitionRequest {
  fieldKey: string;
  fieldLabel: string;
  fieldType: string;
  required: boolean;
  options?: string[];
  validation?: Record<string, unknown>;
  sortOrder: number;
}

export interface FormSubmissionDTO {
  id: string;
  memberUserId: string;
  memberName?: string;
  payload: Record<string, unknown>;
  submittedAt: string;
}

export interface ChangeEventDTO {
  id: string;
  actorId: string;
  actorName?: string;
  changeType: string;
  summary: string;
  diffJson?: unknown;
  visibleFrom: string;
  createdAt: string;
}

export interface BookingDTO {
  id: string;
  organizerId: string;
  itineraryId?: string;
  title: string;
  description?: string;
  status: BookingStatus;
  totalAmount: string;
  discountAmount: string;
  escrowAmount: string;
  items: BookingItemDTO[];
  createdAt: string;
  updatedAt: string;
}

export interface BookingItemDTO {
  id: string;
  itemType: string;
  itemName: string;
  description?: string;
  unitPrice: string;
  quantity: number;
  subtotal: string;
  category: string;
}

export interface CouponDTO {
  id: string;
  code: string;
  name: string;
  discountType: DiscountType;
  amount?: string;
  minSpend?: string;
  percentOff?: string;
  validFrom: string;
  validTo: string;
  exclusive: boolean;
  active: boolean;
}

export interface PricePreviewRequest {
  bookingId: string;
  couponCodes?: string[];
  idempotencyKey?: string;
}

export interface PricePreviewResponse {
  eligibleCoupons: AppliedCouponDTO[];
  ineligibleCoupons: IneligibleCouponDTO[];
  appliedDiscounts: AppliedDiscountDTO[];
  subtotal: string;
  totalDiscount: string;
  escrowHoldAmount: string;
  finalPayable: string;
  pricingSnapshotId: string;
}

export interface AppliedCouponDTO {
  couponId: string;
  code: string;
  name: string;
  discountAmount: string;
}

export interface IneligibleCouponDTO {
  couponId: string;
  code: string;
  name: string;
  reason: CouponIneligibilityReason;
  message: string;
}

export interface AppliedDiscountDTO {
  type: string;
  description: string;
  amount: string;
}

export interface CheckoutRequest {
  bookingId: string;
  pricingSnapshotId: string;
  couponCodes?: string[];
  idempotencyKey: string;
}

export interface CheckoutResponse {
  bookingId: string;
  status: BookingStatus;
  escrowId: string;
  totalCharged: string;
  pricingSnapshotId: string;
}

export interface WalletDTO {
  id: string;
  ownerId?: string;
  walletType: string;
  balance: string;
  currency: string;
}

export interface TenderRecordRequest {
  orderType: string;
  orderId: string;
  tenderType: TenderType;
  amount: string;
  referenceText?: string;
}

export interface RefundRequest {
  orderType: string;
  orderId: string;
  refundAmount: string;
  refundReason: string;
}

export interface WithdrawalRequestDTO {
  id: string;
  courierId: string;
  requestAmount: string;
  status: WithdrawalStatus;
  requestedAt: string;
  settledAt?: string;
}

export interface EscrowDTO {
  id: string;
  orderType: string;
  orderId: string;
  amountHeld: string;
  amountReleased: string;
  amountRefunded: string;
  status: EscrowStatus;
}

export interface JournalEntryDTO {
  id: string;
  entryType: string;
  referenceType: string;
  referenceId: string;
  description: string;
  effectiveAt: string;
  lines: JournalLineDTO[];
}

export interface JournalLineDTO {
  id: string;
  accountCode: string;
  direction: 'debit' | 'credit';
  amount: string;
  counterpartyId?: string;
}

export interface ReconciliationReportDTO {
  runDate: string;
  openingBalance: string;
  inflows: string;
  outflows: string;
  heldInEscrow: string;
  released: string;
  refunded: string;
  netPayable: string;
  unreconciledItems: number;
}

export interface RFQDTO {
  id: string;
  createdBy: string;
  title: string;
  description?: string;
  deadline: string;
  status: RFQStatus;
  items: RFQItemDTO[];
  quotes: RFQQuoteDTO[];
  createdAt: string;
}

export interface RFQItemDTO {
  id: string;
  itemName: string;
  specifications?: string;
  quantity: number;
  unit: string;
  sortOrder: number;
}

export interface RFQQuoteDTO {
  id: string;
  supplierId: string;
  supplierName?: string;
  totalAmount: string;
  leadTimeDays: number;
  notes?: string;
  items: RFQQuoteItemDTO[];
  submittedAt: string;
}

export interface RFQQuoteItemDTO {
  rfqItemId: string;
  unitPrice: string;
  quantity: number;
  subtotal: string;
  notes?: string;
}

export interface PurchaseOrderDTO {
  id: string;
  supplierId: string;
  supplierName?: string;
  poNumber: string;
  promisedDate: string;
  status: POStatus;
  totalAmount: string;
  items: POItemDTO[];
  createdAt: string;
}

export interface POItemDTO {
  id: string;
  itemName: string;
  specifications?: string;
  unitPrice: string;
  quantity: number;
  subtotal: string;
}

export interface InspectionDTO {
  id: string;
  deliveryId?: string;
  poId: string;
  inspectorId: string;
  status: InspectionStatus;
  notes?: string;
  inspectedAt?: string;
}

export interface ExceptionCaseDTO {
  id: string;
  referenceType: string;
  referenceId: string;
  status: ExceptionStatus;
  openedReason: string;
  hasSettlementAdjustment: boolean;
  hasWaiver: boolean;
  openedAt: string;
  closedAt?: string;
}

export interface FileUploadResponse {
  id: string;
  originalFilename: string;
  mimeType: string;
  byteSize: number;
  encrypted: boolean;
}

export interface DownloadTokenResponse {
  token: string;
  expiresAt: string;
  downloadUrl: string;
}

export interface InvoiceRequestDTO {
  id: string;
  requesterId: string;
  orderType: string;
  orderId: string;
  status: InvoiceRequestStatus;
  createdAt: string;
}

export interface InvoiceDTO {
  id: string;
  invoiceNumber: string;
  orderType: string;
  orderId: string;
  amount: string;
  fileId?: string;
  generatedAt?: string;
}

export interface ReviewDTO {
  id: string;
  reviewerId: string;
  reviewerName?: string;
  subjectId: string;
  orderType: string;
  orderId: string;
  overallRating: number;
  comment?: string;
  scores: ReviewScoreDTO[];
  createdAt: string;
}

export interface ReviewScoreDTO {
  dimensionName: string;
  dimensionLabel: string;
  score: number;
}

export interface CreditSnapshotDTO {
  userId: string;
  tier: CreditTier;
  totalTransactions: number;
  avgRating: number;
  violationCount: number;
  computedAt: string;
}

export interface NotificationDTO {
  id: string;
  eventType: string;
  sourceType: string;
  sourceId: string;
  channel: string;
  status: string;
  payload?: unknown;
  deliveredAt?: string;
  readAt?: string;
  createdAt: string;
}

export interface MessageDTO {
  id: string;
  senderId?: string;
  subject: string;
  body: string;
  readAt?: string;
  createdAt: string;
}
