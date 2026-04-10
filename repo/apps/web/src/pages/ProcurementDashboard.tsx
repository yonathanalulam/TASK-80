import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import api from "@/lib/api";
import { useAuthStore } from "@/lib/auth";

interface RFQ {
  id: string;
  title: string;
  status: string;
  deadline: string;
  createdAt: string;
  itemCount: number;
}

interface PurchaseOrder {
  id: string;
  rfqId: string;
  supplierId: string;
  supplierName: string;
  status: string;
  totalAmount: number;
  deliveryDate: string;
  createdAt: string;
}

interface Delivery {
  id: string;
  poId: string;
  status: string;
  expectedDate: string;
  actualDate?: string;
  notes: string;
}

interface ProcException {
  id: string;
  reference: string;
  type: string;
  status: string;
  description: string;
  createdAt: string;
}

interface SupplierQuote {
  id: string;
  rfqId: string;
  rfqTitle: string;
  rfqStatus: string;
  deadline: string;
  totalQuoted: number;
  submittedAt?: string;
  status: string;
}

type TabKey = "rfqs" | "pos" | "deliveries" | "exceptions";

const STATUS_COLORS: Record<string, string> = {
  open: "bg-blue-100 text-blue-700",
  pending: "bg-yellow-100 text-yellow-800",
  awarded: "bg-green-100 text-green-700",
  closed: "bg-gray-100 text-gray-700",
  cancelled: "bg-red-100 text-red-700",
  draft: "bg-gray-100 text-gray-700",
  submitted: "bg-blue-100 text-blue-700",
  approved: "bg-green-100 text-green-700",
  rejected: "bg-red-100 text-red-700",
  in_transit: "bg-amber-100 text-amber-700",
  delivered: "bg-green-100 text-green-700",
  resolved: "bg-green-100 text-green-700",
  unresolved: "bg-red-100 text-red-700",
};

