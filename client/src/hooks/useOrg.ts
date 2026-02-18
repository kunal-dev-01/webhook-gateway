import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { orgService } from "@/services/org.service";
import type {
  ListMembersParams,
  UpdateOrgNameRequest,
  UpdateMemberRoleRequest,
} from "@/types/org.types";
import { ApiError } from "@/types/api.types";

// ============================================================================
// QUERY KEYS
// ============================================================================

export const orgKeys = {
  all: ["organizations"] as const,
  lists: () => [...orgKeys.all, "list"] as const,
  detail: (orgId: string) => [...orgKeys.all, orgId] as const,
  members: (orgId: string) => [...orgKeys.all, orgId, "members"] as const,
};

// ============================================================================
// LIST MY ORGANIZATIONS
// ============================================================================

export const useListOrganizations = () => {
  const query = useQuery({
    queryKey: orgKeys.lists(),
    queryFn: () => orgService.listMyOrganizations(),
    staleTime: 5 * 60 * 1000,
  });

  return {
    organizations: query.data?.organizations ?? [],
    total: query.data?.total ?? 0,
    isLoading: query.isLoading,
    isError: query.isError,
    error: query.error as ApiError | null,
    refetch: query.refetch,
  };
};

// ============================================================================
// GET ORGANIZATION
// ============================================================================

export const useGetOrganization = (orgId: string) => {
  const query = useQuery({
    queryKey: orgKeys.detail(orgId),
    queryFn: () => orgService.getOrganization(orgId),
    staleTime: 5 * 60 * 1000,
    enabled: !!orgId,
  });

  return {
    organization: query.data?.organization,
    isLoading: query.isLoading,
    isError: query.isError,
    error: query.error as ApiError | null,
    refetch: query.refetch,
  };
};

// ============================================================================
// UPDATE ORG NAME
// ============================================================================

export const useUpdateOrgName = (orgId: string) => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (data: UpdateOrgNameRequest) => orgService.updateName(orgId, data),
    onSuccess: (res) => {
      // Update detail cache directly
      queryClient.setQueryData(orgKeys.detail(orgId), res);
      // Invalidate list so name reflects there too
      queryClient.invalidateQueries({ queryKey: orgKeys.lists() });
    },
    onError: (error: ApiError) => {
      console.error("Update org name failed:", error);
    },
  });

  return {
    updateName: mutation.mutate,
    updateNameAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    reset: mutation.reset,
  };
};

// ============================================================================
// LIST MEMBERS
// ============================================================================

export const useListMembers = (orgId: string, params?: ListMembersParams) => {
  const query = useQuery({
    queryKey: [...orgKeys.members(orgId), params],
    queryFn: () => orgService.listMembers(orgId, params),
    staleTime: 2 * 60 * 1000,
    enabled: !!orgId,
  });

  return {
    members: query.data?.members ?? [],
    total: query.data?.total ?? 0,
    limit: query.data?.limit ?? 20,
    offset: query.data?.offset ?? 0,
    isLoading: query.isLoading,
    isError: query.isError,
    error: query.error as ApiError | null,
    refetch: query.refetch,
  };
};

// ============================================================================
// UPDATE MEMBER ROLE
// ============================================================================

export const useUpdateMemberRole = (orgId: string) => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: ({ userId, data }: { userId: string; data: UpdateMemberRoleRequest }) =>
      orgService.updateMemberRole(orgId, userId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: orgKeys.members(orgId) });
    },
    onError: (error: ApiError) => {
      console.error("Update member role failed:", error);
    },
  });

  return {
    updateRole: mutation.mutate,
    updateRoleAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    reset: mutation.reset,
  };
};

// ============================================================================
// REMOVE MEMBER
// ============================================================================

export const useRemoveMember = (orgId: string) => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (userId: string) => orgService.removeMember(orgId, userId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: orgKeys.members(orgId) });
    },
    onError: (error: ApiError) => {
      console.error("Remove member failed:", error);
    },
  });

  return {
    removeMember: mutation.mutate,
    removeMemberAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    reset: mutation.reset,
  };
};