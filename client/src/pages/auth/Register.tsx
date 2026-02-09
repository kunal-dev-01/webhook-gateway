// src/pages/Register.tsx

import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { useForm, useWatch } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { registerSchema, type RegisterFormData } from "@/schemas/auth.schema";
import { useRegister } from "@/hooks/useAuth";
import { useResendVerification } from "@/hooks/useAuth";
import AuthInput from "@/common/AuthInput";
import SubmitButton from "@/common/SubmitButton";
import Toast from "@/common/Toast";
import PasswordStrength from "@/common/PasswordStrength";
import { User } from "@/svg/Auth";
import { ChevronDown, Mail } from "@/svg/Common";

const Register = () => {
  const [toastStatus, setToastStatus] = useState<"success" | "error" | null>(
    null,
  );
  const [toastMessage, setToastMessage] = useState<string | null>(null);
  const [showVerification, setShowVerification] = useState(false);
  const [registeredEmail, setRegisteredEmail] = useState("");
  const [cooldown, setCooldown] = useState(0);

  const {
    register,
    handleSubmit,
    control,
    formState: { errors },
  } = useForm<RegisterFormData>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      full_name: "",
      email: "",
      password: "",
      confirmPassword: "",
    },
  });

  // Watch password for strength meter
  const password = useWatch({ control, name: "password" });

  const registerMutation = useRegister();
  const resendMutation = useResendVerification();

  // Cooldown timer
  useEffect(() => {
    if (cooldown > 0) {
      const timer = setTimeout(() => setCooldown(cooldown - 1), 1000);
      return () => clearTimeout(timer);
    }
  }, [cooldown]);

  // Register success handling
  useEffect(() => {
    if (registerMutation.isSuccess) {
      setToastStatus("success");
      setToastMessage(
        "Registration successful! Please check your email to verify your account.",
      );
      setShowVerification(true);
    }
  }, [registerMutation.isSuccess]);

  // Register error handling
  useEffect(() => {
    if (registerMutation.isError && registerMutation.error) {
      const errorMessage =
        (registerMutation.error as any)?.response?.data?.error ||
        "Registration failed. Please try again.";
      setToastStatus("error");
      setToastMessage(errorMessage);
    }
  }, [registerMutation.isError, registerMutation.error]);

  // Resend verification success
  useEffect(() => {
    if (resendMutation.isSuccess) {
      setToastStatus("success");
      setToastMessage("Verification email sent! Please check your inbox.");
      setCooldown(60); // 60 seconds cooldown
    }
  }, [resendMutation.isSuccess]);

  // Resend verification error
  useEffect(() => {
    if (resendMutation.isError && resendMutation.error) {
      const errorMessage =
        (resendMutation.error as any)?.response?.data?.error ||
        "Failed to resend email. Please try again.";
      setToastStatus("error");
      setToastMessage(errorMessage);
    }
  }, [resendMutation.isError, resendMutation.error]);

  const onSubmit = (data: RegisterFormData) => {
    setRegisteredEmail(data.email);
    registerMutation.register({
      full_name: data.full_name,
      email: data.email,
      password: data.password,
    });
  };

  const handleResendVerification = () => {
    if (cooldown > 0) return;
    resendMutation.resendVerification({ email: registeredEmail });
  };

  const handleToastClose = () => {
    setToastStatus(null);
    setToastMessage(null);
  };

  // Show verification message after successful registration
  if (showVerification) {
    return (
      <div className="w-full max-w-md mx-auto">
        {/* Success Icon */}
        <div className="text-center mb-8">
          <div className="w-20 h-20 bg-indigo-100 dark:bg-indigo-900/30 rounded-full flex items-center justify-center mx-auto mb-4">
            <Mail className="w-10 h-10 text-indigo-500" />
          </div>
          <h1 className="text-2xl font-bold text-zinc-900 dark:text-white mb-2">
            Check Your Email
          </h1>
          <p className="text-sm text-zinc-600 dark:text-zinc-400 mb-4">
            We've sent a verification link to
            <br />
            <span className="font-semibold text-zinc-900 dark:text-white">
              {registeredEmail}
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
          <h3 className="text-sm font-semibold text-indigo-900 dark:text-indigo-300 mb-2">
            Next Steps:
          </h3>
          <ol className="text-sm text-indigo-700 dark:text-indigo-400 space-y-1 list-decimal list-inside">
            <li>Open your email inbox</li>
            <li>Click the verification link we sent</li>
            <li>You'll be redirected to login</li>
          </ol>
        </div>

        {/* Resend Button */}
        <button
          onClick={handleResendVerification}
          disabled={cooldown > 0 || resendMutation.isPending}
          className="w-full border border-zinc-200 dark:border-zinc-800 text-zinc-700 dark:text-zinc-200 font-semibold py-2.5 rounded-xl hover:bg-zinc-50 dark:hover:bg-zinc-800 transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed mb-4"
        >
          {resendMutation.isPending
            ? "Sending..."
            : cooldown > 0
              ? `Resend in ${cooldown}s`
              : "Resend Verification Email"}
        </button>

        {/* Back to Login */}
        <div className="text-center">
          <Link
            to="/auth/login"
            aria-label="sign in"
            className="text-sm text-indigo-600 dark:text-indigo-400 hover:text-indigo-700 dark:hover:text-indigo-300 font-medium transition-colors flex items-center justify-center gap-1"
          >
            <ChevronDown className="w-4 h-4 rotate-90" /> Back to login
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="w-full max-w-md mx-auto">
      {/* Header */}
      <div className="text-center mb-8">
        <div className="w-16 h-16 bg-linear-to-br from-indigo-500 to-purple-600 rounded-full flex items-center justify-center mx-auto mb-4">
          <User className="w-8 h-8 text-white" />
        </div>
        <h1 className="text-2xl font-bold text-zinc-900 dark:text-white mb-2">
          Create Account
        </h1>
        <p className="text-sm text-zinc-600 dark:text-zinc-400">
          Sign up to get started
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
          label="Full Name"
          type="text"
          name="full_name"
          placeholder="Enter your full name"
          register={register}
          error={errors.full_name?.message}
          disabled={registerMutation.isPending}
        />

        <AuthInput
          label="Email Address"
          type="email"
          name="email"
          placeholder="Enter your email"
          register={register}
          error={errors.email?.message}
          disabled={registerMutation.isPending}
        />

        <AuthInput
          label="Password"
          type="password"
          name="password"
          placeholder="Create a strong password"
          register={register}
          error={errors.password?.message}
          disabled={registerMutation.isPending}
        />

        <AuthInput
          label="Confirm Password"
          type="password"
          name="confirmPassword"
          placeholder="Confirm your password"
          register={register}
          error={errors.confirmPassword?.message}
          disabled={registerMutation.isPending}
        />

        <PasswordStrength password={password || ""} />

        <div className="mt-8">
          <SubmitButton loading={registerMutation.isPending} type="submit">
            {registerMutation.isPending
              ? "Creating account..."
              : "Create Account"}
          </SubmitButton>
        </div>

        {/* Divider */}
        <div className="relative my-6">
          <div className="absolute inset-0 flex items-center">
            <div className="w-full border-t border-zinc-200 dark:border-zinc-800"></div>
          </div>
          <div className="relative flex justify-center text-sm">
            <span className="px-4 bg-white dark:bg-zinc-900 text-zinc-500 dark:text-zinc-400 rounded-full">
              Already have an account?
            </span>
          </div>
        </div>

        {/* Login Link */}
        <Link aria-label="sign in" to="/auth/login">
          <button
            type="button"
            className="w-full border border-zinc-200 dark:border-zinc-800 text-zinc-700 dark:text-zinc-200 font-semibold py-2.5 rounded-xl hover:bg-zinc-50 dark:hover:bg-zinc-800 transition-all duration-200"
          >
            Sign In
          </button>
        </Link>
      </form>

      {/* Footer */}
      <p className="text-center text-xs text-zinc-500 dark:text-zinc-500 mt-6">
        By creating an account, you agree to our{" "}
        <Link
          to="/legal/terms-of-service"
          aria-label="terms of service"
          target="_blank"
          className="text-indigo-600 dark:text-indigo-400 hover:underline"
        >
          Terms of Service
        </Link>{" "}
        and{" "}
        <Link
          to="https://upnixo.com/legal/privacy-policy"
          aria-label="privacy policy"
          target="_blank"
          className="text-indigo-600 dark:text-indigo-400 hover:underline"
        >
          Privacy Policy
        </Link>
      </p>
    </div>
  );
};

export default Register;
