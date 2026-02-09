import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { authService } from "@/services/auth.service";
import { useAuthStore } from "@/store/auth.store";
import type {
  RegisterRequest,
  LoginRequest,
  Verify2FARequest,
  VerifyEmailRequest,
  ResendVerificationRequest,
  ForgotPasswordRequest,
  ResetPasswordRequest,
  ChangePasswordRequest,
  Enable2FARequest,
  Disable2FARequest,
} from "@/types/auth.types";
import { ApiError } from "@/types/api.types";

// ============================================================================
// Registration
// ============================================================================

export const useRegister = () => {
  const mutation = useMutation({
    mutationFn: (data: RegisterRequest) => authService.register(data),
    onError: (error: ApiError) => {
      console.error("Registration failed:", error);
    },
  });

  return {
    register: mutation.mutate,
    registerAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    data: mutation.data,
    reset: mutation.reset,
  };
};

// ============================================================================
// Login - ENHANCED ✅
// ============================================================================

export const useLogin = () => {
  const setAuth = useAuthStore((state) => state.setAuth);
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (data: LoginRequest) => authService.login(data),
    onSuccess: (data) => {
      if (!data.require_2fa && data.user && data.access_token) {
        setAuth(data.user, data.access_token);
        // ✅ Prefetch user data after login
        queryClient.invalidateQueries({ queryKey: ["users", "me"] });
      }
    },
    onError: (error: ApiError) => {
      console.error("Login failed:", error);
    },
  });

  return {
    login: mutation.mutate,
    loginAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    data: mutation.data,
    reset: mutation.reset,
  };
};

// ============================================================================
// Verify 2FA - ENHANCED ✅
// ============================================================================

export const useVerify2FA = () => {
  const setAuth = useAuthStore((state) => state.setAuth);
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (data: Verify2FARequest) => authService.verify2FA(data),
    onSuccess: (data) => {
      if (data.user && data.access_token) {
        setAuth(data.user, data.access_token);
        // ✅ Prefetch user data after 2FA verification
        queryClient.invalidateQueries({ queryKey: ["users", "me"] });
      }
    },
    onError: (error: ApiError) => {
      console.error("2FA verification failed:", error);
    },
  });

  return {
    verify2FA: mutation.mutate,
    verify2FAAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    data: mutation.data,
    reset: mutation.reset,
  };
};

// ============================================================================
// Email Verification
// ============================================================================

export const useVerifyEmail = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (data: VerifyEmailRequest) => authService.verifyEmail(data),
    onSuccess: () => {
      // ✅ Refresh user data to show email_verified: true
      queryClient.invalidateQueries({ queryKey: ["users", "me"] });
    },
    onError: (error: ApiError) => {
      console.error("Email verification failed:", error);
    },
  });

  return {
    verifyEmail: mutation.mutate,
    verifyEmailAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    data: mutation.data,
    reset: mutation.reset,
  };
};

// ============================================================================
// Resend Verification
// ============================================================================

export const useResendVerification = () => {
  const mutation = useMutation({
    mutationFn: (data: ResendVerificationRequest) =>
      authService.resendVerification(data),
    onError: (error: ApiError) => {
      console.error("Resend verification failed:", error);
    },
  });

  return {
    resendVerification: mutation.mutate,
    resendVerificationAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    data: mutation.data,
    reset: mutation.reset,
  };
};

// ============================================================================
// Forgot Password
// ============================================================================

export const useForgotPassword = () => {
  const mutation = useMutation({
    mutationFn: (data: ForgotPasswordRequest) =>
      authService.forgotPassword(data),
    onError: (error: ApiError) => {
      console.error("Forgot password failed:", error);
    },
  });

  return {
    forgotPassword: mutation.mutate,
    forgotPasswordAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    data: mutation.data,
    reset: mutation.reset,
  };
};

// ============================================================================
// Reset Password
// ============================================================================

export const useResetPassword = () => {
  const mutation = useMutation({
    mutationFn: (data: ResetPasswordRequest) => authService.resetPassword(data),
    onError: (error: ApiError) => {
      console.error("Reset password failed:", error);
    },
  });

  return {
    resetPassword: mutation.mutate,
    resetPasswordAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    data: mutation.data,
    reset: mutation.reset,
  };
};

