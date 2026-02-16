import { useEffect, useState } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { useForm, useWatch } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { toast } from "sonner";
import { registerSchema, type RegisterFormData } from "@/schemas/auth.schema";
import { useRegister, useResendVerification } from "@/hooks/useAuth";
import AuthInput from "@/common/AuthInput";
import SubmitButton from "@/common/SubmitButton";
import PasswordStrength from "@/common/PasswordStrength";
import { User } from "@/svg/Auth";
import { ChevronDown, Mail } from "@/svg/Common";

// --- Plan Configuration (Matches your SQL) ---
// src/pages/Register.tsx

const PLANS = [
  {
    id: "developer",
    name: "Developer",
    price: { monthly: 0, annual: 0 },
    description: "For testing & hobbies",
    popular: false, // Added this
  },
  {
    id: "starter",
    name: "Starter",
    price: { monthly: 19, annual: 190 },
    description: "For solo founders",
    popular: false, // Added this
  },
  {
    id: "growth",
    name: "Growth",
    price: { monthly: 49, annual: 490 },
    description: "For growing startups",
    popular: true,
  },
  {
    id: "scale",
    name: "Scale",
    price: { monthly: 149, annual: 1490 },
    description: "For high volume",
    popular: false, // Added this
  },
] as const;

const Register = () => {
  const [searchParams] = useSearchParams();
  const [showVerification, setShowVerification] = useState(false);
  const [registeredEmail, setRegisteredEmail] = useState("");
  const [cooldown, setCooldown] = useState(0);

  // Initialize form
  const {
    register,
    handleSubmit,
    control,
    setValue,
    formState: { errors },
  } = useForm<RegisterFormData>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      full_name: "",
      email: "",
      password: "",
      confirmPassword: "",
      // Default to what's in URL or fallbacks
      plan_tier: (searchParams.get("plan_tier") as any) || "developer",
      billing_period: (searchParams.get("billing_period") as any) || "monthly",
    },
  });

  // Watchers for real-time UI updates
  const password = useWatch({ control, name: "password" });
  const selectedPlan = useWatch({ control, name: "plan_tier" });
  const billingPeriod = useWatch({ control, name: "billing_period" });

  const registerMutation = useRegister();
  const resendMutation = useResendVerification();

  // Sync URL params if they change after mount (optional, but good for UX)
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

  // Success/Error Handling
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
    registerMutation.register({
      full_name: data.full_name,
      email: data.email,
      password: data.password,
      plan_tier: data.plan_tier,
      billing_period: data.billing_period,
    });
  };

  const handleResendVerification = () => {
    if (cooldown > 0) return;
    resendMutation.resendVerification({ email: registeredEmail });
  };

  // Verification View
  if (showVerification) {
    return (
      <div className="w-full max-w-md mx-auto">
        <div className="text-center mb-8">
          <div className="w-20 h-20 bg-indigo-100 dark:bg-indigo-900/30 rounded-full flex items-center justify-center mx-auto mb-4">
            <Mail className="w-10 h-10 text-indigo-500" />
          </div>
          <h1 className="text-2xl font-bold text-zinc-900 dark:text-white mb-2">
            Check Your Email
          </h1>
          <p className="text-sm text-zinc-600 dark:text-zinc-400 mb-4">
            We've sent a verification link to <br />
            <span className="font-semibold text-zinc-900 dark:text-white">{registeredEmail}</span>
          </p>
        </div>
        
        {/* ... (Existing Verification UI) ... */}
        
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

        <div className="text-center">
           <Link to="/auth/login" className="text-sm text-indigo-600 dark:text-indigo-400 flex items-center justify-center gap-1">
            <ChevronDown className="w-4 h-4 rotate-90" /> Back to login
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="w-full max-w-md mx-auto">
      {/* Header */}
      <div className="text-center mb-6">
        <div className="w-16 h-16 bg-linear-to-br from-indigo-500 to-purple-600 rounded-full flex items-center justify-center mx-auto mb-4">
          <User className="w-8 h-8 text-white" />
        </div>
        <h1 className="text-2xl font-bold text-zinc-900 dark:text-white mb-2">
          Create Account
        </h1>
        <p className="text-sm text-zinc-600 dark:text-zinc-400">
          Setup your {selectedPlan.charAt(0).toUpperCase() + selectedPlan.slice(1)} plan
        </p>
      </div>

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        
        {/* --- PLAN SELECTOR SECTION --- */}
        <div className="bg-zinc-50 dark:bg-zinc-900/50 rounded-2xl p-1 border border-zinc-200 dark:border-zinc-800 mb-6">
          
          {/* Billing Toggle (Hidden if Developer) */}
          {selectedPlan !== 'developer' && (
            <div className="flex justify-center mb-4 pt-3">
              <div className="relative flex bg-zinc-200 dark:bg-zinc-800 p-1 rounded-lg">
                <button
                  type="button"
                  onClick={() => setValue("billing_period", "monthly")}
                  className={`relative z-10 px-4 py-1.5 text-xs font-semibold rounded-md transition-all duration-200 ${
                    billingPeriod === "monthly"
                      ? "bg-white dark:bg-zinc-700 text-zinc-900 dark:text-white shadow-sm"
                      : "text-zinc-500 dark:text-zinc-400 hover:text-zinc-700"
                  }`}
                >
                  Monthly
                </button>
                <button
                  type="button"
                  onClick={() => setValue("billing_period", "annual")}
                  className={`relative z-10 px-4 py-1.5 text-xs font-semibold rounded-md transition-all duration-200 flex items-center gap-1.5 ${
                    billingPeriod === "annual"
                      ? "bg-white dark:bg-zinc-700 text-zinc-900 dark:text-white shadow-sm"
                      : "text-zinc-500 dark:text-zinc-400 hover:text-zinc-700"
                  }`}
                >
                  Annual
                  <span className="text-[10px] bg-green-100 dark:bg-green-900/40 text-green-600 dark:text-green-400 px-1.5 py-0.5 rounded-full">
                    -17%
                  </span>
                </button>
              </div>
            </div>
          )}

          {/* Plan Cards Grid */}
          <div className="space-y-2 p-2">
            {PLANS.map((plan) => {
              const isSelected = selectedPlan === plan.id;
              
              // Calculate Price Display
              let priceDisplay = "Free";
              if (plan.price.monthly > 0) {
                 const amount = billingPeriod === 'annual' ? plan.price.annual : plan.price.monthly;
                 const period = billingPeriod === 'annual' ? '/yr' : '/mo';
                 priceDisplay = `$${amount}${period}`;
              }

              return (
                <div
                  key={plan.id}
                  onClick={() => setValue("plan_tier", plan.id as any)}
                  className={`relative cursor-pointer group rounded-xl border-2 p-3 transition-all duration-200 ${
                    isSelected
                      ? "border-indigo-600 bg-indigo-50/50 dark:bg-indigo-900/10 dark:border-indigo-500"
                      : "border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 hover:border-zinc-300 dark:hover:border-zinc-700"
                  }`}
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      {/* Radio Indicator */}
                      <div className={`w-4 h-4 rounded-full border flex items-center justify-center ${
                        isSelected ? "border-indigo-600" : "border-zinc-300 dark:border-zinc-600"
                      }`}>
                         {isSelected && <div className="w-2 h-2 bg-indigo-600 rounded-full" />}
                      </div>
                      
                      <div>
                        <div className="flex items-center gap-2">
                          <span className={`text-sm font-bold ${isSelected ? 'text-indigo-900 dark:text-indigo-100' : 'text-zinc-700 dark:text-zinc-200'}`}>
                            {plan.name}
                          </span>
                          {plan.popular && (
                            <span className="text-[10px] uppercase font-bold tracking-wider bg-indigo-100 text-indigo-600 dark:bg-indigo-500/20 dark:text-indigo-300 px-1.5 py-0.5 rounded-md">
                              Popular
                            </span>
                          )}
                        </div>
                        <p className="text-xs text-zinc-500 dark:text-zinc-400">
                          {plan.description}
                        </p>
                      </div>
                    </div>

                    <div className="text-right">
                      <span className={`text-sm font-bold ${isSelected ? 'text-indigo-700 dark:text-indigo-300' : 'text-zinc-900 dark:text-white'}`}>
                        {priceDisplay}
                      </span>
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        </div>

        {/* --- EXISTING INPUTS --- */}
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
              : `Create ${selectedPlan.charAt(0).toUpperCase() + selectedPlan.slice(1)} Account`}
          </SubmitButton>
        </div>

        {/* Divider & Footer Links (Same as before) */}
        <div className="relative my-6">
          <div className="absolute inset-0 flex items-center">
            <div className="w-full border-t border-zinc-200 dark:border-zinc-800"></div>
          </div>
          <div className="relative flex justify-center text-sm">
            <span className="px-4 bg-white dark:bg-zinc-950 text-zinc-500 dark:text-zinc-400 rounded-full">
              Already have an account?
            </span>
          </div>
        </div>

        <Link aria-label="sign in" to="/auth/login">
          <button
            type="button"
            className="w-full border border-zinc-200 dark:border-zinc-800 text-zinc-700 dark:text-zinc-200 font-semibold py-2.5 rounded-xl hover:bg-zinc-50 dark:hover:bg-zinc-800 transition-all duration-200"
          >
            Sign In
          </button>
        </Link>
      </form>

      <p className="text-center text-xs text-zinc-500 dark:text-zinc-500 mt-6">
        By creating an account, you agree to our{" "}
        <Link to="/legal/terms" className="text-indigo-600 dark:text-indigo-400 hover:underline">
          Terms
        </Link>{" "}
        and{" "}
        <Link to="/legal/privacy" className="text-indigo-600 dark:text-indigo-400 hover:underline">
          Privacy Policy
        </Link>
      </p>
    </div>
  );
};

export default Register;