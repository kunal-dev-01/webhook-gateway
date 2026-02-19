import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { useUpdateOrgName } from "@/hooks/useOrg";
import { type Organization } from "@/types/org.types";
import SubmitButton from "@/common/SubmitButton";
import AuthInput from "@/common/AuthInput"; // 👈 Update path as per your folder structure
import { Settings } from "@/svg/Common";

// Form type define kar liya
interface FormValues {
  name: string;
}

export const OrgGeneralSettings = ({ org }: { org: Organization }) => {
  const { updateName, isPending } = useUpdateOrgName(org.id);

  // react-hook-form setup
  const {
    register,
    handleSubmit,
    watch,
    reset,
    formState: { errors },
  } = useForm<FormValues>({
    defaultValues: {
      name: org.name,
    },
  });

  // Dekhne ke liye ki name change hua hai ya nahi, taki Save button disable kar sakein
  const currentName = watch("name");

  // Agar prop se organization ka naam change hota hai (e.g. refetch hone pe), toh form reset kar do
  useEffect(() => {
    reset({ name: org.name });
  }, [org.name, reset]);

  const onSubmit = (data: FormValues) => {
    if (!data.name.trim()) {
      return toast.error("Organization name cannot be empty");
    }
    
    updateName(
      { name: data.name },
      {
        onSuccess: () => toast.success("Organization updated successfully!"),
        onError: () => toast.error("Failed to update organization name."),
      }
    );
  };

  return (
    <section className="bg-white dark:bg-zinc-900/70 border border-zinc-200 dark:border-zinc-800 rounded-4xl p-6 md:p-8 space-y-6">

      <div className="flex items-start gap-4 mb-6">
        <div className="p-3 bg-zinc-100 dark:bg-zinc-800 rounded-xl hidden sm:block">
          <Settings className="w-6 h-6 text-zinc-500" />
        </div>
        <div>
          <h2 className="text-xl font-bold text-zinc-900 dark:text-white">General Settings</h2>
          <p className="text-sm text-zinc-500">
            Update your workspace identity and core settings.
          </p>
        </div>
      </div>

      <div className="w-full h-px bg-zinc-100 dark:bg-zinc-800/80 mb-6"></div>

      {/* Form tag use kar rahe hain react-hook-form ke liye */}
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        
        {/* Tumhara Custom AuthInput */}
        <AuthInput
          label="Organization Name"
          type="text"
          name="name"
          placeholder="e.g. Acme Corp"
          register={register}
          error={errors.name?.message}
          disabled={org.role === "viewer" || isPending}
        />

        {/* Workspace Slug (Ye read-only hai isliye isko custom hi rakha hai) */}
        <div className="pt-2">
          <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-200 mb-1.5 ml-1">
            Workspace Slug
          </label>
          <div className="flex items-center px-4 py-2.5 rounded-xl border border-zinc-200 dark:border-zinc-800 bg-zinc-100 dark:bg-zinc-900/50 text-zinc-500 cursor-not-allowed transition-all duration-200">
            <span className="text-zinc-900 dark:text-zinc-300">{org.slug}</span>
          </div>
          <p className="text-xs text-zinc-500 mt-1.5 ml-1">
            The slug is auto-generated and cannot be changed.
          </p>
        </div>

        {/* Save Button */}
        {org.role !== "viewer" && (
          <div className="pt-4 flex justify-end">
            <SubmitButton
              loading={isPending}
              disabled={isPending || currentName === org.name || !currentName.trim()}
              onClick={handleSubmit(onSubmit)} 
              className="px-6"
            >
              {isPending ? "Saving..." : "Save Changes"}
            </SubmitButton>
          </div>
        )}
      </form>
    </section>
  );
};