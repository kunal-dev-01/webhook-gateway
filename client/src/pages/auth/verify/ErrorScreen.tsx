import { Link } from "react-router-dom";
import { Error } from "@/svg/Common";

type Props = {
  errorMessage: string;
};

export default function ErrorScreen({ errorMessage }: Props) {
  return (
    <div className="w-full max-w-md mx-auto anim-fade-in">
      <div className="text-center">
        {/* Error Icon */}
        <div className="w-20 h-20 bg-red-100 dark:bg-red-900/30 rounded-full flex items-center justify-center mx-auto mb-6">
          <Error className="w-10 h-10 text-red-600 dark:text-red-500" />
        </div>

        {/* Title */}
        <h1 className="text-2xl font-bold text-zinc-900 dark:text-white mb-3">
          Verification Failed
        </h1>

        {/* Description */}
        <p className="text-sm text-zinc-600 dark:text-zinc-400 mb-8">
          {errorMessage || "We could not verify your email address."}
        </p>

        {/* Error Message Box */}
        <div className="bg-red-50 dark:bg-red-900/20 border border-red-100 dark:border-red-900 rounded-xl p-4 mb-6">
          <div className="flex items-start gap-3">
            <Error className="w-5 h-5 text-red-600 dark:text-red-500 mt-0.5 shrink-0" />
            <div className="text-left">
              <p className="text-sm font-semibold text-red-900 dark:text-red-300 mb-1">
                Possible Reasons:
              </p>
              <ul className="text-xs text-red-700 dark:text-red-400 space-y-1 list-disc list-inside">
                <li>The verification link has expired</li>
                <li>The link has already been used</li>
                <li>The link is invalid or corrupted</li>
              </ul>
            </div>
          </div>
        </div>

        {/* Action Buttons */}
        <div className="space-y-3">
          <Link aria-label="register" to="/auth/register">
            <button className="w-full bg-indigo-500 hover:bg-indigo-600 text-white font-semibold py-2.5 rounded-xl transition-all duration-200">
              Register Again
            </button>
          </Link>

          <Link aria-label="login" to="/auth/login">
            <button className="w-full border border-zinc-200 dark:border-zinc-800 text-zinc-700 dark:text-zinc-200 font-semibold py-2.5 rounded-xl hover:bg-zinc-50 dark:hover:bg-zinc-800 transition-all duration-200 mt-4">
              Back to Login
            </button>
          </Link>
        </div>

        {/* Help Text */}
        <p className="text-xs text-zinc-500 dark:text-zinc-500 mt-6">
          Need help?{" "}
          <Link
            aria-label="support"
            to="https://upnixo.com/company/support"
            target="_blank"
            className="text-indigo-600 dark:text-indigo-400 hover:underline"
          >
            Contact Support
          </Link>
        </p>
      </div>
    </div>
  );
}