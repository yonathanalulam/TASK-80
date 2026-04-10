import { useState } from "react";
import { useForm, useFieldArray } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { useNavigate } from "react-router-dom";
import api from "@/lib/api";

const memberFieldSchema = z.object({
  key: z.string().min(1, "Key is required"),
  label: z.string().min(1, "Label is required"),
  type: z.enum(["text", "number", "email", "select", "checkbox", "date"]),
  required: z.boolean(),
  options: z.string().optional(),
});

const checkpointSchema = z.object({
  name: z.string().min(1, "Checkpoint name is required"),
  description: z.string().optional(),
});

const wizardSchema = z.object({
  title: z.string().min(1, "Title is required"),
  meetupDate: z
    .string()
    .min(1, "Meetup date is required")
    .regex(
      /^\d{2}\/\d{2}\/\d{4} \d{1,2}:\d{2} (AM|PM)$/,
      "Format: MM/DD/YYYY h:mm AM/PM",
    ),
  location: z.string().min(1, "Location is required"),
  checkpoints: z
    .array(checkpointSchema)
    .min(1, "At least one checkpoint is required"),
  notes: z.string().optional(),
  memberFields: z.array(memberFieldSchema),
});

type WizardForm = z.infer<typeof wizardSchema>;

const STEPS = [
  "Basic Details",
  "Meetup Date/Time",
  "Location",
  "Checkpoints",
  "Notes",
  "Member Form Config",
  "Review & Publish",
];

