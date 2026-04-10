import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import api from "@/lib/api";
import { useAuthStore } from "@/lib/auth";

interface Notification {
  id: string;
  eventType: string;
  title: string;
  message: string;
  category: string;
  read: boolean;
  createdAt: string;
}

interface PaginatedNotifications {
  items: Notification[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

interface DndSettings {
  enabled: boolean;
  startTime: string;
  endTime: string;
}

type FilterTab = "all" | "unread";

const CATEGORIES = [
  { value: "all", label: "All Categories" },
  { value: "itinerary_changes", label: "Itinerary Changes" },
  { value: "booking_updates", label: "Booking Updates" },
  { value: "procurement", label: "Procurement" },
  { value: "settlement", label: "Settlement" },
  { value: "risk", label: "Risk" },
  { value: "documents", label: "Documents" },
  { value: "marketing", label: "Marketing" },
];

const EVENT_TYPE_STYLES: Record<string, string> = {
  itinerary_changes: "bg-blue-100 text-blue-700",
  booking_updates: "bg-emerald-100 text-emerald-700",
  procurement: "bg-amber-100 text-amber-700",
  settlement: "bg-purple-100 text-purple-700",
  risk: "bg-red-100 text-red-700",
  documents: "bg-gray-100 text-gray-700",
  marketing: "bg-pink-100 text-pink-700",
};

function formatTimestamp(iso: string): string {
  const date = new Date(iso);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);

  if (diffMins < 1) return "Just now";
  if (diffMins < 60) return `${diffMins}m ago`;
  const diffHours = Math.floor(diffMins / 60);
  if (diffHours < 24) return `${diffHours}h ago`;
  const diffDays = Math.floor(diffHours / 24);
  if (diffDays < 7) return `${diffDays}d ago`;

  return date.toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

export default function NotificationCenter() {
  const [tab, setTab] = useState<FilterTab>("all");
  const [categoryFilter, setCategoryFilter] = useState("all");
  const [page, setPage] = useState(1);
  const pageSize = 20;
  const qc = useQueryClient();
  const { user, hasRole } = useAuthStore();
  const isAdmin = hasRole("administrator");

  const unreadOnly = tab === "unread";

  const { data, isLoading, isError } = useQuery<PaginatedNotifications>({
    queryKey: ["notifications", page, unreadOnly, categoryFilter],
    queryFn: async () => {
      const params: Record<string, unknown> = {
        page,
        pageSize,
        unreadOnly,
      };
      if (categoryFilter !== "all") params.category = categoryFilter;
      const { data } = await api.get("/notifications", { params });
      return data;
    },
  });

  const markAsReadMutation = useMutation({
    mutationFn: async (notifId: string) => {
      await api.post(`/notifications/${notifId}/read`);
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["notifications"] });
    },
  });

  const exportCallbackMutation = useMutation({
    mutationFn: async () => {
      const { data } = await api.post("/messages/callback-queue/export");
      return data;
    },
  });

  const { data: dndSettings } = useQuery<DndSettings>({
    queryKey: ["dnd-settings", user?.id],
    queryFn: async () => {
      const { data } = await api.get(`/users/${user!.id}/dnd`);
      return data;
    },
    enabled: !!user?.id,
  });

