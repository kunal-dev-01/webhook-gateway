import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { notificationService } from "@/services/notification.service";
import type { ListNotificationsParams } from "@/types/notification.types";
import type { ApiError } from "@/types/api.types";

// ============================================================================
// QUERY KEYS
// ============================================================================

export const notificationKeys = {
  all: ["notifications"] as const,
  lists: (params?: ListNotificationsParams) => [...notificationKeys.all, "list", params] as const,
  unreadCount: () => [...notificationKeys.all, "unread-count"] as const,
};

// ============================================================================
// LIST NOTIFICATIONS
// ============================================================================

export const useListNotifications = (params?: ListNotificationsParams) => {
  const query = useQuery({
    queryKey: notificationKeys.lists(params),
    queryFn: () => notificationService.listNotifications(params),
    staleTime: 30 * 1000, // 30s — notifications change frequently
  });

  return {
    notifications: query.data?.notifications ?? [],
    total: query.data?.total ?? 0,
    limit: query.data?.limit ?? 20,
    offset: query.data?.offset ?? 0,
    isLoading: query.isLoading,
    isError: query.isError,
    error: query.error as ApiError | null,
    refetch: query.refetch,
  };
};

// ============================================================================
// UNREAD COUNT
// ============================================================================

export const useUnreadCount = () => {
  const query = useQuery({
    queryKey: notificationKeys.unreadCount(),
    queryFn: () => notificationService.getUnreadCount(),
    staleTime: 30 * 1000,
    refetchInterval: 60 * 1000, // Poll every 60s for badge updates
  });

  return {
    unreadCount: query.data?.unread_count ?? 0,
    isLoading: query.isLoading,
    error: query.error as ApiError | null,
    refetch: query.refetch,
  };
};

// ============================================================================
// MARK READ
// ============================================================================

export const useMarkRead = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (notificationId: string) => notificationService.markRead(notificationId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: notificationKeys.all });
    },
  });

  return {
    markRead: mutation.mutate,
    markReadAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    error: mutation.error as ApiError | null,
  };
};

// ============================================================================
// MARK ALL READ
// ============================================================================

export const useMarkAllRead = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: () => notificationService.markAllRead(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: notificationKeys.all });
    },
  });

  return {
    markAllRead: mutation.mutate,
    markAllReadAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    error: mutation.error as ApiError | null,
  };
};

// ============================================================================
// DELETE NOTIFICATION
// ============================================================================

export const useDeleteNotification = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (notificationId: string) => notificationService.deleteNotification(notificationId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: notificationKeys.all });
    },
  });

  return {
    deleteNotification: mutation.mutate,
    deleteNotificationAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    error: mutation.error as ApiError | null,
  };
};