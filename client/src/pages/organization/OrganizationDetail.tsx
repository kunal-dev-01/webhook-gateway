import { useState } from "react";
import { useParams } from "react-router-dom";
import { useGetOrganization } from "@/hooks/useOrg"; 
import { Loader } from "@/common/Loader";
import { Error, Settings } from "@/svg/Common"; 
// Sub-components
import { OrgGeneralSettings } from "./OrgGeneralSettings";
import { OrgMembersList } from "./OrgMembersList";
import { OrgInvitationsList } from "./OrgInvitationsList";
import { OrgUsageLimits } from "./OrgUsageLimits";
import { OrgBilling } from "./OrgBilling"; // <--- Updated Import
import { Users } from "@/svg/Auth";
import { BarChart, Payment } from "@/svg/More";

type TabType = "overview" | "members" | "billing" | "settings";

const OrganizationDetail = () => {
  const { orgId } = useParams<{ orgId: string }>(); 
  const { organization, isLoading, isError } = useGetOrganization(orgId || "");
  const [activeTab, setActiveTab] = useState<TabType>("overview");

  if (!orgId) return <div className="text-center py-10 text-red-500">Invalid Organization ID</div>;

  if (isLoading) {
    return (
      <div className="flex justify-center items-center min-h-[60vh]">
        <Loader className="w-8 h-8 text-indigo-500" />
      </div>
    );
  }

  if (isError || !organization) {
    return (
      <div className="flex flex-col justify-center items-center gap-5 min-h-[60vh]">
        <Error className="w-20 h-20 text-red-500" />
        <h1 className="text-3xl font-bold tracking-tight text-zinc-700 dark:text-white">Organization not found.</h1>
      </div>
    );
  }

  // Check permissions for Settings/Billing
  const isOwnerOrAdmin = organization.role === 'owner' || organization.role === 'admin';

  return (
    <div className="mx-auto w-full max-w-7xl py-8 md:py-12">
      
      {/* 1. Header & Navigation */}
      <div className="mb-8">
        <h1 className="text-3xl font-bold tracking-tight text-zinc-900 dark:text-white capitalize">
          {organization.name}
        </h1>
        <p className="text-zinc-500 dark:text-zinc-400 mt-1 text-lg">
          @{organization.slug}
        </p>
        
        {/* TABS NAVIGATION */}
        <div className="mt-8 flex items-center gap-1 overflow-x-auto border-b border-zinc-200 dark:border-zinc-800 pb-0.5 no-scrollbar">
           <TabButton 
             active={activeTab === 'overview'} 
             onClick={() => setActiveTab('overview')} 
             label="Overview" 
             icon={<BarChart className="w-4 h-4" />}
           />
           <TabButton 
             active={activeTab === 'members'} 
             onClick={() => setActiveTab('members')} 
             label="Team & Members" 
             icon={<Users className="w-4 h-4" />}
           />
           <TabButton 
             active={activeTab === 'billing'} 
             onClick={() => setActiveTab('billing')} 
             label="Billing & Plans" 
             icon={<Payment className="w-4 h-4" />}
           />
           {isOwnerOrAdmin && (
             <TabButton 
               active={activeTab === 'settings'} 
               onClick={() => setActiveTab('settings')} 
               label="Settings" 
               icon={<Settings className="w-4 h-4" />}
             />
           )}
        </div>
      </div>

      {/* 2. Content Area based on Tab */}
      <div className="min-h-100">
        
        {/* VIEW: OVERVIEW */}
        {activeTab === "overview" && (
          <div className="space-y-8 animate-in fade-in slide-in-from-bottom-2 duration-300">
             <OrgUsageLimits org={organization} />
             {/* You can add a "Recent Activity" component here later */}
          </div>
        )}

        {/* VIEW: MEMBERS */}
        {activeTab === "members" && (
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start animate-in fade-in slide-in-from-bottom-2 duration-300">
            <div className="lg:col-span-8">
              <OrgMembersList orgId={organization.id} currentUserRole={organization.role} />
            </div>
            <div className="lg:col-span-4 lg:sticky lg:top-8">
              <OrgInvitationsList orgId={organization.id} currentUserRole={organization.role} />
            </div>
          </div>
        )}

        {/* VIEW: BILLING (Now Context Aware) */}
        {activeTab === "billing" && (
           <OrgBilling orgId={organization.id} /> 
        )}

        {/* VIEW: SETTINGS */}
        {activeTab === "settings" && isOwnerOrAdmin && (
          <div className="max-w-3xl animate-in fade-in slide-in-from-bottom-2 duration-300">
             <OrgGeneralSettings org={organization} />
          </div>
        )}

      </div>
    </div>
  );
};

// Small utility component for consistent tabs
const TabButton = ({ active, onClick, label, icon }: any) => (
  <button
    onClick={onClick}
    className={`
      flex items-center gap-2 px-4 py-3 text-sm font-medium border-b-2 transition-all duration-200 whitespace-nowrap
      ${active 
        ? "border-indigo-500 text-indigo-600 dark:text-indigo-400" 
        : "border-transparent text-zinc-500 hover:text-zinc-700 dark:hover:text-zinc-300 hover:border-zinc-300"
      }
    `}
  >
    {icon}
    {label}
  </button>
);

export default OrganizationDetail;