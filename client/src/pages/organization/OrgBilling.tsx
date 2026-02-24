import { useState } from "react";
import { toast } from "sonner";
import { 
  useListPlans, 
  useGetSubscription, 
  useGetUsage, 
  useListPayments,
  useCreateCheckout 
} from "@/hooks/useBilling"; 

import { Loader } from "@/common/Loader";
import { Switch } from "@/common/Switch"; 
import type { BillingPeriod } from "@/types/org.types";
import { CircleCheck } from "@/svg/Common";
import { Bolt, FileText } from "@/svg/More";

export const OrgBilling = ({ orgId }: { orgId: string }) => {
  // 1. Fetch Data specific to this Org
  const { plans, isLoading: isPlansLoading } = useListPlans();
  const { subscription, plan: currentPlan, customerPortalUrl, isLoading: isSubLoading } = useGetSubscription(orgId);
  const { usage, isLoading: isUsageLoading } = useGetUsage(orgId);
  const { payments, isLoading: isPaymentsLoading, total: totalPayments } = useListPayments(orgId, { limit: 5 });

  // 2. Local State
  const [billingPeriod, setBillingPeriod] = useState<BillingPeriod>("monthly");
  const [showHistory, setShowHistory] = useState(false);

  // Loading State
  if (isPlansLoading || isSubLoading || isUsageLoading) {
    return (
      <div className="flex justify-center items-center py-20">
        <Loader className="w-8 h-8 text-indigo-500" />
      </div>
    );
  }

  const currentTier = currentPlan?.tier || "developer";

  return (
    <div className="space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-500">
      
      {/* SECTION 1: Current Plan Overview */}
      <section className="bg-white dark:bg-zinc-900/70 border border-zinc-200 dark:border-zinc-800 rounded-4xl p-6 md:p-8 relative overflow-hidden">
          <div className="flex flex-col lg:flex-row gap-8 lg:gap-12">
              
              {/* Left: Subscription Status */}
              <div className="lg:w-1/3 flex flex-col justify-between">
                  <div>
                      <div className="flex items-center gap-2 mb-2">
                          <span className={`px-2.5 py-0.5 rounded-full text-xs font-bold uppercase tracking-wide border ${
                            subscription?.status === 'active' 
                            ? 'bg-green-100 text-green-700 border-green-200 dark:bg-green-500/10 dark:text-green-400 dark:border-green-500/20'
                            : 'bg-indigo-50 text-indigo-700 border-indigo-200 dark:bg-indigo-500/10 dark:text-indigo-300 dark:border-indigo-500/20'
                          }`}>
                              {subscription?.status || "Free Tier"}
                          </span>
                      </div>
                      <h2 className="text-3xl font-bold text-zinc-900 dark:text-white capitalize mb-2">
                          {currentTier} Plan
                      </h2>
                      <p className="text-zinc-500 dark:text-zinc-400 text-sm leading-relaxed">
                          {currentTier === 'developer' 
                              ? "You are currently on the free Developer plan." 
                              : `Your plan renews on ${new Date(subscription?.current_period_end || "").toLocaleDateString()}.`
                          }
                      </p>
                  </div>

                  {/* Manage Button (Stripe/LemonSqueezy Portal) */}
                  {currentTier !== 'developer' && customerPortalUrl && (
                      <div className="mt-6">
                          <a 
                              href={customerPortalUrl}
                              target="_blank"
                              rel="noreferrer"
                              className="inline-flex items-center justify-center px-5 py-2.5 text-sm font-medium text-zinc-700 dark:text-zinc-200 bg-white dark:bg-zinc-800 border border-zinc-200 dark:border-zinc-700 rounded-xl hover:bg-zinc-50 dark:hover:bg-zinc-700 transition-colors shadow-sm"
                          >
                              Manage Subscription
                          </a>
                      </div>
                  )}
              </div>

              {/* Right: Usage Mini-Dashboard */}
              <div className="lg:w-2/3 w-full border-t lg:border-t-0 lg:border-l border-zinc-100 dark:border-zinc-800 pt-6 lg:pt-0 lg:pl-10">
                  <h3 className="text-sm font-bold text-zinc-900 dark:text-white uppercase tracking-wider mb-4 flex items-center gap-2">
                    <Bolt className="w-4 h-4 text-indigo-500" /> Usage Overview
                  </h3>
                  
                  {usage && (
                    <div className="space-y-5">
                      {/* Events Bar */}
                      <div>
                        <div className="flex justify-between items-center mb-1.5">
                          <span className="text-sm font-medium text-zinc-600 dark:text-zinc-400">Events</span>
                          <span className="text-xs font-mono text-zinc-500">{usage.monthly_events_used.toLocaleString()} / {usage.monthly_events_limit.toLocaleString()}</span>
                        </div>
                        <div className="h-2 w-full bg-zinc-100 dark:bg-zinc-800 rounded-full overflow-hidden">
                          <div className="h-full bg-indigo-500 rounded-full" style={{ width: `${Math.min(usage.usage_percentage, 100)}%` }}></div>
                        </div>
                      </div>
                      
                      {/* Stat Grid */}
                      <div className="grid grid-cols-2 gap-4 mt-4">
                         <div className="p-3 bg-zinc-50 dark:bg-zinc-800/50 rounded-xl border border-zinc-100 dark:border-zinc-800">
                            <div className="text-xs text-zinc-500">Retention</div>
                            <div className="font-bold text-zinc-900 dark:text-white">{usage.retention_days} Days</div>
                         </div>
                         <div className="p-3 bg-zinc-50 dark:bg-zinc-800/50 rounded-xl border border-zinc-100 dark:border-zinc-800">
                            <div className="text-xs text-zinc-500">Seats Used</div>
                            <div className="font-bold text-zinc-900 dark:text-white">1 / {usage.team_seats_limit}</div>
                         </div>
                      </div>
                    </div>
                  )}
              </div>
          </div>
      </section>

      {/* SECTION 2: Change Plan */}
      <section className="space-y-6">
          <div className="flex flex-col sm:flex-row justify-between items-end sm:items-center gap-4">
             <div>
                <h3 className="text-xl font-bold text-zinc-900 dark:text-white">Available Plans</h3>
                <p className="text-sm text-zinc-500">Upgrade your workspace to unlock higher limits.</p>
             </div>
             
             {/* Billing Toggle */}
             <div className="flex items-center gap-3 bg-zinc-100 dark:bg-zinc-800/50 p-1.5 rounded-xl border border-zinc-200 dark:border-zinc-800">
                <span className={`text-sm font-medium px-3 py-1.5 rounded-lg cursor-pointer transition-colors ${billingPeriod === 'monthly' ? 'bg-white dark:bg-zinc-800 text-zinc-900 dark:text-white shadow-sm' : 'text-zinc-500'}`} onClick={() => setBillingPeriod('monthly')}>
                  Monthly
                </span>
                <Switch 
                  checked={billingPeriod === 'annual'} 
                  onChange={(checked) => setBillingPeriod(checked ? 'annual' : 'monthly')}
                />
                <span className={`text-sm font-medium px-3 py-1.5 rounded-lg cursor-pointer transition-colors ${billingPeriod === 'annual' ? 'bg-white dark:bg-zinc-800 text-zinc-900 dark:text-white shadow-sm' : 'text-zinc-500'}`} onClick={() => setBillingPeriod('annual')}>
                  Yearly <span className="text-[10px] text-green-600 bg-green-100 dark:bg-green-900/30 px-1 rounded ml-1">-20%</span>
                </span>
             </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            {plans.map((plan) => (
              <OrgPlanCard 
                key={plan.id}
                plan={plan}
                currentTier={currentTier}
                selectedPeriod={billingPeriod}
                orgId={orgId}
              />
            ))}
          </div>
      </section>

      {/* SECTION 3: Invoice History (Accordion Style) */}
      <section className="bg-white dark:bg-zinc-900/70 border border-zinc-200 dark:border-zinc-800 rounded-4xl overflow-hidden">
           <button 
             className="w-full flex items-center justify-between p-6 hover:bg-zinc-50 dark:hover:bg-zinc-800/30 transition-colors text-left"
             onClick={() => setShowHistory(!showHistory)}
           >
              <div className="flex items-center gap-4">
                <div className="p-2.5 bg-zinc-100 dark:bg-zinc-800 rounded-xl">
                  <FileText className="w-5 h-5 text-zinc-500" />
                </div>
                <div>
                  <h3 className="text-lg font-bold text-zinc-900 dark:text-white">Invoice History</h3>
                  <p className="text-sm text-zinc-500">View past payments and download receipts.</p>
                </div>
              </div>
              
              <div className={`transition-transform duration-300 ${showHistory ? 'rotate-180' : ''}`}>
                 <svg width="20" height="20" viewBox="0 0 20 20" fill="none" className="text-zinc-400"><path d="M5 7.5L10 12.5L15 7.5" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/></svg>
              </div>
           </button>

           {showHistory && (
             <div className="px-6 pb-6 pt-0 border-t border-zinc-100 dark:border-zinc-800">
                {isPaymentsLoading ? (
                  <div className="py-4 text-center text-sm text-zinc-500">Loading history...</div>
                ) : !payments || totalPayments === 0 ? (
                  <div className="py-4 text-center text-sm text-zinc-500">No payment history found.</div>
                ) : (
                  <div className="mt-4 overflow-x-auto">
                    <table className="w-full text-left">
                      <thead className="text-xs text-zinc-400 uppercase font-medium">
                        <tr>
                          <th className="pb-3 pl-2">Date</th>
                          <th className="pb-3">Amount</th>
                          <th className="pb-3">Status</th>
                          <th className="pb-3 text-right pr-2">Invoice</th>
                        </tr>
                      </thead>
                      <tbody className="text-sm divide-y divide-zinc-100 dark:divide-zinc-800">
                        {payments.map((p) => (
                          <tr key={p.id}>
                            <td className="py-3 pl-2 text-zinc-600 dark:text-zinc-300">{new Date(p.created_at).toLocaleDateString()}</td>
                            <td className="py-3 font-medium text-zinc-900 dark:text-white">{p.currency.toUpperCase()} {(p.amount / 100).toFixed(2)}</td>
                            <td className="py-3"><span className="bg-green-100 text-green-700 px-2 py-0.5 rounded text-xs font-bold uppercase">{p.status}</span></td>
                            <td className="py-3 text-right pr-2">
                              {p.invoice_url && <a href={p.invoice_url} target="_blank" className="text-indigo-600 hover:underline">Download</a>}
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                )}
             </div>
           )}
      </section>

    </div>
  );
};

// Internal Sub-component for clean file
const OrgPlanCard = ({ plan, currentTier, selectedPeriod, orgId }: any) => {
  const { createCheckout, isPending } = useCreateCheckout(orgId);
  if (plan.billing_period !== selectedPeriod) return null;
  
  const isCurrent = plan.tier === currentTier;
  const isFree = plan.tier === 'developer';

  const handleAction = () => {
    if(isCurrent) return;
    createCheckout({ plan_tier: plan.tier, billing_period: plan.billing_period }, {
      onError: () => toast.error("Could not start checkout")
    });
  }

  return (
    <div className={`flex flex-col p-6 rounded-3xl border transition-all ${isCurrent ? 'border-indigo-500 ring-1 ring-indigo-500 bg-indigo-50/10' : 'border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900/40'}`}>
       <div className="mb-4">
         <h3 className="font-bold text-lg text-zinc-900 dark:text-white capitalize">{plan.name}</h3>
         <div className="flex items-baseline gap-1 mt-2">
            <span className="text-3xl font-bold text-zinc-900 dark:text-white">{isFree ? "Free" : plan.formatted_price}</span>
            {!isFree && <span className="text-zinc-500 text-sm">/{selectedPeriod === 'monthly' ? 'mo' : 'yr'}</span>}
         </div>
       </div>
       <ul className="space-y-2 mb-8 flex-1">
          <li className="flex gap-2 text-sm text-zinc-600 dark:text-zinc-400"><CircleCheck className="w-4 h-4 text-indigo-500 shrink-0"/> {plan.monthly_events_limit.toLocaleString()} Events</li>
          <li className="flex gap-2 text-sm text-zinc-600 dark:text-zinc-400"><CircleCheck className="w-4 h-4 text-indigo-500 shrink-0"/> {plan.team_seats_limit} Team Seats</li>
          <li className="flex gap-2 text-sm text-zinc-600 dark:text-zinc-400"><CircleCheck className="w-4 h-4 text-indigo-500 shrink-0"/> {plan.retention_days} Days Retention</li>
       </ul>
       <button 
         onClick={handleAction}
         disabled={isCurrent || isPending}
         className={`w-full py-2.5 rounded-xl font-medium text-sm transition-all flex justify-center items-center ${isCurrent ? 'bg-zinc-100 text-zinc-400 cursor-default' : 'bg-zinc-900 text-white hover:bg-zinc-800'}`}
       >
         {isPending ? <Loader className="w-4 h-4"/> : isCurrent ? "Current Plan" : "Upgrade"}
       </button>
    </div>
  )
}