import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import api from "@/lib/api";

interface Booking {
  id: string;
  title: string;
  status: string;
  totalAmount: number;
  discountAmount: number;
  createdAt: string;
}

interface PaginatedResponse {
  items: Booking[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
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

function formatMoney(cents: number): string {
  return `$${(cents / 100).toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

function statusLabel(status: string): string {
  return status.replace(/_/g, " ").replace(/\b\w/g, (c) => c.toUpperCase());
}

export default function BookingList() {
  const [page, setPage] = useState(1);
  const pageSize = 15;
  const navigate = useNavigate();

  const { data: pageData, isLoading, isError } = useQuery<PaginatedResponse>({
    queryKey: ["bookings", page],
    queryFn: async () => {
      const { data } = await api.get("/bookings", {
        params: { page, pageSize },
      });
      return data;
    },
  });

  const bookings = pageData?.items ?? [];
  const totalPages = pageData?.totalPages ?? 0;
  const total = pageData?.total ?? 0;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Bookings</h1>
          <p className="mt-1 text-gray-500">
            Manage your bookings and reservations.
          </p>
        </div>
        <Link
          to="/bookings/new"
          className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-indigo-700"
        >
          Create Booking
        </Link>
      </div>

      {isLoading ? (
        <div className="py-12 text-center text-gray-500">Loading bookings...</div>
      ) : isError ? (
        <div className="rounded-xl border border-red-200 bg-red-50 py-12 text-center text-red-600">
          Failed to load bookings. Please try again.
        </div>
      ) : !bookings.length ? (
        <div className="rounded-xl border border-gray-200 bg-white py-12 text-center">
          <p className="text-gray-500">No bookings yet.</p>
          <Link
            to="/bookings/new"
            className="mt-2 inline-block text-sm font-medium text-indigo-600 hover:text-indigo-700"
          >
            Create your first booking
          </Link>
        </div>
      ) : (
        <>
          <div className="overflow-hidden rounded-xl border border-gray-200 bg-white">
            <table className="w-full text-left text-sm">
              <thead className="border-b border-gray-200 bg-gray-50">
                <tr>
                  <th className="px-4 py-3 font-medium text-gray-600">Title</th>
                  <th className="px-4 py-3 font-medium text-gray-600">Status</th>
                  <th className="px-4 py-3 text-right font-medium text-gray-600">
                    Total
                  </th>
                  <th className="px-4 py-3 text-right font-medium text-gray-600">
                    Discount
                  </th>
                  <th className="px-4 py-3 font-medium text-gray-600">Created</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {bookings.map((booking) => (
                  <tr
                    key={booking.id}
                    onClick={() => navigate(`/bookings/${booking.id}`)}
                    className="cursor-pointer transition-colors hover:bg-gray-50"
                  >
                    <td className="px-4 py-3 font-medium text-gray-900">
                      {booking.title}
                    </td>
                    <td className="px-4 py-3">
                      <span
                        className={`inline-block rounded-full px-2.5 py-0.5 text-xs font-medium ${STATUS_COLORS[booking.status] ?? "bg-gray-100 text-gray-700"}`}
                      >
                        {statusLabel(booking.status)}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-right text-gray-700">
                      {formatMoney(booking.totalAmount)}
                    </td>
                    <td className="px-4 py-3 text-right text-gray-700">
                      {booking.discountAmount ? formatMoney(booking.discountAmount) : "--"}
                    </td>
                    <td className="px-4 py-3 text-gray-500">
                      {formatDate(booking.createdAt)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {totalPages > 1 && (
            <div className="flex items-center justify-between">
              <p className="text-sm text-gray-500">
                Page {page} of {totalPages} ({total} total)
              </p>
              <div className="flex gap-2">
                <button
                  onClick={() => setPage((p) => Math.max(1, p - 1))}
                  disabled={page <= 1}
                  className="rounded-lg border border-gray-300 px-3 py-1.5 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-100 disabled:opacity-50"
                >
                  Previous
                </button>
                <button
                  onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                  disabled={page >= totalPages}
                  className="rounded-lg border border-gray-300 px-3 py-1.5 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-100 disabled:opacity-50"
                >
                  Next
                </button>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  );
}
