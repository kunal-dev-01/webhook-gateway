import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { userService } from "@/services/user.service";
import { useAuthStore } from "@/store/auth.store";
import type {
  UpdateUserRequest,
  UpdatePreferencesRequest,
} from "@/types/user.types";
import { ApiError } from "@/types/api.types";

// ============================================================================
// GET ME
// ============================================================================

export const useGetMe = () => {
  const query = useQuery({
    queryKey: ["users", "me"],
    queryFn: () => userService.getMe(),
    staleTime: 5 * 60 * 1000, // 5 min
  });

  return {
    user: query.data?.user,
    isLoading: query.isLoading,
    isError: query.isError,
    error: query.error as ApiError | null,
    refetch: query.refetch,
  };
};

// ============================================================================
// UPDATE FULL NAME
// ============================================================================

export const useUpdateUser = () => {
  const queryClient = useQueryClient();
  const setUser = useAuthStore((s) => s.setUser);
  const currentUser = useAuthStore((s) => s.user);

  const mutation = useMutation({
    mutationFn: (data: UpdateUserRequest) => userService.updateUser(data),
    onSuccess: (res) => {
      // Update store
      if (currentUser) {
        setUser({
          ...currentUser,
          full_name: res.full_name
        });
      }
      queryClient.invalidateQueries({ queryKey: ["users", "me"] });
    },
    onError: (error: ApiError) => {
      console.error("Update user failed:", error);
    },
  });

  return {
    updateUser: mutation.mutate,
    updateUserAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    reset: mutation.reset,
  };
};

// ============================================================================
// UPDATE AVATAR
// ============================================================================

export const useUpdateAvatar = () => {
  const queryClient = useQueryClient();
  const setUser = useAuthStore((s) => s.setUser);
  const currentUser = useAuthStore((s) => s.user);

  const mutation = useMutation({
    mutationFn: (file: File) => userService.updateAvatar(file),
    onSuccess: (res) => {
      // Update store
      if (currentUser) {
        setUser({ ...currentUser, avatar_url: res.avatar_url });
      }
      queryClient.invalidateQueries({ queryKey: ["users", "me"] });
    },
    onError: (error: ApiError) => {
      console.error("Update avatar failed:", error);
    },
  });

  return {
    updateAvatar: mutation.mutate,
    updateAvatarAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    reset: mutation.reset,
  };
};

// ============================================================================
// UPDATE PREFERENCES
// ============================================================================

export const useUpdatePreferences = () => {
  const queryClient = useQueryClient();
  const setUser = useAuthStore((s) => s.setUser);
  const currentUser = useAuthStore((s) => s.user);

  const mutation = useMutation({
    mutationFn: (data: UpdatePreferencesRequest) =>
      userService.updatePreferences(data),
    onSuccess: (res) => {
      // Update store
      if (currentUser) {
        setUser({ ...currentUser, preferences: res.preferences });
      }
      queryClient.invalidateQueries({ queryKey: ["users", "me"] });
    },
    onError: (error: ApiError) => {
      console.error("Update preferences failed:", error);
    },
  });

  return {
    updatePreferences: mutation.mutate,
    updatePreferencesAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    data: mutation.data,
    reset: mutation.reset,
  };
};