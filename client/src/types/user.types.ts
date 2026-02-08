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