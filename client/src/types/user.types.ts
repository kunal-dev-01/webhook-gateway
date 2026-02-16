export interface UserPreferences {
  theme: string;
  email_notifications: boolean;
  timezone: string;
}

export interface User {
  id: string;
  email: string;
  full_name: string;
  avatar_url: string;
  email_verified: boolean;
  two_factor_enabled: boolean;
  two_factor_confirmed_at: Date;
  preferences: UserPreferences;
  created_at: Date; 
}