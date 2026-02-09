import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { userService } from "@/services/user.service";
import { useAuthStore } from "@/store/auth.store";
import type {
  UpdateUserRequest,
  ListUsersParams,
} from "@/types/user.types";
import { ApiError } from "@/types/api.types";

// ============================================================================
// Get Current User (Me) - PRODUCTION READY ✅
// ============================================================================

export const useMe = () => {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);

  const query = useQuery({
    queryKey: ["users", "me"],
    queryFn: async () => {
      const data = await userService.getMe();
      // ✅ Update store directly after fetch (runs only on success)
      useAuthStore.getState().setUser(data);
      return data;
    },
    enabled: isAuthenticated,
    staleTime: 5 * 60 * 1000, // 5 minutes
    gcTime: 10 * 60 * 1000,   // 10 minutes (previously cacheTime)
    retry: 1,
    refetchOnWindowFocus: false, // Don't refetch on window focus
    refetchOnReconnect: true,    // Refetch on reconnect
  });

  return {
    user: query.data,
    isPending: query.isPending,
    isSuccess: query.isSuccess,
    isError: query.isError,
    error: query.error as ApiError | null,
    refetch: query.refetch,
  };
};

// ============================================================================
// Update Current User Profile - ENHANCED ✅
// ============================================================================

export const useUpdateProfile = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (data: UpdateUserRequest) => userService.updateProfile(data),
    onSuccess: async () => {
      // ✅ Invalidate queries to trigger refetch
      await queryClient.invalidateQueries({ queryKey: ["users", "me"] });
    },
    onError: (error: ApiError) => {
      console.error("Profile update failed:", error);
    },
  });

  return {
    updateProfile: mutation.mutate,
    updateProfileAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    data: mutation.data,
    reset: mutation.reset,
  };
};

// ============================================================================
// Get User by ID
// ============================================================================

export const useUser = (userId: string, enabled = true) => {
  const query = useQuery({
    queryKey: ["users", userId],
    queryFn: () => userService.getUserById(userId),
    enabled: enabled && !!userId,
    staleTime: 5 * 60 * 1000,
    refetchOnWindowFocus: false,
  });

  return {
    user: query.data,
    isPending: query.isPending,
    isSuccess: query.isSuccess,
    isError: query.isError,
    error: query.error as ApiError | null,
    refetch: query.refetch,
  };
};

// ============================================================================
// List Users - ENHANCED ✅
// ============================================================================

export const useUsers = (params?: ListUsersParams) => {
  const query = useQuery({
    queryKey: ["users", "list", params],
    queryFn: () => userService.listUsers(params),
    staleTime: 2 * 60 * 1000,
    refetchOnWindowFocus: false,
    // ✅ Keep previous data while fetching new page
    placeholderData: (previousData) => previousData,
  });

  return {
    users: query.data?.users ?? [],
    total: query.data?.total ?? 0,
    page: query.data?.page ?? 1,
    limit: query.data?.limit ?? 10,
    isPending: query.isPending,
    isSuccess: query.isSuccess,
    isError: query.isError,
    error: query.error as ApiError | null,
    refetch: query.refetch,
  };
};

// ============================================================================
// Update User (Admin/HR) - ALREADY GOOD ✅
// ============================================================================

export const useUpdateUser = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: ({ userId, data }: { userId: string; data: UpdateUserRequest }) =>
      userService.updateUser(userId, data),
    onSuccess: async (_, variables) => {
      // Invalidate specific user
      await queryClient.invalidateQueries({ queryKey: ["users", variables.userId] });
      // Invalidate users list
      await queryClient.invalidateQueries({ queryKey: ["users", "list"] });
    },
  });

  return {
    updateUser: mutation.mutate,
    updateUserAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    data: mutation.data,
    reset: mutation.reset,
  };
};

// ============================================================================
// Delete User (Admin/HR)
// ============================================================================

export const useDeleteUser = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (userId: string) => userService.deleteUser(userId),
    onSuccess: async (_, userId) => {
      // Remove from cache
      queryClient.removeQueries({ queryKey: ["users", userId] });
      // Invalidate users list
      await queryClient.invalidateQueries({ queryKey: ["users", "list"] });
    },
  });

  return {
    deleteUser: mutation.mutate,
    deleteUserAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    data: mutation.data,
    reset: mutation.reset,
  };
};