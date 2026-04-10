import { useState } from "react";
import { useParams } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import api from "@/lib/api";

interface Checkpoint {
  id: string;
  checkpointText: string;
  sortOrder: number;
  eta?: string;
}

interface Itinerary {
  id: string;
  title: string;
  status: string;
  meetupAt: string;
  meetupLocationText: string;
  notes: string;
  checkpoints: Checkpoint[];
  membersCount: number;
  members?: { id: string; userId: string; role: string }[];
  formDefinitions?: { id: string; fieldKey: string; fieldLabel: string; fieldType: string; required: boolean }[];
  createdAt: string;
}

const TABS = ["Overview", "Checkpoints", "Members", "Forms", "Change History"];

export default function ItineraryDetail() {
  const { id } = useParams<{ id: string }>();
  const [activeTab, setActiveTab] = useState("Overview");

  const { data: itinerary, isLoading } = useQuery<Itinerary>({
    queryKey: ["itinerary", id],
    queryFn: async () => {
      const { data } = await api.get(`/itineraries/${id}`);
      return data;
    },
  });

  if (isLoading) {
    return <div className="py-12 text-center text-gray-500">Loading...</div>;
  }

  if (!itinerary) {
    return (
      <div className="py-12 text-center text-gray-500">
        Itinerary not found.
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">{itinerary.title}</h1>
        <p className="mt-1 text-gray-500">
          {itinerary.meetupLocationText} &middot; {itinerary.meetupAt}
        </p>
      </div>

      <div className="border-b border-gray-200">
        <nav className="-mb-px flex gap-6">
          {TABS.map((tab) => (
            <button
              key={tab}
              onClick={() => setActiveTab(tab)}
              className={`border-b-2 pb-3 text-sm font-medium transition-colors ${
                activeTab === tab
                  ? "border-indigo-600 text-indigo-600"
                  : "border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700"
              }`}
            >
              {tab}
            </button>
          ))}
        </nav>
      </div>

      <div className="rounded-xl border border-gray-200 bg-white p-6">
        {activeTab === "Overview" && (
          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <span className="text-sm font-medium text-gray-500">
                  Status
                </span>
                <p className="mt-1 capitalize text-gray-900">
                  {itinerary.status}
                </p>
              </div>
              <div>
                <span className="text-sm font-medium text-gray-500">
                  Members
                </span>
                <p className="mt-1 text-gray-900">{itinerary.membersCount}</p>
              </div>
              <div>
                <span className="text-sm font-medium text-gray-500">
                  Location
                </span>
                <p className="mt-1 text-gray-900">{itinerary.meetupLocationText}</p>
              </div>
              <div>
                <span className="text-sm font-medium text-gray-500">
                  Meetup Date
                </span>
                <p className="mt-1 text-gray-900">{itinerary.meetupAt}</p>
              </div>
            </div>
            {itinerary.notes && (
              <div>
                <span className="text-sm font-medium text-gray-500">
                  Notes
                </span>
                <p className="mt-1 text-gray-900">{itinerary.notes}</p>
              </div>
            )}
          </div>
        )}

        {activeTab === "Checkpoints" && (
          <div className="space-y-3">
            {itinerary.checkpoints?.length ? (
              itinerary.checkpoints.map((cp) => (
                <div
                  key={cp.id}
                  className="flex items-start gap-3 rounded-lg border border-gray-100 p-3"
                >
                  <span className="flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full bg-indigo-100 text-xs font-medium text-indigo-700">
                    {cp.sortOrder}
                  </span>
                  <div>
                    <p className="font-medium text-gray-900">{cp.checkpointText}</p>
                  </div>
                </div>
              ))
            ) : (
              <p className="text-sm text-gray-500">No checkpoints defined.</p>
            )}
          </div>
        )}

        {activeTab === "Members" && (
          <p className="text-sm text-gray-500">
            Member management coming soon.
          </p>
        )}

        {activeTab === "Forms" && (
          <p className="text-sm text-gray-500">
            Form configuration coming soon.
          </p>
        )}

        {activeTab === "Change History" && (
          <p className="text-sm text-gray-500">
            Change history coming soon.
          </p>
        )}
      </div>
    </div>
  );
}
