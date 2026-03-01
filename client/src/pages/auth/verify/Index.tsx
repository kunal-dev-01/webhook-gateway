import { useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { useVerifyEmail } from "@/hooks/useAuth";

// Import Split Components
import VerifyingScreen from "./VerifyingScreen";
import SuccessScreen from "./SuccessScreen";
import ErrorScreen from "./ErrorScreen";

type VerificationState = "verifying" | "success" | "error" | "invalid";

const VerifyEmail = () => {
  const [searchParams] = useSearchParams();

  const [verificationState, setVerificationState] = useState<VerificationState>("verifying");
  const [errorMessage, setErrorMessage] = useState<string>("");
  const [countdown, setCountdown] = useState(2);

  const { verifyEmail, isPending, isSuccess, isError, error } = useVerifyEmail();

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
  }, [verifyEmail, searchParams]);

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

  // Render Logic
  if (verificationState === "verifying" || isPending) {
    return <VerifyingScreen />;
  }

  if (verificationState === "success") {
    return <SuccessScreen countdown={countdown} />;
  }

  return <ErrorScreen errorMessage={errorMessage} />;
};

export default VerifyEmail;