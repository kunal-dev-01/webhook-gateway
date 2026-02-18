import { axiosClient } from "@/libs/axios";
import type {
  UpdateUserRequest,
  UpdateUserResponse,
  UpdateAvatarResponse,
  UpdatePreferencesRequest,
  UpdatePreferencesResponse,
  GetMeResponse,
} from "@/types/user.types";

function unwrapResponse<T>(response: { data: { success: boolean; data: T } }): T {
  return response.data.data;
}

export const userService = {
  // ============================================================================
  // GET ME
  // ============================================================================
  getMe: async (): Promise<GetMeResponse> => {
    const response = await axiosClient.get("/user/me");
    return unwrapResponse<GetMeResponse>(response);
  },

  // ============================================================================
  // UPDATE FULL NAME
  // ============================================================================
  updateUser: async (data: UpdateUserRequest): Promise<UpdateUserResponse> => {
    const response = await axiosClient.patch("/user/me", data);
    return unwrapResponse<UpdateUserResponse>(response);
  },

  // ============================================================================
  // UPDATE AVATAR (multipart/form-data)
  // ============================================================================
  updateAvatar: async (file: File): Promise<UpdateAvatarResponse> => {
    const formData = new FormData();
    formData.append("avatar", file);

    const response = await axiosClient.put("/user/me/avatar", formData, {
      headers: { "Content-Type": "multipart/form-data" },
    });
    return unwrapResponse<UpdateAvatarResponse>(response);
  },

  // ============================================================================
  // UPDATE PREFERENCES
  // ============================================================================
  updatePreferences: async (
    data: UpdatePreferencesRequest,
  ): Promise<UpdatePreferencesResponse> => {
    const response = await axiosClient.patch("/user/me/preferences", data);
    return unwrapResponse<UpdatePreferencesResponse>(response);
  },
};