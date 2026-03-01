import { Link } from "react-router-dom";
import { CircleCheck } from "@/svg/Common";

type Props = {
  countdown: number;
};

export default function SuccessScreen({ countdown }: Props) {
  return (
    <div className="w-full max-w-md mx-auto anim-fade-in">
      <div className="text-center">
        {/* Success Icon */}
        <div className="w-20 h-20 bg-green-100 dark:bg-green-900/30 rounded-full flex items-center justify-center mx-auto mb-6">
          <CircleCheck className="w-10 h-10 text-green-600 dark:text-green-500" />
        </div>

        {/* Title */}
        <h1 className="text-2xl font-bold text-zinc-900 dark:text-white mb-3">
          Email Verified Successfully!
        </h1>

        {/* Description */}
        <p className="text-sm text-zinc-600 dark:text-zinc-400 mb-8">
          Your email has been verified. You can now sign in to your account.
        </p>

        {/* Success Message Box */}
        <div className="bg-green-50 dark:bg-green-900/20 border border-green-100 dark:border-green-900 rounded-xl p-4 mb-6">
          <div className="flex items-start gap-3">
            <CircleCheck className="w-5 h-5 text-green-600 dark:text-green-500 mt-0.5 shrink-0" />
            <div className="text-left">
              <p className="text-sm font-semibold text-green-900 dark:text-green-300 mb-1">
                Account Activated
              </p>
              <p className="text-xs text-green-700 dark:text-green-400">
                Redirecting you to login in {countdown} second{countdown !== 1 ? "s" : ""}...
              </p>
            </div>
          </div>
        </div>

        {/* Manual Redirect Button */}
        <Link aria-label="login" to="/auth/login">
          <button className="w-full bg-indigo-500 hover:bg-indigo-600 text-white font-semibold py-2.5 rounded-xl transition-all duration-200">
            Continue to Login
          </button>
        </Link>
      </div>
    </div>
  );
}