import { Link } from "react-router-dom";
import SubmitButton from "@/common/SubmitButton";
import { ChevronDown, Mail, TriangleAlert } from "@/svg/Common";

type Props = {
  submittedEmail: string;
  cooldown: number;
  isPending: boolean;
  handleResend: () => void;
  handleUseAnotherEmail: () => void;
};

export default function SuccessScreen({
  submittedEmail,
  cooldown,
  isPending,
  handleResend,
  handleUseAnotherEmail,
}: Props) {
  return (
    <div className="w-full max-w-md mx-auto anim-fade-in">
      {/* Success Icon & Header */}
      <div className="text-center mb-8">
        <div className="w-20 h-20 bg-indigo-100 dark:bg-indigo-900/30 rounded-full flex items-center justify-center mx-auto mb-4">
          <Mail className="w-10 h-10 text-indigo-600 dark:text-indigo-400" />
        </div>
        <h1 className="text-2xl font-bold text-zinc-900 dark:text-white mb-2">
          Check Your Email
        </h1>
        <p className="text-sm text-zinc-600 dark:text-zinc-400 mb-4">
          We've sent a password reset link to
          <br />
          <span className="font-semibold text-zinc-900 dark:text-white">
            {submittedEmail}
          </span>
        </p>
      </div>

      {/* Instructions */}
      <div className="bg-indigo-50 dark:bg-indigo-900/20 border border-indigo-100 dark:border-indigo-900 rounded-xl p-4 mb-6">
        <h3 aria-label="steps" className="text-sm font-semibold text-indigo-900 dark:text-indigo-300 mb-2">
          Next Steps:
        </h3>
        <ol className="text-sm text-indigo-700 dark:text-indigo-400 space-y-1 list-decimal list-inside">
          <li>Open your email inbox</li>
          <li>Click the password reset link</li>
          <li>Create a new password</li>
          <li>Sign in with your new password</li>
        </ol>
      </div>

      {/* Important Note */}
      <div className="bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-100 dark:border-yellow-900 rounded-xl p-4 mb-6">
        <div className="flex gap-3">
          <TriangleAlert className="w-5 h-5 text-yellow-700 dark:text-yellow-400 mt-0.5 shrink-0" />
          <div>
            <p className="text-xs font-semibold text-yellow-900 dark:text-yellow-300 mb-1">
              Link expires in 1 hour
            </p>
            <p className="text-xs text-yellow-700 dark:text-yellow-400">
              If you don't see the email, check your spam folder
            </p>
          </div>
        </div>
      </div>

      {/* Action Buttons */}
      <div className="space-y-3">
        <SubmitButton
          onClick={handleResend}
          disabled={cooldown > 0 || isPending}
          loading={isPending}
        >
          {isPending
            ? "Sending..."
            : cooldown > 0
              ? `Resend in ${cooldown}s`
              : "Resend Reset Link"}
        </SubmitButton>

        <button
          onClick={handleUseAnotherEmail}
          className="w-full border border-zinc-200 dark:border-zinc-800 text-zinc-700 dark:text-zinc-200 font-semibold py-2.5 rounded-xl hover:bg-zinc-50 dark:hover:bg-zinc-800 transition-all duration-200"
        >
          Use Another Email
        </button>

        {/* Back to Login */}
        <div className="text-center pt-2">
          <Link
            aria-label="sign in"
            to="/auth/login"
            className="text-sm text-indigo-600 dark:text-indigo-400 hover:text-indigo-700 dark:hover:text-indigo-300 font-medium transition-colors flex items-center justify-center gap-3"
          >
            <ChevronDown className="w-5 h-5 rotate-90" /> Back to login
          </Link>
        </div>
      </div>
    </div>
  );
}