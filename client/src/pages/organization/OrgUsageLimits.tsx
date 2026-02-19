import { type Organization } from "@/types/org.types";
import { Monitor, DotPattern } from "@/svg/Auth"; // Reusing SVGs

export const OrgUsageLimits = ({ org }: { org: Organization }) => {
  const usagePercent = org.monthly_events_limit > 0 
    ? Math.min((org.monthly_events_used / org.monthly_events_limit) * 100, 100) 
    : 0;

  return (
    <section className="bg-white dark:bg-zinc-900/70 border border-zinc-200 dark:border-zinc-800 rounded-4xl p-6 md:p-8 space-y-6 relative overflow-hidden">
      <div className="absolute top-0 right-0 opacity-20 dark:opacity-10 pointer-events-none">
        <DotPattern />
      </div>

      <div className="flex items-start gap-4 mb-6 relative z-10">
        <div className="p-3 bg-indigo-50 dark:bg-indigo-500/10 rounded-xl hidden sm:block">
          <Monitor className="w-6 h-6 text-indigo-500" />
        </div>
        <div>
          <h2 className="text-xl font-bold text-zinc-900 dark:text-white">Plan & Usage Limits</h2>
          <p className="text-sm text-zinc-500">
            Current subscription is <span className="font-semibold capitalize text-indigo-500">{org.plan_tier} ({org.billing_period})</span>
          </p>
        </div>
      </div>

      <div className="w-full h-px bg-zinc-100 dark:bg-zinc-800/80 relative z-10 mb-6"></div>

      <div className="space-y-6 relative z-10">
        {/* Events Progress */}
        <div>
          <div className="flex justify-between items-end mb-2">
            <p className="text-sm text-zinc-700 dark:text-zinc-300 font-medium">Monthly Webhook Events</p>
            <div className="text-right">
              <span className="text-zinc-900 dark:text-white font-bold text-sm">
                {org.monthly_events_used.toLocaleString()}
              </span>
              <span className="text-zinc-500 text-xs ml-1">
                / {org.monthly_events_limit.toLocaleString()}
              </span>
            </div>
          </div>
          <div className="w-full bg-zinc-100 dark:bg-zinc-800/50 rounded-full h-3 overflow-hidden border border-zinc-200 dark:border-zinc-800/50">
            <div
              className={`h-full rounded-full transition-all duration-500 ${
                usagePercent > 85 ? "bg-red-500" : "bg-indigo-500"
              }`}
              style={{ width: `${usagePercent}%` }}
            ></div>
          </div>
        </div>

        {/* Technical Specs Grid */}
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
          <StatBox label="Rate Limit" value={`${org.rate_limit_per_second}/s`} />
          <StatBox label="Retention" value={`${org.retention_days} Days`} />
          <StatBox label="Projects Limit" value={org.projects_limit.toString()} />
          <StatBox label="Team Seats" value={org.team_seats_limit.toString()} />
        </div>
      </div>
    </section>
  );
};

// Reusable micro-component for stats
const StatBox = ({ label, value }: { label: string; value: string }) => (
  <div className="p-4 rounded-2xl border border-zinc-200 dark:border-zinc-800 bg-zinc-50/50 dark:bg-zinc-900/30 text-center">
    <p className="text-[11px] text-zinc-500 uppercase tracking-wider font-semibold mb-1">{label}</p>
    <p className="font-bold text-zinc-900 dark:text-white text-lg">{value}</p>
  </div>
);