import { type UseFormReturn } from "react-hook-form";
import { Link } from "react-router-dom";
import AuthInput from "@/common/AuthInput";
import SubmitButton from "@/common/SubmitButton";
import { ChevronDown, Error } from "@/svg/Common";
import { Key } from "@/svg/Auth";
import { type ForgotPasswordFormData } from "@/schemas/auth.schema";

type Props = {
  form: UseFormReturn<ForgotPasswordFormData>;
  onSubmit: (data: ForgotPasswordFormData) => void;
  isPending: boolean;
};

export default function ForgotPasswordForm({ form, onSubmit, isPending }: Props) {
  return (
    <div className="w-full max-w-md mx-auto anim-fade-in">
      {/* Header */}
      <div className="text-center mb-8">
        <div className="w-16 h-16 dark:bg-zinc-900 bg-zinc-100 border border-zinc-200 dark:border-zinc-800 rounded-full flex items-center justify-center mx-auto mb-4">
          <Key className="w-10 h-10 text-zinc-900 dark:text-zinc-100" />
        </div>
        <h1 className="text-2xl font-bold text-zinc-900 dark:text-white mb-2">
          Forgot Password?
        </h1>
        <p className="text-sm text-zinc-600 dark:text-zinc-400">
          No worries! Enter your email and we'll send you reset instructions
        </p>
      </div>

      {/* Form */}
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
        <AuthInput
          label="Email Address"
          type="email"
          name="email"
          placeholder="Enter your registered email"
          register={form.register}
          error={form.formState.errors.email?.message}
          disabled={isPending}
        />

        <SubmitButton loading={isPending} type="submit">
          {isPending ? "Sending link..." : "Send Reset Link"}
        </SubmitButton>

        {/* Info Box */}
        <div className="bg-zinc-50 dark:bg-zinc-800/50 border border-zinc-200 dark:border-zinc-700 rounded-xl p-4">
          <div className="flex gap-3">
            <Error className="w-5 h-5 text-zinc-600 dark:text-zinc-400 mt-0.5 shrink-0" />
            <p className="text-xs text-zinc-600 dark:text-zinc-400">
              We'll send you an email with a link to reset your password. The
              link will expire in 1 hour.
            </p>
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

      {/* Footer */}
      <p className="text-center text-xs text-zinc-500 dark:text-zinc-500 mt-6">
        Need help?{" "}
        <Link
          aria-label="contact support"
          to="https://upnixo.com/company/support"
          className="text-indigo-600 dark:text-indigo-400 hover:underline"
        >
          Contact Support
        </Link>
      </p>
    </div>
  );
}