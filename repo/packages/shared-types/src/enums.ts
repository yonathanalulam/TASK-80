
export enum UserStatus {
  Active = 'active',
  Suspended = 'suspended',
  Banned = 'banned',
}

export enum Role {
  Traveler = 'traveler',
  GroupOrganizer = 'group_organizer',
  Supplier = 'supplier',
  CourierRunner = 'courier_runner',
  Accountant = 'accountant',
  Administrator = 'administrator',
}

export enum ItineraryStatus {
  Draft = 'draft',
  Published = 'published',
  Revised = 'revised',
  InProgress = 'in_progress',
  Completed = 'completed',
  Cancelled = 'cancelled',
  Archived = 'archived',
}

export enum BookingStatus {
  Draft = 'draft',
  PendingPricing = 'pending_pricing',
  PendingPaymentRecord = 'pending_payment_record',
  PaidHeldInEscrow = 'paid_held_in_escrow',
  Confirmed = 'confirmed',
  PartiallyFulfilled = 'partially_fulfilled',
  Fulfilled = 'fulfilled',
  Completed = 'completed',
  Cancelled = 'cancelled',
  RefundedPartial = 'refunded_partial',
  RefundedFull = 'refunded_full',
  Closed = 'closed',
}

export enum DiscountType {
  ThresholdFixed = 'threshold_fixed',
  Percentage = 'percentage',
  NewUserGift = 'new_user_gift',
}

export enum CouponIneligibilityReason {
  RoomTypeRestricted = 'ROOM_TYPE_RESTRICTED',
  DateOutOfRange = 'DATE_OUT_OF_RANGE',
  MinSpendNotMet = 'MIN_SPEND_NOT_MET',
  MembershipRequired = 'MEMBERSHIP_REQUIRED',
  NewUserOnly = 'NEW_USER_ONLY',
  StackingNotAllowed = 'STACKING_NOT_ALLOWED',
  UsageLimitReached = 'USAGE_LIMIT_REACHED',
  Expired = 'EXPIRED',
  Inactive = 'INACTIVE',
  CategoryRestricted = 'CATEGORY_RESTRICTED',
  AlreadyRedeemed = 'ALREADY_REDEEMED',
}

export enum TenderType {
  Cash = 'cash',
  CardOnFileRecorded = 'card_on_file_recorded',
  BankTransferRecorded = 'bank_transfer_recorded',
  OtherManual = 'other_manual',
}

export enum EscrowStatus {
  Held = 'held',
  PartiallyReleased = 'partially_released',
  Released = 'released',
  Refunded = 'refunded',
}

export enum WithdrawalStatus {
  Requested = 'requested',
  UnderReview = 'under_review',
  Approved = 'approved',
  Rejected = 'rejected',
  Settled = 'settled',
}

export enum WalletType {
  Customer = 'customer',
  Supplier = 'supplier',
  Courier = 'courier',
  PlatformClearing = 'platform_clearing',
  EscrowControl = 'escrow_control',
  RefundClearing = 'refund_clearing',
  FeeRevenue = 'fee_revenue',
}

export enum RFQStatus {
  Draft = 'draft',
  Issued = 'issued',
  Responded = 'responded',
  ComparisonReady = 'comparison_ready',
  Selected = 'selected',
  ClosedNoAward = 'closed_no_award',
  ConvertedToPO = 'converted_to_po',
}

export enum POStatus {
  Draft = 'draft',
  Issued = 'issued',
  Accepted = 'accepted',
  PartiallyDelivered = 'partially_delivered',
  Delivered = 'delivered',
  InspectionPending = 'inspection_pending',
  ExceptionOpen = 'exception_open',
  Closed = 'closed',
}

export enum InspectionStatus {
  Pending = 'pending',
  Passed = 'passed',
  Failed = 'failed',
}

export enum DiscrepancyType {
  Shortage = 'shortage',
  Damage = 'damage',
  WrongItem = 'wrong_item',
  LateDelivery = 'late_delivery',
  ServiceDeviation = 'service_deviation',
  Other = 'other',
}

export enum ExceptionStatus {
  Open = 'open',
  PendingFinancialResolution = 'pending_financial_resolution',
  PendingWaiver = 'pending_waiver',
  ReadyToClose = 'ready_to_close',
  Closed = 'closed',
}

export enum InvoiceRequestStatus {
  Requested = 'requested',
  Approved = 'approved',
  Generated = 'generated',
  Delivered = 'delivered',
  Cancelled = 'cancelled',
}

export enum CreditTier {
  Bronze = 'bronze',
  Silver = 'silver',
  Gold = 'gold',
  Platinum = 'platinum',
  Restricted = 'restricted',
}

export enum NotificationChannel {
  InApp = 'in_app',
  CallbackQueue = 'callback_queue',
}
