import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { invitationService } from "@/services/invitation.service";
import { orgKeys } from "@/hooks/useOrg";
import type { InviteMemberRequest, AcceptInvitationRequest } from "@/types/invitation.types";
import type { ApiError } from "@/types/api.types";

// ============================================================================
// QUERY KEYS
// ============================================================================

export const invitationKeys = {
  all: ["invitations"] as const,
  byOrg: (orgId: string) => [...invitationKeys.all, orgId] as const,
};

// ============================================================================
// LIST INVITATIONS
// ============================================================================

export const useListInvitations = (orgId: string) => {
  const query = useQuery({
    queryKey: invitationKeys.byOrg(orgId),
    queryFn: () => invitationService.listInvitations(orgId),
    staleTime: 2 * 60 * 1000,
    enabled: !!orgId,
  });

  return {
    invitations: query.data?.invitations ?? [],
    total: query.data?.total ?? 0,
    isLoading: query.isLoading,
    isError: query.isError,
    error: query.error as ApiError | null,
    refetch: query.refetch,
  };
};

// ============================================================================
// INVITE MEMBER
// ============================================================================

export const useInviteMember = (orgId: string) => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (data: InviteMemberRequest) => invitationService.inviteMember(orgId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: invitationKeys.byOrg(orgId) });
    },
  });

  return {
    inviteMember: mutation.mutate,
    inviteMemberAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    reset: mutation.reset,
  };
};

// ============================================================================
// REVOKE INVITATION
// ============================================================================

export const useRevokeInvitation = (orgId: string) => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (invitationId: string) => invitationService.revokeInvitation(orgId, invitationId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: invitationKeys.byOrg(orgId) });
    },
  });

  return {
    revokeInvitation: mutation.mutate,
    revokeInvitationAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    reset: mutation.reset,
  };
};

// ============================================================================
// ACCEPT INVITATION
// ============================================================================

export const useAcceptInvitation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (data: AcceptInvitationRequest) => invitationService.acceptInvitation(data),
    onSuccess: () => {
      // Invalidate orgs list so newly joined org appears
      queryClient.invalidateQueries({ queryKey: orgKeys.lists() });
    },
  });

  return {
    acceptInvitation: mutation.mutate,
    acceptInvitationAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    data: mutation.data,
    reset: mutation.reset,
  };
};