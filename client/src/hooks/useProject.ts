import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { projectService } from "@/services/project.service";
import type {
  CreateProjectRequest,
  UpdateProjectRequest,
  CreateEndpointRequest,
  UpdateEndpointRequest,
  AddDomainRequest,
} from "@/types/project.types";
import { ApiError } from "@/types/api.types";

// ============================================================================
// QUERY KEYS
// ============================================================================

export const projectKeys = {
  all: ["projects"] as const,

  // Projects
  lists: (orgId: string) => [...projectKeys.all, orgId, "list"] as const,
  detail: (orgId: string, projectId: string) =>
    [...projectKeys.all, orgId, projectId] as const,

  // Endpoints (nested under project)
  endpoints: (orgId: string, projectId: string) =>
    [...projectKeys.detail(orgId, projectId), "endpoints"] as const,
  endpointDetail: (orgId: string, projectId: string, endpointId: string) =>
    [...projectKeys.endpoints(orgId, projectId), endpointId] as const,

  // Domains (org-level)
  domains: (orgId: string) => ["domains", orgId, "list"] as const,
  domainDetail: (orgId: string, domainId: string) =>
    ["domains", orgId, domainId] as const,
};

// ============================================================================
// LIST PROJECTS
// ============================================================================

export const useListProjects = (orgId: string) => {
  const query = useQuery({
    queryKey: projectKeys.lists(orgId),
    queryFn: () => projectService.listProjects(orgId),
    staleTime: 2 * 60 * 1000,
    enabled: !!orgId,
  });

  return {
    projects: query.data?.projects ?? [],
    total: query.data?.total ?? 0,
    isLoading: query.isLoading,
    isError: query.isError,
    error: query.error as ApiError | null,
    refetch: query.refetch,
  };
};

// ============================================================================
// GET PROJECT
// ============================================================================

export const useGetProject = (orgId: string, projectId: string) => {
  const query = useQuery({
    queryKey: projectKeys.detail(orgId, projectId),
    queryFn: () => projectService.getProject(orgId, projectId),
    staleTime: 2 * 60 * 1000,
    enabled: !!orgId && !!projectId,
  });

  return {
    project: query.data?.project,
    isLoading: query.isLoading,
    isError: query.isError,
    error: query.error as ApiError | null,
    refetch: query.refetch,
  };
};

// ============================================================================
// CREATE PROJECT
// ============================================================================

export const useCreateProject = (orgId: string) => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (data: CreateProjectRequest) =>
      projectService.createProject(orgId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: projectKeys.lists(orgId) });
    },
    onError: (error: ApiError) => {
      console.error("Create project failed:", error);
    },
  });

  return {
    createProject: mutation.mutate,
    createProjectAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    data: mutation.data,
    reset: mutation.reset,
  };
};

// ============================================================================
// UPDATE PROJECT
// ============================================================================

export const useUpdateProject = (orgId: string, projectId: string) => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (data: UpdateProjectRequest) =>
      projectService.updateProject(orgId, projectId, data),
    onSuccess: (res) => {
      queryClient.setQueryData(projectKeys.detail(orgId, projectId), res);
      queryClient.invalidateQueries({ queryKey: projectKeys.lists(orgId) });
    },
    onError: (error: ApiError) => {
      console.error("Update project failed:", error);
    },
  });

  return {
    updateProject: mutation.mutate,
    updateProjectAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    reset: mutation.reset,
  };
};

// ============================================================================
// DELETE PROJECT
// ============================================================================

export const useDeleteProject = (orgId: string) => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (projectId: string) =>
      projectService.deleteProject(orgId, projectId),
    onSuccess: (_, projectId) => {
      // Remove detail cache + invalidate list
      queryClient.removeQueries({
        queryKey: projectKeys.detail(orgId, projectId),
      });
      queryClient.invalidateQueries({ queryKey: projectKeys.lists(orgId) });
    },
    onError: (error: ApiError) => {
      console.error("Delete project failed:", error);
    },
  });

  return {
    deleteProject: mutation.mutate,
    deleteProjectAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    reset: mutation.reset,
  };
};

