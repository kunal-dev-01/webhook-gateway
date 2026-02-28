import { axiosClient } from "@/libs/axios";
import {unwrapResponse} from "@/utils/api";
import type {
  Invitation,
  InviteMemberRequest,
  AcceptInvitationRequest,
  ListInvitationsResponse,
  AcceptInvitationResponse,
} from "@/types/invitation.types";
import type { MessageResponse } from "@/types/api.types";

export const invitationService = {
  // POST /user/organizations/:orgId/invitations
  inviteMember: async (orgId: string, data: InviteMemberRequest): Promise<Invitation> => {
    const response = await axiosClient.post(`/user/organizations/${orgId}/invitations`, data);
    return unwrapResponse<Invitation>(response);
  },

  // GET /user/organizations/:orgId/invitations
  listInvitations: async (orgId: string): Promise<ListInvitationsResponse> => {
    const response = await axiosClient.get(`/user/organizations/${orgId}/invitations`);
    return unwrapResponse<ListInvitationsResponse>(response);
  },

  // DELETE /user/organizations/:orgId/invitations/:invitationId
  revokeInvitation: async (orgId: string, invitationId: string): Promise<MessageResponse> => {
    const response = await axiosClient.delete(
      `/user/organizations/${orgId}/invitations/${invitationId}`,
    );
    return unwrapResponse<MessageResponse>(response);
  },

  // POST /invitations/accept
  acceptInvitation: async (data: AcceptInvitationRequest): Promise<AcceptInvitationResponse> => {
    const response = await axiosClient.post("/user/invitations/accept", data);
    return unwrapResponse<AcceptInvitationResponse>(response);
  },
};