import { useQuery } from "@tanstack/react-query";
import api from "@/lib/api";

export default function AdminSettings() {
  const { data: config, isLoading } = useQuery<Record<string, unknown>>({
    queryKey: ["admin-config"],
    queryFn: async () => {
      const { data } = await api.get("/admin/config");
      return data;
    },
  });

  return (
    <div className="rounded-xl border border-gray-200 bg-white p-6">
      <h2 className="mb-4 text-lg font-semibold text-gray-900">System Settings</h2>
      {isLoading ? (
        <p className="text-sm text-gray-500">Loading...</p>
      ) : config ? (
        <div className="space-y-3">
          {Object.entries(config).map(([key, value]) => (
            <div key={key} className="flex items-center justify-between rounded-lg border border-gray-100 px-4 py-3">
              <span className="text-sm font-medium text-gray-700">{key.replace(/_/g, " ")}</span>
              <span className="text-sm text-gray-900">{String(value)}</span>
            </div>
          ))}
        </div>
      ) : null}
    </div>
  );
}
