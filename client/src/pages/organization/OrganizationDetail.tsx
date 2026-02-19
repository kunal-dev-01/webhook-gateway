import { useParams } from "react-router-dom"; // 👈 Hook import kiya
import { useGetOrganization } from "@/hooks/useOrg"; 
import { Loader } from "@/common/Loader";
import { Error } from "@/svg/Common";
import { OrgGeneralSettings } from "./OrgGeneralSettings";
import { OrgMembersList } from "./OrgMembersList";
import { OrgUsageLimits } from "./OrgUsageLimits";
import { OrgInvitationsList } from "./OrgInvitationsList";

// 👈 Prop nikal diya, ab component parameters ke through id lega
const OrganizationDetail = () => {
  // 👈 URL param se orgId nikal rahe hain
  const { orgId } = useParams<{ orgId: string }>(); 
  
  // Agar kisi wajah se orgId undefined ho, hum query run nahi karenge
  const { organization, isLoading, isError } = useGetOrganization(orgId || "");

  // Fallback agar URL mein parameter nahi hai (waise ye case aayega nahi routing setup ki wajah se)
  if (!orgId) {
     return <div className="text-center py-10 text-red-500">Invalid Organization ID</div>;
  }

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
        <h1 className="text-3xl font-bold tracking-tight text-zinc-700 dark:text-white">
          Organization not found.
        </h1>
      </div>
    );
  }

  return (
    <div className="mx-auto w-full max-w-7xl py-8 md:py-12">
      <div className="mb-8 md:mb-12">
        <h1 className="text-3xl font-bold tracking-tight text-zinc-700 dark:text-white capitalize">
          {organization.name} Workspace
        </h1>
        <p className="text-zinc-500 dark:text-zinc-400 mt-2 text-lg">
          Manage your webhook gateway configurations, team, and billing limits.
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
        {/* Left Column: Settings & Usage */}
        <div className="lg:col-span-7 space-y-8">
          <OrgGeneralSettings org={organization} />
          <OrgUsageLimits org={organization} />
        </div>

        {/* Right Column: Members Management */}
        <div className="lg:col-span-5 space-y-8 lg:sticky lg:top-8">
          <OrgMembersList orgId={organization.id} currentUserRole={organization.role} />
          <OrgInvitationsList orgId={organization.id} currentUserRole={organization.role} />
        </div>
      </div>
    </div>
  );
};

export default OrganizationDetail;