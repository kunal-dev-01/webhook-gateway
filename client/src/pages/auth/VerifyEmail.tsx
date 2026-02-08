// src/pages/VerifyEmail.tsx

import { useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { useVerifyEmail } from "@/hooks/useAuth";
import { Loader } from "@/common/Loader";
import { CircleCheck, Error } from "@/svg/Common";
import { Link } from "react-router-dom";

type VerificationState = "verifying" | "success" | "error" | "invalid";

const VerifyEmail = () => {

  const [searchParams] = useSearchParams();

  const [verificationState, setVerificationState] =
    useState<VerificationState>("verifying");
  const [errorMessage, setErrorMessage] = useState<string>("");
  const [countdown, setCountdown] = useState(5);

  const { verifyEmail, isPending, isSuccess, isError, error } =
    useVerifyEmail();

  // Extract token from URL on mount
  useEffect(() => {
    const token = searchParams.get("token");

    if (!token) {
      setVerificationState("invalid");
      setErrorMessage("Verification link is invalid or expired.");
      return;
    }

    // Auto-verify with token
    verifyEmail({ token });
  }, [verifyEmail]);

  // Handle success
  useEffect(() => {
    if (isSuccess) {
      setVerificationState("success");
    }
  }, [isSuccess]);

  // Handle error
  useEffect(() => {
    if (isError && error) {
      setVerificationState("error");
      const errorMsg =
        (error as any)?.response?.data?.error ||
        "Verification failed. The link may be invalid or expired.";
      setErrorMessage(errorMsg);
    }
  }, [isError, error]);

  // Countdown timer for success redirect
  useEffect(() => {
    if (verificationState === "success" && countdown > 0) {
      const timer = setTimeout(() => setCountdown(countdown - 1), 1000);
      return () => clearTimeout(timer);
    }

    if (verificationState === "success" && countdown === 0) {
      window.location.href = "/auth/login";
    }
  }, [verificationState, countdown]);

  // Verifying state
  if (verificationState === "verifying" || isPending) {
    return (
      <div className="w-full max-w-md mx-auto">
        <div className="text-center">
          {/* Loading Icon */}
          <div className="w-20 h-20 bg-indigo-100 dark:bg-indigo-900/30 rounded-full flex items-center justify-center mx-auto mb-6">
            <Loader className="w-10 h-10 text-indigo-600 dark:text-indigo-400" />
          </div>

          {/* Title */}
          <h1 className="text-2xl font-bold text-zinc-900 dark:text-white mb-3">
            Verifying Your Email
          </h1>

          {/* Description */}
          <p className="text-sm text-zinc-600 dark:text-zinc-400 mb-8">
            Please wait while we verify your email address...
          </p>

          {/* Loading Animation */}
          <div className="flex justify-center gap-2">
            <div
              className="w-2 h-2 bg-indigo-600 dark:bg-indigo-400 rounded-full animate-bounce"
              style={{ animationDelay: "0ms" }}
            ></div>
            <div
              className="w-2 h-2 bg-indigo-600 dark:bg-indigo-400 rounded-full animate-bounce"
              style={{ animationDelay: "150ms" }}
            ></div>
            <div
              className="w-2 h-2 bg-indigo-600 dark:bg-indigo-400 rounded-full animate-bounce"
              style={{ animationDelay: "300ms" }}
            ></div>
          </div>
        </div>
      </div>
    );
  }

  // Success state
  if (verificationState === "success") {
    return (
      <div className="w-full max-w-md mx-auto">
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
              <svg
                className="w-5 h-5 text-green-600 dark:text-green-500 mt-0.5 shrink-0"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M5 13l4 4L19 7"
                />
              </svg>
              <div className="text-left">
                <p className="text-sm font-semibold text-green-900 dark:text-green-300 mb-1">
                  Account Activated
                </p>
                <p className="text-xs text-green-700 dark:text-green-400">
                  Redirecting you to login in {countdown} second
                  {countdown !== 1 ? "s" : ""}...
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

  // Error or Invalid state
  return (
    <div className="w-full max-w-md mx-auto">
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
            <svg
              className="w-5 h-5 text-red-600 dark:text-red-500 mt-0.5 shrink-0"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
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
};

export default VerifyEmail;
