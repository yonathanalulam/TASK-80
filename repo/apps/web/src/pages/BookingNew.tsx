import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useMutation } from "@tanstack/react-query";
import api from "@/lib/api";

interface LineItemInput {
  key: number;
  type: string;
  name: string;
  description: string;
  unitPrice: string;
  quantity: string;
  category: string;
}

const ITEM_TYPES = ["lodging", "transport", "activity", "other"];

let nextKey = 1;

function emptyLineItem(): LineItemInput {
  return {
    key: nextKey++,
    type: "lodging",
    name: "",
    description: "",
    unitPrice: "",
    quantity: "1",
    category: "",
  };
}

function formatMoney(cents: number): string {
  return `$${(cents / 100).toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
}

export default function BookingNew() {
  const navigate = useNavigate();
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [itineraryLink, setItineraryLink] = useState("");
  const [lineItems, setLineItems] = useState<LineItemInput[]>([emptyLineItem()]);
  const [errors, setErrors] = useState<string[]>([]);

  const createMutation = useMutation({
    mutationFn: async (payload: {
      title: string;
      description: string;
      itineraryId?: string;
      items: {
        itemType: string;
        itemName: string;
        description: string;
        unitPrice: number;
        quantity: number;
        category: string;
      }[];
    }) => {
      const { data } = await api.post("/bookings", payload);
      return data;
    },
    onSuccess: (data: { id: string }) => {
      navigate(`/bookings/${data.id}`);
    },
  });

  const addLineItem = () => {
    setLineItems((prev) => [...prev, emptyLineItem()]);
  };

  const removeLineItem = (key: number) => {
    setLineItems((prev) => prev.filter((item) => item.key !== key));
  };

  const updateLineItem = (
    key: number,
    field: keyof LineItemInput,
    value: string,
  ) => {
    setLineItems((prev) =>
      prev.map((item) => (item.key === key ? { ...item, [field]: value } : item)),
    );
  };

  const getSubtotal = (item: LineItemInput): number => {
    const price = parseFloat(item.unitPrice) || 0;
    const qty = parseInt(item.quantity, 10) || 0;
    return Math.round(price * 100) * qty;
  };

  const total = lineItems.reduce((sum, item) => sum + getSubtotal(item), 0);

  const validate = (): boolean => {
    const errs: string[] = [];
    if (!title.trim()) errs.push("Title is required.");
    if (lineItems.length === 0) errs.push("At least one line item is required.");
    lineItems.forEach((item, i) => {
      if (!item.name.trim()) errs.push(`Line item ${i + 1}: Name is required.`);
      if (!item.unitPrice || parseFloat(item.unitPrice) <= 0)
        errs.push(`Line item ${i + 1}: Unit price must be greater than 0.`);
      if (!item.quantity || parseInt(item.quantity, 10) <= 0)
        errs.push(`Line item ${i + 1}: Quantity must be greater than 0.`);
    });
    setErrors(errs);
    return errs.length === 0;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!validate()) return;

    createMutation.mutate({
      title: title.trim(),
      description: description.trim(),
      itineraryId: itineraryLink.trim() || undefined,
      items: lineItems.map((item) => ({
        itemType: item.type,
        itemName: item.name.trim(),
        description: item.description.trim(),
        unitPrice: Math.round(parseFloat(item.unitPrice) * 100),
        quantity: parseInt(item.quantity, 10),
        category: item.category.trim(),
      })),
    });
  };

  return (
    <div className="mx-auto max-w-4xl space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Create Booking</h1>
        <p className="mt-1 text-gray-500">
          Fill in the details below to create a new booking.
        </p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        <div className="rounded-xl border border-gray-200 bg-white p-6">
          <h2 className="mb-4 text-lg font-semibold text-gray-900">
            Booking Information
          </h2>
          <div className="space-y-4">
            <div>
              <label className="mb-1 block text-sm font-medium text-gray-700">
                Title <span className="text-red-500">*</span>
              </label>
              <input
                type="text"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder="e.g., Safari Package - Group A"
                className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none"
              />
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-gray-700">
                Description
              </label>
              <textarea
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                rows={3}
                placeholder="Optional description of this booking"
                className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none"
              />
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-gray-700">
                Itinerary Link
              </label>
              <input
                type="text"
                value={itineraryLink}
                onChange={(e) => setItineraryLink(e.target.value)}
                placeholder="Optional link to associated itinerary"
                className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none"
              />
            </div>
          </div>
        </div>

        <div className="rounded-xl border border-gray-200 bg-white p-6">
          <div className="mb-4 flex items-center justify-between">
            <h2 className="text-lg font-semibold text-gray-900">Line Items</h2>
            <button
              type="button"
              onClick={addLineItem}
              className="rounded-lg border border-indigo-300 px-3 py-1.5 text-sm font-medium text-indigo-600 transition-colors hover:bg-indigo-50"
            >
              + Add Item
            </button>
          </div>

          <div className="space-y-4">
            {lineItems.map((item, index) => (
              <div
                key={item.key}
                className="rounded-lg border border-gray-200 bg-gray-50 p-4"
              >
                <div className="mb-3 flex items-center justify-between">
                  <span className="text-sm font-medium text-gray-600">
                    Item {index + 1}
                  </span>
                  {lineItems.length > 1 && (
                    <button
                      type="button"
                      onClick={() => removeLineItem(item.key)}
                      className="text-sm text-red-500 hover:text-red-700"
                    >
                      Remove
                    </button>
                  )}
                </div>
                <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
                  <div>
                    <label className="mb-1 block text-xs font-medium text-gray-600">
                      Type
                    </label>
                    <select
                      value={item.type}
                      onChange={(e) =>
                        updateLineItem(item.key, "type", e.target.value)
                      }
                      className="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none"
                    >
                      {ITEM_TYPES.map((t) => (
                        <option key={t} value={t}>
                          {t.charAt(0).toUpperCase() + t.slice(1)}
                        </option>
                      ))}
                    </select>
                  </div>
                  <div>
                    <label className="mb-1 block text-xs font-medium text-gray-600">
                      Name <span className="text-red-500">*</span>
                    </label>
                    <input
                      type="text"
                      value={item.name}
                      onChange={(e) =>
                        updateLineItem(item.key, "name", e.target.value)
                      }
                      placeholder="Item name"
                      className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none"
                    />
                  </div>
                  <div>
                    <label className="mb-1 block text-xs font-medium text-gray-600">
                      Category
                    </label>
                    <input
                      type="text"
                      value={item.category}
                      onChange={(e) =>
                        updateLineItem(item.key, "category", e.target.value)
                      }
                      placeholder="e.g., premium"
                      className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none"
                    />
                  </div>
                  <div className="sm:col-span-2 lg:col-span-3">
                    <label className="mb-1 block text-xs font-medium text-gray-600">
                      Description
                    </label>
                    <input
                      type="text"
                      value={item.description}
                      onChange={(e) =>
                        updateLineItem(item.key, "description", e.target.value)
                      }
                      placeholder="Optional description"
                      className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none"
                    />
                  </div>
                  <div>
                    <label className="mb-1 block text-xs font-medium text-gray-600">
                      Unit Price ($) <span className="text-red-500">*</span>
                    </label>
                    <input
                      type="number"
                      step="0.01"
                      min="0"
                      value={item.unitPrice}
                      onChange={(e) =>
                        updateLineItem(item.key, "unitPrice", e.target.value)
                      }
                      placeholder="0.00"
                      className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none"
                    />
                  </div>
                  <div>
                    <label className="mb-1 block text-xs font-medium text-gray-600">
                      Quantity <span className="text-red-500">*</span>
                    </label>
                    <input
                      type="number"
                      min="1"
                      value={item.quantity}
                      onChange={(e) =>
                        updateLineItem(item.key, "quantity", e.target.value)
                      }
                      className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none"
                    />
                  </div>
                  <div className="flex items-end">
                    <p className="pb-2 text-sm font-medium text-gray-700">
                      Subtotal: {formatMoney(getSubtotal(item))}
                    </p>
                  </div>
                </div>
              </div>
            ))}
          </div>

          <div className="mt-4 flex justify-end border-t border-gray-200 pt-4">
            <p className="text-lg font-semibold text-gray-900">
              Total: {formatMoney(total)}
            </p>
          </div>
        </div>

        {errors.length > 0 && (
          <div className="rounded-lg border border-red-200 bg-red-50 p-4">
            <ul className="list-inside list-disc space-y-1 text-sm text-red-600">
              {errors.map((err, i) => (
                <li key={i}>{err}</li>
              ))}
            </ul>
          </div>
        )}

        {createMutation.isError && (
          <div className="rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-600">
            Failed to create booking. Please try again.
          </div>
        )}

        <div className="flex justify-end gap-3">
          <button
            type="button"
            onClick={() => navigate("/bookings")}
            className="rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={createMutation.isPending}
            className="rounded-lg bg-indigo-600 px-6 py-2 text-sm font-medium text-white transition-colors hover:bg-indigo-700 disabled:opacity-50"
          >
            {createMutation.isPending ? "Creating..." : "Create Booking"}
          </button>
        </div>
      </form>
    </div>
  );
}