function formatMoney(cents: number): string {
  return `$${(cents / 100).toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

function statusLabel(status: string): string {
  return status.replace(/_/g, " ").replace(/\b\w/g, (c) => c.toUpperCase());
}

export default function ProcurementDashboard() {
  const [activeTab, setActiveTab] = useState<TabKey>("rfqs");
  const [showCreateRFQ, setShowCreateRFQ] = useState(false);
  const [rfqTitle, setRfqTitle] = useState("");
  const [rfqDeadline, setRfqDeadline] = useState("");
  const [expandedRfq, setExpandedRfq] = useState<string | null>(null);

  const { hasRole } = useAuthStore();
  const isSupplier = hasRole("supplier");
  const canCreate = hasRole("group_organizer") || hasRole("administrator") || hasRole("accountant");
  const qc = useQueryClient();

  const tabs: { key: TabKey; label: string }[] = [
    { key: "rfqs", label: "RFQs" },
    { key: "pos", label: "Purchase Orders" },
    { key: "deliveries", label: "Deliveries" },
    { key: "exceptions", label: "Exceptions" },
  ];

  const { data: rfqs, isLoading: rfqsLoading } = useQuery<RFQ[]>({
    queryKey: ["procurement-rfqs"],
    queryFn: async () => {
      const { data } = await api.get("/rfqs");
      return data;
    },
    enabled: activeTab === "rfqs",
  });

  const { data: pos, isLoading: posLoading } = useQuery<PurchaseOrder[]>({
    queryKey: ["procurement-pos"],
    queryFn: async () => {
      const { data } = await api.get("/purchase-orders");
      return data;
    },
    enabled: activeTab === "pos",
  });

  const { data: deliveries, isLoading: deliveriesLoading } = useQuery<Delivery[]>(
    {
      queryKey: ["procurement-deliveries"],
      queryFn: async () => {
        const { data } = await api.get("/deliveries");
        return data;
      },
      enabled: activeTab === "deliveries",
    },
  );

  const { data: exceptions, isLoading: exceptionsLoading } = useQuery<
    ProcException[]
  >({
    queryKey: ["procurement-exceptions"],
    queryFn: async () => {
      const { data } = await api.get("/exceptions");
      return data;
    },
    enabled: activeTab === "exceptions",
  });

  const { data: myQuotes } = useQuery<SupplierQuote[]>({
    queryKey: ["supplier-quotes"],
    queryFn: async () => {
      const { data } = await api.get("/supplier-quotes");
      return data;
    },
    enabled: isSupplier,
  });

  const createRfqMutation = useMutation({
    mutationFn: async (payload: { title: string; deadline: string }) => {
      const { data } = await api.post("/rfqs", payload);
      return data;
    },
    onSuccess: () => {
      setShowCreateRFQ(false);
      setRfqTitle("");
      setRfqDeadline("");
      qc.invalidateQueries({ queryKey: ["procurement-rfqs"] });
    },
  });

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Procurement</h1>
          <p className="mt-1 text-gray-500">
            Track and manage procurement items.
          </p>
        </div>
        <div className="flex gap-2">
          {canCreate && (
            <button
              onClick={() => setShowCreateRFQ(true)}
              className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-indigo-700"
            >
              Create RFQ
            </button>
          )}
        </div>
      </div>

      {isSupplier && myQuotes && myQuotes.length > 0 && (
        <div className="rounded-xl border border-amber-200 bg-amber-50 p-6">
          <h2 className="mb-3 text-lg font-semibold text-amber-900">My Quotes</h2>
          <div className="space-y-2">
            {myQuotes.map((quote) => (
              <div
                key={quote.id}
                className="flex items-center justify-between rounded-lg border border-amber-200 bg-white p-3"
              >
                <div>
                  <p className="text-sm font-medium text-gray-900">
                    {quote.rfqTitle}
                  </p>
                  <p className="text-xs text-gray-500">
                    Deadline: {formatDate(quote.deadline)}
                  </p>
                </div>
                <div className="flex items-center gap-3">
                  {quote.totalQuoted > 0 && (
                    <span className="text-sm font-medium text-gray-700">
                      {formatMoney(quote.totalQuoted)}
                    </span>
                  )}
                  <span
                    className={`rounded-full px-2.5 py-0.5 text-xs font-medium ${STATUS_COLORS[quote.status] ?? "bg-gray-100 text-gray-700"}`}
                  >
                    {statusLabel(quote.status)}
                  </span>
                  {quote.status !== "submitted" && (
                    <button className="rounded-lg bg-indigo-600 px-3 py-1 text-xs font-medium text-white hover:bg-indigo-700">
                      Submit Quote
                    </button>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      <div className="border-b border-gray-200">
        <nav className="-mb-px flex gap-6">
          {tabs.map((t) => (
            <button
              key={t.key}
              onClick={() => setActiveTab(t.key)}
              className={`border-b-2 pb-3 text-sm font-medium transition-colors ${
                activeTab === t.key
                  ? "border-indigo-600 text-indigo-600"
                  : "border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700"
              }`}
            >
              {t.label}
            </button>
          ))}
        </nav>
      </div>

      {activeTab === "rfqs" && (
        <div className="rounded-xl border border-gray-200 bg-white">
          {rfqsLoading ? (
            <div className="p-6 text-center text-gray-500">Loading RFQs...</div>
          ) : !rfqs?.length ? (
            <div className="p-6 text-center text-gray-500">No RFQs found.</div>
          ) : (
            <div className="divide-y divide-gray-100">
              {rfqs.map((rfq) => (
                <div key={rfq.id}>
                  <div
                    onClick={() =>
                      setExpandedRfq(expandedRfq === rfq.id ? null : rfq.id)
                    }
                    className="flex cursor-pointer items-center justify-between px-6 py-4 transition-colors hover:bg-gray-50"
                  >
                    <div>
                      <p className="text-sm font-medium text-gray-900">
                        {rfq.title}
                      </p>
                      <p className="text-xs text-gray-500">
                        {rfq.itemCount} items | Deadline:{" "}
                        {formatDate(rfq.deadline)}
                      </p>
                    </div>
                    <div className="flex items-center gap-3">
                      <span
                        className={`rounded-full px-2.5 py-0.5 text-xs font-medium ${STATUS_COLORS[rfq.status] ?? "bg-gray-100 text-gray-700"}`}
                      >
                        {statusLabel(rfq.status)}
                      </span>
                      <span className="text-gray-400">
                        {expandedRfq === rfq.id ? "\u25B2" : "\u25BC"}
                      </span>
                    </div>
                  </div>
                  {expandedRfq === rfq.id && (
                    <div className="border-t border-gray-100 bg-gray-50 px-6 py-4">
                      <div className="grid grid-cols-2 gap-4 text-sm">
                        <div>
                          <span className="font-medium text-gray-500">
                            Created
                          </span>
                          <p className="text-gray-900">
                            {formatDate(rfq.createdAt)}
                          </p>
                        </div>
                        <div>
                          <span className="font-medium text-gray-500">
                            Status
                          </span>
                          <p className="capitalize text-gray-900">
                            {statusLabel(rfq.status)}
                          </p>
                        </div>
                      </div>
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {activeTab === "pos" && (
        <div className="rounded-xl border border-gray-200 bg-white">
          {posLoading ? (
            <div className="p-6 text-center text-gray-500">
              Loading purchase orders...
            </div>
          ) : !pos?.length ? (
            <div className="p-6 text-center text-gray-500">
              No purchase orders found.
            </div>
          ) : (
            <table className="w-full text-left text-sm">
              <thead className="border-b border-gray-200 bg-gray-50">
                <tr>
                  <th className="px-6 py-3 font-medium text-gray-600">
                    Supplier
                  </th>
                  <th className="px-6 py-3 font-medium text-gray-600">Status</th>
                  <th className="px-6 py-3 text-right font-medium text-gray-600">
                    Amount
                  </th>
                  <th className="px-6 py-3 font-medium text-gray-600">
                    Delivery Date
                  </th>
                  <th className="px-6 py-3 font-medium text-gray-600">Created</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {pos.map((po) => (
                  <tr key={po.id} className="hover:bg-gray-50">
                    <td className="px-6 py-3 font-medium text-gray-900">
                      {po.supplierName}
                    </td>
                    <td className="px-6 py-3">
                      <span
                        className={`rounded-full px-2.5 py-0.5 text-xs font-medium ${STATUS_COLORS[po.status] ?? "bg-gray-100 text-gray-700"}`}
                      >
                        {statusLabel(po.status)}
                      </span>
                    </td>
                    <td className="px-6 py-3 text-right text-gray-700">
                      {formatMoney(po.totalAmount)}
                    </td>
                    <td className="px-6 py-3 text-gray-600">
                      {formatDate(po.deliveryDate)}
                    </td>
                    <td className="px-6 py-3 text-gray-500">
                      {formatDate(po.createdAt)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}

      {activeTab === "deliveries" && (
        <div className="rounded-xl border border-gray-200 bg-white">
          {deliveriesLoading ? (
            <div className="p-6 text-center text-gray-500">
              Loading deliveries...
            </div>
          ) : !deliveries?.length ? (
            <div className="p-6 text-center text-gray-500">
              No deliveries found.
            </div>
          ) : (
            <table className="w-full text-left text-sm">
              <thead className="border-b border-gray-200 bg-gray-50">
                <tr>
                  <th className="px-6 py-3 font-medium text-gray-600">PO ID</th>
                  <th className="px-6 py-3 font-medium text-gray-600">Status</th>
                  <th className="px-6 py-3 font-medium text-gray-600">
                    Expected
                  </th>
                  <th className="px-6 py-3 font-medium text-gray-600">Actual</th>
                  <th className="px-6 py-3 font-medium text-gray-600">Notes</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {deliveries.map((del) => (
                  <tr key={del.id} className="hover:bg-gray-50">
                    <td className="px-6 py-3 font-mono text-xs text-gray-700">
                      {del.poId}
                    </td>
                    <td className="px-6 py-3">
                      <span
                        className={`rounded-full px-2.5 py-0.5 text-xs font-medium ${STATUS_COLORS[del.status] ?? "bg-gray-100 text-gray-700"}`}
                      >
                        {statusLabel(del.status)}
                      </span>
                    </td>
                    <td className="px-6 py-3 text-gray-600">
                      {formatDate(del.expectedDate)}
                    </td>
                    <td className="px-6 py-3 text-gray-600">
                      {del.actualDate ? formatDate(del.actualDate) : "--"}
                    </td>
                    <td className="max-w-xs truncate px-6 py-3 text-gray-500">
                      {del.notes || "--"}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}

      {activeTab === "exceptions" && (
        <div className="rounded-xl border border-gray-200 bg-white">
          {exceptionsLoading ? (
            <div className="p-6 text-center text-gray-500">
              Loading exceptions...
            </div>
          ) : !exceptions?.length ? (
            <div className="p-6 text-center text-gray-500">
              No exceptions found.
            </div>
          ) : (
            <div className="divide-y divide-gray-100">
              {exceptions.map((exc) => (
                <div
                  key={exc.id}
                  className="flex items-start justify-between px-6 py-4"
                >
                  <div>
                    <div className="flex items-center gap-2">
                      <span
                        className={`rounded-full px-2.5 py-0.5 text-xs font-medium ${STATUS_COLORS[exc.status] ?? "bg-gray-100 text-gray-700"}`}
                      >
                        {statusLabel(exc.status)}
                      </span>
                      <span className="text-xs text-gray-500">{exc.type}</span>
                    </div>
                    <p className="mt-1 text-sm text-gray-900">{exc.description}</p>
                    <p className="mt-1 font-mono text-xs text-gray-400">
                      Ref: {exc.reference}
                    </p>
                  </div>
                  <span className="text-xs text-gray-400">
                    {formatDate(exc.createdAt)}
                  </span>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {showCreateRFQ && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="mx-4 w-full max-w-md rounded-xl bg-white p-6 shadow-xl">
            <h3 className="mb-4 text-lg font-semibold text-gray-900">
              Create RFQ
            </h3>
            <div className="space-y-4">
              <div>
                <label className="mb-1 block text-sm font-medium text-gray-700">
                  Title
                </label>
                <input
                  type="text"
                  value={rfqTitle}
                  onChange={(e) => setRfqTitle(e.target.value)}
                  placeholder="RFQ title"
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none"
                />
              </div>
              <div>
                <label className="mb-1 block text-sm font-medium text-gray-700">
                  Deadline
                </label>
                <input
                  type="date"
                  value={rfqDeadline}
                  onChange={(e) => setRfqDeadline(e.target.value)}
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none"
                />
              </div>
              {createRfqMutation.isError && (
                <p className="text-sm text-red-600">
                  Failed to create RFQ. Please try again.
                </p>
              )}
            </div>
            <div className="mt-6 flex justify-end gap-3">
              <button
                onClick={() => setShowCreateRFQ(false)}
                className="rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100"
              >
                Cancel
              </button>
              <button
                onClick={() =>
                  createRfqMutation.mutate({
                    title: rfqTitle,
                    deadline: rfqDeadline,
                  })
                }
                disabled={
                  !rfqTitle.trim() ||
                  !rfqDeadline ||
                  createRfqMutation.isPending
                }
                className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-indigo-700 disabled:opacity-50"
              >
                {createRfqMutation.isPending ? "Creating..." : "Create"}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
