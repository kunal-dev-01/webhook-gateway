import { useState, useEffect } from "react";
import { useListOrganizations } from "@/hooks/useOrg";
import { Loader } from "@/common/Loader";
import { Bell, Error } from "@/svg/Common";
import { OrgSelector } from "./OrgSelector"; // Path adjust kar lena

const Dashboard = () => {
  const { organizations, isLoading, isError } = useListOrganizations();
  const [selectedOrgId, setSelectedOrgId] = useState<string>("");

  // Jaise hi organizations load ho, by default pehle wale ko select kar lo
  useEffect(() => {
    if (organizations.length > 0 && !selectedOrgId) {
      setSelectedOrgId(organizations[0].id);
    }
  }, [organizations, selectedOrgId]);

  if (isLoading) {
    return (
      <div className="flex justify-center items-center min-h-[60vh]">
        <Loader className="w-8 h-8 text-indigo-500" />
      </div>
    );
  }

  if (isError) {
    return (
      <div className="flex flex-col justify-center items-center gap-5 min-h-[60vh]">
        <Error className="w-16 h-16 text-red-500" />
        <h1 className="text-2xl font-bold tracking-tight text-zinc-700 dark:text-white">
          Failed to load dashboard data.
        </h1>
      </div>
    );
  }

  return (
    <div className="mx-auto w-full max-w-7xl py-8 md:py-12">
      
      {/* ================= HEADER & BELL ICON ================= */}
      <div className="flex items-center justify-between mb-8 md:mb-12">
        <div>
          <h1 className="text-3xl font-bold tracking-tight text-zinc-900 dark:text-white">
            Dashboard
          </h1>
          <p className="text-zinc-500 dark:text-zinc-400 mt-1">
            Overview of your webhook gateway activity.
          </p>
        </div>

        {/* Bell Notification Button */}
        <button className="relative p-2.5 rounded-full border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 text-zinc-500 hover:text-indigo-600 dark:hover:text-indigo-400 hover:bg-zinc-50 dark:hover:bg-zinc-800 transition-colors cursor-pointer shadow-sm">
          <Bell className="w-6 h-6" />
          <span className="absolute top-2.5 right-2.5 w-2 h-2 bg-indigo-500 rounded-full border border-white dark:border-zinc-900"></span>
        </button>
      </div>

      {/* ================= ORGANIZATION SELECTOR COMPONENT ================= */}
      <OrgSelector 
        organizations={organizations} 
        selectedOrgId={selectedOrgId} 
        onSelect={setSelectedOrgId} 
      />

      {/* ================= WIP DASHBOARD CONTENT AREA ================= */}
      <div className="w-full h-64 border-2 border-dashed border-zinc-200 dark:border-zinc-800 rounded-3xl flex flex-col items-center justify-center text-zinc-400 dark:text-zinc-500 bg-zinc-50/50 dark:bg-zinc-900/20">
        <svg className="w-8 h-8 mb-3 opacity-50" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
        </svg>
        <p className="font-medium">Gateway Analytics & Charts</p>
        <p className="text-xs mt-1">Coming soon...</p>
      </div>

    </div>
  );
};

export default Dashboard;