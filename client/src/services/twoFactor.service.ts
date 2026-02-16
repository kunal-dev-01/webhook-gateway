import { axiosClient } from "@/libs/axios";
import type {
  TwoFactorStatusResponse,
  Enable2FAResponse,
  Confirm2FARequest,
  Confirm2FAResponse,
  Disable2FARequest,
  Disable2FAResponse,
  RegenerateBackupCodesRequest,
  RegenerateBackupCodesResponse,
} from "@/types/twoFactor.types";

// Helper to unwrap backend response format
function unwrapResponse<T>(response: { data: { success: boolean; data: T } }): T {
  return response.data.data;
}

export const twoFactorService = {
  // ============================================================================
  // GET 2FA STATUS
  // ============================================================================
  getStatus: async (): Promise<TwoFactorStatusResponse> => {
    const response = await axiosClient.get("/user/2fa/status");
    return unwrapResponse<TwoFactorStatusResponse>(response);
  },

  // ============================================================================
  // ENABLE 2FA (Step 1: Get QR Code)
  // ============================================================================
  enable: async (): Promise<Enable2FAResponse> => {
    const response = await axiosClient.post("/user/2fa/enable");
    return unwrapResponse<Enable2FAResponse>(response);
  },

  // ============================================================================
  // CONFIRM 2FA (Step 2: Verify TOTP code and get backup codes)
  // ============================================================================
  confirm: async (data: Confirm2FARequest): Promise<Confirm2FAResponse> => {
    const response = await axiosClient.post("/user/2fa/confirm", data);
    return unwrapResponse<Confirm2FAResponse>(response);
  },

  // ============================================================================
  // DISABLE 2FA
  // ============================================================================
  disable: async (data: Disable2FARequest): Promise<Disable2FAResponse> => {
    const response = await axiosClient.post("/user/2fa/disable", data);
    return unwrapResponse<Disable2FAResponse>(response);
  },

  // ============================================================================
  // REGENERATE BACKUP CODES
  // ============================================================================
  regenerateBackupCodes: async (
    data: RegenerateBackupCodesRequest,
  ): Promise<RegenerateBackupCodesResponse> => {
    const response = await axiosClient.post("/user/2fa/regenerate-codes", data);
    return unwrapResponse<RegenerateBackupCodesResponse>(response);
  },
};