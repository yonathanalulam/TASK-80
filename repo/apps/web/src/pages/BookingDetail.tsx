import { useState } from "react";
import { useParams, Link } from "react-router-dom";
import { useQuery, useMutation } from "@tanstack/react-query";
import api from "@/lib/api";

interface LineItem {
  id: string;
  itemName: string;
  itemType: string;
  description: string;
  category: string;
  unitPrice: number;
  quantity: number;
  subtotal: number;
}

interface PaymentRecord {
  id: string;
  tenderType: string;
  amount: number;
  reference: string;
  createdAt: string;
}

interface Booking {
  id: string;
  title: string;
  description: string;
  status: string;
  totalAmount: number;
  discountAmount: number;
  escrowAmount: number;
  items: LineItem[];
  paymentRecords: PaymentRecord[];
  createdAt: string;
  updatedAt: string;
}

interface PricePreview {
  snapshotId: string;
  subtotal: number;
  totalDiscount: number;
  escrowHoldAmount: number;
  finalPayable: number;
  eligibleCoupons: { couponId: string; code: string; name: string; discountAmount: number }[];
  ineligibleCoupons: { couponId: string; code: string; name: string; reasonCode: string; message: string }[];
}

const STATUS_COLORS: Record<string, string> = {
  draft: "bg-gray-100 text-gray-700",
  pending_pricing: "bg-yellow-100 text-yellow-800",
  priced: "bg-blue-100 text-blue-700",
  paid_held_in_escrow: "bg-blue-100 text-blue-700",
  completed: "bg-green-100 text-green-700",
  cancelled: "bg-red-100 text-red-700",
  refunded: "bg-orange-100 text-orange-700",
  disputed: "bg-red-100 text-red-700",
};

const TENDER_TYPES = [
  { value: "cash", label: "Cash" },
  { value: "card_on_file_recorded", label: "Card on File" },
  { value: "bank_transfer_recorded", label: "Bank Transfer" },
  { value: "other_manual", label: "Other (Manual)" },
];

