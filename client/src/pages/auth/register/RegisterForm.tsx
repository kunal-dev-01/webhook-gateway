import { type UseFormReturn } from "react-hook-form";
import { Link } from "react-router-dom";
import AuthInput from "@/common/AuthInput";
import SubmitButton from "@/common/SubmitButton";
import PasswordStrength from "@/common/PasswordStrength";
import { UserPlus } from "@/svg/Auth";
import { type RegisterFormData } from "@/schemas/auth.schema";
import { PLANS } from "./plans";

type Props = {
  form: UseFormReturn<RegisterFormData>;
  onSubmit: (data: RegisterFormData) => void;
  isPending: boolean;
  password: string;
  selectedPlanId: string;
  billingPeriod: string;
  onOpenPlanModal: () => void;
};

export default function RegisterForm({
  form, onSubmit, isPending, password, selectedPlanId, billingPeriod, onOpenPlanModal
}: Props) {
  const selectedPlanDetails = PLANS.find(p => p.id === selectedPlanId) || PLANS[0];
  
  // Format price display
  let priceDisplay = "Free forever";
  if (selectedPlanDetails.price.monthly > 0) {
    const amount = billingPeriod === 'annual' ? selectedPlanDetails.price.annual : selectedPlanDetails.price.monthly;
    const period = billingPeriod === 'annual' ? '/yr' : '/mo';
    priceDisplay = `$${amount}${period}`;
  }

  return (
    <div className="w-full max-w-md mx-auto animate-in fade-in slide-in-from-bottom-4 duration-300">
      {/* Header */}
      <div className="text-center mb-6">
        <div className="w-16 h-16 dark:bg-zinc-900 bg-zinc-100 border border-zinc-200 dark:border-zinc-800 rounded-full flex items-center justify-center mx-auto mb-4">
          <UserPlus className="w-10 h-10 text-zinc-900 dark:text-zinc-100" />
        </div>
        <h1 className="text-2xl font-bold text-zinc-900 dark:text-white mb-2">Create Account</h1>
        <p className="text-sm text-zinc-600 dark:text-zinc-400">Join us to get started</p>
      </div>

      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
        
        {/* COMPACT PLAN SUMMARY CARD */}
        <div className="bg-zinc-50 dark:bg-zinc-900/50 rounded-xl p-4 border border-zinc-200 dark:border-zinc-800 flex items-center justify-between">
          <div>
            <p className="text-xs text-zinc-500 dark:text-zinc-400 font-medium mb-0.5">Selected Plan</p>
            <div className="flex items-baseline gap-2">
              <span className="font-bold text-zinc-900 dark:text-white">{selectedPlanDetails.name}</span>
              <span className="text-xs text-zinc-500 dark:text-zinc-400">• {priceDisplay}</span>
            </div>
          </div>
          <button
            type="button"
            onClick={onOpenPlanModal}
            className="text-xs font-semibold px-3 py-1.5 bg-white dark:bg-zinc-800 border border-zinc-200 dark:border-zinc-700 rounded-lg hover:bg-zinc-50 dark:hover:bg-zinc-700 transition-colors"
          >
            Change
          </button>
        </div>

        {/* Inputs */}
        <AuthInput label="Full Name" type="text" name="full_name" placeholder="Enter your full name" register={form.register} error={form.formState.errors.full_name?.message} disabled={isPending} />
        <AuthInput label="Email Address" type="email" name="email" placeholder="Enter your email" register={form.register} error={form.formState.errors.email?.message} disabled={isPending} />
        <AuthInput label="Password" type="password" name="password" placeholder="Create a strong password" register={form.register} error={form.formState.errors.password?.message} disabled={isPending} />
        <AuthInput label="Confirm Password" type="password" name="confirmPassword" placeholder="Confirm your password" register={form.register} error={form.formState.errors.confirmPassword?.message} disabled={isPending} />
        
        <PasswordStrength password={password} />

        <div className="mt-6">
          <SubmitButton loading={isPending} type="submit">
            {isPending ? "Creating account..." : "Create Account"}
          </SubmitButton>
        </div>

        {/* Footer Links */}
        <div className="relative my-6">
          <div className="absolute inset-0 flex items-center">
            <div className="w-full border-t border-zinc-200 dark:border-zinc-800"></div>
          </div>
          <div className="relative flex justify-center text-sm">
            <span className="px-4 bg-white dark:bg-zinc-950 text-zinc-500 dark:text-zinc-400 rounded-full">Already have an account?</span>
          </div>
        </div>

        <Link aria-label="sign in" to="/auth/login">
          <button type="button" className="w-full border border-zinc-200 dark:border-zinc-800 text-zinc-700 dark:text-zinc-200 font-semibold py-2.5 rounded-xl hover:bg-zinc-50 dark:hover:bg-zinc-800 transition-all duration-200">
            Sign In
          </button>
        </Link>
      </form>

      <p className="text-center text-xs text-zinc-500 dark:text-zinc-500 mt-6">
        By creating an account, you agree to our{" "}
        <Link to="/legal/terms" className="text-indigo-600 dark:text-indigo-400 hover:underline">Terms</Link> and{" "}
        <Link to="/legal/privacy" className="text-indigo-600 dark:text-indigo-400 hover:underline">Privacy Policy</Link>
      </p>
    </div>
  );
}