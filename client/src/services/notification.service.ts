import { axiosClient } from "@/libs/axios";
import type {
  ListNotificationsParams,
  ListNotificationsResponse,
  UnreadCountResponse,
} from "@/types/notification.types";
import type { MessageResponse } from "@/types/api.types";

function unwrapResponse<T>(response: { data: { success: boolean; data: T } }): T {
  return response.data.data;
}

export const notificationService = {
  // GET /notifications
  listNotifications: async (params?: ListNotificationsParams): Promise<ListNotificationsResponse> => {
    const response = await axiosClient.get("/user/notifications", { params });
    return unwrapResponse<ListNotificationsResponse>(response);
  },

  // GET /notifications/unread-count
  getUnreadCount: async (): Promise<UnreadCountResponse> => {
    const response = await axiosClient.get("/user/notifications/unread-count");
    return unwrapResponse<UnreadCountResponse>(response);
  },

  // PATCH /notifications/:notificationId/read
  markRead: async (notificationId: string): Promise<MessageResponse> => {
    const response = await axiosClient.patch(`/user/notifications/${notificationId}/read`);
    return unwrapResponse<MessageResponse>(response);
  },

  // PATCH /notifications/read-all
  markAllRead: async (): Promise<MessageResponse> => {
    const response = await axiosClient.patch("/user/notifications/read-all");
    return unwrapResponse<MessageResponse>(response);
  },

  // DELETE /notifications/:notificationId
  deleteNotification: async (notificationId: string): Promise<MessageResponse> => {
    const response = await axiosClient.delete(`/user/notifications/${notificationId}`);
    return unwrapResponse<MessageResponse>(response);
  },
};