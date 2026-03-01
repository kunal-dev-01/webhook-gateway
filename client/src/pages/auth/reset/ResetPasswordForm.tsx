import { type UseFormReturn } from "react-hook-form";
import { Link } from "react-router-dom";
import AuthInput from "@/common/AuthInput";
import SubmitButton from "@/common/SubmitButton";
import PasswordStrength from "@/common/PasswordStrength";
import { ChevronDown, Shield } from "@/svg/Common";
import { Key } from "@/svg/Auth";
import { type ResetPasswordFormData } from "@/schemas/auth.schema";

type Props = {
  form: UseFormReturn<ResetPasswordFormData>;
  onSubmit: (data: ResetPasswordFormData) => void;
  isPending: boolean;
  password: string;
};

export default function ResetPasswordForm({ form, onSubmit, isPending, password }: Props) {
  return (
    <div className="w-full max-w-md mx-auto anim-fade-in">
      {/* Header */}
      <div className="text-center mb-8">
        <div className="w-16 h-16 dark:bg-zinc-900 bg-zinc-100 border border-zinc-200 dark:border-zinc-800 rounded-full flex items-center justify-center mx-auto mb-4">
          <Key className="w-10 h-10 text-zinc-900 dark:text-zinc-100" />
        </div>
        <h1 className="text-2xl font-bold text-zinc-900 dark:text-white mb-2">
          Reset Your Password
        </h1>
        <p className="text-sm text-zinc-600 dark:text-zinc-400">
          Enter your new password below
        </p>
      </div>

      {/* Form */}
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
        <AuthInput
          label="New Password"
          type="password"
          name="password"
          placeholder="Create a strong password"
          register={form.register}
          error={form.formState.errors.password?.message}
          disabled={isPending}
        />

        <AuthInput
          label="Confirm New Password"
          type="password"
          name="confirmPassword"
          placeholder="Confirm your new password"
          register={form.register}
          error={form.formState.errors.confirmPassword?.message}
          disabled={isPending}
        />

        <PasswordStrength password={password} />

        <div className="mt-4">
          <SubmitButton loading={isPending} type="submit">
            {isPending ? "Resetting password..." : "Reset Password"}
          </SubmitButton>
        </div>

        {/* Security Info */}
        <div className="bg-zinc-50 dark:bg-zinc-900/50 border border-zinc-200 dark:border-zinc-800 rounded-xl p-4">
          <div className="flex gap-3">
            <Shield className="w-6 h-6 text-zinc-600 dark:text-zinc-400 shrink-0 mt-0.5" />
            <div className="text-left">
              <p className="text-xs font-semibold text-zinc-700 dark:text-zinc-300 mb-1">
                Password Security Tips
              </p>
              <ul className="text-xs text-zinc-600 dark:text-zinc-400 space-y-0.5">
                <li>• Use a unique password you haven't used before</li>
                <li>• Don't share your password with anyone</li>
                <li>• Consider using a password manager</li>
              </ul>
            </div>
          </div>
        </div>

        {/* Back to Login */}
        <div className="text-center pt-4 border-t border-zinc-200 dark:border-zinc-800">
          <Link
            aria-label="sign in"
            to="/auth/login"
            className="text-sm text-indigo-600 dark:text-indigo-400 hover:text-indigo-700 dark:hover:text-indigo-300 font-medium transition-colors flex items-center justify-center gap-3"
          >
            <ChevronDown className="w-5 h-5 rotate-90" /> Back to login
          </Link>
        </div>
      </form>
    </div>
  );
}