export default function ItineraryWizard() {
  const [step, setStep] = useState(0);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const navigate = useNavigate();

  const {
    register,
    handleSubmit,
    control,
    watch,
    trigger,
    formState: { errors, isSubmitting },
  } = useForm<WizardForm>({
    resolver: zodResolver(wizardSchema),
    defaultValues: {
      title: "",
      meetupDate: "",
      location: "",
      checkpoints: [{ name: "", description: "" }],
      notes: "",
      memberFields: [],
    },
  });

  const {
    fields: checkpointFields,
    append: addCheckpoint,
    remove: removeCheckpoint,
    move: moveCheckpoint,
  } = useFieldArray({ control, name: "checkpoints" });

  const {
    fields: memberFieldFields,
    append: addMemberField,
    remove: removeMemberField,
  } = useFieldArray({ control, name: "memberFields" });

  const formValues = watch();

  const validateStep = async (): Promise<boolean> => {
    switch (step) {
      case 0:
        return trigger("title");
      case 1:
        return trigger("meetupDate");
      case 2:
        return trigger("location");
      case 3:
        return trigger("checkpoints");
      case 4:
        return true;
      case 5:
        return trigger("memberFields");
      default:
        return true;
    }
  };

  const nextStep = async () => {
    const valid = await validateStep();
    if (valid && step < STEPS.length - 1) {
      setStep(step + 1);
    }
  };

  const prevStep = () => {
    if (step > 0) setStep(step - 1);
  };

  const onSubmit = async (data: WizardForm) => {
    try {
      setSubmitError(null);
      const meetupDate = new Date(data.meetupDate);

      const { data: created } = await api.post("/itineraries", {
        title: data.title,
        meetupAt: meetupDate.toISOString(),
        meetupLocationText: data.location,
        notes: data.notes || "",
      });

      const itineraryId = created.id;

      for (let i = 0; i < data.checkpoints.length; i++) {
        await api.post(`/itineraries/${itineraryId}/checkpoints`, {
          checkpointText: data.checkpoints[i].name,
          sortOrder: i + 1,
        });
      }

      for (let i = 0; i < data.memberFields.length; i++) {
        const f = data.memberFields[i];
        await api.post(`/itineraries/${itineraryId}/form-definitions`, {
          fieldKey: f.key,
          fieldLabel: f.label,
          fieldType: f.type,
          required: f.required,
          options: f.options ? JSON.stringify(f.options.split(",").map((o: string) => o.trim())) : null,
          sortOrder: i + 1,
        });
      }

      navigate(`/itineraries/${itineraryId}`);
    } catch {
      setSubmitError("Failed to create itinerary. Please try again.");
    }
  };

  const inputClass =
    "w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500";
  const labelClass = "mb-1 block text-sm font-medium text-gray-700";

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Create Itinerary</h1>
        <p className="mt-1 text-gray-500">
          Step {step + 1} of {STEPS.length}: {STEPS[step]}
        </p>
      </div>

      <div className="flex gap-1">
        {STEPS.map((_, i) => (
          <div
            key={i}
            className={`h-1.5 flex-1 rounded-full ${
              i <= step ? "bg-indigo-600" : "bg-gray-200"
            }`}
          />
        ))}
      </div>

      <form
        onSubmit={handleSubmit(onSubmit)}
        className="rounded-xl border border-gray-200 bg-white p-6"
      >
        {submitError && (
          <div className="mb-4 rounded-lg bg-red-50 p-3 text-sm text-red-700">
            {submitError}
          </div>
        )}

        {step === 0 && (
          <div>
            <label htmlFor="title" className={labelClass}>
              Title
            </label>
            <input
              id="title"
              {...register("title")}
              className={inputClass}
              placeholder="Enter itinerary title"
            />
            {errors.title && (
              <p className="mt-1 text-xs text-red-600">
                {errors.title.message}
              </p>
            )}
          </div>
        )}

        {step === 1 && (
          <div>
            <label htmlFor="meetupDate" className={labelClass}>
              Meetup Date & Time
            </label>
            <input
              id="meetupDate"
              {...register("meetupDate")}
              className={inputClass}
              placeholder="MM/DD/YYYY h:mm AM"
            />
            <p className="mt-1 text-xs text-gray-400">
              Format: MM/DD/YYYY h:mm AM/PM (e.g. 01/15/2026 9:00 AM)
            </p>
            {errors.meetupDate && (
              <p className="mt-1 text-xs text-red-600">
                {errors.meetupDate.message}
              </p>
            )}
          </div>
        )}

        {step === 2 && (
          <div>
            <label htmlFor="location" className={labelClass}>
              Location
            </label>
            <input
              id="location"
              {...register("location")}
              className={inputClass}
              placeholder="Enter meetup location"
            />
            {errors.location && (
              <p className="mt-1 text-xs text-red-600">
                {errors.location.message}
              </p>
            )}
          </div>
        )}

        {step === 3 && (
          <div className="space-y-4">
            {checkpointFields.map((field, index) => (
              <div
                key={field.id}
                className="rounded-lg border border-gray-200 p-4"
              >
                <div className="mb-2 flex items-center justify-between">
                  <span className="text-sm font-medium text-gray-700">
                    Checkpoint {index + 1}
                  </span>
                  <div className="flex gap-1">
                    {index > 0 && (
                      <button
                        type="button"
                        onClick={() => moveCheckpoint(index, index - 1)}
                        className="rounded px-2 py-1 text-xs text-gray-500 hover:bg-gray-100"
                      >
                        Up
                      </button>
                    )}
                    {index < checkpointFields.length - 1 && (
                      <button
                        type="button"
                        onClick={() => moveCheckpoint(index, index + 1)}
                        className="rounded px-2 py-1 text-xs text-gray-500 hover:bg-gray-100"
                      >
                        Down
                      </button>
                    )}
                    {checkpointFields.length > 1 && (
                      <button
                        type="button"
                        onClick={() => removeCheckpoint(index)}
                        className="rounded px-2 py-1 text-xs text-red-500 hover:bg-red-50"
                      >
                        Remove
                      </button>
                    )}
                  </div>
                </div>
                <input
                  {...register(`checkpoints.${index}.name`)}
                  className={inputClass}
                  placeholder="Checkpoint name"
                />
                {errors.checkpoints?.[index]?.name && (
                  <p className="mt-1 text-xs text-red-600">
                    {errors.checkpoints[index]?.name?.message}
                  </p>
                )}
                <input
                  {...register(`checkpoints.${index}.description`)}
                  className={`${inputClass} mt-2`}
                  placeholder="Description (optional)"
                />
              </div>
            ))}
            <button
              type="button"
              onClick={() => addCheckpoint({ name: "", description: "" })}
              className="w-full rounded-lg border-2 border-dashed border-gray-300 py-2 text-sm text-gray-500 hover:border-indigo-400 hover:text-indigo-600"
            >
              + Add Checkpoint
            </button>
          </div>
        )}

        {step === 4 && (
          <div>
            <label htmlFor="notes" className={labelClass}>
              Notes
            </label>
            <textarea
              id="notes"
              {...register("notes")}
              className={`${inputClass} min-h-[120px]`}
              placeholder="Additional notes (optional)"
            />
          </div>
        )}

        {step === 5 && (
          <div className="space-y-4">
            <p className="text-sm text-gray-500">
              Configure fields that members must fill out when joining.
            </p>
            {memberFieldFields.map((field, index) => (
              <div
                key={field.id}
                className="rounded-lg border border-gray-200 p-4"
              >
                <div className="mb-2 flex items-center justify-between">
                  <span className="text-sm font-medium text-gray-700">
                    Field {index + 1}
                  </span>
                  <button
                    type="button"
                    onClick={() => removeMemberField(index)}
                    className="rounded px-2 py-1 text-xs text-red-500 hover:bg-red-50"
                  >
                    Remove
                  </button>
                </div>
                <div className="grid grid-cols-2 gap-2">
                  <div>
                    <input
                      {...register(`memberFields.${index}.key`)}
                      className={inputClass}
                      placeholder="Key (e.g. phone)"
                    />
                    {errors.memberFields?.[index]?.key && (
                      <p className="mt-1 text-xs text-red-600">
                        {errors.memberFields[index]?.key?.message}
                      </p>
                    )}
                  </div>
                  <div>
                    <input
                      {...register(`memberFields.${index}.label`)}
                      className={inputClass}
                      placeholder="Label (e.g. Phone Number)"
                    />
                    {errors.memberFields?.[index]?.label && (
                      <p className="mt-1 text-xs text-red-600">
                        {errors.memberFields[index]?.label?.message}
                      </p>
                    )}
                  </div>
                  <select
                    {...register(`memberFields.${index}.type`)}
                    className={inputClass}
                  >
                    <option value="text">Text</option>
                    <option value="number">Number</option>
                    <option value="email">Email</option>
                    <option value="select">Select</option>
                    <option value="checkbox">Checkbox</option>
                    <option value="date">Date</option>
                  </select>
                  <label className="flex items-center gap-2 text-sm text-gray-700">
                    <input
                      type="checkbox"
                      {...register(`memberFields.${index}.required`)}
                      className="rounded border-gray-300"
                    />
                    Required
                  </label>
                </div>
                <input
                  {...register(`memberFields.${index}.options`)}
                  className={`${inputClass} mt-2`}
                  placeholder="Options (comma-separated, for select type)"
                />
              </div>
            ))}
            <button
              type="button"
              onClick={() =>
                addMemberField({
                  key: "",
                  label: "",
                  type: "text",
                  required: false,
                  options: "",
                })
              }
              className="w-full rounded-lg border-2 border-dashed border-gray-300 py-2 text-sm text-gray-500 hover:border-indigo-400 hover:text-indigo-600"
            >
              + Add Field
            </button>
          </div>
        )}

        {step === 6 && (
          <div className="space-y-4">
            <h3 className="font-semibold text-gray-900">Review</h3>
            <div className="space-y-3 text-sm">
              <div>
                <span className="font-medium text-gray-700">Title:</span>{" "}
                {formValues.title}
              </div>
              <div>
                <span className="font-medium text-gray-700">
                  Meetup Date:
                </span>{" "}
                {formValues.meetupDate}
              </div>
              <div>
                <span className="font-medium text-gray-700">Location:</span>{" "}
                {formValues.location}
              </div>
              <div>
                <span className="font-medium text-gray-700">
                  Checkpoints:
                </span>{" "}
                {formValues.checkpoints?.length ?? 0}
                <ul className="ml-4 mt-1 list-disc text-gray-500">
                  {formValues.checkpoints?.map((cp, i) => (
                    <li key={i}>{cp.name || "(unnamed)"}</li>
                  ))}
                </ul>
              </div>
              {formValues.notes && (
                <div>
                  <span className="font-medium text-gray-700">Notes:</span>{" "}
                  {formValues.notes}
                </div>
              )}
              <div>
                <span className="font-medium text-gray-700">
                  Member Fields:
                </span>{" "}
                {formValues.memberFields?.length ?? 0}
                {(formValues.memberFields?.length ?? 0) > 0 && (
                  <ul className="ml-4 mt-1 list-disc text-gray-500">
                    {formValues.memberFields?.map((f, i) => (
                      <li key={i}>
                        {f.label || f.key} ({f.type}
                        {f.required ? ", required" : ""})
                      </li>
                    ))}
                  </ul>
                )}
              </div>
            </div>
          </div>
        )}

        <div className="mt-6 flex justify-between">
          <button
            type="button"
            onClick={prevStep}
            disabled={step === 0}
            className="rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-100 disabled:opacity-50"
          >
            Back
          </button>

          {step < STEPS.length - 1 ? (
            <button
              type="button"
              onClick={nextStep}
              className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-indigo-700"
            >
              Next
            </button>
          ) : (
            <button
              type="submit"
              disabled={isSubmitting}
              className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-indigo-700 disabled:opacity-50"
            >
              {isSubmitting ? "Publishing..." : "Publish Itinerary"}
            </button>
          )}
        </div>
      </form>
    </div>
  );
}
