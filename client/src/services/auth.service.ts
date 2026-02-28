import { axiosClient } from "@/libs/axios";
import {unwrapResponse} from "@/utils/api";
import type {
  RegisterRequest,
  RegisterResponse,
  LoginRequest,
  LoginResponse,
  Verify2FARequest,
  Verify2FAResponse,
  VerifyEmailRequest,
  VerifyEmailResponse,
  ResendVerificationRequest,
  ResendVerificationResponse,
  ForgotPasswordRequest,
  ForgotPasswordResponse,
  ResetPasswordRequest,
  ResetPasswordResponse,
  ChangePasswordRequest,
  ChangePasswordResponse,
  RefreshTokenResponse,
  LogoutRequest,
  LogoutResponse,
  ListSessionsResponse,
} from "@/types/auth.types";

export const authService = {
  // ============================================================================
  // REGISTER
  // ============================================================================
  register: async (data: RegisterRequest): Promise<RegisterResponse> => {
    const response = await axiosClient.post("/auth/register", data);
    return unwrapResponse<RegisterResponse>(response);
  },

  // ============================================================================
  // LOGIN
  // ============================================================================
  login: async (data: LoginRequest): Promise<LoginResponse> => {
    const response = await axiosClient.post("/auth/login", data);
    return unwrapResponse<LoginResponse>(response);
  },

  // ============================================================================
  // 2FA VERIFICATION (during login)
  // ============================================================================
  verify2FA: async (data: Verify2FARequest): Promise<Verify2FAResponse> => {
    const response = await axiosClient.post("/auth/2fa/verify", data);
    return unwrapResponse<Verify2FAResponse>(response);
  },

  // ============================================================================
  // EMAIL VERIFICATION
  // ============================================================================
  verifyEmail: async (data: VerifyEmailRequest): Promise<VerifyEmailResponse> => {
    const response = await axiosClient.post("/auth/verify-email", data);
    return unwrapResponse<VerifyEmailResponse>(response);
  },

  resendVerification: async (
    data: ResendVerificationRequest,
  ): Promise<ResendVerificationResponse> => {
    const response = await axiosClient.post("/auth/resend-verification", data);
    return unwrapResponse<ResendVerificationResponse>(response);
  },

  // ============================================================================
  // PASSWORD MANAGEMENT
  // ============================================================================
  forgotPassword: async (
    data: ForgotPasswordRequest,
  ): Promise<ForgotPasswordResponse> => {
    const response = await axiosClient.post("/auth/forgot-password", data);
    return unwrapResponse<ForgotPasswordResponse>(response);
  },

  resetPassword: async (
    data: ResetPasswordRequest,
  ): Promise<ResetPasswordResponse> => {
    const response = await axiosClient.post("/auth/reset-password", data);
    return unwrapResponse<ResetPasswordResponse>(response);
  },

  changePassword: async (
    data: ChangePasswordRequest,
  ): Promise<ChangePasswordResponse> => {
    const response = await axiosClient.post("/auth/change-password", data);
    return unwrapResponse<ChangePasswordResponse>(response);
  },

  // ============================================================================
  // REFRESH TOKEN
  // ============================================================================
  refreshToken: async (): Promise<RefreshTokenResponse> => {
    const response = await axiosClient.post("/auth/refresh");
    return unwrapResponse<RefreshTokenResponse>(response);
  },

  // ============================================================================
  // LOGOUT
  // ============================================================================
  logout: async (data?: LogoutRequest): Promise<LogoutResponse> => {
    const response = await axiosClient.post("/auth/logout", data || {});
    return unwrapResponse<LogoutResponse>(response);
  },

  // ============================================================================
  // SESSIONS
  // ============================================================================
  getSessions: async (): Promise<ListSessionsResponse> => {
    const response = await axiosClient.get("/auth/sessions");
    return unwrapResponse<ListSessionsResponse>(response);
  },

  deleteSession: async (sessionId: string): Promise<{ message: string }> => {
    const response = await axiosClient.delete(`/auth/sessions/${sessionId}`);
    return unwrapResponse<{ message: string }>(response);
  },
};