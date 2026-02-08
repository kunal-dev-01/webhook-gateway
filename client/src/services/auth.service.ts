import { axiosClient } from "@/libs/axios";
import type {
  RegisterRequest,
  LoginRequest,
  LoginResponse,
  Verify2FARequest,
  VerifyEmailRequest,
  ResendVerificationRequest,
  ForgotPasswordRequest,
  ResetPasswordRequest,
  ChangePasswordRequest,
  RefreshTokenResponse,
  Setup2FAResponse,
  Enable2FARequest,
  Disable2FARequest,
  ListSessionsResponse,
} from "@/types/auth.types";
import { type User } from "@/types/user.types";
import { type MessageResponse } from "@/types/api.types";

export const authService = {
  register: async (data: RegisterRequest): Promise<MessageResponse> => {
    const response = await axiosClient.post<MessageResponse>(
      "/auth/register",
      data,
    );
    return response.data;
  },

  login: async (data: LoginRequest): Promise<LoginResponse> => {
    const response = await axiosClient.post<LoginResponse>("/auth/login", data);
    return response.data;
  },

  verify2FA: async (data: Verify2FARequest): Promise<LoginResponse> => {
    const response = await axiosClient.post<LoginResponse>(
      "/auth/verify-2fa",
      data,
    );
    return response.data;
  },

  verifyEmail: async (data: VerifyEmailRequest): Promise<MessageResponse> => {
    const response = await axiosClient.post<MessageResponse>(
      "/auth/verify-email",
      data,
    );
    return response.data;
  },

  resendVerification: async (
    data: ResendVerificationRequest,
  ): Promise<MessageResponse> => {
    const response = await axiosClient.post<MessageResponse>(
      "/auth/resend-verification",
      data,
    );
    return response.data;
  },

  forgotPassword: async (
    data: ForgotPasswordRequest,
  ): Promise<MessageResponse> => {
    const response = await axiosClient.post<MessageResponse>(
      "/auth/forgot-password",
      data,
    );
    return response.data;
  },

  resetPassword: async (
    data: ResetPasswordRequest,
  ): Promise<MessageResponse> => {
    const response = await axiosClient.post<MessageResponse>(
      "/auth/reset-password",
      data,
    );
    return response.data;
  },

  changePassword: async (
    data: ChangePasswordRequest,
  ): Promise<MessageResponse> => {
    const response = await axiosClient.patch<MessageResponse>(
      "/auth/change-password",
      data,
    );
    return response.data;
  },

  getMe: async (): Promise<User> => {
    const response = await axiosClient.get<User>("/auth/me");
    return response.data;
  },

  refreshToken: async (): Promise<RefreshTokenResponse> => {
    const response =
      await axiosClient.post<RefreshTokenResponse>("/auth/refresh");
    return response.data;
  },

  logout: async (): Promise<MessageResponse> => {
    const response = await axiosClient.post<MessageResponse>("/auth/logout");
    return response.data;
  },

  logoutAll: async (): Promise<MessageResponse> => {
    const response =
      await axiosClient.post<MessageResponse>("/auth/logout-all");
    return response.data;
  },

  setup2FA: async (): Promise<Setup2FAResponse> => {
    const response =
      await axiosClient.post<Setup2FAResponse>("/auth/2fa/setup");
    return response.data;
  },

  enable2FA: async (data: Enable2FARequest): Promise<MessageResponse> => {
    const response = await axiosClient.post<MessageResponse>(
      "/auth/2fa/enable",
      data,
    );
    return response.data;
  },

  disable2FA: async (data: Disable2FARequest): Promise<MessageResponse> => {
    const response = await axiosClient.post<MessageResponse>(
      "/auth/2fa/disable",
      data,
    );
    return response.data;
  },

  getSessions: async (): Promise<ListSessionsResponse> => {
    const response =
      await axiosClient.get<ListSessionsResponse>("/auth/sessions");
    return response.data;
  },

  deleteSession: async (sessionId: string): Promise<MessageResponse> => {
    const response = await axiosClient.delete<MessageResponse>(
      `/auth/sessions/${sessionId}`,
    );
    return response.data;
  },
};