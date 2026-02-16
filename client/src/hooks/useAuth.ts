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
  LogoutRequest,
} from "@/types/auth.types";
import { ApiError } from "@/types/api.types";

// ============================================================================
// REGISTER
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
// LOGIN
// ============================================================================

export const useLogin = () => {
  const setAuth = useAuthStore((state) => state.setAuth);
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (data: LoginRequest) => authService.login(data),
    onSuccess: (data) => {
      // Only set auth if 2FA is not required
      if (!data.two_factor_required && data.user && data.access_token) {
        setAuth(data.user, data.access_token);
        queryClient.invalidateQueries({ queryKey: ["users", "me"] });
      }
      // If 2FA required, frontend should redirect to 2FA page with temp_token
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
// VERIFY 2FA (during login)
// ============================================================================

export const useVerify2FA = () => {
  const setAuth = useAuthStore((state) => state.setAuth);
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (data: Verify2FARequest) => authService.verify2FA(data),
    onSuccess: (data) => {
      if (data.user && data.access_token) {
        setAuth(data.user, data.access_token);
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
// EMAIL VERIFICATION
// ============================================================================

export const useVerifyEmail = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (data: VerifyEmailRequest) => authService.verifyEmail(data),
    onSuccess: () => {
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
// RESEND VERIFICATION
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
// FORGOT PASSWORD
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
// RESET PASSWORD
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
// CHANGE PASSWORD
// ============================================================================

export const useChangePassword = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (data: ChangePasswordRequest) =>
      authService.changePassword(data),
    onSuccess: () => {
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
// LOGOUT
// ============================================================================

export const useLogout = () => {
  const clearAuth = useAuthStore((state) => state.clearAuth);
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (data?: LogoutRequest) => authService.logout(data),
    onSuccess: () => {
      clearAuth();
      queryClient.clear();
      window.location.replace("/auth/login");
    },
    onError: (error: ApiError) => {
      console.error("Logout failed:", error);
      // Clear auth even on error
      clearAuth();
      queryClient.clear();
      window.location.replace("/auth/login");
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
// SESSIONS
// ============================================================================

export const useSessions = () => {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);

  const query = useQuery({
    queryKey: ["sessions"],
    queryFn: authService.getSessions,
    enabled: isAuthenticated,
    staleTime: 5 * 60 * 1000,
    gcTime: 10 * 60 * 1000,
    refetchOnWindowFocus: false,
    retry: 1,
  });

  return {
    sessions: query.data?.sessions ?? [],
    total: query.data?.total ?? 0,
    isPending: query.isPending,
    isSuccess: query.isSuccess,
    isError: query.isError,
    error: query.error as ApiError | null,
    refetch: query.refetch,
  };
};

// ============================================================================
// DELETE SESSION
// ============================================================================

export const useDeleteSession = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (sessionId: string) => authService.deleteSession(sessionId),
    onSuccess: () => {
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