import { axiosClient } from "@/libs/axios";
import type {
  ListOrganizationsResponse,
  GetOrganizationResponse,
  ListMembersResponse,
  ListMembersParams,
  UpdateOrgNameRequest,
  UpdateMemberRoleRequest,
} from "@/types/org.types";

function unwrapResponse<T>(response: { data: { success: boolean; data: T } }): T {
  return response.data.data;
}

export const orgService = {
  // ============================================================================
  // LIST MY ORGANIZATIONS
  // ============================================================================
  listMyOrganizations: async (): Promise<ListOrganizationsResponse> => {
    const response = await axiosClient.get("/user/organizations");
    return unwrapResponse<ListOrganizationsResponse>(response);
  },

  // ============================================================================
  // GET ORGANIZATION
  // ============================================================================
  getOrganization: async (orgId: string): Promise<GetOrganizationResponse> => {
    const response = await axiosClient.get(`/user/organizations/${orgId}`);
    return unwrapResponse<GetOrganizationResponse>(response);
  },

  // ============================================================================
  // UPDATE ORG NAME
  // ============================================================================
  updateName: async (orgId: string, data: UpdateOrgNameRequest): Promise<GetOrganizationResponse> => {
    const response = await axiosClient.patch(`/user/organizations/${orgId}`, data);
    return unwrapResponse<GetOrganizationResponse>(response);
  },

  // ============================================================================
  // LIST MEMBERS
  // ============================================================================
  listMembers: async (orgId: string, params?: ListMembersParams): Promise<ListMembersResponse> => {
    const response = await axiosClient.get(`/user/organizations/${orgId}/members`, { params });
    return unwrapResponse<ListMembersResponse>(response);
  },

  // ============================================================================
  // UPDATE MEMBER ROLE
  // ============================================================================
  updateMemberRole: async (
    orgId: string,
    userId: string,
    data: UpdateMemberRoleRequest,
  ): Promise<{ member: import("@/types/org.types").OrgMember }> => {
    const response = await axiosClient.patch(
      `/user/organizations/${orgId}/members/${userId}/role`,
      data,
    );
    return unwrapResponse(response);
  },

  // ============================================================================
  // REMOVE MEMBER
  // ============================================================================
  removeMember: async (orgId: string, userId: string): Promise<{ message: string }> => {
    const response = await axiosClient.delete(`/user/organizations/${orgId}/members/${userId}`);
    return unwrapResponse(response);
  },
};