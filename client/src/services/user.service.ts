import { axiosClient } from "@/libs/axios";
import type {
  UpdateUserRequest,
  ListUsersParams,
  UserResponse,
  MeResponse,
  ListUsersResponse,
} from "@/types/user.types";
import { type MessageResponse } from "@/types/api.types";

export const userService = {
  // ============================================================================
  // Get Current User
  // ============================================================================

  getMe: async (): Promise<MeResponse> => {
    const response = await axiosClient.get<MeResponse>("/users/me");
    return response.data;
  },

  // ============================================================================
  // Update Current User Profile
  // ============================================================================

  updateProfile: async (data: UpdateUserRequest): Promise<MessageResponse> => {
    const response = await axiosClient.patch<MessageResponse>(
      "/users/me",
      data,
    );
    return response.data;
  },

  // ============================================================================
  // Get User by ID (Admin/HR/Accountant)
  // ============================================================================

  getUserById: async (userId: string): Promise<UserResponse> => {
    const response = await axiosClient.get<UserResponse>(`/users/${userId}`);
    return response.data;
  },

  // ============================================================================
  // List Users (Admin/HR/Accountant)
  // ============================================================================

  listUsers: async (params?: ListUsersParams): Promise<ListUsersResponse> => {
    // Convert page to offset if page is provided
    const queryParams = { ...params };
    if (queryParams.page && queryParams.limit) {
      queryParams.offset = (queryParams.page - 1) * queryParams.limit;
      delete queryParams.page;
    }

    const response = await axiosClient.get<ListUsersResponse>("/users", {
      params: queryParams,
    });
    return response.data;
  },

  // ============================================================================
  // Update User (Admin/HR)
  // ============================================================================

  updateUser: async (
    userId: string,
    data: UpdateUserRequest,
  ): Promise<MessageResponse> => {
    const response = await axiosClient.patch<MessageResponse>(
      `/users/${userId}`,
      data,
    );
    return response.data;
  },

  // ============================================================================
  // Delete User (Admin/HR)
  // ============================================================================

  deleteUser: async (userId: string): Promise<MessageResponse> => {
    const response = await axiosClient.delete<MessageResponse>(
      `/users/${userId}`,
    );
    return response.data;
  },
};