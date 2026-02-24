import type { PlanTier, BillingPeriod, SubscriptionStatus } from "./org.types";

// Re-export for convenience
export type { PlanTier, BillingPeriod, SubscriptionStatus };

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