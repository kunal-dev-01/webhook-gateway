//=========================== USER TYPES ===========================//


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

export interface UpdateUserRequest {
  full_name: string;
}

export interface UpdateUserResponse {
  full_name: string;
}

export interface UpdateAvatarResponse {
  avatar_url: string;
}

export interface UpdatePreferencesRequest {
  preferences: {
    theme: string;
    email_notifications: boolean;
    timezone: string;
  };
}

export interface UpdatePreferencesResponse {
  preferences: {
    theme: string;
    email_notifications: boolean;
    timezone: string;
  };
}

export interface GetMeResponse {
  user: User;
}