// ============================================================================
// Change Password - FIXED ✅
// ============================================================================

export const useChangePassword = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (data: ChangePasswordRequest) =>
      authService.changePassword(data),
    onSuccess: () => {
      // ✅ FIXED: Invalidate user data to refresh password_changed_at
      queryClient.invalidateQueries({ queryKey: ["users", "me"] });
    },
    onError: (error: ApiError) => {
      console.error("Change password failed:", error);
    },
  });

  return {
    changePassword: mutation.mutate,
    changePasswordAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    data: mutation.data,
    reset: mutation.reset,
  };
};

// ============================================================================
// Logout - ENHANCED ✅
// ============================================================================

export const useLogout = () => {
  const clearAuth = useAuthStore((state) => state.clearAuth);
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: authService.logout,
    onSuccess: () => {
      clearAuth();
      // ✅ Clear all cached queries
      queryClient.clear();
    },
    onError: (error: ApiError) => {
      // ✅ Even on error, clear auth (cookie might be invalid)
      console.error("Logout failed:", error);
      clearAuth();
      queryClient.clear();
    },
  });

  return {
    logout: mutation.mutate,
    logoutAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    reset: mutation.reset,
  };
};

// ============================================================================
// Logout All - ENHANCED ✅
// ============================================================================

export const useLogoutAll = () => {
  const clearAuth = useAuthStore((state) => state.clearAuth);
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: authService.logoutAll,
    onSuccess: () => {
      clearAuth();
      queryClient.clear();
    },
    onError: (error: ApiError) => {
      // ✅ Even on error, clear auth
      console.error("Logout all failed:", error);
      clearAuth();
      queryClient.clear();
    },
  });

  return {
    logoutAll: mutation.mutate,
    logoutAllAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    reset: mutation.reset,
  };
};

// ============================================================================
// Setup 2FA
// ============================================================================

export const useSetup2FA = () => {
  const mutation = useMutation({
    mutationFn: authService.setup2FA,
    onError: (error: ApiError) => {
      console.error("Setup 2FA failed:", error);
    },
  });

  return {
    setup2FA: mutation.mutate,
    setup2FAAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    data: mutation.data,
    reset: mutation.reset,
  };
};

// ============================================================================
// Enable 2FA - FIXED ✅
// ============================================================================

export const useEnable2FA = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (data: Enable2FARequest) => authService.enable2FA(data),
    onSuccess: () => {
      // ✅ FIXED: Correct query key to match useMe hook
      queryClient.invalidateQueries({ queryKey: ["users", "me"] });
    },
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
    data: mutation.data,
    reset: mutation.reset,
  };
};

// ============================================================================
// Disable 2FA - FIXED ✅
// ============================================================================

export const useDisable2FA = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (data: Disable2FARequest) => authService.disable2FA(data),
    onSuccess: () => {
      // ✅ FIXED: Correct query key to match useMe hook
      queryClient.invalidateQueries({ queryKey: ["users", "me"] });
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
// Sessions - ENHANCED ✅
// ============================================================================

export const useSessions = () => {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);

  const query = useQuery({
    queryKey: ["sessions"],
    queryFn: authService.getSessions,
    enabled: isAuthenticated,
    staleTime: 5 * 60 * 1000, // ✅ 5 minutes
    gcTime: 10 * 60 * 1000,   // ✅ 10 minutes (cache time)
    refetchOnWindowFocus: false, // ✅ Don't refetch on focus
    retry: 1, // ✅ Only retry once on failure
  });

  return {
    sessions: query.data?.sessions ?? [], // ✅ Safe default
    total: query.data?.total ?? 0,        // ✅ Safe default
    isPending: query.isPending,
    isSuccess: query.isSuccess,
    isError: query.isError,
    error: query.error as ApiError | null,
    refetch: query.refetch,
  };
};

// ============================================================================
// Delete Session
// ============================================================================

export const useDeleteSession = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (sessionId: string) => authService.deleteSession(sessionId),
    onSuccess: () => {
      // ✅ Refetch sessions list
      queryClient.invalidateQueries({ queryKey: ["sessions"] });
    },
    onError: (error: ApiError) => {
      console.error("Delete session failed:", error);
    },
  });

  return {
    deleteSession: mutation.mutate,
    deleteSessionAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    reset: mutation.reset,
  };
};