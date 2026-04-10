import { Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import api from "@/lib/api";

interface Itinerary {
  id: string;
  title: string;
  status: string;
  meetupAt: string;
  meetupLocationText: string;
  membersCount: number;
  createdAt: string;
}

interface PaginatedItineraries {
  items: Itinerary[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

const statusColors: Record<string, string> = {
  draft: "bg-gray-100 text-gray-700",
  published: "bg-emerald-100 text-emerald-700",
  archived: "bg-amber-100 text-amber-700",
};

export default function ItineraryList() {
  const { data: pageData, isLoading } = useQuery<PaginatedItineraries>({
    queryKey: ["itineraries"],
    queryFn: async () => {
      const { data } = await api.get("/itineraries");
      return data;
    },
  });
  const itineraries = pageData?.items ?? [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Itineraries</h1>
          <p className="mt-1 text-gray-500">
            Manage your trip itineraries and tours.
          </p>
        </div>
        <Link
          to="/itineraries/new"
          className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-indigo-700"
        >
          Create Itinerary
        </Link>
      </div>

      {isLoading ? (
        <div className="py-12 text-center text-gray-500">Loading...</div>
      ) : !itineraries.length ? (
        <div className="rounded-xl border border-gray-200 bg-white py-12 text-center">
          <p className="text-gray-500">No itineraries found.</p>
          <Link
            to="/itineraries/new"
            className="mt-2 inline-block text-sm font-medium text-indigo-600 hover:text-indigo-700"
          >
            Create your first itinerary
          </Link>
        </div>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {itineraries.map((itinerary) => (
            <Link
              key={itinerary.id}
              to={`/itineraries/${itinerary.id}`}
              className="rounded-xl border border-gray-200 bg-white p-5 transition-shadow hover:shadow-md"
            >
              <div className="mb-2 flex items-center justify-between">
                <h3 className="font-semibold text-gray-900">
                  {itinerary.title}
                </h3>
                <span
                  className={`rounded-full px-2 py-0.5 text-xs font-medium ${statusColors[itinerary.status]}`}
                >
                  {itinerary.status}
                </span>
              </div>
              <p className="text-sm text-gray-500">{itinerary.meetupLocationText}</p>
              <div className="mt-3 flex items-center gap-4 text-xs text-gray-400">
                <span>{itinerary.meetupAt}</span>
                <span>{itinerary.membersCount} members</span>
              </div>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
