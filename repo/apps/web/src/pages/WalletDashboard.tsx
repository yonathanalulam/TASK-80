import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import api from "@/lib/api";
import { useAuthStore } from "@/lib/auth";

interface Wallet {
  id: string;
  ownerId: string;
  balance: number;
  currency: string;
  status: string;
}

interface Transaction {
  id: string;
  amount: number;
  direction: string;
  referenceType: string;
  referenceId: string;
  description: string;
  createdAt: string;
}

interface Escrow {
  id: string;
  orderType: string;
  orderId: string;
  amountHeld: number;
  amountReleased: number;
  status: string;
  createdAt: string;
}

interface PaginatedTransactions {
  items: Transaction[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

function formatMoney(cents: number, currency = "USD"): string {
  const symbol = currency === "USD" ? "$" : currency;
  return `${symbol}${(cents / 100).toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

const TX_COLORS: Record<string, string> = {
  credit: "text-green-600",
  debit: "text-red-600",
};

export default function WalletDashboard() {
  const { user, hasRole } = useAuthStore();
  const ownerId = user?.id ?? "";
  const isCourier = hasRole("courier_runner");
  const isAccountant = hasRole("accountant");
  const isAdmin = hasRole("administrator");
  const canRefund = isAccountant || isAdmin;

  const [txPage, setTxPage] = useState(1);
  const [showWithdrawModal, setShowWithdrawModal] = useState(false);
  const [withdrawAmount, setWithdrawAmount] = useState("");
  const [showRefundModal, setShowRefundModal] = useState(false);
  const [refundBookingId, setRefundBookingId] = useState("");
  const [refundAmount, setRefundAmount] = useState("");
  const [refundReason, setRefundReason] = useState("");
  const qc = useQueryClient();

  const {
    data: wallet,
    isLoading: walletLoading,
    isError: walletError,
  } = useQuery<Wallet>({
    queryKey: ["wallet", ownerId],
    queryFn: async () => {
      const { data } = await api.get(`/wallets/${ownerId}`);
      return data;
    },
    enabled: !!ownerId,
  });

  const { data: txData, isLoading: txLoading } = useQuery<PaginatedTransactions>({
    queryKey: ["wallet-transactions", ownerId, txPage],
    queryFn: async () => {
      const { data } = await api.get(`/wallets/${ownerId}/transactions`, {
        params: { page: txPage, pageSize: 20 },
      });
      return data;
    },
    enabled: !!ownerId,
  });

  const { data: escrowData } = useQuery<Escrow[]>({
    queryKey: ["wallet-escrows", ownerId],
    queryFn: async () => {
      const { data } = await api.get(`/escrows/${ownerId}`);
      return data;
    },
    enabled: !!ownerId,
  });
  const escrows = escrowData ?? [];

  const withdrawMutation = useMutation({
    mutationFn: async (amount: number) => {
      const { data } = await api.post("/withdrawals", { amount, ownerId });
      return data;
    },
    onSuccess: () => {
      setShowWithdrawModal(false);
      setWithdrawAmount("");
      qc.invalidateQueries({ queryKey: ["wallet"] });
      qc.invalidateQueries({ queryKey: ["wallet-transactions"] });
    },
  });

  const refundMutation = useMutation({
    mutationFn: async (payload: {
      orderType: string;
      orderId: string;
      amount: number;
      reason: string;
    }) => {
      const { data } = await api.post("/refunds", payload);
      return data;
    },
    onSuccess: () => {
      setShowRefundModal(false);
      setRefundBookingId("");
      setRefundAmount("");
      setRefundReason("");
      qc.invalidateQueries({ queryKey: ["wallet"] });
      qc.invalidateQueries({ queryKey: ["wallet-transactions"] });
    },
  });

  const transactions = txData?.items ?? [];

  const activeEscrows = (escrows ?? []).filter((e) => e.status === "held" || e.status === "partially_released");

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Wallet</h1>
        <p className="mt-1 text-gray-500">
          View your balance and transaction history.
        </p>
      </div>

      {walletLoading ? (
        <div className="rounded-xl border border-gray-200 bg-white p-6">
          <p className="text-gray-500">Loading wallet...</p>
        </div>
      ) : walletError ? (
        <div className="rounded-xl border border-red-200 bg-red-50 p-6 text-red-600">
          Failed to load wallet information.
        </div>
      ) : wallet ? (
        <div className="rounded-xl border border-gray-200 bg-gradient-to-r from-indigo-600 to-indigo-700 p-6 text-white">
          <p className="text-sm font-medium text-indigo-200">Available Balance</p>
          <p className="mt-2 text-3xl font-bold">
            {formatMoney(wallet.balance, wallet.currency)}
          </p>
          <div className="mt-4 flex gap-3">
            {isCourier && (
              <button
                onClick={() => setShowWithdrawModal(true)}
                className="rounded-lg bg-white/20 px-4 py-2 text-sm font-medium text-white backdrop-blur-sm transition-colors hover:bg-white/30"
              >
                Request Withdrawal
              </button>
            )}
            {canRefund && (
              <button
                onClick={() => setShowRefundModal(true)}
                className="rounded-lg bg-white/20 px-4 py-2 text-sm font-medium text-white backdrop-blur-sm transition-colors hover:bg-white/30"
              >
                Issue Refund
              </button>
            )}
          </div>
        </div>
      ) : null}

      {activeEscrows.length > 0 && (
        <div className="rounded-xl border border-gray-200 bg-white">
          <div className="border-b border-gray-200 px-6 py-4">
            <h2 className="text-lg font-semibold text-gray-900">
              Active Escrows
            </h2>
          </div>
          <div className="divide-y divide-gray-100">
            {activeEscrows.map((escrow) => (
              <div
                key={escrow.id}
                className="flex items-center justify-between px-6 py-3"
              >
                <div>
                  <p className="text-sm font-medium text-gray-900">
                    {escrow.orderType}: {escrow.orderId.slice(0, 8)}...
                  </p>
                  <p className="text-xs text-gray-500">
                    {formatDate(escrow.createdAt)}
                  </p>
                </div>
                <div className="text-right">
                  <p className="text-sm font-semibold text-blue-600">
                    {formatMoney(escrow.amountHeld)}
                  </p>
                  <span className="text-xs capitalize text-gray-500">
                    {escrow.status}
                  </span>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      <div className="rounded-xl border border-gray-200 bg-white">
        <div className="border-b border-gray-200 px-6 py-4">
          <h2 className="text-lg font-semibold text-gray-900">
            Transaction History
          </h2>
        </div>
        {txLoading ? (
          <div className="p-6 text-center text-gray-500">
            Loading transactions...
          </div>
        ) : !transactions.length ? (
          <div className="p-6 text-center text-gray-500">
            No transactions yet.
          </div>
        ) : (
          <>
            <table className="w-full text-left text-sm">
              <thead className="border-b border-gray-200 bg-gray-50">
                <tr>
                  <th className="px-6 py-3 font-medium text-gray-600">Direction</th>
                  <th className="px-6 py-3 font-medium text-gray-600">
                    Description
                  </th>
                  <th className="px-6 py-3 font-medium text-gray-600">
                    Reference
                  </th>
                  <th className="px-6 py-3 text-right font-medium text-gray-600">
                    Amount
                  </th>
                  <th className="px-6 py-3 font-medium text-gray-600">Date</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {transactions.map((tx) => (
                  <tr key={tx.id}>
                    <td className="px-6 py-3">
                      <span
                        className={`text-xs font-medium capitalize ${TX_COLORS[tx.direction] ?? "text-gray-600"}`}
                      >
                        {tx.direction}
                      </span>
                    </td>
                    <td className="px-6 py-3 text-gray-700">{tx.description}</td>
                    <td className="px-6 py-3 font-mono text-xs text-gray-500">
                      {tx.referenceType && tx.referenceId
                        ? `${tx.referenceType}:${tx.referenceId.slice(0, 8)}`
                        : "--"}
                    </td>
                    <td
                      className={`px-6 py-3 text-right font-medium ${
                        tx.direction === "credit"
                          ? "text-green-600"
                          : "text-red-600"
                      }`}
                    >
                      {tx.direction === "credit" ? "+" : "-"}
                      {formatMoney(Math.abs(tx.amount))}
                    </td>
                    <td className="px-6 py-3 text-gray-500">
                      {formatDate(tx.createdAt)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>

            {txData && txData.totalPages > 1 && (
              <div className="flex items-center justify-between border-t border-gray-200 px-6 py-3">
                <p className="text-sm text-gray-500">
                  Page {txData.page} of {txData.totalPages}
                </p>
                <div className="flex gap-2">
                  <button
                    onClick={() => setTxPage((p) => Math.max(1, p - 1))}
                    disabled={txPage <= 1}
                    className="rounded-lg border border-gray-300 px-3 py-1 text-sm text-gray-700 hover:bg-gray-100 disabled:opacity-50"
                  >
                    Previous
                  </button>
                  <button
                    onClick={() =>
                      setTxPage((p) => Math.min(txData.totalPages, p + 1))
                    }
                    disabled={txPage >= txData.totalPages}
                    className="rounded-lg border border-gray-300 px-3 py-1 text-sm text-gray-700 hover:bg-gray-100 disabled:opacity-50"
                  >
                    Next
                  </button>
                </div>
              </div>
            )}
          </>
        )}
      </div>

      {canRefund && (
        <div className="rounded-xl border border-gray-200 bg-white p-6">
          <h2 className="mb-2 text-lg font-semibold text-gray-900">
            Reconciliation
          </h2>
          <p className="text-sm text-gray-500">
            Generate and view reconciliation reports for accounting purposes.
          </p>
          <button className="mt-3 rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-100">
            View Reconciliation Report
          </button>
        </div>
      )}

      {showWithdrawModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="mx-4 w-full max-w-md rounded-xl bg-white p-6 shadow-xl">
            <h3 className="mb-4 text-lg font-semibold text-gray-900">
              Request Withdrawal
            </h3>
            <div className="space-y-4">
              <div>
                <label className="mb-1 block text-sm font-medium text-gray-700">
                  Amount ($)
                </label>
                <input
                  type="number"
                  step="0.01"
                  min="0"
                  value={withdrawAmount}
                  onChange={(e) => setWithdrawAmount(e.target.value)}
                  placeholder="0.00"
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none"
                />
              </div>
              <div className="rounded-lg bg-amber-50 p-3 text-sm text-amber-700">
                Note: Daily withdrawal caps may apply. Contact admin if you need
                an increased limit.
              </div>
              {withdrawMutation.isError && (
                <p className="text-sm text-red-600">
                  Withdrawal request failed. Please try again.
                </p>
              )}
            </div>
            <div className="mt-6 flex justify-end gap-3">
              <button
                onClick={() => setShowWithdrawModal(false)}
                className="rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100"
              >
                Cancel
              </button>
              <button
                onClick={() =>
                  withdrawMutation.mutate(
                    Math.round(parseFloat(withdrawAmount) * 100),
                  )
                }
                disabled={
                  !withdrawAmount ||
                  parseFloat(withdrawAmount) <= 0 ||
                  withdrawMutation.isPending
                }
                className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-indigo-700 disabled:opacity-50"
              >
                {withdrawMutation.isPending ? "Processing..." : "Submit Request"}
              </button>
            </div>
          </div>
        </div>
      )}

      {showRefundModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="mx-4 w-full max-w-md rounded-xl bg-white p-6 shadow-xl">
            <h3 className="mb-4 text-lg font-semibold text-gray-900">
              Issue Refund
            </h3>
            <div className="space-y-4">
              <div>
                <label className="mb-1 block text-sm font-medium text-gray-700">
                  Booking ID
                </label>
                <input
                  type="text"
                  value={refundBookingId}
                  onChange={(e) => setRefundBookingId(e.target.value)}
                  placeholder="Enter booking ID"
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none"
                />
              </div>
              <div>
                <label className="mb-1 block text-sm font-medium text-gray-700">
                  Amount ($)
                </label>
                <input
                  type="number"
                  step="0.01"
                  min="0"
                  value={refundAmount}
                  onChange={(e) => setRefundAmount(e.target.value)}
                  placeholder="0.00"
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none"
                />
              </div>
              <div>
                <label className="mb-1 block text-sm font-medium text-gray-700">
                  Reason
                </label>
                <textarea
                  value={refundReason}
                  onChange={(e) => setRefundReason(e.target.value)}
                  rows={2}
                  placeholder="Reason for refund"
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none"
                />
              </div>
              {refundMutation.isError && (
                <p className="text-sm text-red-600">
                  Refund failed. Please check the details and try again.
                </p>
              )}
            </div>
            <div className="mt-6 flex justify-end gap-3">
              <button
                onClick={() => setShowRefundModal(false)}
                className="rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100"
              >
                Cancel
              </button>
              <button
                onClick={() =>
                  refundMutation.mutate({
                    orderType: "booking",
                    orderId: refundBookingId,
                    amount: Math.round(parseFloat(refundAmount) * 100),
                    reason: refundReason,
                  })
                }
                disabled={
                  !refundBookingId ||
                  !refundAmount ||
                  parseFloat(refundAmount) <= 0 ||
                  refundMutation.isPending
                }
                className="rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-red-700 disabled:opacity-50"
              >
                {refundMutation.isPending ? "Processing..." : "Issue Refund"}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
