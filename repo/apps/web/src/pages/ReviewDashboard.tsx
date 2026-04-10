import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import api from "@/lib/api";
import { useAuthStore } from "@/lib/auth";

interface Review {
  id: string;
  subjectId: string;
  reviewerId: string;
  orderType: string;
  orderId: string;
  overallRating: number;
  comment: string;
  scores: { dimensionName: string; score: number }[];
  createdAt: string;
}

interface CreditTier {
  tier: string;
  score: number;
  label: string;
}

interface PaginatedReviews {
  items: Review[];
  total: number;
  page: number;
}

type TabKey = "my-reviews" | "about-me";

const DIMENSIONS = [
  { key: "punctuality", label: "Punctuality" },
  { key: "communication", label: "Communication" },
  { key: "quality", label: "Quality" },
  { key: "reliability", label: "Reliability" },
  { key: "value", label: "Value" },
];

const TIER_COLORS: Record<string, string> = {
  gold: "bg-yellow-100 text-yellow-800 border-yellow-300",
  silver: "bg-gray-100 text-gray-700 border-gray-300",
  bronze: "bg-orange-100 text-orange-800 border-orange-300",
  platinum: "bg-indigo-100 text-indigo-800 border-indigo-300",
  standard: "bg-gray-50 text-gray-600 border-gray-200",
};

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

function StarRating({
  value,
  onChange,
  readonly = false,
}: {
  value: number;
  onChange?: (v: number) => void;
  readonly?: boolean;
}) {
  return (
    <div className="flex gap-1">
      {[1, 2, 3, 4, 5].map((star) => (
        <button
          key={star}
          type="button"
          disabled={readonly}
          onClick={() => onChange?.(star)}
          className={`text-lg ${
            star <= value ? "text-yellow-400" : "text-gray-300"
          } ${readonly ? "cursor-default" : "cursor-pointer hover:text-yellow-500"}`}
        >
          {"\u2605"}
        </button>
      ))}
    </div>
  );
}