// ============================================================================
// TOGGLE PROJECT
// ============================================================================

export const useToggleProject = (orgId: string, projectId: string) => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: () => projectService.toggleProject(orgId, projectId),
    onSuccess: (res) => {
      queryClient.setQueryData(projectKeys.detail(orgId, projectId), res);
      queryClient.invalidateQueries({ queryKey: projectKeys.lists(orgId) });
    },
    onError: (error: ApiError) => {
      console.error("Toggle project failed:", error);
    },
  });

  return {
    toggleProject: mutation.mutate,
    toggleProjectAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    reset: mutation.reset,
  };
};

// ============================================================================
// LIST ENDPOINTS
// ============================================================================

export const useListEndpoints = (orgId: string, projectId: string) => {
  const query = useQuery({
    queryKey: projectKeys.endpoints(orgId, projectId),
    queryFn: () => projectService.listEndpoints(orgId, projectId),
    staleTime: 2 * 60 * 1000,
    enabled: !!orgId && !!projectId,
  });

  return {
    endpoints: query.data?.endpoints ?? [],
    total: query.data?.total ?? 0,
    isLoading: query.isLoading,
    isError: query.isError,
    error: query.error as ApiError | null,
    refetch: query.refetch,
  };
};

// ============================================================================
// GET ENDPOINT
// ============================================================================

export const useGetEndpoint = (
  orgId: string,
  projectId: string,
  endpointId: string,
) => {
  const query = useQuery({
    queryKey: projectKeys.endpointDetail(orgId, projectId, endpointId),
    queryFn: () => projectService.getEndpoint(orgId, projectId, endpointId),
    staleTime: 2 * 60 * 1000,
    enabled: !!orgId && !!projectId && !!endpointId,
  });

  return {
    endpoint: query.data?.endpoint,
    isLoading: query.isLoading,
    isError: query.isError,
    error: query.error as ApiError | null,
    refetch: query.refetch,
  };
};

// ============================================================================
// CREATE ENDPOINT
// ============================================================================

export const useCreateEndpoint = (orgId: string, projectId: string) => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (data: CreateEndpointRequest) =>
      projectService.createEndpoint(orgId, projectId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: projectKeys.endpoints(orgId, projectId),
      });
    },
    onError: (error: ApiError) => {
      console.error("Create endpoint failed:", error);
    },
  });

  return {
    createEndpoint: mutation.mutate,
    createEndpointAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    data: mutation.data,
    reset: mutation.reset,
  };
};

// ============================================================================
// UPDATE ENDPOINT
// ============================================================================

export const useUpdateEndpoint = (
  orgId: string,
  projectId: string,
  endpointId: string,
) => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (data: UpdateEndpointRequest) =>
      projectService.updateEndpoint(orgId, projectId, endpointId, data),
    onSuccess: (res) => {
      queryClient.setQueryData(
        projectKeys.endpointDetail(orgId, projectId, endpointId),
        res,
      );
      queryClient.invalidateQueries({
        queryKey: projectKeys.endpoints(orgId, projectId),
      });
    },
    onError: (error: ApiError) => {
      console.error("Update endpoint failed:", error);
    },
  });

  return {
    updateEndpoint: mutation.mutate,
    updateEndpointAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    reset: mutation.reset,
  };
};

// ============================================================================
// DELETE ENDPOINT
// ============================================================================

export const useDeleteEndpoint = (orgId: string, projectId: string) => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (endpointId: string) =>
      projectService.deleteEndpoint(orgId, projectId, endpointId),
    onSuccess: (_, endpointId) => {
      queryClient.removeQueries({
        queryKey: projectKeys.endpointDetail(orgId, projectId, endpointId),
      });
      queryClient.invalidateQueries({
        queryKey: projectKeys.endpoints(orgId, projectId),
      });
    },
    onError: (error: ApiError) => {
      console.error("Delete endpoint failed:", error);
    },
  });

  return {
    deleteEndpoint: mutation.mutate,
    deleteEndpointAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    reset: mutation.reset,
  };
};

