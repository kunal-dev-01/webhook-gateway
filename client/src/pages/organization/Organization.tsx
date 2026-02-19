import { useNavigate } from "react-router-dom";
import { useListOrganizations } from "@/hooks/useOrg";
import { Loader } from "@/common/Loader";
import { ChevronDown, Error, X } from "@/svg/Common";
import { DotPattern, Globe, Users } from "@/svg/Auth";
import { type Organization } from "@/types/org.types";
import SubmitButton from "@/common/SubmitButton";

const Organization = () => {
  const { organizations, isLoading, isError } = useListOrganizations();

  // 🔥 Loader while fetching organizations
  if (isLoading) {
    return (
      <div className="flex justify-center items-center min-h-[60vh]">
        <Loader className="w-8 h-8 text-indigo-500" />
      </div>
    );
  }

  // ❌ Error state
  if (isError) {
    return (
      <div className="flex flex-col justify-center items-center gap-5 min-h-[80vh]">
        <Error className="w-20 h-20 text-red-500" />
        <h1 className="text-3xl font-bold tracking-tight text-zinc-700 dark:text-white">
          Failed to load organizations.
        </h1>
      </div>
    );
  }

  return (
    <div className="mx-auto w-full max-w-7xl py-8 md:py-12">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-8 md:mb-12">
        <div>
          <h1 className="text-3xl font-bold tracking-tight text-zinc-700 dark:text-white">
            Organizations
          </h1>
          <p className="text-zinc-500 dark:text-zinc-400 mt-2 text-lg">
            Manage your webhook gateway workspaces, team members, and usage
            limits.
          </p>
        </div>

        {/* Create Organization Button */}
        <div className="inline-block">
          <SubmitButton className="px-6">
            <X className="w-5 h-5 rotate-45" />
            Create Organization
          </SubmitButton>
        </div>
      </div>

      {/* Main Content Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8 items-start">
        {organizations.length === 0 ? (
          <div className="col-span-full bg-zinc-50/50 dark:bg-zinc-900/30 border border-zinc-200 dark:border-zinc-800 rounded-4xl p-10 text-center space-y-4">
            <Users className="w-12 h-12 text-zinc-400 mx-auto opacity-50" />
            <div>
              <h3 className="text-zinc-700 dark:text-zinc-300 font-medium text-lg">
                No Organizations Found
              </h3>
              <p className="text-sm text-zinc-500 mt-1">
                You are not part of any organization yet. Create one to get
                started.
              </p>
            </div>
          </div>
        ) : (
          organizations.map((org) => <OrgCard key={org.id} org={org} />)
        )}
      </div>
    </div>
  );
};

// ============================================================================
// ORGANIZATION CARD COMPONENT
// ============================================================================

const OrgCard = ({ org }: { org: Organization }) => {
  const navigate = useNavigate(); // 👈 Navigation hook initialize kiya

  // Calculate percentage for webhook events usage bar
  const usagePercent =
    org.monthly_events_limit > 0
      ? Math.min(
          (org.monthly_events_used / org.monthly_events_limit) * 100,
          100,
        )
      : 0;

  // Determine color based on usage limits
  const isNearLimit = usagePercent > 85;

  return (
    <section className="bg-white dark:bg-zinc-900/70 border border-zinc-200 dark:border-zinc-800 rounded-4xl p-6 md:p-8 space-y-6 relative overflow-hidden group hover:border-zinc-300 dark:hover:border-zinc-700 transition-colors">
      {/* Decorative SVG Pattern */}
      <div className="absolute top-0 right-0 opacity-20 dark:opacity-70 pointer-events-none">
        <DotPattern />
      </div>

      {/* Header Info: Name, Slug, Badges */}
      <div className="flex justify-between items-start relative z-10">
        <div className="flex gap-4 items-center">
          <div className="p-3 bg-zinc-100 dark:bg-zinc-800 rounded-xl hidden sm:block">
            <Globe className="w-6 h-6 text-indigo-500" />
          </div>
          <div>
            <h2 className="text-xl font-bold text-zinc-900 dark:text-white capitalize truncate max-w-37.5 sm:max-w-50">
              {org.name}
            </h2>
            <p className="text-sm text-zinc-500">@{org.slug}</p>
          </div>
        </div>

        {/* Badges */}
        <div className="flex flex-col sm:flex-row gap-2 items-end sm:items-center">
          <span className="px-2.5 py-1 text-[11px] font-bold uppercase tracking-wider bg-indigo-50 text-indigo-600 dark:bg-indigo-500/10 dark:text-indigo-400 rounded-full border border-indigo-100 dark:border-indigo-500/20">
            {org.plan_tier}
          </span>
          <span className="px-2.5 py-1 text-[11px] font-bold uppercase tracking-wider bg-zinc-100 text-zinc-600 dark:bg-zinc-800 dark:text-zinc-400 rounded-full border border-zinc-200 dark:border-zinc-700">
            {org.role}
          </span>
        </div>
      </div>

      <div className="w-full h-px bg-zinc-100 dark:bg-zinc-800/80"></div>

      {/* Stats & Usage Limits */}
      <div className="space-y-5 relative z-10">
        {/* Events Progress Bar */}
        <div>
          <div className="flex justify-between items-end mb-2">
            <div>
              <p className="text-sm text-zinc-600 dark:text-zinc-400 font-medium">
                Monthly Webhook Events
              </p>
            </div>
            <div className="text-right">
              <span className="text-zinc-900 dark:text-white font-bold text-sm">
                {org.monthly_events_used.toLocaleString()}
              </span>
              <span className="text-zinc-500 text-xs ml-1">
                / {org.monthly_events_limit.toLocaleString()}
              </span>
            </div>
          </div>

          <div className="w-full bg-zinc-100 dark:bg-zinc-800/50 rounded-full h-2.5 overflow-hidden border border-zinc-200 dark:border-zinc-800/50">
            <div
              className={`h-2.5 rounded-full transition-all duration-500 ${
                isNearLimit ? "bg-red-500" : "bg-indigo-500"
              }`}
              style={{ width: `${usagePercent}%` }}
            ></div>
          </div>
          {isNearLimit && (
            <p className="text-[11px] text-red-500 mt-2 font-medium">
              Approaching monthly events limit. Consider upgrading your tier.
            </p>
          )}
        </div>

        {/* Gateway Technical Limits Info */}
        <div className="grid grid-cols-2 gap-4">
          <div className="p-4 rounded-2xl border border-zinc-200 dark:border-zinc-800 bg-zinc-50/50 dark:bg-zinc-900/30">
            <p className="text-xs text-zinc-500 dark:text-zinc-400 font-medium">
              Log Retention
            </p>
            <p className="font-semibold text-zinc-900 dark:text-white mt-1 text-lg">
              {org.retention_days}{" "}
              <span className="text-sm font-normal text-zinc-500">Days</span>
            </p>
          </div>
          <div className="p-4 rounded-2xl border border-zinc-200 dark:border-zinc-800 bg-zinc-50/50 dark:bg-zinc-900/30">
            <p className="text-xs text-zinc-500 dark:text-zinc-400 font-medium">
              Rate Limit
            </p>
            <p className="font-semibold text-zinc-900 dark:text-white mt-1 text-lg">
              {org.rate_limit_per_second}{" "}
              <span className="text-sm font-normal text-zinc-500">Req/s</span>
            </p>
          </div>
        </div>
      </div>

      {/* Manage Action */}
      <div className="pt-2 relative z-10">
        <button
          onClick={() => navigate(`/organizations/${org.id}`)}
          className="group relative w-full flex items-center justify-center gap-2 py-3 px-4 rounded-xl border border-zinc-200 dark:border-zinc-800 bg-transparent overflow-hidden cursor-pointer"
        >
          {/* Hover Scaling Background Div */}
          <div className="absolute inset-0 bg-zinc-100 dark:bg-zinc-800/80 rounded-xl scale-95 opacity-0 transition-all duration-300 ease-out group-hover:scale-100 group-hover:opacity-100"></div>

          {/* Button Text */}
          <span className="relative z-10 text-md font-medium text-zinc-600 dark:text-zinc-400 group-hover:text-zinc-900 dark:group-hover:text-white transition-colors duration-300">
            Manage Workspace Settings
          </span>

          {/* Moving Arrow Icon */}
          <ChevronDown className="relative z-10 w-4 h-4 text-zinc-400 group-hover:text-zinc-900 dark:group-hover:text-white transition-all duration-300 transform group-hover:translate-x-1 -rotate-90" />
        </button>
      </div>
    </section>
  );
};

export default Organization;
