import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { userService } from "@/services/user.service";
import { useAuthStore } from "@/store/auth.store";
import type {
  UpdateUserRequest,
  ListUsersParams,
} from "@/types/user.types";
import { ApiError } from "@/types/api.types";

// ============================================================================
// Get Current User (Me)
// ============================================================================

export const useMe = () => {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  const setUser = useAuthStore((state) => state.setUser);

  const query = useQuery({
    queryKey: ["users", "me"],
    queryFn: userService.getMe,
    enabled: isAuthenticated,
    staleTime: 5 * 60 * 1000, // 5 minutes
    retry: 1,
  });

  // Update auth store with user data
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

// ============================================================================
// Update Current User Profile
// ============================================================================

export const useUpdateProfile = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (data: UpdateUserRequest) => userService.updateProfile(data),
    onSuccess: () => {
      // Invalidate current user queries
      queryClient.invalidateQueries({ queryKey: ["users", "me"] });
      queryClient.invalidateQueries({ queryKey: ["me"] }); // Auth me endpoint
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
    staleTime: 5 * 60 * 1000, // 5 minutes
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
// List Users
// ============================================================================

export const useUsers = (params?: ListUsersParams) => {
  const query = useQuery({
    queryKey: ["users", "list", params],
    queryFn: () => userService.listUsers(params),
    staleTime: 2 * 60 * 1000, // 2 minutes
  });

  return {
    users: query.data?.users,
    total: query.data?.total,
    page: query.data?.page,
    limit: query.data?.limit,
    isPending: query.isPending,
    isSuccess: query.isSuccess,
    isError: query.isError,
    error: query.error as ApiError | null,
    refetch: query.refetch,
  };
};

// ============================================================================
// Update User (Admin/HR)
// ============================================================================

export const useUpdateUser = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: ({ userId, data }: { userId: string; data: UpdateUserRequest }) =>
      userService.updateUser(userId, data),
    onSuccess: (_, variables) => {
      // Invalidate specific user
      queryClient.invalidateQueries({ queryKey: ["users", variables.userId] });
      // Invalidate users list
      queryClient.invalidateQueries({ queryKey: ["users", "list"] });
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
    onSuccess: (_, userId) => {
      // Remove from cache
      queryClient.removeQueries({ queryKey: ["users", userId] });
      // Invalidate users list
      queryClient.invalidateQueries({ queryKey: ["users", "list"] });
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