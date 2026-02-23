import { useState } from "react";
import { toast } from "sonner";
import { useSessions, useDeleteSession } from "@/hooks/useAuth";
import DeleteButton from "@/common/DeleteButton";
import { Monitor, Globe, Users, Clock } from "@/svg/Auth";

export const SessionsSection = () => {
  const [deletingId, setDeletingId] = useState<string | null>(null);

  // Getting data from custom hooks
  const { sessions, isPending: isLoadingSessions } = useSessions();
  const { deleteSession } = useDeleteSession();

  // Handle Delete Session
  const handleDelete = (sessionId: string) => {
    setDeletingId(sessionId);
    deleteSession(sessionId, {
      onSuccess: () => {
        toast.success("Session revoked successfully.");
        setDeletingId(null);
      },
      onError: () => {
        toast.error("Failed to revoke session.");
        setDeletingId(null);
      }
    });
  };

  // Modern Relative Time Formatter (e.g., "Just now", "2h ago")
  const formatRelativeTime = (dateString: string) => {
    const date = new Date(dateString);
    const now = new Date();
    const diffInSeconds = Math.floor((now.getTime() - date.getTime()) / 1000);

    if (diffInSeconds < 60) return "Just now";
    if (diffInSeconds < 3600) return `${Math.floor(diffInSeconds / 60)}m ago`;
    if (diffInSeconds < 86400) return `${Math.floor(diffInSeconds / 3600)}h ago`;
    if (diffInSeconds < 2592000) return `${Math.floor(diffInSeconds / 86400)}d ago`;
    return date.toLocaleDateString("en-US", { month: "short", day: "numeric" });
  };

  // Basic device parser fallback
  const getDeviceName = (userAgent: string = "") => {
    if (userAgent.includes("Windows")) return "Windows PC";
    if (userAgent.includes("Mac")) return "MacBook / iMac";
    if (userAgent.includes("iPhone") || userAgent.includes("iPad")) return "iOS Device";
    if (userAgent.includes("Android")) return "Android Device";
    if (userAgent.includes("Linux")) return "Linux PC";
    return "Unknown Device";
  };

  return (
    <section className="bg-white dark:bg-zinc-900/70 border border-zinc-200 dark:border-zinc-800 rounded-4xl p-6 md:p-8 space-y-6 relative overflow-hidden mt-8">
      
      {/* Header Section */}
      <div className="flex items-start gap-4 mb-6">
        <div className="p-3 bg-zinc-100 dark:bg-zinc-800 rounded-xl hidden sm:block">
          <Users className="w-6 h-6 text-zinc-500" />
        </div>
        <div>
          <h2 className="text-xl font-bold text-zinc-900 dark:text-white">
            Active Sessions
          </h2>
          <p className="text-sm text-zinc-500 max-w-2xl mt-1">
            Review devices currently logged into your account. Revoke access immediately for any unfamiliar devices to secure your account.
          </p>
        </div>
      </div>
      
      <div className="w-full h-px bg-zinc-100 dark:bg-zinc-800/80 mb-6"></div>

      {/* SESSIONS LIST */}
      <div className="space-y-3">
        {isLoadingSessions ? (
          // SKELETON LOADER (Professional Loading State)
          <>
            {[1, 2].map((i) => (
              <div key={i} className="flex flex-col sm:flex-row sm:items-center justify-between p-5 rounded-2xl border border-zinc-200 dark:border-zinc-800/60 bg-zinc-50/50 dark:bg-zinc-900/20 gap-4 animate-pulse">
                <div className="flex items-start gap-4 w-full">
                  <div className="w-10 h-10 rounded-full bg-zinc-200 dark:bg-zinc-800 shrink-0"></div>
                  <div className="space-y-3 w-full max-w-sm mt-1">
                    <div className="h-4 bg-zinc-200 dark:bg-zinc-800 rounded w-1/2"></div>
                    <div className="flex gap-2">
                      <div className="h-4 bg-zinc-200 dark:bg-zinc-800 rounded w-20"></div>
                      <div className="h-4 bg-zinc-200 dark:bg-zinc-800 rounded w-20"></div>
                    </div>
                  </div>
                </div>
                <div className="w-24 h-9 bg-zinc-200 dark:bg-zinc-800 rounded-lg shrink-0"></div>
              </div>
            ))}
          </>
        ) : sessions.length > 0 ? (
          // ACTIVE SESSIONS RENDER
          <div className="grid gap-3">
            {sessions.map((session) => (
              <div 
                key={session.session_id} 
                className="group flex flex-col sm:flex-row sm:items-center justify-between p-5 rounded-2xl border border-zinc-200 dark:border-zinc-800/60 bg-white dark:bg-zinc-900/40 hover:bg-zinc-50 dark:hover:bg-zinc-800/80 hover:border-zinc-300 dark:hover:border-zinc-700 transition-all duration-300 gap-4"
              >
                {/* Left: Device & OS Info */}
                <div className="flex items-start gap-4 w-full">
                  <div className="mt-0.5 w-10 h-10 rounded-full bg-zinc-100 dark:bg-zinc-800 flex items-center justify-center border border-zinc-200 dark:border-zinc-700 shrink-0 text-zinc-600 dark:text-zinc-400 group-hover:bg-indigo-50 dark:group-hover:bg-indigo-500/10 group-hover:text-indigo-600 dark:group-hover:text-indigo-400 group-hover:border-indigo-100 dark:group-hover:border-indigo-500/20 transition-colors duration-300">
                     <Monitor className="w-5 h-5" />
                  </div>
                  
                  <div className="flex flex-col w-full">
                    <h4 className="font-semibold text-zinc-900 dark:text-zinc-100 text-[15px]">
                      {session.device_info || getDeviceName(session.user_agent)}
                    </h4>
                    
                    {/* Minimalist Badges for Meta Data */}
                    <div className="flex flex-wrap items-center gap-2 mt-2 text-[11px] font-medium text-zinc-500 dark:text-zinc-400">
                      <div className="flex items-center gap-1.5 bg-zinc-100 dark:bg-zinc-800/60 px-2.5 py-1 rounded-md border border-zinc-200/50 dark:border-zinc-700/50">
                        <Globe className="w-3 h-3" />
                        <span>{session.ip_address || "Unknown IP"}</span>
                      </div>
                      <div className="flex items-center gap-1.5 bg-zinc-100 dark:bg-zinc-800/60 px-2.5 py-1 rounded-md border border-zinc-200/50 dark:border-zinc-700/50">
                        <Clock className="w-3 h-3" />
                        <span>Active {formatRelativeTime(session.last_activity)}</span>
                      </div>
                    </div>
                    
                    {/* 👇 UPDATED: User Agent - Full Text, No Truncation, Graceful Wrapping */}
                    <p className="text-[11px] text-zinc-500 dark:text-zinc-400/80 mt-3 font-mono wrap-break-word leading-relaxed w-full pr-4">
                      {session.user_agent}
                    </p>
                  </div>
                </div>

                {/* 👇 UPDATED: Action Button - Perfect for Light & Dark Mode */}
                <div className="sm:shrink-0 flex justify-end mt-4 sm:mt-0">
                  <DeleteButton 
                    loading={deletingId === session.session_id}
                    onClick={() => handleDelete(session.session_id)}
                    className="px-6"
                  >
                    Revoke Access
                  </DeleteButton>
                </div>
              </div>
            ))}
          </div>
        ) : (
          // EMPTY STATE (When only current session exists)
          <div className="text-center py-12 px-4 rounded-2xl border border-dashed border-zinc-200 dark:border-zinc-800 bg-zinc-50/50 dark:bg-zinc-900/20">
            <div className="w-12 h-12 bg-green-50 dark:bg-green-500/10 text-green-600 dark:text-green-400 rounded-full flex items-center justify-center mx-auto mb-3">
              <Monitor className="w-6 h-6" />
            </div>
            <h3 className="text-zinc-900 dark:text-zinc-100 font-medium text-[15px]">Account is Secure</h3>
            <p className="text-sm text-zinc-500 mt-1 max-w-sm mx-auto">
              You are currently only logged in on this device. No other active sessions were found.
            </p>
          </div>
        )}
      </div>
    </section>
  );
};