export default function ReviewDashboard() {
  const [activeTab, setActiveTab] = useState<TabKey>("my-reviews");
  const [showReviewForm, setShowReviewForm] = useState(false);
  const [subjectId, setSubjectId] = useState("");
  const [overallRating, setOverallRating] = useState(0);
  const [dimensionRatings, setDimensionRatings] = useState<Record<string, number>>(
    {},
  );
  const [comment, setComment] = useState("");
  const [reviewOrderId, setReviewOrderId] = useState("");
  const [showFlagModal, setShowFlagModal] = useState(false);
  const [flagReviewId, setFlagReviewId] = useState("");
  const [flagReason, setFlagReason] = useState("");

  const { user, hasRole } = useAuthStore();
  const isAdmin = hasRole("administrator");
  const qc = useQueryClient();

  const { data: myReviewsData, isLoading: myReviewsLoading } = useQuery<PaginatedReviews>({
    queryKey: ["my-reviews"],
    queryFn: async () => {
      const { data } = await api.get(`/reviews/subject/${user!.id}`, {
        params: { role: "reviewer" },
      });
      return data;
    },
    enabled: activeTab === "my-reviews" && !!user?.id,
  });
  const myReviews = myReviewsData?.items ?? [];

  const { data: aboutMeData, isLoading: aboutMeLoading } = useQuery<PaginatedReviews>({
    queryKey: ["reviews-about-me"],
    queryFn: async () => {
      const { data } = await api.get(`/reviews/subject/${user!.id}`);
      return data;
    },
    enabled: activeTab === "about-me" && !!user?.id,
  });
  const aboutMeReviews = aboutMeData?.items ?? [];

  const { data: creditTier } = useQuery<CreditTier>({
    queryKey: ["credit-tier"],
    queryFn: async () => {
      const { data } = await api.get(`/credit-tiers/${user!.id}`);
      return data;
    },
    enabled: !!user?.id,
  });

  const submitReviewMutation = useMutation({
    mutationFn: async (payload: {
      subjectId: string;
      orderType: string;
      orderId: string;
      overallRating: number;
      comment: string;
      scores: { dimensionName: string; score: number }[];
    }) => {
      const { data } = await api.post("/reviews", payload);
      return data;
    },
    onSuccess: () => {
      setShowReviewForm(false);
      setSubjectId("");
      setReviewOrderId("");
      setOverallRating(0);
      setDimensionRatings({});
      setComment("");
      qc.invalidateQueries({ queryKey: ["my-reviews"] });
    },
  });

  const flagMutation = useMutation({
    mutationFn: async (payload: { reviewId: string; reason: string }) => {
      const { data } = await api.post("/harassment-flags", payload);
      return data;
    },
    onSuccess: () => {
      setShowFlagModal(false);
      setFlagReviewId("");
      setFlagReason("");
    },
  });

  const reviews = activeTab === "my-reviews" ? myReviews : aboutMeReviews;
  const isLoading =
    activeTab === "my-reviews" ? myReviewsLoading : aboutMeLoading;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Reviews</h1>
          <p className="mt-1 text-gray-500">
            View and manage reviews and feedback.
          </p>
        </div>
        <button
          onClick={() => setShowReviewForm(true)}
          className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-indigo-700"
        >
          Write Review
        </button>
      </div>

      {creditTier && (
        <div
          className={`inline-flex items-center gap-2 rounded-lg border px-4 py-2 ${TIER_COLORS[creditTier.tier] ?? TIER_COLORS.standard}`}
        >
          <span className="text-sm font-medium">
            Credit Tier: {creditTier.label}
          </span>
          <span className="text-xs opacity-75">Score: {creditTier.score}</span>
        </div>
      )}

      <div className="border-b border-gray-200">
        <nav className="-mb-px flex gap-6">
          <button
            onClick={() => setActiveTab("my-reviews")}
            className={`border-b-2 pb-3 text-sm font-medium transition-colors ${
              activeTab === "my-reviews"
                ? "border-indigo-600 text-indigo-600"
                : "border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700"
            }`}
          >
            My Reviews
          </button>
          <button
            onClick={() => setActiveTab("about-me")}
            className={`border-b-2 pb-3 text-sm font-medium transition-colors ${
              activeTab === "about-me"
                ? "border-indigo-600 text-indigo-600"
                : "border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700"
            }`}
          >
            Reviews About Me
          </button>
        </nav>
      </div>

      {isLoading ? (
        <div className="py-12 text-center text-gray-500">Loading reviews...</div>
      ) : !reviews?.length ? (
        <div className="rounded-xl border border-gray-200 bg-white py-12 text-center">
          <p className="text-gray-500">No reviews yet.</p>
        </div>
      ) : (
        <div className="space-y-3">
          {reviews.map((review) => (
            <div
              key={review.id}
              className="rounded-xl border border-gray-200 bg-white p-5"
            >
              <div className="flex items-start justify-between">
                <div>
                  <div className="flex items-center gap-2">
                    <h3 className="text-sm font-semibold text-gray-900">
                      {activeTab === "my-reviews"
                        ? `Review for ${review.subjectId}`
                        : `By ${review.reviewerId}`}
                    </h3>
                    <StarRating value={review.overallRating} readonly />
                  </div>
                  <p className="mt-1 text-xs text-gray-400">
                    {formatDate(review.createdAt)}
                  </p>
                </div>
                {isAdmin && (
                  <button
                    onClick={() => {
                      setFlagReviewId(review.id);
                      setShowFlagModal(true);
                    }}
                    className="text-xs text-red-500 hover:text-red-700"
                  >
                    Flag
                  </button>
                )}
              </div>

              {review.scores && review.scores.length > 0 && (
                  <div className="mt-3 flex flex-wrap gap-3">
                    {review.scores.map((s) => (
                      <div key={s.dimensionName} className="text-xs text-gray-500">
                        <span className="capitalize">{s.dimensionName}</span>:{" "}
                        <span className="font-medium text-gray-700">
                          {s.score}/5
                        </span>
                      </div>
                    ))}
                  </div>
                )}

              {review.comment && (
                <p className="mt-3 text-sm text-gray-700">{review.comment}</p>
              )}
            </div>
          ))}
        </div>
      )}

      {showReviewForm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="mx-4 w-full max-w-lg rounded-xl bg-white p-6 shadow-xl">
            <h3 className="mb-4 text-lg font-semibold text-gray-900">
              Write a Review
            </h3>
            <div className="space-y-4">
              <div>
                <label className="mb-1 block text-sm font-medium text-gray-700">
                  Subject (User ID)
                </label>
                <input
                  type="text"
                  value={subjectId}
                  onChange={(e) => setSubjectId(e.target.value)}
                  placeholder="Enter user ID to review"
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none"
                />
              </div>

              <div>
                <label className="mb-1 block text-sm font-medium text-gray-700">
                  Order ID
                </label>
                <input
                  type="text"
                  value={reviewOrderId}
                  onChange={(e) => setReviewOrderId(e.target.value)}
                  placeholder="Enter order ID"
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none"
                />
              </div>

              <div>
                <label className="mb-1 block text-sm font-medium text-gray-700">
                  Overall Rating
                </label>
                <StarRating
                  value={overallRating}
                  onChange={setOverallRating}
                />
              </div>

              {DIMENSIONS.map((dim) => (
                <div key={dim.key} className="flex items-center justify-between">
                  <label className="text-sm text-gray-700">{dim.label}</label>
                  <StarRating
                    value={dimensionRatings[dim.key] ?? 0}
                    onChange={(v) =>
                      setDimensionRatings((prev) => ({
                        ...prev,
                        [dim.key]: v,
                      }))
                    }
                  />
                </div>
              ))}

              <div>
                <label className="mb-1 block text-sm font-medium text-gray-700">
                  Comment
                </label>
                <textarea
                  value={comment}
                  onChange={(e) => setComment(e.target.value)}
                  rows={3}
                  placeholder="Share your experience..."
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none"
                />
              </div>

              {submitReviewMutation.isError && (
                <p className="text-sm text-red-600">
                  Failed to submit review. Please try again.
                </p>
              )}
            </div>
            <div className="mt-6 flex justify-end gap-3">
              <button
                onClick={() => setShowReviewForm(false)}
                className="rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100"
              >
                Cancel
              </button>
              <button
                onClick={() =>
                  submitReviewMutation.mutate({
                    subjectId,
                    orderType: "booking",
                    orderId: reviewOrderId,
                    overallRating,
                    comment,
                    scores: Object.entries(dimensionRatings).map(([name, score]) => ({
                      dimensionName: name,
                      score,
                    })),
                  })
                }
                disabled={
                  !subjectId ||
                  !reviewOrderId ||
                  overallRating === 0 ||
                  submitReviewMutation.isPending
                }
                className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-indigo-700 disabled:opacity-50"
              >
                {submitReviewMutation.isPending
                  ? "Submitting..."
                  : "Submit Review"}
              </button>
            </div>
          </div>
        </div>
      )}

      {showFlagModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="mx-4 w-full max-w-md rounded-xl bg-white p-6 shadow-xl">
            <h3 className="mb-4 text-lg font-semibold text-gray-900">
              Flag Review
            </h3>
            <div className="space-y-4">
              <div>
                <label className="mb-1 block text-sm font-medium text-gray-700">
                  Reason
                </label>
                <select
                  value={flagReason}
                  onChange={(e) => setFlagReason(e.target.value)}
                  className="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none"
                >
                  <option value="">Select reason</option>
                  <option value="harassment">Harassment</option>
                  <option value="violation">Policy Violation</option>
                  <option value="spam">Spam</option>
                  <option value="inappropriate">Inappropriate Content</option>
                  <option value="other">Other</option>
                </select>
              </div>
              {flagMutation.isError && (
                <p className="text-sm text-red-600">
                  Failed to flag review. Please try again.
                </p>
              )}
            </div>
            <div className="mt-6 flex justify-end gap-3">
              <button
                onClick={() => setShowFlagModal(false)}
                className="rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100"
              >
                Cancel
              </button>
              <button
                onClick={() =>
                  flagMutation.mutate({
                    reviewId: flagReviewId,
                    reason: flagReason,
                  })
                }
                disabled={!flagReason || flagMutation.isPending}
                className="rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-red-700 disabled:opacity-50"
              >
                {flagMutation.isPending ? "Flagging..." : "Submit Flag"}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
