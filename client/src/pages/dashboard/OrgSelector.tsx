import { ChevronDown, QuestionMark } from "@/svg/Common";
import { type Organization } from "@/types/org.types";

interface OrgSelectorProps {
  organizations: Organization[];
  selectedOrgId: string;
  onSelect: (orgId: string) => void;
}

export const OrgSelector = ({ organizations, selectedOrgId, onSelect }: OrgSelectorProps) => {
  // Selected organization data
  const selectedOrg = organizations.find((org) => org.id === selectedOrgId);

  if (organizations.length === 0) {
    return (
      <div className="p-4 bg-zinc-50 dark:bg-zinc-900/50 border border-zinc-200 dark:border-zinc-800 rounded-xl text-sm text-zinc-500 mb-10">
        No organizations found. Please create one from the Organizations page.
      </div>
    );
  }

  return (
    <div className="flex items-center gap-3 mb-10">
      {/* Custom Select Dropdown Container */}
      <div className="relative w-64 group">
        <select
          value={selectedOrgId}
          onChange={(e) => onSelect(e.target.value)}
          className="
            w-full appearance-none 
            px-4 py-2.5 rounded-xl 
            border border-zinc-200 dark:border-zinc-800 
            bg-white dark:bg-zinc-900/70 
            text-zinc-900 dark:text-white 
            text-sm font-medium 
            outline-none 
            hover:border-zinc-300 dark:hover:border-zinc-700
            cursor-pointer
            transition-all duration-200 ease-in-out
          "
        >
          {organizations.map((org) => (
            <option key={org.id} value={org.id} className="dark:bg-zinc-900">
              {org.name}
            </option>
          ))}
        </select>
        
        {/* Custom Chevron Icon */}
        <div className="pointer-events-none absolute inset-y-0 right-0 flex items-center px-4 text-zinc-400 group-hover:text-zinc-600 dark:group-hover:text-zinc-300 transition-colors duration-200">
          <ChevronDown className="w-5 h-5" />
        </div>
      </div>

      {/* Question Mark Info Tooltip */}
      {selectedOrg && (
        <div className="relative group/tooltip flex items-center justify-center">
          <button className="text-zinc-400 hover:text-indigo-500 dark:hover:text-indigo-400 transition-colors duration-200 cursor-help p-2 outline-none border border-zinc-100 dark:border-zinc-800 rounded-full">
            <QuestionMark className="w-6 h-6" />
          </button>
          
          {/* Tooltip Card */}
          <div className="absolute left-1/2 -translate-x-1/2 top-full mt-2 w-64 opacity-0 invisible group-hover/tooltip:opacity-100 group-hover/tooltip:visible transition-all duration-200 z-50 transform translate-y-2 group-hover/tooltip:translate-y-0">
            <div className="bg-white/50 dark:bg-zinc-900/40 border border-zinc-100 dark:border-zinc-800 rounded-2xl p-4">
              <h4 className="text-sm font-bold text-zinc-900 dark:text-white border-b border-zinc-100 dark:border-zinc-700 pb-2 mb-2 truncate">
                {selectedOrg.name} Details
              </h4>
              <div className="space-y-2 text-xs">
                <div className="flex justify-between">
                  <span className="text-zinc-500">Plan Tier:</span>
                  <span className="font-semibold text-indigo-500 capitalize">{selectedOrg.plan_tier}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-zinc-500">Your Role:</span>
                  <span className="font-semibold text-zinc-700 dark:text-zinc-300 capitalize">{selectedOrg.role}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-zinc-500">Events Used:</span>
                  <span className="font-semibold text-zinc-700 dark:text-zinc-300">
                    {selectedOrg.monthly_events_used.toLocaleString()} / {selectedOrg.monthly_events_limit.toLocaleString()}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};