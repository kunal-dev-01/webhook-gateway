// ============================================================================
// ENUMS
// ============================================================================

export type PlanTier = "developer" | "starter" | "growth" | "scale";
export type BillingPeriod = "monthly" | "annual";
export type SubscriptionStatus = "active" | "past_due" | "canceled" | "paused" | "expired";
export type OrgMemberRole = "owner" | "admin" | "member" | "viewer";

// ============================================================================
// ORGANIZATION
// ============================================================================

export interface Organization {
  id: string;
  name: string;
  slug: string;
  owner_id: string;
  role: OrgMemberRole; // caller's role in this org
  plan_tier: PlanTier;
  billing_period: BillingPeriod;
  subscription_status: SubscriptionStatus;
  monthly_events_limit: number;
  monthly_events_used: number;
  retention_days: number;
  team_seats_limit: number;
  rate_limit_per_second: number;
  projects_limit: number;
  created_at: string;
}

// ============================================================================
// MEMBER
// ============================================================================

export interface OrgMember {
  id: string;
  organization_id: string;
  user_id: string;
  role: OrgMemberRole;
  invited_by?: string;
  invitation_accepted_at?: string;
  joined_at: string;
}

// ============================================================================
// REQUESTS
// ============================================================================

export interface UpdateOrgNameRequest {
  name: string;
}

export interface UpdateMemberRoleRequest {
  role: Exclude<OrgMemberRole, "owner">; // owner assign nahi hoga
}

export interface ListMembersParams {
  limit?: number;
  offset?: number;
}

// ============================================================================
// RESPONSES
// ============================================================================

export interface ListOrganizationsResponse {
  organizations: Organization[];
  total: number;
}

export interface GetOrganizationResponse {
  organization: Organization;
}

export interface ListMembersResponse {
  members: OrgMember[];
  total: number;
  limit: number;
  offset: number;
}