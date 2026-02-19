import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { toast } from "sonner"; 
import { useChangePassword, useLogout } from "@/hooks/useAuth"; 
import AuthInput from "@/common/AuthInput";
import SubmitButton from "@/common/SubmitButton";
import { LockClosed } from "@/svg/Auth";
import { changePasswordSchema, type ChangePasswordFormData } from "@/schemas/user.schema";
import { QuestionMark } from "@/svg/Common";

// ─────────────────────────── Component ─────────────────────────────
export const ChangePasswordForm = () => {
  // Change Password Hook
  const { 
    changePassword, 
    isPending, 
    isSuccess, 
    isError, 
    error, 
  } = useChangePassword();

  // Logout Hook
  const { logout } = useLogout();

  const {
    register,
    handleSubmit,
    reset: resetForm,
    formState: { errors },
  } = useForm<ChangePasswordFormData>({
    resolver: zodResolver(changePasswordSchema),
    defaultValues: {
      current_password: "",
      new_password: "",
      confirm_password: "",
    },
  });

  // Handle Success/Error Side Effects
  useEffect(() => {
    let timeoutId: ReturnType<typeof setTimeout>;

    if (isSuccess) {
      toast.success("Password changed successfully. Logging out from all devices...");
      
      resetForm();
      
      timeoutId = setTimeout(() => {
        logout({ all_devices: true });
      }, 2000);
    }

    if (isError) {
      const errorMessage = (error as any)?.response?.data?.error || "Failed to change password.";
      toast.error(errorMessage);
    }

    // Cleanup timeout if component unmounts naturally
    return () => {
      if (timeoutId) clearTimeout(timeoutId);
    };
  }, [isSuccess, isError, error, resetForm, logout]); 

  const onSubmit = (data: ChangePasswordFormData) => {
    changePassword({
      current_password: data.current_password,
      new_password: data.new_password,
    });
  };

  return (
    <div className="max-w-3xl">
      <div className="flex items-start gap-4 mb-6">
        <div className="p-3 bg-zinc-100 dark:bg-zinc-800 rounded-xl hidden sm:block">
          <LockClosed className="w-6 h-6 text-zinc-500" />
        </div>
        <div>
          <h2 className="text-xl font-bold text-zinc-900 dark:text-white">
            Change Password
          </h2>
          <p className="text-sm text-zinc-500">
            Ensure your account is using a long, random password to stay secure.
          </p>
        </div>
      </div>

      <div className="w-full h-px bg-zinc-100 dark:bg-zinc-800/80 mb-6"></div>

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {/* Current Password - Full Width */}
          <div className="md:col-span-2">
            <AuthInput
              label="Current Password"
              type="password"
              name="current_password"
              placeholder="Enter current password"
              register={register}
              error={errors.current_password?.message}
              disabled={isPending} 
            />
          </div>

          {/* New Password */}
          <AuthInput
            label="New Password"
            type="password"
            name="new_password"
            placeholder="Min. 8 characters"
            register={register}
            error={errors.new_password?.message}
            disabled={isPending}
          />

          {/* Confirm Password */}
          <AuthInput
            label="Confirm New Password"
            type="password"
            name="confirm_password"
            placeholder="Re-enter new password"
            register={register}
            error={errors.confirm_password?.message}
            disabled={isPending}
          />
        </div>

        {/* Footer Actions */}
        <div className="pt-4 flex flex-col sm:flex-row items-center justify-between gap-4">
          <p className="text-sm text-zinc-400 text-center sm:text-left flex items-center gap-2">
            <QuestionMark className="w-5 h-5" /> You will be logged out from all active sessions.
          </p>
          <div className="w-full sm:w-auto">
            <SubmitButton loading={isPending} className="w-full sm:w-auto px-6">
              {isPending ? "Changing..." : "Change Password"}
            </SubmitButton>
          </div>
        </div>
      </form>
    </div>
  );
};