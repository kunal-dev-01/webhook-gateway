// ============================================================================
// ENUMS
// ============================================================================

export type NotificationType =
  | "invitation"
  | "invitation_accepted"
  | "invitation_declined"
  | "team_member_joined"
  | "team_member_removed"
  | "billing_failed"
  | "subscription_expiring"
  | "subscription_renewed"
  | "subscription_canceled"
  | "payment_succeeded"
  | "payment_failed"
  | "usage_limit_warning"
  | "usage_limit_exceeded";

export type NotificationPriority = "low" | "normal" | "high" | "urgent";

// ============================================================================
// MODELS
// ============================================================================

export interface NotificationMetadata {
  inviter_name?: string;
  org_name?: string;
  plan_name?: string;
  amount?: string;
  days_left?: number;
  usage_percent?: number;
}

export interface Notification {
  id: string;
  type: NotificationType;
  title: string;
  message: string;
  action_url?: string;
  action_label?: string;
  related_organization_id?: string;
  metadata: NotificationMetadata;
  is_read: boolean;
  read_at?: string;
  priority: NotificationPriority;
  expires_at?: string;
  created_at: string;
}

// ============================================================================
// REQUESTS
// ============================================================================

export interface ListNotificationsParams {
  limit?: number;
  offset?: number;
}

// ============================================================================
// RESPONSES
// ============================================================================

export interface ListNotificationsResponse {
  notifications: Notification[];
  total: number;
  limit: number;
  offset: number;
}

export interface UnreadCountResponse {
  unread_count: number;
}