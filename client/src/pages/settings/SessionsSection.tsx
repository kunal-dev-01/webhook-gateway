import { useState } from "react";
import { toast } from "sonner";
import { useSessions, useDeleteSession } from "@/hooks/useAuth";
import SubmitButton from "@/common/SubmitButton";
import DeleteButton from "@/common/DeleteButton";
import { Monitor, Globe, Users, Clock } from "@/svg/Auth";
import { Loader } from "@/common/Loader";
import { Error } from "@/svg/Common";

export const SessionsSection = () => {
  const [showSessions, setShowSessions] = useState(false);
  const [deletingId, setDeletingId] = useState<string | null>(null);

  // Getting data from your custom hooks
  const { sessions, isPending: isLoadingSessions, refetch } = useSessions();
  const { deleteSession } = useDeleteSession();

  // Handle Fetch Sessions Click
  const handleRevealSessions = async () => {
    setShowSessions(true);
    await refetch(); // Fetch fresh data on click
  };

  // Handle Delete Session
  const handleDelete = (sessionId: string) => {
    setDeletingId(sessionId);
    deleteSession(sessionId, {
      onSuccess: () => {
        toast.success("Session terminated successfully.");
        setDeletingId(null);
      },
      onError: () => {
        toast.error("Failed to terminate session.");
        setDeletingId(null);
      }
    });
  };

  // Helper to format date
  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleString("en-US", {
      month: "short", day: "numeric", hour: "2-digit", minute: "2-digit"
    });
  };

  // Basic device parser based on user_agent (Optional, makes UI look good)
  const getDeviceName = (userAgent: string = "") => {
    if (userAgent.includes("Windows")) return "Windows PC";
    if (userAgent.includes("Mac")) return "MacBook / iMac";
    if (userAgent.includes("iPhone") || userAgent.includes("iPad")) return "iOS Device";
    if (userAgent.includes("Android")) return "Android Device";
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
          <p className="text-sm text-zinc-500 max-w-xl">
            Review the devices currently logged into your account. If you see any unfamiliar activity, you can terminate the session immediately to protect your account.
          </p>
        </div>
      </div>
      <div className="w-full h-px bg-zinc-100 dark:bg-zinc-800/80 mb-6"></div>

      {/* BEFORE CLICK: Information & Trigger Button */}
      {!showSessions ? (
        <div className="bg-zinc-50/50 dark:bg-zinc-900/30 border border-zinc-200 dark:border-zinc-800 rounded-2xl p-6 text-center space-y-4">
          <Monitor className="w-10 h-10 text-zinc-400 mx-auto opacity-50" />
          <div>
            <h3 className="text-zinc-700 dark:text-zinc-300 font-medium">Keep your account secure</h3>
            <p className="text-sm text-zinc-500 mt-1 mb-4">Click the button below to fetch and review all your active login sessions securely.</p>
          </div>
          <div className="flex justify-center">
             <SubmitButton onClick={handleRevealSessions} className="px-6 w-auto">
               Get Active Sessions
             </SubmitButton>
          </div>
        </div>
      ) : (
        /* AFTER CLICK: Display Sessions List */
        <div className="space-y-4">
          {isLoadingSessions && !sessions?.length ? (
            <div className="flex justify-center items-center py-10">
              <Loader className="w-8 h-8 text-indigo-500" />
            </div>
          ) : (
            <div className="grid gap-4">
              {sessions.map((session) => (
                <div 
                  key={session.session_id} 
                  className="flex flex-col sm:flex-row sm:items-center justify-between p-5 rounded-2xl border border-zinc-200 dark:border-zinc-800 bg-zinc-50/50 dark:bg-zinc-900/30 hover:border-zinc-300 dark:hover:border-zinc-700 transition-colors gap-4"
                >
                  {/* Left: Device & OS Info */}
                  <div className="flex items-start gap-4">
                    <div className="w-12 h-12 rounded-full bg-white dark:bg-zinc-800 flex items-center justify-center border border-zinc-200 dark:border-zinc-700 shrink-0 shadow-xs">
                       <Monitor className="w-6 h-6 text-indigo-500" />
                    </div>
                    <div>
                      <h4 className="font-semibold text-zinc-900 dark:text-zinc-100 text-[15px]">
                        {session.device_info || getDeviceName(session.user_agent)}
                      </h4>
                      
                      {/* Session Details (IP & Time) */}
                      <div className="flex flex-col sm:flex-row sm:items-center gap-1 sm:gap-4 mt-1.5 text-xs text-zinc-500 dark:text-zinc-400">
                        <div className="flex items-center gap-1.5">
                          <Globe className="w-3.5 h-3.5" />
                          <span>{session.ip_address || "Unknown IP"}</span>
                        </div>
                        <div className="hidden sm:block w-1 h-1 rounded-full bg-zinc-300 dark:bg-zinc-700"></div>
                        <div className="flex items-center gap-1.5">
                          <Clock className="w-3.5 h-3.5" />
                          <span>Last active: {formatDate(session.last_activity)}</span>
                        </div>
                      </div>
                      
                      {/* Browser / User Agent info */}
                      <p className="text-[11px] text-zinc-400 dark:text-zinc-500 mt-2 truncate max-w-62.5 sm:max-w-xs">
                        {session.user_agent}
                      </p>
                    </div>
                  </div>

                  {/* Right: Action Button */}
                  <div className="sm:shrink-0 flex justify-end">
                    <DeleteButton 
                      loading={deletingId === session.session_id}
                      onClick={() => handleDelete(session.session_id)}
                      className="px-4 py-2 text-sm w-auto rounded-lg" // Overriding default padding for a smaller button
                    >
                      Revoke
                    </DeleteButton>
                  </div>
                </div>
              ))}

              {sessions.length === 0 && (
                <div className="text-center py-8 text-zinc-500 dark:text-zinc-400">
                  <Error className="w-10 h-10 mx-auto mb-2" />
                  No active sessions found.
                </div>
              )}
            </div>
          )}
        </div>
      )}
    </section>
  );
};