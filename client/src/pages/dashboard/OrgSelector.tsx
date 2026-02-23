import { useState, useRef, useEffect } from "react";
import { QuestionMark } from "@/svg/Common";
import { type Organization } from "@/types/org.types";
import SelectInput from "@/common/SelectInput";

interface OrgSelectorProps {
  organizations: Organization[];
  selectedOrgId: string;
  onSelect: (orgId: string) => void;
}

export const OrgSelector = ({ organizations, selectedOrgId, onSelect }: OrgSelectorProps) => {
  const selectedOrg = organizations.find((org) => org.id === selectedOrgId);
  
  // Mobile + Desktop state handling
  const [showTooltip, setShowTooltip] = useState(false);
  const tooltipRef = useRef<HTMLDivElement>(null);

  // Close tooltip when clicking outside (crucial for mobile experience)
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (tooltipRef.current && !tooltipRef.current.contains(event.target as Node)) {
        setShowTooltip(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  // Convert organizations to SelectInput Option format
  const orgOptions = organizations.map((org) => ({
    label: org.name,
    value: org.id,
  }));

  if (organizations.length === 0) {
    return (
      <div className="p-4 bg-zinc-50 dark:bg-zinc-900/50 border border-zinc-200 dark:border-zinc-800 rounded-xl text-sm text-zinc-500 mb-10">
        No organizations found. Please create one from the Organizations page.
      </div>
    );
  }

  return (
    <div className="flex items-end gap-3 mb-10 w-full sm:w-auto">
      
      {/* Custom Select Dropdown */}
      <div className="w-full sm:w-72">
        <SelectInput
          label="Active Workspace"
          options={orgOptions}
          value={selectedOrgId}
          onChange={onSelect}
        />
      </div>

      {/* Question Mark Info Tooltip */}
      {selectedOrg && (
        <div 
          ref={tooltipRef}
          className="relative flex items-center justify-center shrink-0"
          onMouseEnter={() => setShowTooltip(true)} // Desktop Hover Open
          onMouseLeave={() => setShowTooltip(false)} // Desktop Hover Close
        >
          <button 
            type="button"
            onClick={() => setShowTooltip(!showTooltip)} // Mobile Tap Toggle
            className="text-zinc-400 hover:text-indigo-500 dark:hover:text-indigo-400 transition-colors duration-200 cursor-help p-2 outline-none border border-zinc-200 dark:border-zinc-800 rounded-xl bg-white dark:bg-zinc-900 hover:border-indigo-200 dark:hover:border-indigo-900/50"
          >
            <QuestionMark className="w-5 h-5" />
          </button>
          
          {/* Tooltip Card (Controlled by State now) */}
          <div 
            className={`absolute top-full mt-2 w-64 z-50 transition-all duration-200 transform right-0 sm:left-1/2 sm:-translate-x-1/2 sm:right-auto ${
              showTooltip ? 'opacity-100 visible translate-y-0' : 'opacity-0 invisible translate-y-2'
            }`}
          >
            <div className="bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-700 rounded-2xl p-4 shadow-xl backdrop-blur-sm">
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