import { useEffect } from "react";
import { PLANS } from "./plans";
import SubmitButton from "@/common/SubmitButton";
import { X } from "@/svg/Common";

type Props = {
  isOpen: boolean;
  onClose: () => void;
  selectedPlanId: string;
  billingPeriod: string;
  setPlanValue: (id: any) => void;
  setBillingValue: (period: any) => void;
};

export default function PlanSelectionModal({
  isOpen, onClose, selectedPlanId, billingPeriod, setPlanValue, setBillingValue
}: Props) {
  
  // Prevent body scroll when modal is open
  useEffect(() => {
    if (isOpen) document.body.style.overflow = "hidden";
    else document.body.style.overflow = "unset";
    return () => { document.body.style.overflow = "unset"; };
  }, [isOpen]);

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-zinc-900/40 backdrop-blur-md">
      
      {/* Click outside to close (Optional, overlay click) */}
      <div className="absolute inset-0" onClick={onClose} />

      <div className="relative bg-white dark:bg-zinc-900 rounded-2xl w-full max-w-md border border-zinc-200 dark:border-zinc-800 p-6">
        
        {/* Header */}
        <div className="flex justify-between items-center mb-6">
          <div>
            <h2 className="text-xl font-bold text-zinc-900 dark:text-white">Choose a Plan</h2>
            <p className="text-xs text-zinc-500 dark:text-zinc-400">Select the best plan for your needs.</p>
          </div>
          <button onClick={onClose} className="px-3 py-1 text-zinc-400 hover:text-zinc-600 dark:hover:text-zinc-200 bg-zinc-100 dark:bg-zinc-800 rounded-full transition-colors">
            <X className="w-6 h-6" />
          </button>
        </div>

        {/* Billing Toggle */}
        <div className="flex justify-center mb-6">
          <div className="flex bg-zinc-100 dark:bg-zinc-800 p-1 rounded-xl">
            <button
              onClick={() => setBillingValue("monthly")}
              className={`px-5 py-2 text-xs font-semibold rounded-lg transition-all duration-200 ${
                billingPeriod === "monthly" ? "bg-white dark:bg-zinc-700 text-zinc-900 dark:text-white shadow-sm" : "text-zinc-500 hover:text-zinc-700 dark:hover:text-zinc-300"
              }`}
            >
              Monthly
            </button>
            <button
              onClick={() => setBillingValue("annual")}
              className={`px-5 py-2 text-xs font-semibold rounded-lg transition-all duration-200 flex items-center gap-1.5 ${
                billingPeriod === "annual" ? "bg-white dark:bg-zinc-700 text-zinc-900 dark:text-white shadow-sm" : "text-zinc-500 hover:text-zinc-700 dark:hover:text-zinc-300"
              }`}
            >
              Annual <span className="text-[10px] bg-green-100 dark:bg-green-900/40 text-green-600 dark:text-green-400 px-1.5 py-0.5 rounded-full">Save 17%</span>
            </button>
          </div>
        </div>

        {/* Plan List */}
        <div className="space-y-3 mb-6 max-h-[40vh] overflow-y-auto pr-2">
          {PLANS.map((plan) => {
            const isSelected = selectedPlanId === plan.id;
            const amount = billingPeriod === 'annual' ? plan.price.annual : plan.price.monthly;
            const priceDisplay = amount > 0 ? `$${amount}` : "Free";
            const period = amount > 0 ? (billingPeriod === 'annual' ? '/yr' : '/mo') : '';

            return (
              <div
                key={plan.id}
                onClick={() => setPlanValue(plan.id)}
                className={`relative cursor-pointer rounded-xl border-2 p-4 transition-all duration-200 ${
                  isSelected
                    ? "border-indigo-600 bg-indigo-50/50 dark:bg-indigo-900/10"
                    : "border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 hover:border-zinc-300 dark:hover:border-zinc-700"
                }`}
              >
                <div className="flex justify-between items-center">
                  <div>
                    <div className="flex items-center gap-2 mb-1">
                      <span className={`font-bold ${isSelected ? 'text-indigo-900 dark:text-indigo-100' : 'text-zinc-900 dark:text-white'}`}>{plan.name}</span>
                      {plan.popular && (
                        <span className="text-[10px] uppercase font-bold tracking-wider bg-indigo-100 text-indigo-600 dark:bg-indigo-500/20 dark:text-indigo-300 px-1.5 py-0.5 rounded-md">Popular</span>
                      )}
                    </div>
                    <p className="text-xs text-zinc-500 dark:text-zinc-400">{plan.description}</p>
                  </div>
                  <div className="text-right">
                    <span className={`text-lg font-bold ${isSelected ? 'text-indigo-700 dark:text-indigo-300' : 'text-zinc-900 dark:text-white'}`}>
                      {priceDisplay}
                    </span>
                    <span className="text-xs text-zinc-500 dark:text-zinc-400">{period}</span>
                  </div>
                </div>
              </div>
            );
          })}
        </div>

        {/* Action Button */}
        <SubmitButton
          type="button"
          onClick={onClose}
        >
          Confirm Plan
        </SubmitButton>
      </div>
    </div>
  );
}