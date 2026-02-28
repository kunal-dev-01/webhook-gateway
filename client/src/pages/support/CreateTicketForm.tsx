// src/pages/support/CreateTicketForm.tsx
import { z } from "zod";
import { useForm, Controller } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { toast } from "sonner";
import { useCreateTicket } from "@/hooks/useSupport";
import AuthInput from "@/common/AuthInput";
import SubmitButton from "@/common/SubmitButton";
import SelectInput from "@/common/SelectInput";

const ticketSchema = z.object({
  subject: z.string().min(5, "Subject must be at least 5 characters"),
  category: z.enum(["general", "billing", "technical", "account", "feature_request", "bug"]),
  priority: z.enum(["low", "normal", "high", "urgent"]),
  body: z.string().min(10, "Please provide more details (min 10 characters)"),
});

type TicketFormValues = z.infer<typeof ticketSchema>;

const CATEGORY_OPTIONS = [
  { label: "General", value: "general" },
  { label: "Billing", value: "billing" },
  { label: "Technical Support", value: "technical" },
  { label: "Bug Report", value: "bug" },
  { label: "Feature Request", value: "feature_request" },
];

const PRIORITY_OPTIONS = [
  { label: "Low", value: "low" },
  { label: "Normal", value: "normal" },
  { label: "High", value: "high" },
  { label: "Urgent", value: "urgent" },
];

export const CreateTicketForm = ({ onSuccess }: { onSuccess: () => void }) => {
  const { createTicketAsync, isPending } = useCreateTicket();

  const { register, handleSubmit, control, reset, formState: { errors } } = useForm<TicketFormValues>({
    resolver: zodResolver(ticketSchema),
    defaultValues: { category: "general", priority: "normal", subject: "", body: "" },
  });

  const onSubmit = async (data: TicketFormValues) => {
    try {
      await createTicketAsync(data);
      toast.success("Ticket created successfully!");
      reset();
      onSuccess();
    } catch (err) {
      toast.error("Failed to create ticket.");
    }
  };

  return (
    <section className="bg-white dark:bg-zinc-900/70 border border-zinc-200 dark:border-zinc-800 rounded-4xl p-6">
      <h2 className="text-xl font-bold text-zinc-900 dark:text-white mb-2">Create New Ticket</h2>
      <p className="text-sm text-zinc-500 mb-6">Describe your issue in detail and we'll help you out.</p>

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
        <AuthInput
          label="Subject"
          type="text"
          name="subject"
          placeholder="E.g. Cannot access dashboard"
          register={register}
          error={errors.subject?.message}
          disabled={isPending}
        />

        <div className="grid grid-cols-2 gap-4">
          <Controller
            name="category"
            control={control}
            render={({ field }) => (
              <SelectInput
                label="Category"
                options={CATEGORY_OPTIONS}
                value={field.value}
                onChange={field.onChange}
                disabled={isPending}
                className="p-3"
              />
            )}
          />
          <Controller
            name="priority"
            control={control}
            render={({ field }) => (
              <SelectInput
                label="Priority"
                options={PRIORITY_OPTIONS}
                value={field.value}
                onChange={field.onChange}
                disabled={isPending}
                className="p-3"
              />
            )}
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-200 ml-1 mb-1.5">
            Description
          </label>
          <textarea
            {...register("body")}
            disabled={isPending}
            className="w-full min-h-30 p-3 border border-zinc-200 dark:border-zinc-800 rounded-xl bg-transparent outline-none focus:border-indigo-500 dark:focus:border-indigo-600 text-sm text-zinc-900 dark:text-white resize-none"
            placeholder="Please explain the issue..."
          />
          {errors.body && <p className="text-xs text-red-500 mt-1.5 ml-1">{errors.body.message}</p>}
        </div>

        <SubmitButton loading={isPending} type="submit">
          {isPending ? "Submitting..." : "Submit Ticket"}
        </SubmitButton>
      </form>
    </section>
  );
};