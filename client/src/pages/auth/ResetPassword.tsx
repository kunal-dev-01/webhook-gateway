// src/pages/ResetPassword.tsx

import { useEffect, useState } from "react";
import { useForm, useWatch } from "react-hook-form";
import { useSearchParams } from "react-router-dom";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  resetPasswordSchema,
  type ResetPasswordFormData,
} from "@/schemas/auth.schema";
import { useResetPassword } from "@/hooks/useAuth";
import AuthInput from "@/common/AuthInput";
import SubmitButton from "@/common/SubmitButton";
import Toast from "@/common/Toast";
import PasswordStrength from "@/common/PasswordStrength";
import { ChevronDown, CircleCheck, Error, Shield, TriangleAlert } from "@/svg/Common";
import { Link } from "react-router-dom";
import { Key } from "@/svg/Auth";

const ResetPassword = () => {

  const [searchParams] = useSearchParams();

  const [toastStatus, setToastStatus] = useState<"success" | "error" | null>(
    null,
  );
  const [toastMessage, setToastMessage] = useState<string | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [isTokenValid, setIsTokenValid] = useState(true);
  const [countdown, setCountdown] = useState(5);

  const {
    register,
    handleSubmit,
    control,
    formState: { errors },
  } = useForm<ResetPasswordFormData>({
    resolver: zodResolver(resetPasswordSchema),
    defaultValues: {
      password: "",
      confirmPassword: "",
    },
  });

  // Watch password for strength meter
  const password = useWatch({ control, name: "password" });

  const { resetPassword, isPending, isSuccess, isError, error } =
    useResetPassword();

  // Extract token from URL on mount
  useEffect(() => {
    const resetToken = searchParams.get("token");

    if (!resetToken) {
      setIsTokenValid(false);
      setToastStatus("error");
      setToastMessage(
        "Invalid or missing reset token. Please request a new password reset link.",
      );
    } else {
      setToken(resetToken);
    }
  }, []);

  // Success handling with countdown
  useEffect(() => {
    if (isSuccess && countdown > 0) {
      const timer = setTimeout(() => setCountdown(countdown - 1), 1000);
      return () => clearTimeout(timer);
    }

    if (isSuccess && countdown === 0) {
      window.location.href = "/auth/login";
    }
  }, [isSuccess, countdown]);

  // Initial success state
  useEffect(() => {
    if (isSuccess) {
      setToastStatus("success");
      setToastMessage("Password reset successful! Redirecting to login...");
    }
  }, [isSuccess]);

  // Error handling
  useEffect(() => {
    if (isError && error) {
      const errorMessage =
        (error as any)?.response?.data?.error ||
        "Password reset failed. The link may be invalid or expired.";
      setToastStatus("error");
      setToastMessage(errorMessage);
    }
  }, [isError, error]);

  const onSubmit = (data: ResetPasswordFormData) => {
    if (!token) return;

    resetPassword({
      token,
      new_password: data.password,
    });
  };

  const handleToastClose = () => {
    setToastStatus(null);
    setToastMessage(null);
  };

  // Success screen after password reset
  if (isSuccess) {
    return (
      <div className="w-full max-w-md mx-auto">
        <div className="text-center">
          {/* Success Icon */}
          <div className="w-20 h-20 bg-green-100 dark:bg-green-900/30 rounded-full flex items-center justify-center mx-auto mb-6">
            <CircleCheck className="w-10 h-10 text-green-600 dark:text-green-400" />
          </div>

          {/* Title */}
          <h1 className="text-2xl font-bold text-zinc-900 dark:text-white mb-3">
            Password Reset Successfully!
          </h1>

          {/* Description */}
          <p className="text-sm text-zinc-600 dark:text-zinc-400 mb-8">
            Your password has been changed. You can now sign in with your new
            password.
          </p>

          {/* Success Message Box */}
          <div className="bg-green-50 dark:bg-green-900/20 border border-green-100 dark:border-green-900 rounded-xl p-4 mb-6">
            <div className="flex items-start gap-3">
              <CircleCheck className="w-5 h-5 text-green-600 dark:text-green-400 mt-0.5 shrink-0" />
              <div className="text-left">
                <p className="text-sm font-semibold text-green-900 dark:text-green-300 mb-1">
                  All Set!
                </p>
                <p className="text-xs text-green-700 dark:text-green-400">
                  Redirecting you to login in {countdown} second
                  {countdown !== 1 ? "s" : ""}...
                </p>
              </div>
            </div>
          </div>

          {/* Manual Redirect Button */}
          <a aria-label="sign in" href="/auth/login">
            <button className="w-full bg-indigo-500 hover:bg-indigo-600 text-white font-semibold py-2.5 rounded-xl transition-all duration-200">
              Continue to Login
            </button>
          </a>
        </div>
      </div>
    );
  }

  // Invalid token screen
  if (!isTokenValid || !token) {
    return (
      <div className="w-full max-w-md mx-auto">
        <div className="text-center">
          {/* Error Icon */}
          <div className="w-20 h-20 bg-red-100 dark:bg-red-900/30 rounded-full flex items-center justify-center mx-auto mb-6">
            <TriangleAlert className="w-10 h-10 text-red-600 dark:text-red-500" />
          </div>

          {/* Title */}
          <h1 className="text-2xl font-bold text-zinc-900 dark:text-white mb-3">
            Invalid Reset Link
          </h1>

          {/* Description */}
          <p className="text-sm text-zinc-600 dark:text-zinc-400 mb-8">
            The password reset link is invalid or has expired.
          </p>

          {/* Toast */}
          <Toast
            status={toastStatus}
            message={toastMessage}
            onClose={handleToastClose}
          />

          {/* Error Message Box */}
          <div className="bg-red-50 dark:bg-red-900/20 border border-red-100 dark:border-red-900 rounded-xl p-4 mb-6">
            <div className="flex items-start gap-3">
              <Error className="w-5 h-5 text-red-600 dark:text-red-500 mt-0.5 shrink-0" />
              <div className="text-left">
                <p className="text-sm font-semibold text-red-900 dark:text-red-300 mb-1">
                  Possible Reasons:
                </p>
                <ul className="text-xs text-red-700 dark:text-red-400 space-y-1 list-disc list-inside">
                  <li>The reset link has expired (valid for 1 hour)</li>
                  <li>The link has already been used</li>
                  <li>The link is invalid or corrupted</li>
                </ul>
              </div>
            </div>
          </div>

          {/* Action Buttons */}
          <div className="space-y-3">
            <a aria-label="Request New Reset Link" href="/auth/forgot-password">
              <button className="w-full bg-indigo-500 hover:bg-indigo-600 text-white font-semibold py-2.5 rounded-xl transition-all duration-200">
                Request New Reset Link
              </button>
            </a>

            <a aria-label="Back to Login" href="/auth/login">
              <button className="w-full border border-zinc-200 dark:border-zinc-800 text-zinc-700 dark:text-zinc-200 font-semibold py-2.5 rounded-xl hover:bg-zinc-50 dark:hover:bg-zinc-800 transition-all duration-200 mt-4">
                Back to Login
              </button>
            </a>
          </div>

          {/* Help Text */}
          <p className="text-xs text-zinc-500 dark:text-zinc-500 mt-6">
            Need help?{" "}
            <a
              aria-label="Contact Support"
              href="/company/support"
              className="text-indigo-600 dark:text-indigo-400 hover:underline"
            >
              Contact Support
            </a>
          </p>
        </div>
      </div>
    );
  }

  // Reset password form
  return (
    <div className="w-full max-w-md mx-auto">
      {/* Header */}
      <div className="text-center mb-8">
        <div className="w-16 h-16 bg-linear-to-br from-indigo-500 to-purple-600 rounded-full flex items-center justify-center mx-auto mb-4">
          <Key />
        </div>
        <h1 className="text-2xl font-bold text-zinc-900 dark:text-white mb-2">
          Reset Your Password
        </h1>
        <p className="text-sm text-zinc-600 dark:text-zinc-400">
          Enter your new password below
        </p>
      </div>

      {/* Toast */}
      <Toast
        status={toastStatus}
        message={toastMessage}
        onClose={handleToastClose}
      />

      {/* Form */}
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        <AuthInput
          label="New Password"
          type="password"
          name="password"
          placeholder="Create a strong password"
          register={register}
          error={errors.password?.message}
          disabled={isPending}
        />

        <AuthInput
          label="Confirm New Password"
          type="password"
          name="confirmPassword"
          placeholder="Confirm your new password"
          register={register}
          error={errors.confirmPassword?.message}
          disabled={isPending}
        />

        <PasswordStrength password={password || ""} />

        <div className="mt-4">
          <SubmitButton loading={isPending} type="submit">
            {isPending ? "Resetting password..." : "Reset Password"}
          </SubmitButton>
        </div>

        {/* Security Info */}
        <div className="bg-zinc-50 dark:bg-zinc-800/50 border border-zinc-200 dark:border-zinc-700 rounded-xl p-4">
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
};

export default ResetPassword;
