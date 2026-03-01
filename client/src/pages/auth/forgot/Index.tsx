import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { toast } from "sonner";
import { forgotPasswordSchema, type ForgotPasswordFormData } from "@/schemas/auth.schema";
import { useForgotPassword } from "@/hooks/useAuth";

// Import Split Components
import SuccessScreen from "./SuccessScreen";
import ForgotPasswordForm from "./ForgotPasswordForm";

const ForgotPassword = () => {
  const [emailSent, setEmailSent] = useState(false);
  const [submittedEmail, setSubmittedEmail] = useState("");
  const [cooldown, setCooldown] = useState(0);

  const form = useForm<ForgotPasswordFormData>({
    resolver: zodResolver(forgotPasswordSchema),
    defaultValues: { email: "" },
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
      toast.success("Password reset link sent! Please check your email.");
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
      toast.error(errorMessage);
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
    form.reset();
    resetMutation();
  };

  if (emailSent) {
    return (
      <SuccessScreen
        submittedEmail={submittedEmail}
        cooldown={cooldown}
        isPending={isPending}
        handleResend={handleResend}
        handleUseAnotherEmail={handleUseAnotherEmail}
      />
    );
  }

  return (
    <ForgotPasswordForm
      form={form}
      onSubmit={onSubmit}
      isPending={isPending}
    />
  );
};

export default ForgotPassword;