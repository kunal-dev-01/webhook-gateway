export type UserRole = "admin" | "hr" | "accountant" | "viewer";

export type UserStatus = "active" | "inactive" | "suspended" | "deleted";

export interface User {
  id: string;
  email: string;
  full_name: string;
  phone?: string;
  role: UserRole;
  status: UserStatus;
  email_verified: boolean;
  two_factor_enabled: boolean;
  last_login_at?: string;
  created_at: string;
  password_changed_at: string;
}

// ============================================================================
// Request Types
// ============================================================================

export interface UpdateUserRequest {
  full_name?: string;
  phone?: string | null;
  role?: UserRole;
  status?: UserStatus;
  timezone?: string;
  language?: string;
}

export interface ListUsersParams {
  search?: string;
  status?: UserStatus;
  role?: UserRole;
  email_verified?: boolean;
  sort_by?: string;
  sort_desc?: boolean;
  limit?: number;
  offset?: number;
  page?: number;
}

// ============================================================================
// Response Types
// ============================================================================

export interface UserResponse {
  id: string;
  email: string;
  full_name: string;
  phone?: string;
  role: UserRole;
  status: UserStatus;
  email_verified: boolean;
  two_factor_enabled: boolean;
  last_login_at?: string;
  created_at: string;
}

export interface MeResponse {
  id: string;
  email: string;
  full_name: string;
  phone?: string;
  role: UserRole;
  status: UserStatus;
  email_verified: boolean;
  two_factor_enabled: boolean;
  last_login_at?: string;
  created_at: string;
  password_changed_at: string;
}

export interface ListUsersResponse {
  users: UserResponse[];
  total: number;
  page: number;
  limit: number;
}