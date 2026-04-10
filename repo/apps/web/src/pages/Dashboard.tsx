import { Link } from "react-router-dom";
import { useAuthStore } from "@/lib/auth";

const quickLinks = [
  {
    to: "/itineraries/new",
    label: "Create Itinerary",
    description: "Plan a new trip or tour",
    color: "bg-indigo-50 text-indigo-700",
  },
  {
    to: "/bookings",
    label: "View Bookings",
    description: "Manage your bookings",
    color: "bg-emerald-50 text-emerald-700",
  },
  {
    to: "/procurement",
    label: "Procurement",
    description: "Track procurement items",
    color: "bg-amber-50 text-amber-700",
  },
  {
    to: "/wallet",
    label: "Wallet",
    description: "View your balance and transactions",
    color: "bg-purple-50 text-purple-700",
  },
];

export default function Dashboard() {
  const { user } = useAuthStore();

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">
          Welcome back, {user?.name ?? "User"}
        </h1>
        <p className="mt-1 text-gray-500">
          Here is an overview of your workspace.
        </p>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {quickLinks.map((link) => (
          <Link
            key={link.to}
            to={link.to}
            className="rounded-xl border border-gray-200 bg-white p-5 transition-shadow hover:shadow-md"
          >
            <div
              className={`mb-3 inline-block rounded-lg px-2.5 py-1 text-xs font-medium ${link.color}`}
            >
              {link.label}
            </div>
            <p className="text-sm text-gray-600">{link.description}</p>
          </Link>
        ))}
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <div className="rounded-xl border border-gray-200 bg-white p-6">
          <h2 className="mb-4 text-lg font-semibold text-gray-900">
            Recent Itineraries
          </h2>
          <p className="text-sm text-gray-500">
            No itineraries yet. Create your first one to get started.
          </p>
        </div>

        <div className="rounded-xl border border-gray-200 bg-white p-6">
          <h2 className="mb-4 text-lg font-semibold text-gray-900">
            Upcoming Trips
          </h2>
          <p className="text-sm text-gray-500">
            No upcoming trips scheduled.
          </p>
        </div>
      </div>
    </div>
  );
}
