// src/pages/ForgotPassword.tsx

import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { forgotPasswordSchema, type ForgotPasswordFormData } from "@/schemas/auth.schema";
import { useForgotPassword } from "@/hooks/useAuth";
import AuthInput from "@/common/AuthInput";
import SubmitButton from "@/common/SubmitButton";
import Toast from "@/common/Toast";
import { ChevronDown, Error, Mail, TriangleAlert } from "@/svg/Common";
import { Link } from "react-router-dom";
import { Key } from "@/svg/Auth";

const ForgotPassword = () => {
  const [toastStatus, setToastStatus] = useState<"success" | "error" | null>(
    null,
  );
  const [toastMessage, setToastMessage] = useState<string | null>(null);
  const [emailSent, setEmailSent] = useState(false);
  const [submittedEmail, setSubmittedEmail] = useState("");
  const [cooldown, setCooldown] = useState(0);

  const {
    register,
    handleSubmit,
    formState: { errors },
    reset: resetForm,
  } = useForm<ForgotPasswordFormData>({
    resolver: zodResolver(forgotPasswordSchema),
    defaultValues: {
      email: "",
    },
  });

  const {
    forgotPassword,
    isPending,
    isSuccess,
    isError,
    error,
    reset: resetMutation,
  } = useForgotPassword();

  // Cooldown timer
  useEffect(() => {
    if (cooldown > 0) {
      const timer = setTimeout(() => setCooldown(cooldown - 1), 1000);
      return () => clearTimeout(timer);
    }
  }, [cooldown]);

  // Success handling
  useEffect(() => {
    if (isSuccess) {
      setToastStatus("success");
      setToastMessage("Password reset link sent! Please check your email.");
      setEmailSent(true);
      setCooldown(60); // 60 seconds cooldown
    }
  }, [isSuccess]);

  // Error handling
  useEffect(() => {
    if (isError && error) {
      const errorMessage =
        (error as any)?.response?.data?.error ||
        "Failed to send reset link. Please try again.";
      setToastStatus("error");
      setToastMessage(errorMessage);
    }
  }, [isError, error]);

  const onSubmit = (data: ForgotPasswordFormData) => {
    setSubmittedEmail(data.email);
    forgotPassword(data);
  };

  const handleResend = () => {
    if (cooldown > 0) return;
    forgotPassword({ email: submittedEmail });
  };

  const handleUseAnotherEmail = () => {
    setEmailSent(false);
    setSubmittedEmail("");
    setCooldown(0);
    resetForm();
    resetMutation();
    setToastStatus(null);
    setToastMessage(null);
  };

  const handleToastClose = () => {
    setToastStatus(null);
    setToastMessage(null);
  };

  // Email sent success screen
  if (emailSent) {
    return (
      <div className="w-full max-w-md mx-auto">
        {/* Success Icon */}
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

        {/* Toast */}
        <Toast
          status={toastStatus}
          message={toastMessage}
          onClose={handleToastClose}
        />

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
          {/* Resend Button */}
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

          {/* Use Another Email Button */}
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

  // Forgot password form
  return (
    <div className="w-full max-w-md mx-auto">
      {/* Header */}
      <div className="text-center mb-8">
        <div className="w-16 h-16 bg-linear-to-br from-indigo-500 to-purple-600 rounded-full flex items-center justify-center mx-auto mb-4">
          <Key className="w-8 h-8 text-white" />
        </div>
        <h1 className="text-2xl font-bold text-zinc-900 dark:text-white mb-2">
          Forgot Password?
        </h1>
        <p className="text-sm text-zinc-600 dark:text-zinc-400">
          No worries! Enter your email and we'll send you reset instructions
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
          label="Email Address"
          type="email"
          name="email"
          placeholder="Enter your registered email"
          register={register}
          error={errors.email?.message}
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
};

export default ForgotPassword;