  const toggleDndMutation = useMutation({
    mutationFn: async (enabled: boolean) => {
      await api.patch(`/users/${user!.id}/dnd`, { enabled });
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["dnd-settings"] });
    },
  });

  const notifications = data?.items ?? [];


  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Notifications</h1>
          <p className="mt-1 text-gray-500">
            Stay up to date with your activity.
          </p>
        </div>
        {isAdmin && (
          <button
            onClick={() => exportCallbackMutation.mutate()}
            disabled={exportCallbackMutation.isPending}
            className="rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-100 disabled:opacity-50"
          >
            {exportCallbackMutation.isPending
              ? "Exporting..."
              : "Export Callback Queue"}
          </button>
        )}
      </div>

      {exportCallbackMutation.isSuccess && (
        <div className="rounded-lg bg-green-50 p-3 text-sm text-green-700">
          Callback queue exported successfully.
        </div>
      )}

      <div className="flex flex-wrap items-center gap-3">
        <div className="flex rounded-lg border border-gray-200 bg-white">
          {(["all", "unread"] as FilterTab[]).map((t) => (
            <button
              key={t}
              onClick={() => {
                setTab(t);
                setPage(1);
              }}
              className={`px-4 py-2 text-sm font-medium capitalize transition-colors ${
                tab === t
                  ? "bg-indigo-50 text-indigo-700"
                  : "text-gray-500 hover:text-gray-700"
              }`}
            >
              {t}
            </button>
          ))}
        </div>

        <select
          value={categoryFilter}
          onChange={(e) => {
            setCategoryFilter(e.target.value);
            setPage(1);
          }}
          className="rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none"
        >
          {CATEGORIES.map((cat) => (
            <option key={cat.value} value={cat.value}>
              {cat.label}
            </option>
          ))}
        </select>
      </div>

      {isLoading ? (
        <div className="py-12 text-center text-gray-500">
          Loading notifications...
        </div>
      ) : isError ? (
        <div className="rounded-xl border border-red-200 bg-red-50 py-12 text-center text-red-600">
          Failed to load notifications.
        </div>
      ) : !notifications.length ? (
        <div className="rounded-xl border border-gray-200 bg-white py-12 text-center">
          <p className="text-gray-500">No notifications.</p>
        </div>
      ) : (
        <div className="space-y-2">
          {notifications.map((notification) => (
            <div
              key={notification.id}
              className={`flex items-start justify-between rounded-xl border bg-white p-4 transition-colors ${
                notification.read
                  ? "border-gray-200"
                  : "border-indigo-200 bg-indigo-50/30"
              }`}
            >
              <div className="flex-1">
                <div className="flex flex-wrap items-center gap-2">
                  {!notification.read && (
                    <span className="h-2 w-2 flex-shrink-0 rounded-full bg-indigo-600" />
                  )}
                  <span
                    className={`rounded-full px-2 py-0.5 text-xs font-medium ${
                      EVENT_TYPE_STYLES[notification.category] ??
                      "bg-gray-100 text-gray-700"
                    }`}
                  >
                    {notification.eventType ?? notification.category}
                  </span>
                  <h3 className="text-sm font-semibold text-gray-900">
                    {notification.title}
                  </h3>
                </div>
                <p className="mt-1 text-sm text-gray-600">
                  {notification.message}
                </p>
                <p className="mt-1 text-xs text-gray-400">
                  {formatTimestamp(notification.createdAt)}
                </p>
              </div>
              {!notification.read && (
                <button
                  onClick={() => markAsReadMutation.mutate(notification.id)}
                  disabled={markAsReadMutation.isPending}
                  className="ml-4 flex-shrink-0 rounded-lg border border-gray-300 px-2.5 py-1 text-xs text-gray-600 transition-colors hover:bg-gray-100"
                >
                  Mark read
                </button>
              )}
            </div>
          ))}
        </div>
      )}

      {data && data.totalPages > 1 && (
        <div className="flex items-center justify-between">
          <p className="text-sm text-gray-500">
            Page {data.page} of {data.totalPages} ({data.total} total)
          </p>
          <div className="flex gap-2">
            <button
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page <= 1}
              className="rounded-lg border border-gray-300 px-3 py-1.5 text-sm font-medium text-gray-700 hover:bg-gray-100 disabled:opacity-50"
            >
              Previous
            </button>
            <button
              onClick={() => setPage((p) => Math.min(data.totalPages, p + 1))}
              disabled={page >= data.totalPages}
              className="rounded-lg border border-gray-300 px-3 py-1.5 text-sm font-medium text-gray-700 hover:bg-gray-100 disabled:opacity-50"
            >
              Next
            </button>
          </div>
        </div>
      )}

      <div className="rounded-xl border border-gray-200 bg-white p-6">
        <h2 className="mb-4 text-lg font-semibold text-gray-900">
          Do Not Disturb
        </h2>
        {dndSettings ? (
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-600">
                DND Window: {dndSettings.startTime} - {dndSettings.endTime}
              </p>
              <p className="mt-1 text-sm text-gray-500">
                Status:{" "}
                <span
                  className={`font-medium ${dndSettings.enabled ? "text-green-600" : "text-gray-500"}`}
                >
                  {dndSettings.enabled ? "Enabled" : "Disabled"}
                </span>
              </p>
            </div>
            <button
              onClick={() => toggleDndMutation.mutate(!dndSettings.enabled)}
              disabled={toggleDndMutation.isPending}
              className={`rounded-lg px-4 py-2 text-sm font-medium transition-colors ${
                dndSettings.enabled
                  ? "border border-gray-300 text-gray-700 hover:bg-gray-100"
                  : "bg-indigo-600 text-white hover:bg-indigo-700"
              }`}
            >
              {dndSettings.enabled ? "Disable DND" : "Enable DND"}
            </button>
          </div>
        ) : (
          <p className="text-sm text-gray-500">
            Loading DND settings...
          </p>
        )}
      </div>
    </div>
  );
}
