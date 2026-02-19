import type { OrgMemberRole, Organization, OrgMember } from "@/types/org.types";

// ============================================================================
// ENUMS
// ============================================================================

export type InvitationStatus = "pending" | "accepted" | "expired" | "revoked";

// ============================================================================
// MODELS
// ============================================================================

export interface InvitationMetadata {
  custom_message?: string;
  inviter_name?: string;
  org_name?: string;
}

export interface Invitation {
  id: string;
  organization_id: string;
  email: string;
  role: OrgMemberRole;
  invited_by: string;
  status: InvitationStatus;
  expires_at: string;
  metadata: InvitationMetadata;
  created_at: string;
}

// ============================================================================
// REQUESTS
// ============================================================================

export interface InviteMemberRequest {
  email: string;
  role: Exclude<OrgMemberRole, "owner">;
  custom_message?: string;
}

export interface AcceptInvitationRequest {
  token: string;
}

// ============================================================================
// RESPONSES
// ============================================================================

export interface ListInvitationsResponse {
  invitations: Invitation[];
  total: number;
}

export interface AcceptInvitationResponse {
  organization: Organization;
  member: OrgMember;
}