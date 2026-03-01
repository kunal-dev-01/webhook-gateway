import { useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { useForm, useWatch } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { toast } from "sonner";
import { registerSchema, type RegisterFormData } from "@/schemas/auth.schema";
import { useRegister, useResendVerification } from "@/hooks/useAuth";

// Components
import RegisterForm from "./RegisterForm";
import VerificationScreen from "./VerificationScreen";
import PlanSelectionModal from "./PlanSelectionModal";

const Register = () => {
  const [searchParams] = useSearchParams();
  const [showVerification, setShowVerification] = useState(false);
  const [registeredEmail, setRegisteredEmail] = useState("");
  const [cooldown, setCooldown] = useState(0);
  
  // Modal State for Plan Selection
  const [isPlanModalOpen, setIsPlanModalOpen] = useState(false);

  const form = useForm<RegisterFormData>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      full_name: "",
      email: "",
      password: "",
      confirmPassword: "",
      plan_tier: (searchParams.get("plan_tier") as any) || "developer",
      billing_period: (searchParams.get("billing_period") as any) || "monthly",
    },
  });

  const { control, setValue } = form;

  // Watchers
  const password = useWatch({ control, name: "password" });
  const selectedPlanId = useWatch({ control, name: "plan_tier" });
  const billingPeriod = useWatch({ control, name: "billing_period" });

  const registerMutation = useRegister();
  const resendMutation = useResendVerification();

  // Sync URL params
  useEffect(() => {
    const planParam = searchParams.get("plan_tier");
    const billingParam = searchParams.get("billing_period");

    if (planParam && ["developer", "starter", "growth", "scale"].includes(planParam)) {
      setValue("plan_tier", planParam as any);
    }
    if (billingParam && ["monthly", "annual"].includes(billingParam)) {
      setValue("billing_period", billingParam as any);
    }
  }, [searchParams, setValue]);

  // Cooldown timer
  useEffect(() => {
    if (cooldown > 0) {
      const timer = setTimeout(() => setCooldown(cooldown - 1), 1000);
      return () => clearTimeout(timer);
    }
  }, [cooldown]);

  // Mutations Handlers
  useEffect(() => {
    if (registerMutation.isSuccess) {
      toast.success("Registration successful! Verify your email.");
      setShowVerification(true);
    }
    if (registerMutation.isError && registerMutation.error) {
      const errorMessage =
        (registerMutation.error as any)?.response?.data?.error ||
        "Registration failed. Please try again.";
      toast.error(errorMessage);
    }
  }, [registerMutation.isSuccess, registerMutation.isError, registerMutation.error]);

  useEffect(() => {
    if (resendMutation.isSuccess) {
      toast.success("Verification email sent!");
      setCooldown(60);
    }
    if (resendMutation.isError) {
      toast.error("Failed to resend email.");
    }
  }, [resendMutation.isSuccess, resendMutation.isError]);

  const onSubmit = (data: RegisterFormData) => {
    setRegisteredEmail(data.email);
    registerMutation.register(data);
  };

  const handleResendVerification = () => {
    if (cooldown > 0) return;
    resendMutation.resendVerification({ email: registeredEmail });
  };

  if (showVerification) {
    return (
      <VerificationScreen
        registeredEmail={registeredEmail}
        cooldown={cooldown}
        isPending={resendMutation.isPending}
        onResend={handleResendVerification}
      />
    );
  }

  return (
    <>
      <RegisterForm
        form={form}
        onSubmit={onSubmit}
        isPending={registerMutation.isPending}
        password={password || ""}
        selectedPlanId={selectedPlanId}
        billingPeriod={billingPeriod}
        onOpenPlanModal={() => setIsPlanModalOpen(true)}
      />

      {/* Professional Blur Modal */}
      <PlanSelectionModal
        isOpen={isPlanModalOpen}
        onClose={() => setIsPlanModalOpen(false)}
        selectedPlanId={selectedPlanId}
        billingPeriod={billingPeriod}
        setPlanValue={(id) => setValue("plan_tier", id)}
        setBillingValue={(period) => setValue("billing_period", period)}
      />
    </>
  );
};

export default Register;