function formatMoney(cents: number): string {
  return `$${(cents / 100).toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function statusLabel(status: string): string {
  return status.replace(/_/g, " ").replace(/\b\w/g, (c) => c.toUpperCase());
}

function generateUUID(): string {
  return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0;
    const v = c === "x" ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
}

export default function BookingDetail() {
  const { id } = useParams<{ id: string }>();
  const [couponCodes, setCouponCodes] = useState("");
  const [pricePreview, setPricePreview] = useState<PricePreview | null>(null);
  const [showCheckoutModal, setShowCheckoutModal] = useState(false);
  const [tenderType, setTenderType] = useState("cash");
  const [tenderAmount, setTenderAmount] = useState("");
  const [tenderReference, setTenderReference] = useState("");
  const [checkoutSuccess, setCheckoutSuccess] = useState(false);
  const [checkoutError, setCheckoutError] = useState("");

  const {
    data: booking,
    isLoading,
    refetch,
  } = useQuery<Booking>({
    queryKey: ["booking", id],
    queryFn: async () => {
      const { data } = await api.get(`/bookings/${id}`);
      return data;
    },
  });

  const previewMutation = useMutation({
    mutationFn: async (codes: string[]) => {
      const { data } = await api.post(`/bookings/${id}/price-preview`, {
        couponCodes: codes,
      });
      return data as PricePreview;
    },
    onSuccess: (data) => setPricePreview(data),
  });

  const checkoutMutation = useMutation({
    mutationFn: async () => {
      const idempotencyKey = generateUUID();
      const { data } = await api.post(`/bookings/${id}/checkout`, {
        pricingSnapshotId: pricePreview?.snapshotId || "",
        couponCodes: couponCodes.split(",").map((c) => c.trim()).filter(Boolean),
        idempotencyKey,
      }, {
        headers: { "Idempotency-Key": idempotencyKey },
      });
      return data;
    },
    onSuccess: () => {
      setCheckoutSuccess(true);
      setShowCheckoutModal(false);
      refetch();
    },
    onError: (err: unknown) => {
      const message =
        err instanceof Error ? err.message : "Checkout failed. Please try again.";
      setCheckoutError(message);
    },
  });

  const recordTenderMutation = useMutation({
    mutationFn: async (payload: {
      tenderType: string;
      amount: number;
      referenceText: string;
    }) => {
      const { data } = await api.post(`/bookings/${id}/record-tender`, payload);
      return data;
    },
    onSuccess: () => {
      setShowCheckoutModal(false);
      refetch();
    },
    onError: (err: unknown) => {
      const message =
        err instanceof Error ? err.message : "Failed to record tender. Please try again.";
      setCheckoutError(message);
    },
  });

  const handlePreviewPrice = () => {
    const codes = couponCodes
      .split(",")
      .map((c) => c.trim())
      .filter(Boolean);
    previewMutation.mutate(codes);
  };

  const handleCheckout = () => {
    setCheckoutError("");
    checkoutMutation.mutate();
  };

  const handleRecordTender = () => {
    setCheckoutError("");
    recordTenderMutation.mutate({
      tenderType,
      amount: Math.round(parseFloat(tenderAmount) * 100),
      referenceText: tenderReference,
    });
  };

  const isDraftOrPending =
    booking?.status === "draft" || booking?.status === "pending_pricing";
  const isCheckedOut =
    booking?.status === "paid_held_in_escrow" ||
    booking?.status === "completed" ||
    booking?.status === "refunded";

  if (isLoading) {
    return <div className="py-12 text-center text-gray-500">Loading booking...</div>;
  }

  if (!booking) {
    return (
      <div className="py-12 text-center text-gray-500">Booking not found.</div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between">
        <div>
          <div className="flex items-center gap-3">
            <Link
              to="/bookings"
              className="text-sm text-gray-500 hover:text-gray-700"
            >
              Bookings
            </Link>
            <span className="text-gray-300">/</span>
            <h1 className="text-2xl font-bold text-gray-900">{booking.title}</h1>
          </div>
          {booking.description && (
            <p className="mt-1 text-gray-500">{booking.description}</p>
          )}
        </div>
        <span
          className={`rounded-full px-3 py-1 text-sm font-medium ${STATUS_COLORS[booking.status] ?? "bg-gray-100 text-gray-700"}`}
        >
          {statusLabel(booking.status)}
        </span>
      </div>

      <div className="rounded-xl border border-gray-200 bg-white p-6">
        <h2 className="mb-4 text-lg font-semibold text-gray-900">
          Booking Details
        </h2>
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
          <div>
            <span className="text-sm font-medium text-gray-500">Total Amount</span>
            <p className="mt-1 text-gray-900">{formatMoney(booking.totalAmount)}</p>
          </div>
          <div>
            <span className="text-sm font-medium text-gray-500">Discount</span>
            <p className="mt-1 text-gray-900">
              {booking.discountAmount ? formatMoney(booking.discountAmount) : "--"}
            </p>
          </div>
          <div>
            <span className="text-sm font-medium text-gray-500">Created</span>
            <p className="mt-1 text-gray-900">{formatDate(booking.createdAt)}</p>
          </div>
          <div>
            <span className="text-sm font-medium text-gray-500">Updated</span>
            <p className="mt-1 text-gray-900">{formatDate(booking.updatedAt)}</p>
          </div>
        </div>
      </div>

      {booking.items?.length > 0 && (
        <div className="rounded-xl border border-gray-200 bg-white">
          <div className="border-b border-gray-200 px-6 py-4">
            <h2 className="text-lg font-semibold text-gray-900">Line Items</h2>
          </div>
          <table className="w-full text-left text-sm">
            <thead className="border-b border-gray-200 bg-gray-50">
              <tr>
                <th className="px-6 py-3 font-medium text-gray-600">Name</th>
                <th className="px-6 py-3 font-medium text-gray-600">Type</th>
                <th className="px-6 py-3 font-medium text-gray-600">Category</th>
                <th className="px-6 py-3 text-right font-medium text-gray-600">
                  Unit Price
                </th>
                <th className="px-6 py-3 text-right font-medium text-gray-600">
                  Qty
                </th>
                <th className="px-6 py-3 text-right font-medium text-gray-600">
                  Subtotal
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {booking.items.map((item) => (
                <tr key={item.id}>
                  <td className="px-6 py-3 font-medium text-gray-900">
                    {item.itemName}
                  </td>
                  <td className="px-6 py-3 capitalize text-gray-600">
                    {item.itemType}
                  </td>
                  <td className="px-6 py-3 text-gray-600">{item.category}</td>
                  <td className="px-6 py-3 text-right text-gray-700">
                    {formatMoney(item.unitPrice)}
                  </td>
                  <td className="px-6 py-3 text-right text-gray-700">
                    {item.quantity}
                  </td>
                  <td className="px-6 py-3 text-right font-medium text-gray-900">
                    {formatMoney(item.subtotal)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {isDraftOrPending && (
        <div className="rounded-xl border border-gray-200 bg-white p-6">
          <h2 className="mb-4 text-lg font-semibold text-gray-900">
            Apply Coupons
          </h2>
          <div className="flex gap-3">
            <input
              type="text"
              value={couponCodes}
              onChange={(e) => setCouponCodes(e.target.value)}
              placeholder="Enter coupon codes (comma-separated)"
              className="flex-1 rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none"
            />
            <button
              onClick={handlePreviewPrice}
              disabled={!couponCodes.trim() || previewMutation.isPending}
              className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-indigo-700 disabled:opacity-50"
            >
              {previewMutation.isPending ? "Loading..." : "Preview Price"}
            </button>
          </div>

          {previewMutation.isError && (
            <p className="mt-3 text-sm text-red-600">
              Failed to preview pricing. Please check your coupon codes and try again.
            </p>
          )}

          {pricePreview && (
            <div className="mt-4 space-y-4">
              <div className="rounded-lg border border-gray-200 bg-gray-50 p-4">
                <h3 className="mb-3 text-sm font-semibold text-gray-700">
                  Pricing Breakdown
                </h3>
                <div className="space-y-2 text-sm">
                  <div className="flex justify-between">
                    <span className="text-gray-600">Subtotal</span>
                    <span className="text-gray-900">
                      {formatMoney(pricePreview.subtotal)}
                    </span>
                  </div>

                  {pricePreview.eligibleCoupons.map((coupon) => (
                    <div key={coupon.code} className="flex justify-between">
                      <span className="flex items-center gap-2">
                        <span className="inline-block h-2 w-2 rounded-full bg-green-500" />
                        <span className="text-gray-600">{coupon.code}</span>
                      </span>
                      <span className="text-green-700">
                        -{formatMoney(coupon.discountAmount)}
                      </span>
                    </div>
                  ))}

                  {pricePreview.ineligibleCoupons.map((coupon) => (
                    <div key={coupon.code} className="flex justify-between">
                      <span className="flex items-center gap-2">
                        <span className="inline-block h-2 w-2 rounded-full bg-red-500" />
                        <span className="text-gray-600">{coupon.code}</span>
                        <span className="text-xs text-red-500">[{coupon.reasonCode}] {coupon.message}</span>
                      </span>
                      <span className="text-gray-400">--</span>
                    </div>
                  ))}

                  <div className="border-t border-gray-200 pt-2">
                    <div className="flex justify-between">
                      <span className="text-gray-600">Total Discount</span>
                      <span className="font-medium text-green-700">
                        -{formatMoney(pricePreview.totalDiscount)}
                      </span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-600">Escrow Hold</span>
                      <span className="text-gray-900">
                        {formatMoney(pricePreview.escrowHoldAmount)}
                      </span>
                    </div>
                    <div className="flex justify-between text-base font-semibold">
                      <span className="text-gray-900">Final Payable</span>
                      <span className="text-gray-900">
                        {formatMoney(pricePreview.finalPayable)}
                      </span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>
      )}

      {isDraftOrPending && (
        <div className="rounded-xl border border-gray-200 bg-white p-6">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold text-gray-900">Checkout</h2>
            <div className="flex gap-2">
              <button
                onClick={handleCheckout}
                disabled={!pricePreview || checkoutMutation.isPending}
                className="rounded-lg bg-green-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-green-700 disabled:opacity-50"
              >
                {checkoutMutation.isPending ? "Processing..." : "Proceed to Checkout"}
              </button>
              <button
                onClick={() => {
                  setCheckoutError("");
                  setShowCheckoutModal(true);
                }}
                className="rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-100"
              >
                Record Tender
              </button>
            </div>
          </div>
          {checkoutSuccess && (
            <div className="mt-3 rounded-lg bg-green-50 p-3 text-sm text-green-700">
              Checkout completed successfully! Payment has been recorded.
            </div>
          )}
          {checkoutError && (
            <div className="mt-3 rounded-lg bg-red-50 p-3 text-sm text-red-700">
              {checkoutError}
            </div>
          )}
        </div>
      )}

      {showCheckoutModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="mx-4 w-full max-w-md rounded-xl bg-white p-6 shadow-xl">
            <h3 className="mb-4 text-lg font-semibold text-gray-900">
              Record Tender
            </h3>
            <div className="space-y-4">
              <div>
                <label className="mb-1 block text-sm font-medium text-gray-700">
                  Tender Type
                </label>
                <select
                  value={tenderType}
                  onChange={(e) => setTenderType(e.target.value)}
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none"
                >
                  {TENDER_TYPES.map((t) => (
                    <option key={t.value} value={t.value}>
                      {t.label}
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label className="mb-1 block text-sm font-medium text-gray-700">
                  Amount ($)
                </label>
                <input
                  type="number"
                  step="0.01"
                  value={tenderAmount}
                  onChange={(e) => setTenderAmount(e.target.value)}
                  placeholder="0.00"
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none"
                />
              </div>
              <div>
                <label className="mb-1 block text-sm font-medium text-gray-700">
                  Reference
                </label>
                <input
                  type="text"
                  value={tenderReference}
                  onChange={(e) => setTenderReference(e.target.value)}
                  placeholder="Transaction reference or receipt number"
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none"
                />
              </div>
              {checkoutError && (
                <p className="text-sm text-red-600">{checkoutError}</p>
              )}
            </div>
            <div className="mt-6 flex justify-end gap-3">
              <button
                onClick={() => setShowCheckoutModal(false)}
                className="rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100"
              >
                Cancel
              </button>
              <button
                onClick={handleRecordTender}
                disabled={
                  !tenderAmount || parseFloat(tenderAmount) <= 0 || recordTenderMutation.isPending
                }
                className="rounded-lg bg-green-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-green-700 disabled:opacity-50"
              >
                {recordTenderMutation.isPending ? "Processing..." : "Record Tender"}
              </button>
            </div>
          </div>
        </div>
      )}

      {isCheckedOut && (
        <div className="space-y-6">
          {booking.escrowAmount > 0 && (
            <div className="rounded-xl border border-blue-200 bg-blue-50 p-6">
              <h2 className="mb-2 text-lg font-semibold text-blue-900">
                Escrow Status
              </h2>
              <p className="text-sm text-blue-700">
                Amount held in escrow:{" "}
                <span className="font-semibold">
                  {formatMoney(booking.escrowAmount)}
                </span>
              </p>
              <p className="mt-1 text-sm text-blue-600">
                Status:{" "}
                <span className="font-medium capitalize">
                  {statusLabel(booking.status)}
                </span>
              </p>
            </div>
          )}

          {booking.paymentRecords?.length > 0 && (
            <div className="rounded-xl border border-gray-200 bg-white">
              <div className="border-b border-gray-200 px-6 py-4">
                <h2 className="text-lg font-semibold text-gray-900">
                  Payment Records
                </h2>
              </div>
              <table className="w-full text-left text-sm">
                <thead className="border-b border-gray-200 bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 font-medium text-gray-600">
                      Tender Type
                    </th>
                    <th className="px-6 py-3 text-right font-medium text-gray-600">
                      Amount
                    </th>
                    <th className="px-6 py-3 font-medium text-gray-600">
                      Reference
                    </th>
                    <th className="px-6 py-3 font-medium text-gray-600">Date</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-100">
                  {booking.paymentRecords.map((record) => (
                    <tr key={record.id}>
                      <td className="px-6 py-3 capitalize text-gray-900">
                        {record.tenderType.replace(/_/g, " ")}
                      </td>
                      <td className="px-6 py-3 text-right font-medium text-gray-900">
                        {formatMoney(record.amount)}
                      </td>
                      <td className="px-6 py-3 text-gray-600">
                        {record.reference || "--"}
                      </td>
                      <td className="px-6 py-3 text-gray-500">
                        {formatDate(record.createdAt)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
