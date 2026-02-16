import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { twoFactorService } from "@/services/twoFactor.service";
import { useAuthStore } from "@/store/auth.store";
import type {
  Confirm2FARequest,
  Disable2FARequest,
  RegenerateBackupCodesRequest,
} from "@/types/twoFactor.types";
import { ApiError } from "@/types/api.types";

// ============================================================================
// GET 2FA STATUS
// ============================================================================

export const use2FAStatus = () => {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);

  const query = useQuery({
    queryKey: ["2fa", "status"],
    queryFn: twoFactorService.getStatus,
    enabled: isAuthenticated,
    staleTime: 5 * 60 * 1000,
    gcTime: 10 * 60 * 1000,
    refetchOnWindowFocus: false,
    retry: 1,
  });

  return {
    status: query.data,
    isPending: query.isPending,
    isSuccess: query.isSuccess,
    isError: query.isError,
    error: query.error as ApiError | null,
    refetch: query.refetch,
  };
};

// ============================================================================
// ENABLE 2FA (Step 1: Get QR Code)
// ============================================================================

export const useEnable2FA = () => {
  const mutation = useMutation({
    mutationFn: () => twoFactorService.enable(),
    onError: (error: ApiError) => {
      console.error("Enable 2FA failed:", error);
    },
  });

  return {
    enable2FA: mutation.mutate,
    enable2FAAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    data: mutation.data, // { qr_code_url, secret }
    reset: mutation.reset,
  };
};

// ============================================================================
// CONFIRM 2FA (Step 2: Verify TOTP and get backup codes)
// ============================================================================

export const useConfirm2FA = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (data: Confirm2FARequest) => twoFactorService.confirm(data),
    onSuccess: () => {
      // Refresh user data and 2FA status
      queryClient.invalidateQueries({ queryKey: ["users", "me"] });
      queryClient.invalidateQueries({ queryKey: ["2fa", "status"] });
    },
    onError: (error: ApiError) => {
      console.error("Confirm 2FA failed:", error);
    },
  });

  return {
    confirm2FA: mutation.mutate,
    confirm2FAAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    data: mutation.data, // { message, backup_codes }
    reset: mutation.reset,
  };
};

// ============================================================================
// DISABLE 2FA
// ============================================================================

export const useDisable2FA = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (data: Disable2FARequest) => twoFactorService.disable(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["users", "me"] });
      queryClient.invalidateQueries({ queryKey: ["2fa", "status"] });
    },
    onError: (error: ApiError) => {
      console.error("Disable 2FA failed:", error);
    },
  });

  return {
    disable2FA: mutation.mutate,
    disable2FAAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    data: mutation.data,
    reset: mutation.reset,
  };
};

// ============================================================================
// REGENERATE BACKUP CODES
// ============================================================================

export const useRegenerateBackupCodes = () => {
  const mutation = useMutation({
    mutationFn: (data: RegenerateBackupCodesRequest) =>
      twoFactorService.regenerateBackupCodes(data),
    onError: (error: ApiError) => {
      console.error("Regenerate backup codes failed:", error);
    },
  });

  return {
    regenerateBackupCodes: mutation.mutate,
    regenerateBackupCodesAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    data: mutation.data, // { backup_codes }
    reset: mutation.reset,
  };
};