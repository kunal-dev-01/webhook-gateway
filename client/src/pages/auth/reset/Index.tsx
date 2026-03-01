import { useEffect, useState } from "react";
import { useForm, useWatch } from "react-hook-form";
import { useSearchParams } from "react-router-dom";
import { zodResolver } from "@hookform/resolvers/zod";
import { toast } from "sonner"; 
import { resetPasswordSchema, type ResetPasswordFormData } from "@/schemas/auth.schema";
import { useResetPassword } from "@/hooks/useAuth";

// Import Split Components
import ResetPasswordForm from "./ResetPasswordForm";
import SuccessScreen from "./SuccessScreen";
import InvalidTokenScreen from "./InvalidTokenScreen";

const ResetPassword = () => {
  const [searchParams] = useSearchParams();
  
  const [token, setToken] = useState<string | null>(null);
  const [isTokenValid, setIsTokenValid] = useState(true);
  const [countdown, setCountdown] = useState(5);

  const form = useForm<ResetPasswordFormData>({
    resolver: zodResolver(resetPasswordSchema),
    defaultValues: {
      password: "",
      confirmPassword: "",
    },
  });

  // Watch password for strength meter
  const password = useWatch({ control: form.control, name: "password" });

  const { resetPassword, isPending, isSuccess, isError, error } = useResetPassword();

  // Extract token from URL on mount
  useEffect(() => {
    const resetToken = searchParams.get("token");

    if (!resetToken) {
      setIsTokenValid(false);
      toast.error("Invalid or missing reset token. Please request a new password reset link.");
    } else {
      setToken(resetToken);
    }
  }, [searchParams]);

  // Success handling with countdown
  useEffect(() => {
    if (isSuccess && countdown > 0) {
      const timer = setTimeout(() => setCountdown(countdown - 1), 500);
      return () => clearTimeout(timer);
    }

    if (isSuccess && countdown === 0) {
      window.location.href = "/auth/login";
    }
  }, [isSuccess, countdown]);

  // Initial success state
  useEffect(() => {
    if (isSuccess) {
      toast.success("Password reset successful! Redirecting to login...");
    }
  }, [isSuccess]);

  // Error handling
  useEffect(() => {
    if (isError && error) {
      const errorMessage =
        (error as any)?.response?.data?.error ||
        "Password reset failed. The link may be invalid or expired.";
      toast.error(errorMessage);
    }
  }, [isError, error]);

  const onSubmit = (data: ResetPasswordFormData) => {
    if (!token) return;

    resetPassword({
      token,
      new_password: data.password,
    });
  };

  // Render Logic
  if (!isSuccess) {
    return <SuccessScreen countdown={countdown} />;
  }

  if (!isTokenValid || !token) {
    return <InvalidTokenScreen />;
  }

  return (
    <ResetPasswordForm
      form={form}
      onSubmit={onSubmit}
      isPending={isPending}
      password={password || ""}
    />
  );
};

export default ResetPassword;