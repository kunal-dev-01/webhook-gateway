// src/hooks/useAddon.ts

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { addonService } from "@/services/addon.service";
import { billingKeys } from "@/hooks/useBilling";
import type { CreateAddonCheckoutRequest } from "@/types/billing.types";
import { ApiError } from "@/types/api.types";

// ============================================================================
// QUERY KEYS
// ============================================================================

export const addonKeys = {
  all: ["addons"] as const,
  available: (orgId: string) => [...addonKeys.all, orgId, "available"] as const,
  orgAddons: (orgId: string) => [...addonKeys.all, orgId, "list"] as const,
  limits: (orgId: string) => [...addonKeys.all, orgId, "limits"] as const,
};

// ============================================================================
// LIST AVAILABLE ADDONS (pricing/purchase page ke liye)
// ============================================================================

export const useListAvailableAddons = (orgId: string) => {
  const query = useQuery({
    queryKey: addonKeys.available(orgId),
    queryFn: () => addonService.listAvailableAddons(orgId),
    staleTime: 10 * 60 * 1000, // Addon catalog rarely changes
    enabled: !!orgId,
    refetchOnWindowFocus: false,
  });

  return {
    addons: query.data?.addons ?? [],
    isLoading: query.isLoading,
    isError: query.isError,
    error: query.error as ApiError | null,
    refetch: query.refetch,
  };
};

// ============================================================================
// LIST ORG ADDONS (dashboard — active + past)
// ============================================================================

export const useListOrgAddons = (orgId: string) => {
  const query = useQuery({
    queryKey: addonKeys.orgAddons(orgId),
    queryFn: () => addonService.listOrgAddons(orgId),
    staleTime: 5 * 60 * 1000,
    enabled: !!orgId,
  });

  const activeAddons = query.data?.addons.filter((a) => a.status === "active") ?? [];
  const canceledAddons = query.data?.addons.filter((a) => a.status === "canceled") ?? [];
  const expiredAddons = query.data?.addons.filter((a) => a.status === "expired") ?? [];

  return {
    addons: query.data?.addons ?? [],
    activeAddons,
    canceledAddons,
    expiredAddons,
    hasActiveAddons: activeAddons.length > 0,
    isLoading: query.isLoading,
    isError: query.isError,
    error: query.error as ApiError | null,
    refetch: query.refetch,
  };
};

// ============================================================================
// GET EFFECTIVE LIMITS (base plan + addons)
// ============================================================================

export const useEffectiveLimits = (orgId: string) => {
  const query = useQuery({
    queryKey: addonKeys.limits(orgId),
    queryFn: () => addonService.getEffectiveLimits(orgId),
    staleTime: 2 * 60 * 1000,
    enabled: !!orgId,
    refetchInterval: 5 * 60 * 1000,
  });

  const data = query.data;

  return {
    limits: data,

    // Effective values (jo actually enforce hote hain)
    effectiveEventsLimit: data?.effective_monthly_events_limit ?? 0,
    effectiveRetentionDays: data?.effective_retention_days ?? 0,
    effectiveSeatsLimit: data?.effective_team_seats_limit ?? 0,
    effectiveProjectsLimit: data?.effective_projects_limit ?? 0,

    // Addon bonuses
    eventsBonus: data?.addon_events_bonus ?? 0,
    retentionBonus: data?.addon_retention_bonus ?? 0,
    seatsBonus: data?.addon_seats_bonus ?? 0,
    projectsBonus: data?.addon_projects_bonus ?? 0,

    // Has any active addon bonus
    hasAddons: (data?.active_addons.length ?? 0) > 0,
    activeAddons: data?.active_addons ?? [],

    isLoading: query.isLoading,
    isError: query.isError,
    error: query.error as ApiError | null,
    refetch: query.refetch,
  };
};

// ============================================================================
// CREATE ADDON CHECKOUT
// ============================================================================

export const useCreateAddonCheckout = (orgId: string) => {
  const mutation = useMutation({
    mutationFn: (data: CreateAddonCheckoutRequest) =>
      addonService.createAddonCheckout(orgId, data),
    onSuccess: (data) => {
      // Main subscription checkout jaisa hi — LS checkout pe redirect
      window.location.href = data.checkout_url;
    },
    onError: (error: ApiError) => {
      console.error("Create addon checkout failed:", error);
    },
  });

  return {
    createCheckout: mutation.mutate,
    createCheckoutAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    reset: mutation.reset,
  };
};

// ============================================================================
// CANCEL ADDON
// ============================================================================

export const useCancelAddon = (orgId: string) => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (addonId: string) => addonService.cancelAddon(orgId, addonId),
    onSuccess: () => {
      // Addons list refresh karo
      queryClient.invalidateQueries({ queryKey: addonKeys.orgAddons(orgId) });
      // Limits bhi update honge (cancel_at_period_end set hua)
      queryClient.invalidateQueries({ queryKey: addonKeys.limits(orgId) });
      // Billing usage bhi invalidate karo
      queryClient.invalidateQueries({ queryKey: billingKeys.usage(orgId) });
    },
    onError: (error: ApiError) => {
      console.error("Cancel addon failed:", error);
    },
  });

  return {
    cancelAddon: mutation.mutate,
    cancelAddonAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    reset: mutation.reset,
  };
};