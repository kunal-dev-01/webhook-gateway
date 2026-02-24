import { axiosClient } from "@/libs/axios";
import type {
  ListPlansResponse,
  CreateCheckoutRequest,
  CreateCheckoutResponse,
  SubscriptionResponse,
  ChangePlanRequest,
  ListPaymentsParams,
  ListPaymentsResponse,
  UsageResponse,
} from "@/types/billing.types";

// Same unwrap helper as org/auth service
function unwrapResponse<T>(response: { data: { success: boolean; data: T } }): T {
  return response.data.data;
}

export const billingService = {
  // ============================================================================
  // PLANS (Public — no auth needed)
  // ============================================================================

  listPlans: async (): Promise<ListPlansResponse> => {
    const response = await axiosClient.get("/billing/plans");
    return unwrapResponse<ListPlansResponse>(response);
  },

  // ============================================================================
  // CHECKOUT
  // ============================================================================

  createCheckout: async (
    orgId: string,
    data: CreateCheckoutRequest,
  ): Promise<CreateCheckoutResponse> => {
    const response = await axiosClient.post(`/billing/${orgId}/checkout`, data);
    return unwrapResponse<CreateCheckoutResponse>(response);
  },

  // ============================================================================
  // SUBSCRIPTION
  // ============================================================================

  getSubscription: async (orgId: string): Promise<SubscriptionResponse> => {
    const response = await axiosClient.get(`/billing/${orgId}/subscription`);
    return unwrapResponse<SubscriptionResponse>(response);
  },

  cancelSubscription: async (orgId: string): Promise<SubscriptionResponse> => {
    const response = await axiosClient.delete(`/billing/${orgId}/subscription`);
    return unwrapResponse<SubscriptionResponse>(response);
  },

  resumeSubscription: async (orgId: string): Promise<SubscriptionResponse> => {
    const response = await axiosClient.post(`/billing/${orgId}/subscription/resume`);
    return unwrapResponse<SubscriptionResponse>(response);
  },

  changePlan: async (
    orgId: string,
    data: ChangePlanRequest,
  ): Promise<SubscriptionResponse> => {
    const response = await axiosClient.patch(`/billing/${orgId}/subscription/plan`, data);
    return unwrapResponse<SubscriptionResponse>(response);
  },

  // ============================================================================
  // PAYMENTS
  // ============================================================================

  listPayments: async (
    orgId: string,
    params?: ListPaymentsParams,
  ): Promise<ListPaymentsResponse> => {
    const response = await axiosClient.get(`/billing/${orgId}/payments`, { params });
    return unwrapResponse<ListPaymentsResponse>(response);
  },

  // ============================================================================
  // USAGE
  // ============================================================================

  getUsage: async (orgId: string): Promise<UsageResponse> => {
    const response = await axiosClient.get(`/billing/${orgId}/usage`);
    return unwrapResponse<UsageResponse>(response);
  },
};