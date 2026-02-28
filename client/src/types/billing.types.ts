import type { PlanTier, BillingPeriod, SubscriptionStatus } from "./org.types";

// Re-export for convenience
export type { PlanTier, BillingPeriod, SubscriptionStatus };


export type AddonType = "events" | "retention" | "seats" | "projects";
export type AddonStatus = "active" | "canceled" | "expired" | "paused";

// ============================================================================
// PLAN FEATURES
// ============================================================================

export interface PlanFeatures {
  retry_policy: "manual" | "auto_3x" | "smart_backoff" | "smart_custom";
  support: "community" | "email_48h" | "priority_email" | "slack_priority";
  custom_webhooks: boolean;
  analytics: boolean;
  custom_domains: boolean;
  dedicated_ip: boolean;
}

// ============================================================================
// SUBSCRIPTION PLAN
// ============================================================================

export interface SubscriptionPlan {
  id: string;
  tier: PlanTier;
  name: string;
  description: string;
  billing_period: BillingPeriod;
  price: number;
  price_per_month: number;
  currency: string;
  discount_percentage: number;
  savings_amount: number;
  monthly_events_limit: number;
  retention_days: number;
  team_seats_limit: number;
  rate_limit_per_second: number;
  projects_limit: number;
  features: PlanFeatures;
  is_popular: boolean;
  // Enriched fields from backend
  formatted_price: string;
  formatted_price_per_month: string;
  savings_text?: string;
}

// ============================================================================
// SUBSCRIPTION
// ============================================================================

export interface Subscription {
  id: string;
  organization_id: string;
  plan_id: string;
  status: SubscriptionStatus;
  current_period_start: string;
  current_period_end: string;
  cancel_at_period_end: boolean;
  canceled_at?: string;
  paused_at?: string;
  price_amount: number;
  currency: string;
  created_at: string;
}

// ============================================================================
// PAYMENT
// ============================================================================

export type PaymentStatus =
  | "pending"
  | "paid"
  | "failed"
  | "refunded"
  | "partially_refunded";

export interface Payment {
  id: string;
  organization_id: string;
  subscription_id?: string;
  amount: number;
  currency: string;
  status: PaymentStatus;
  payment_method?: string;
  invoice_url?: string;
  receipt_url?: string;
  refunded_amount: number;
  refunded_at?: string;
  paid_at?: string;
  created_at: string;
}

// ============================================================================
// USAGE
// ============================================================================

export interface UsageInfo {
  organization_id: string;
  monthly_events_limit: number;
  monthly_events_used: number;
  remaining_events: number;
  usage_percentage: number;
  retention_days: number;
  team_seats_limit: number;
  projects_limit: number;
  rate_limit_per_second: number;
}

// ============================================================================
// REQUESTS
// ============================================================================

export interface CreateCheckoutRequest {
  plan_tier: PlanTier;
  billing_period: BillingPeriod;
}

export interface ChangePlanRequest {
  plan_tier: PlanTier;
  billing_period: BillingPeriod;
}

export interface ListPaymentsParams {
  limit?: number;
  offset?: number;
}

// ============================================================================
// RESPONSES
// ============================================================================

export interface ListPlansResponse {
  plans: SubscriptionPlan[];
}

export interface CreateCheckoutResponse {
  checkout_url: string;
  expires_at: string;
}

export interface SubscriptionResponse {
  subscription?: Subscription;
  plan?: SubscriptionPlan;
  customer_portal_url?: string;
}

export interface ListPaymentsResponse {
  payments: Payment[];
  total: number;
  limit: number;
  offset: number;
}

export interface UsageResponse {
  organization_id: string;
  monthly_events_limit: number;
  monthly_events_used: number;
  remaining_events: number;
  usage_percentage: number;
  retention_days: number;
  team_seats_limit: number;
  projects_limit: number;
  rate_limit_per_second: number;
}


// ============================================================================
// ADDON TYPES 
// ============================================================================


// Addon catalog (available addons)
export interface AddonPlan {
  id: string;
  name: string;
  description?: string;
  addon_type: AddonType;
  quantity: number;
  quantity_label: string; // "+100K events", "+7 days" etc.
  price: number;
  currency: string;
  compatible_plans: PlanTier[];
  display_order: number;
}

// Org ka active addon subscription
export interface OrganizationAddon {
  id: string;
  organization_id: string;
  addon_plan_id: string;
  addon_type: AddonType;
  quantity: number;
  price_amount: number;
  currency: string;
  status: AddonStatus;
  current_period_start: string;
  current_period_end: string;
  cancel_at_period_end: boolean;
  canceled_at?: string;
  created_at: string;
}

// Base plan + active addons ka total
export interface EffectiveLimits {
  organization_id: string;
  base_plan_tier: PlanTier;

  // Base limits
  base_monthly_events_limit: number;
  base_retention_days: number;
  base_team_seats_limit: number;
  base_projects_limit: number;

  // Addon bonuses
  addon_events_bonus: number;
  addon_retention_bonus: number;
  addon_seats_bonus: number;
  addon_projects_bonus: number;

  // Effective totals (enforce karne wale)
  effective_monthly_events_limit: number;
  effective_retention_days: number;
  effective_team_seats_limit: number;
  effective_projects_limit: number;

  active_addons: OrganizationAddon[];
}

// ── Requests ─────────────────────────────────────────────────────────────────

export interface CreateAddonCheckoutRequest {
  addon_plan_id: string;
}

// ── Responses ─────────────────────────────────────────────────────────────────

export interface ListAddonPlansResponse {
  addons: AddonPlan[];
}

export interface ListOrgAddonsResponse {
  addons: OrganizationAddon[];
}

export interface EffectiveLimitsResponse {
  organization_id: string;
  base_plan_tier: PlanTier;
  base_monthly_events_limit: number;
  base_retention_days: number;
  base_team_seats_limit: number;
  base_projects_limit: number;
  addon_events_bonus: number;
  addon_retention_bonus: number;
  addon_seats_bonus: number;
  addon_projects_bonus: number;
  effective_monthly_events_limit: number;
  effective_retention_days: number;
  effective_team_seats_limit: number;
  effective_projects_limit: number;
  active_addons: OrganizationAddon[];
}