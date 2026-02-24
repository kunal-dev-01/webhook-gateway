import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { billingService } from "@/services/billing.service";
import type {
  CreateCheckoutRequest,
  ChangePlanRequest,
  ListPaymentsParams,
} from "@/types/billing.types";
import { ApiError } from "@/types/api.types";

// ============================================================================
// QUERY KEYS
// ============================================================================

export const billingKeys = {
  all: ["billing"] as const,
  plans: () => [...billingKeys.all, "plans"] as const,
  subscription: (orgId: string) => [...billingKeys.all, orgId, "subscription"] as const,
  payments: (orgId: string) => [...billingKeys.all, orgId, "payments"] as const,
  paymentsWithParams: (orgId: string, params?: ListPaymentsParams) =>
    [...billingKeys.payments(orgId), params] as const,
  usage: (orgId: string) => [...billingKeys.all, orgId, "usage"] as const,
};

// ============================================================================
// LIST PLANS (Public — pricing page)
// ============================================================================

export const useListPlans = () => {
  const query = useQuery({
    queryKey: billingKeys.plans(),
    queryFn: () => billingService.listPlans(),
    staleTime: 10 * 60 * 1000, // Plans rarely change — 10 min cache
    gcTime: 30 * 60 * 1000,
    refetchOnWindowFocus: false,
  });

  return {
    plans: query.data?.plans ?? [],
    isLoading: query.isLoading,
    isError: query.isError,
    error: query.error as ApiError | null,
    refetch: query.refetch,
  };
};

// ============================================================================
// CREATE CHECKOUT
// ============================================================================

export const useCreateCheckout = (orgId: string) => {
  const mutation = useMutation({
    mutationFn: (data: CreateCheckoutRequest) =>
      billingService.createCheckout(orgId, data),
    onSuccess: (data) => {
      // Redirect to LemonSqueezy checkout
      window.location.href = data.checkout_url;
    },
    onError: (error: ApiError) => {
      console.error("Create checkout failed:", error);
    },
  });

  return {
    createCheckout: mutation.mutate,
    createCheckoutAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    data: mutation.data,
    reset: mutation.reset,
  };
};

// ============================================================================
// GET SUBSCRIPTION
// ============================================================================

export const useGetSubscription = (orgId: string) => {
  const query = useQuery({
    queryKey: billingKeys.subscription(orgId),
    queryFn: () => billingService.getSubscription(orgId),
    staleTime: 5 * 60 * 1000,
    enabled: !!orgId,
  });

  return {
    subscription: query.data?.subscription,
    plan: query.data?.plan,
    customerPortalUrl: query.data?.customer_portal_url,
    isLoading: query.isLoading,
    isError: query.isError,
    error: query.error as ApiError | null,
    refetch: query.refetch,
  };
};

// ============================================================================
// CANCEL SUBSCRIPTION
// ============================================================================

export const useCancelSubscription = (orgId: string) => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: () => billingService.cancelSubscription(orgId),
    onSuccess: (data) => {
      // Update cache directly with new subscription state
      queryClient.setQueryData(billingKeys.subscription(orgId), data);
    },
    onError: (error: ApiError) => {
      console.error("Cancel subscription failed:", error);
    },
  });

  return {
    cancelSubscription: mutation.mutate,
    cancelSubscriptionAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    reset: mutation.reset,
  };
};

// ============================================================================
// RESUME SUBSCRIPTION
// ============================================================================

export const useResumeSubscription = (orgId: string) => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: () => billingService.resumeSubscription(orgId),
    onSuccess: (data) => {
      queryClient.setQueryData(billingKeys.subscription(orgId), data);
    },
    onError: (error: ApiError) => {
      console.error("Resume subscription failed:", error);
    },
  });

  return {
    resumeSubscription: mutation.mutate,
    resumeSubscriptionAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    reset: mutation.reset,
  };
};

// ============================================================================
// CHANGE PLAN (Upgrade / Downgrade)
// ============================================================================

export const useChangePlan = (orgId: string) => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (data: ChangePlanRequest) =>
      billingService.changePlan(orgId, data),
    onSuccess: (data) => {
      // Update subscription cache
      queryClient.setQueryData(billingKeys.subscription(orgId), data);
      // Usage limits changed — invalidate usage
      queryClient.invalidateQueries({ queryKey: billingKeys.usage(orgId) });
    },
    onError: (error: ApiError) => {
      console.error("Change plan failed:", error);
    },
  });

  return {
    changePlan: mutation.mutate,
    changePlanAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    data: mutation.data,
    reset: mutation.reset,
  };
};

// ============================================================================
// LIST PAYMENTS
// ============================================================================

export const useListPayments = (orgId: string, params?: ListPaymentsParams) => {
  const query = useQuery({
    queryKey: billingKeys.paymentsWithParams(orgId, params),
    queryFn: () => billingService.listPayments(orgId, params),
    staleTime: 5 * 60 * 1000,
    enabled: !!orgId,
  });

  return {
    payments: query.data?.payments ?? [],
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
// GET USAGE
// ============================================================================

export const useGetUsage = (orgId: string) => {
  const query = useQuery({
    queryKey: billingKeys.usage(orgId),
    queryFn: () => billingService.getUsage(orgId),
    staleTime: 2 * 60 * 1000, // Usage changes frequently — 2 min
    enabled: !!orgId,
    refetchInterval: 5 * 60 * 1000, // Auto-refresh every 5 min
  });

  return {
    usage: query.data,
    usagePercentage: query.data?.usage_percentage ?? 0,
    isNearLimit: (query.data?.usage_percentage ?? 0) >= 80,  // Warn at 80%
    isAtLimit: (query.data?.usage_percentage ?? 0) >= 100,
    isLoading: query.isLoading,
    isError: query.isError,
    error: query.error as ApiError | null,
    refetch: query.refetch,
  };
};