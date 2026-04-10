import { useQuery } from "@tanstack/react-query";
import api from "@/lib/api";

interface SystemConfig {
  dnd_default_start: string;
  dnd_default_end: string;
  courier_daily_cap: number;
  refund_minimum_unit: number;
  download_token_ttl_min: number;
  max_cancellations_24h: number;
  max_rfqs_10min: number;
  coupon_max_threshold: number;
  coupon_max_percentage: number;
  new_user_gift_exclusive: boolean;
}

export default function AdminOverview() {
  const { data: config, isLoading } = useQuery<SystemConfig>({
    queryKey: ["admin-config"],
    queryFn: async () => {
      const { data } = await api.get("/admin/config");
      return data;
    },
  });

  return (
    <div className="space-y-6">
      <div className="rounded-xl border border-gray-200 bg-white p-6">
        <h2 className="text-lg font-semibold text-gray-900">System Overview</h2>
        {isLoading ? (
          <p className="mt-2 text-sm text-gray-500">Loading...</p>
        ) : config ? (
          <div className="mt-4 grid grid-cols-2 gap-4">
            <div className="rounded-lg border border-gray-100 p-4">
              <p className="text-sm text-gray-500">DND Window</p>
              <p className="text-lg font-semibold text-gray-900">{config.dnd_default_start} - {config.dnd_default_end}</p>
            </div>
            <div className="rounded-lg border border-gray-100 p-4">
              <p className="text-sm text-gray-500">Courier Daily Cap</p>
              <p className="text-lg font-semibold text-gray-900">${config.courier_daily_cap.toLocaleString()}</p>
            </div>
            <div className="rounded-lg border border-gray-100 p-4">
              <p className="text-sm text-gray-500">Refund Minimum</p>
              <p className="text-lg font-semibold text-gray-900">${config.refund_minimum_unit}</p>
            </div>
            <div className="rounded-lg border border-gray-100 p-4">
              <p className="text-sm text-gray-500">Max Cancellations (24h)</p>
              <p className="text-lg font-semibold text-gray-900">{config.max_cancellations_24h}</p>
            </div>
            <div className="rounded-lg border border-gray-100 p-4">
              <p className="text-sm text-gray-500">Max RFQs (10min)</p>
              <p className="text-lg font-semibold text-gray-900">{config.max_rfqs_10min}</p>
            </div>
            <div className="rounded-lg border border-gray-100 p-4">
              <p className="text-sm text-gray-500">Token TTL</p>
              <p className="text-lg font-semibold text-gray-900">{config.download_token_ttl_min} min</p>
            </div>
          </div>
        ) : null}
      </div>
    </div>
  );
}