// ============================================================================
// TOGGLE ENDPOINT
// ============================================================================

export const useToggleEndpoint = (
  orgId: string,
  projectId: string,
  endpointId: string,
) => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: () =>
      projectService.toggleEndpoint(orgId, projectId, endpointId),
    onSuccess: (res) => {
      queryClient.setQueryData(
        projectKeys.endpointDetail(orgId, projectId, endpointId),
        res,
      );
      queryClient.invalidateQueries({
        queryKey: projectKeys.endpoints(orgId, projectId),
      });
    },
    onError: (error: ApiError) => {
      console.error("Toggle endpoint failed:", error);
    },
  });

  return {
    toggleEndpoint: mutation.mutate,
    toggleEndpointAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    reset: mutation.reset,
  };
};

// ============================================================================
// LIST DOMAINS
// ============================================================================

export const useListDomains = (orgId: string) => {
  const query = useQuery({
    queryKey: projectKeys.domains(orgId),
    queryFn: () => projectService.listDomains(orgId),
    staleTime: 5 * 60 * 1000,
    enabled: !!orgId,
  });

  return {
    domains: query.data?.domains ?? [],
    total: query.data?.total ?? 0,
    isLoading: query.isLoading,
    isError: query.isError,
    error: query.error as ApiError | null,
    refetch: query.refetch,
  };
};

// ============================================================================
// GET DOMAIN
// ============================================================================

export const useGetDomain = (orgId: string, domainId: string) => {
  const query = useQuery({
    queryKey: projectKeys.domainDetail(orgId, domainId),
    queryFn: () => projectService.getDomain(orgId, domainId),
    staleTime: 1 * 60 * 1000,
    enabled: !!orgId && !!domainId,
  });

  return {
    domain: query.data?.domain,
    cnameTarget: query.data?.cname_target,
    instruction: query.data?.instruction,
    isLoading: query.isLoading,
    isError: query.isError,
    error: query.error as ApiError | null,
    refetch: query.refetch,
  };
};

// ============================================================================
// ADD DOMAIN
// ============================================================================

export const useAddDomain = (orgId: string) => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (data: AddDomainRequest) =>
      projectService.addDomain(orgId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: projectKeys.domains(orgId) });
    },
    onError: (error: ApiError) => {
      console.error("Add domain failed:", error);
    },
  });

  return {
    addDomain: mutation.mutate,
    addDomainAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    data: mutation.data,
    reset: mutation.reset,
  };
};

// ============================================================================
// DELETE DOMAIN
// ============================================================================

export const useDeleteDomain = (orgId: string) => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (domainId: string) =>
      projectService.deleteDomain(orgId, domainId),
    onSuccess: (_, domainId) => {
      queryClient.removeQueries({
        queryKey: projectKeys.domainDetail(orgId, domainId),
      });
      queryClient.invalidateQueries({ queryKey: projectKeys.domains(orgId) });
    },
    onError: (error: ApiError) => {
      console.error("Delete domain failed:", error);
    },
  });

  return {
    deleteDomain: mutation.mutate,
    deleteDomainAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    reset: mutation.reset,
  };
};

// ============================================================================
// TRIGGER DNS VERIFICATION
// ============================================================================

export const useTriggerVerification = (orgId: string, domainId: string) => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: () => projectService.triggerVerification(orgId, domainId),
    onSuccess: (res) => {
      // Update domain detail cache directly with new verification state
      queryClient.setQueryData(
        projectKeys.domainDetail(orgId, domainId),
        { domain: res.domain, cname_target: "", instruction: "" },
      );
      queryClient.invalidateQueries({ queryKey: projectKeys.domains(orgId) });
    },
    onError: (error: ApiError) => {
      console.error("Trigger verification failed:", error);
    },
  });

  return {
    triggerVerification: mutation.mutate,
    triggerVerificationAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    data: mutation.data, // contains message + updated domain
    reset: mutation.reset,
  };
};