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

export const useRegister = () => {
  const mutation = useMutation({
    mutationFn: (data: RegisterRequest) => authService.register(data),
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

export const useLogin = () => {
  const setAuth = useAuthStore((state) => state.setAuth);

  const mutation = useMutation({
    mutationFn: (data: LoginRequest) => authService.login(data),
    onSuccess: (data) => {
      if (!data.require_2fa && data.user && data.access_token) {
        setAuth(data.user, data.access_token);
      }
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

export const useVerify2FA = () => {
  const setAuth = useAuthStore((state) => state.setAuth);

  const mutation = useMutation({
    mutationFn: (data: Verify2FARequest) => authService.verify2FA(data),
    onSuccess: (data) => {
      if (data.user && data.access_token) {
        setAuth(data.user, data.access_token);
      }
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

export const useVerifyEmail = () => {
  const mutation = useMutation({
    mutationFn: (data: VerifyEmailRequest) => authService.verifyEmail(data),
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

export const useResendVerification = () => {
  const mutation = useMutation({
    mutationFn: (data: ResendVerificationRequest) =>
      authService.resendVerification(data),
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

export const useForgotPassword = () => {
  const mutation = useMutation({
    mutationFn: (data: ForgotPasswordRequest) =>
      authService.forgotPassword(data),
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

export const useResetPassword = () => {
  const mutation = useMutation({
    mutationFn: (data: ResetPasswordRequest) => authService.resetPassword(data),
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

export const useChangePassword = () => {
  const mutation = useMutation({
    mutationFn: (data: ChangePasswordRequest) =>
      authService.changePassword(data),
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

export const useMe = () => {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  const setUser = useAuthStore((state) => state.setUser);

  const query = useQuery({
    queryKey: ["me"],
    queryFn: authService.getMe,
    enabled: isAuthenticated,
    staleTime: 5 * 60 * 1000,
    retry: 1,
  });

  if (query.data && query.isSuccess) {
    setUser(query.data);
  }

  return {
    user: query.data,
    isPending: query.isPending,
    isSuccess: query.isSuccess,
    isError: query.isError,
    error: query.error as ApiError | null,
    refetch: query.refetch,
  };
};

export const useLogout = () => {
  const clearAuth = useAuthStore((state) => state.clearAuth);
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: authService.logout,
    onSuccess: () => {
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

export const useLogoutAll = () => {
  const clearAuth = useAuthStore((state) => state.clearAuth);
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: authService.logoutAll,
    onSuccess: () => {
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

export const useSetup2FA = () => {
  const mutation = useMutation({
    mutationFn: authService.setup2FA,
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

export const useEnable2FA = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (data: Enable2FARequest) => authService.enable2FA(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["me"] });
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

export const useDisable2FA = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (data: Disable2FARequest) => authService.disable2FA(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["me"] });
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

export const useSessions = () => {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);

  const query = useQuery({
    queryKey: ["sessions"],
    queryFn: authService.getSessions,
    enabled: isAuthenticated,
  });

  return {
    sessions: query.data,
    isPending: query.isPending,
    isSuccess: query.isSuccess,
    isError: query.isError,
    error: query.error as ApiError | null,
    refetch: query.refetch,
  };
};

export const useDeleteSession = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (sessionId: string) => authService.deleteSession(sessionId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["sessions"] });
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