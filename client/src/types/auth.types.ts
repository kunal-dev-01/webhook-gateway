import { type User } from "./user.types";

// ====================== AUTHENTICATION TYPES ======================

export interface RegisterRequest {
  full_name: string;
  email: string;
  password: string;
  plan_tier?: string;
  billing_period?: string;
}

export interface RegisterResponse {
  message: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  // If 2FA is DISABLED
  user?: User;
  access_token?: string;
  expires_in?: number;

  // If 2FA is ENABLED
  two_factor_required: boolean;
  temp_token?: string; // Valid for 5 minutes
}

export interface Verify2FARequest {
  temp_token: string;
  code: string;
  is_recovery?: boolean; // true if using backup code
}

export interface Verify2FAResponse {
  user: User;
  access_token: string;
  expires_in: number;
}

export interface RefreshTokenResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

// ====================== EMAIL VERIFICATION ======================

export interface VerifyEmailRequest {
  token: string;
}

export interface VerifyEmailResponse {
  message: string;
}

export interface ResendVerificationRequest {
  email: string;
}

export interface ResendVerificationResponse {
  message: string;
}

// ====================== PASSWORD MANAGEMENT ======================

export interface ForgotPasswordRequest {
  email: string;
}

export interface ForgotPasswordResponse {
  message: string;
}

export interface ResetPasswordRequest {
  token: string;
  new_password: string;
}

export interface ResetPasswordResponse {
  message: string;
}

export interface ChangePasswordRequest {
  current_password: string;
  new_password: string;
}

export interface ChangePasswordResponse {
  message: string;
}

// ====================== LOGOUT ======================

export interface LogoutRequest {
  all_devices?: boolean;
}

export interface LogoutResponse {
  message: string;
}

// ====================== SESSIONS ======================

export interface SessionResponse {
  session_id: string;
  device_info?: string;
  ip_address?: string;
  user_agent?: string;
  created_at: string;
  last_activity: string;
}

export interface ListSessionsResponse {
  sessions: SessionResponse[];
  total: number;
}