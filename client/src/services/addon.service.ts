// src/services/addon.service.ts

import { axiosClient } from "@/libs/axios";
import {unwrapResponse} from "@/utils/api";
import type {
  ListAddonPlansResponse,
  ListOrgAddonsResponse,
  CreateAddonCheckoutRequest,
  CreateCheckoutResponse,
  EffectiveLimitsResponse,
  OrganizationAddon,
} from "@/types/billing.types";

export const addonService = {
  // ============================================================================
  // AVAILABLE ADDONS (compatible with org's current plan)
  // ============================================================================
  listAvailableAddons: async (orgId: string): Promise<ListAddonPlansResponse> => {
    const response = await axiosClient.get(`/billing/${orgId}/addons/available`);
    return unwrapResponse<ListAddonPlansResponse>(response);
  },

  // ============================================================================
  // ORG KE ACTIVE/PAST ADDONS
  // ============================================================================
  listOrgAddons: async (orgId: string): Promise<ListOrgAddonsResponse> => {
    const response = await axiosClient.get(`/billing/${orgId}/addons`);
    return unwrapResponse<ListOrgAddonsResponse>(response);
  },

  // ============================================================================
  // EFFECTIVE LIMITS (base + addons)
  // ============================================================================
  getEffectiveLimits: async (orgId: string): Promise<EffectiveLimitsResponse> => {
    const response = await axiosClient.get(`/billing/${orgId}/addons/limits`);
    return unwrapResponse<EffectiveLimitsResponse>(response);
  },

  // ============================================================================
  // ADDON CHECKOUT
  // ============================================================================
  createAddonCheckout: async (
    orgId: string,
    data: CreateAddonCheckoutRequest,
  ): Promise<CreateCheckoutResponse> => {
    const response = await axiosClient.post(`/billing/${orgId}/addons/checkout`, data);
    return unwrapResponse<CreateCheckoutResponse>(response);
  },

  // ============================================================================
  // CANCEL ADDON
  // ============================================================================
  cancelAddon: async (
    orgId: string,
    addonId: string,
  ): Promise<{ addon: OrganizationAddon }> => {
    const response = await axiosClient.delete(`/billing/${orgId}/addons/${addonId}`);
    return unwrapResponse<{ addon: OrganizationAddon }>(response);
  },
};