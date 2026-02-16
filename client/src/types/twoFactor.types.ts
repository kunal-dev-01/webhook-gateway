// ====================== TWO-FACTOR AUTHENTICATION TYPES ======================

export interface TwoFactorStatusResponse {
  enabled: boolean;
  confirmed_at?: string;
}

export interface Enable2FAResponse {
  qr_code_url: string; // Base64 data URL
  secret: string; // For manual entry
}

export interface Confirm2FARequest {
  code: string; // TOTP code to confirm setup
}

export interface Confirm2FAResponse {
  message: string;
  backup_codes: string[]; // Show once, user must save
}

export interface Disable2FARequest {
  password: string; // Require password for security
  code?: string; // TOTP code
  recovery_code?: string; // OR recovery code
}

export interface Disable2FAResponse {
  message: string;
}

export interface RegenerateBackupCodesRequest {
  password: string;
}

export interface RegenerateBackupCodesResponse {
  backup_codes: string[];
}