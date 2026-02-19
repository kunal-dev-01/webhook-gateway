import { useEffect } from "react";
import { toast } from "sonner";
import { 
  useListNotifications, 
  useMarkRead, 
  useMarkAllRead, 
  useDeleteNotification 
} from "@/hooks/useNotification";
import { useAcceptInvitation } from "@/hooks/useInvitation";
import { Loader } from "@/common/Loader";
import { Bell, Error } from "@/svg/Common"; 

interface NotificationSidebarProps {
  isOpen: boolean;
  onClose: () => void;
}

export const NotificationSidebar = ({ isOpen, onClose }: NotificationSidebarProps) => {
  // Jab drawer open hoga tabhi fetch karega background mein via React Query
  const { notifications, isLoading, isError, refetch } = useListNotifications();
  const { markRead } = useMarkRead();
  const { markAllRead, isPending: isMarkingAll } = useMarkAllRead();
  const { deleteNotification } = useDeleteNotification();
  const { acceptInvitation, isPending: isAccepting } = useAcceptInvitation();

  useEffect(() => {
    if (isOpen) {
      refetch();
    }
  }, [isOpen, refetch]);

  const handleAcceptInvite = (actionUrl: string | undefined, notificationId: string) => {
    if (!actionUrl) return toast.error("Invalid invitation link");
    
    const urlParams = new URLSearchParams(actionUrl.split("?")[1]);
    const token = urlParams.get("token");

    if (!token) return toast.error("Token missing in invitation");

    acceptInvitation(
      { token },
      {
        onSuccess: () => {
          toast.success("Invitation accepted successfully!");
          markRead(notificationId);
        },
        onError: () => toast.error("Failed to accept invitation."),
      }
    );
  };

  const getNotificationStyle = (type: string) => {
    switch (type) {
      case "invitation":
        return "border-indigo-200 dark:border-indigo-500/30 bg-indigo-50/50 dark:bg-indigo-500/10";
      case "usage_limit_warning":
      case "billing_failed":
        return "border-yellow-200 dark:border-yellow-500/30 bg-yellow-50/50 dark:bg-yellow-500/10";
      case "usage_limit_exceeded":
      case "payment_failed":
        return "border-red-200 dark:border-red-500/30 bg-red-50/50 dark:bg-red-500/10";
      case "payment_succeeded":
        return "border-green-200 dark:border-green-500/30 bg-green-50/50 dark:bg-green-500/10";
      default:
        return "border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900/50";
    }
  };

  return (
    <>
      {/* Backdrop */}
      <div 
        className={`fixed inset-0 bg-zinc-900/40 dark:bg-black/60 backdrop-blur-sm z-90 transition-opacity duration-300 ${
          isOpen ? "opacity-100" : "opacity-0 pointer-events-none"
        }`}
        onClick={onClose}
      />

      {/* Drawer (Left Slide) */}
      <div 
        className={`fixed top-0 left-0 h-full w-[85vw] sm:w-100 bg-white dark:bg-zinc-900 z-100 transform transition-transform duration-300 ease-in-out flex flex-col ${
          isOpen ? "translate-x-0" : "-translate-x-full"
        }`}
      >
        {/* Header */}
        <div className="flex items-center justify-between p-5 border-b border-zinc-100 dark:border-zinc-800">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-indigo-50 dark:bg-indigo-500/10 rounded-lg">
              <Bell className="w-5 h-5 text-indigo-500" />
            </div>
            <h2 className="text-lg font-bold text-zinc-900 dark:text-white">Notifications</h2>
          </div>
          <button 
            onClick={onClose}
            className="p-2 text-zinc-400 hover:text-zinc-700 dark:hover:text-zinc-300 rounded-full hover:bg-zinc-100 dark:hover:bg-zinc-800 transition-colors"
          >
            <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        {/* Mark All Read Action */}
        {notifications.some(n => !n.is_read) && (
          <div className="px-5 py-3 border-b border-zinc-100 dark:border-zinc-800 flex justify-end">
            <button
              onClick={() => markAllRead()}
              disabled={isMarkingAll}
              className="text-xs font-semibold text-indigo-600 dark:text-indigo-400 hover:text-indigo-700 transition-colors"
            >
              Mark all as read
            </button>
          </div>
        )}

        {/* Content Area */}
        <div className="flex-1 overflow-y-auto p-5 space-y-4 custom-scrollbar">
          {isLoading ? (
            <div className="flex justify-center py-10">
              <Loader className="w-6 h-6 text-indigo-500" />
            </div>
          ) : isError ? (
            <div className="text-center py-10 text-red-500">
              <Error className="w-8 h-8 mx-auto mb-2" /> 
              <p className="text-sm">Failed to load notifications.</p>
            </div>
          ) : notifications.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-16 text-center">
              <div className="w-16 h-16 bg-zinc-50 dark:bg-zinc-900 rounded-full flex items-center justify-center mb-4">
                <Bell className="w-8 h-8 text-zinc-300 dark:text-zinc-700" />
              </div>
              <p className="text-zinc-500 dark:text-zinc-400 font-medium text-sm">You're all caught up!</p>
              <p className="text-zinc-400 dark:text-zinc-600 text-xs mt-1">No new notifications right now.</p>
            </div>
          ) : (
            notifications.map((notif) => (
              <div
                key={notif.id}
                className={`relative p-4 rounded-2xl border transition-all duration-200 ${getNotificationStyle(notif.type)} ${
                  !notif.is_read ? "shadow-sm" : "opacity-75"
                }`}
              >
                {!notif.is_read && (
                  <span className="absolute top-4 right-4 w-2 h-2 bg-indigo-500 rounded-full animate-pulse"></span>
                )}

                <h3 className={`text-sm font-semibold pr-4 ${!notif.is_read ? "text-zinc-900 dark:text-white" : "text-zinc-700 dark:text-zinc-300"}`}>
                  {notif.title}
                </h3>
                <p className="text-xs text-zinc-600 dark:text-zinc-400 mt-1.5 leading-relaxed">
                  {notif.message}
                </p>
                <p className="text-[10px] text-zinc-400 dark:text-zinc-500 mt-3 font-medium">
                  {new Date(notif.created_at).toLocaleString()}
                </p>

                {/* Actions */}
                <div className="flex flex-wrap items-center gap-3 mt-4 pt-3 border-t border-zinc-200/50 dark:border-zinc-700/50">
                  {notif.type === "invitation" && notif.action_url && (
                    <button
                      onClick={() => handleAcceptInvite(notif.action_url, notif.id)}
                      disabled={isAccepting}
                      className="px-3 py-1.5 bg-indigo-600 hover:bg-indigo-700 text-white text-xs font-medium rounded-lg transition-all disabled:opacity-50"
                    >
                      {isAccepting ? "Accepting..." : notif.action_label || "Accept"}
                    </button>
                  )}

                  {!notif.is_read && (
                    <button
                      onClick={() => markRead(notif.id)}
                      className="text-xs font-semibold text-zinc-500 hover:text-zinc-900 dark:hover:text-white transition-colors"
                    >
                      Mark Read
                    </button>
                  )}

                  <button
                    onClick={() => deleteNotification(notif.id)}
                    className="text-xs font-semibold text-red-400 hover:text-red-600 transition-colors ml-auto"
                  >
                    Delete
                  </button>
                </div>
              </div>
            ))
          )}
        </div>
      </div>
    </>
  );
};