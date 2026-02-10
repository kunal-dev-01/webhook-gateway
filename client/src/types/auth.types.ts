import { type User } from "./user.types";

export interface RegisterRequest {
  email: string;
  password: string;
  full_name: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  access_token: string;
  expires_in: number;
  token_type: string;
  user: User;
  require_2fa?: boolean;
  temp_token?: string;
}

export interface Verify2FARequest {
  temp_token: string;
  code: string;
}

export interface VerifyEmailRequest {
  token: string;
}

export interface ResendVerificationRequest {
  email: string;
}

export interface ForgotPasswordRequest {
  email: string;
}

export interface ResetPasswordRequest {
  token: string;
  new_password: string;
}

export interface ChangePasswordRequest {
  current_password: string;
  new_password: string;
}

export interface RefreshTokenResponse {
  access_token: string;
  expires_in: number;
  token_type: string;
}

export interface Setup2FAResponse {
  secret: string;
  qr_code_url: string;
}

export interface Enable2FAResponse {
  success: boolean;
  message: string;
  recovery_codes: string[]; 
}

// Keep existing
export interface Enable2FARequest {
  code: string;
}

export interface Disable2FARequest {
  password: string;
}

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