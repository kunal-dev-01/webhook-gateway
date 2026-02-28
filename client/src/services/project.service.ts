import { axiosClient } from "@/libs/axios";
import { unwrapResponse } from "@/utils/api";
import type {
  CreateProjectRequest,
  CreateProjectResponse,
  GetProjectResponse,
  ListProjectsResponse,
  UpdateProjectRequest,
  UpdateProjectResponse,
  DeleteProjectResponse,
  ToggleProjectResponse,
  CreateEndpointRequest,
  CreateEndpointResponse,
  GetEndpointResponse,
  ListEndpointsResponse,
  UpdateEndpointRequest,
  UpdateEndpointResponse,
  DeleteEndpointResponse,
  ToggleEndpointResponse,
  AddDomainRequest,
  AddDomainResponse,
  GetDomainResponse,
  ListDomainsResponse,
  DeleteDomainResponse,
  TriggerVerificationResponse,
} from "@/types/project.types";

const orgBase = (orgId: string) => `/user/organizations/${orgId}`;
const projectBase = (orgId: string, projectId: string) =>
  `${orgBase(orgId)}/projects/${projectId}`;
const endpointBase = (orgId: string, projectId: string, endpointId: string) =>
  `${projectBase(orgId, projectId)}/endpoints/${endpointId}`;

export const projectService = {
  // ============================================================================
  // PROJECTS
  // ============================================================================

  createProject: async (
    orgId: string,
    data: CreateProjectRequest,
  ): Promise<CreateProjectResponse> => {
    const response = await axiosClient.post(
      `${orgBase(orgId)}/projects`,
      data,
    );
    return unwrapResponse<CreateProjectResponse>(response);
  },

  listProjects: async (orgId: string): Promise<ListProjectsResponse> => {
    const response = await axiosClient.get(`${orgBase(orgId)}/projects`);
    return unwrapResponse<ListProjectsResponse>(response);
  },

  getProject: async (
    orgId: string,
    projectId: string,
  ): Promise<GetProjectResponse> => {
    const response = await axiosClient.get(projectBase(orgId, projectId));
    return unwrapResponse<GetProjectResponse>(response);
  },

  updateProject: async (
    orgId: string,
    projectId: string,
    data: UpdateProjectRequest,
  ): Promise<UpdateProjectResponse> => {
    const response = await axiosClient.patch(projectBase(orgId, projectId), data);
    return unwrapResponse<UpdateProjectResponse>(response);
  },

  deleteProject: async (
    orgId: string,
    projectId: string,
  ): Promise<DeleteProjectResponse> => {
    const response = await axiosClient.delete(projectBase(orgId, projectId));
    return unwrapResponse<DeleteProjectResponse>(response);
  },

  toggleProject: async (
    orgId: string,
    projectId: string,
  ): Promise<ToggleProjectResponse> => {
    const response = await axiosClient.post(
      `${projectBase(orgId, projectId)}/toggle`,
    );
    return unwrapResponse<ToggleProjectResponse>(response);
  },

  // ============================================================================
  // ENDPOINTS
  // ============================================================================

  createEndpoint: async (
    orgId: string,
    projectId: string,
    data: CreateEndpointRequest,
  ): Promise<CreateEndpointResponse> => {
    const response = await axiosClient.post(
      `${projectBase(orgId, projectId)}/endpoints`,
      data,
    );
    return unwrapResponse<CreateEndpointResponse>(response);
  },

  listEndpoints: async (
    orgId: string,
    projectId: string,
  ): Promise<ListEndpointsResponse> => {
    const response = await axiosClient.get(
      `${projectBase(orgId, projectId)}/endpoints`,
    );
    return unwrapResponse<ListEndpointsResponse>(response);
  },

  getEndpoint: async (
    orgId: string,
    projectId: string,
    endpointId: string,
  ): Promise<GetEndpointResponse> => {
    const response = await axiosClient.get(
      endpointBase(orgId, projectId, endpointId),
    );
    return unwrapResponse<GetEndpointResponse>(response);
  },

  updateEndpoint: async (
    orgId: string,
    projectId: string,
    endpointId: string,
    data: UpdateEndpointRequest,
  ): Promise<UpdateEndpointResponse> => {
    const response = await axiosClient.patch(
      endpointBase(orgId, projectId, endpointId),
      data,
    );
    return unwrapResponse<UpdateEndpointResponse>(response);
  },

  deleteEndpoint: async (
    orgId: string,
    projectId: string,
    endpointId: string,
  ): Promise<DeleteEndpointResponse> => {
    const response = await axiosClient.delete(
      endpointBase(orgId, projectId, endpointId),
    );
    return unwrapResponse<DeleteEndpointResponse>(response);
  },

  toggleEndpoint: async (
    orgId: string,
    projectId: string,
    endpointId: string,
  ): Promise<ToggleEndpointResponse> => {
    const response = await axiosClient.post(
      `${endpointBase(orgId, projectId, endpointId)}/toggle`,
    );
    return unwrapResponse<ToggleEndpointResponse>(response);
  },

  // ============================================================================
  // CUSTOM DOMAINS
  // ============================================================================

  addDomain: async (
    orgId: string,
    data: AddDomainRequest,
  ): Promise<AddDomainResponse> => {
    const response = await axiosClient.post(`${orgBase(orgId)}/domains`, data);
    return unwrapResponse<AddDomainResponse>(response);
  },

  listDomains: async (orgId: string): Promise<ListDomainsResponse> => {
    const response = await axiosClient.get(`${orgBase(orgId)}/domains`);
    return unwrapResponse<ListDomainsResponse>(response);
  },

  getDomain: async (
    orgId: string,
    domainId: string,
  ): Promise<GetDomainResponse> => {
    const response = await axiosClient.get(
      `${orgBase(orgId)}/domains/${domainId}`,
    );
    return unwrapResponse<GetDomainResponse>(response);
  },

  deleteDomain: async (
    orgId: string,
    domainId: string,
  ): Promise<DeleteDomainResponse> => {
    const response = await axiosClient.delete(
      `${orgBase(orgId)}/domains/${domainId}`,
    );
    return unwrapResponse<DeleteDomainResponse>(response);
  },

  triggerVerification: async (
    orgId: string,
    domainId: string,
  ): Promise<TriggerVerificationResponse> => {
    const response = await axiosClient.post(
      `${orgBase(orgId)}/domains/${domainId}/verify`,
    );
    return unwrapResponse<TriggerVerificationResponse>(response);
  },
};