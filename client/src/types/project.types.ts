// ============================================================================
// ENUMS
// ============================================================================

export type ProjectEnvironment = "production" | "development";

export type RetryStrategy =
  | "no_retry"
  | "linear"
  | "exponential_backoff"
  | "smart_backoff";

export type CircuitBreakerStatus = "closed" | "open" | "half_open";

export type DomainVerificationStatus =
  | "pending"
  | "verified"
  | "failed"
  | "expired";

export type DomainSSLStatus =
  | "pending"
  | "provisioning"
  | "active"
  | "failed"
  | "expired"
  | "renewing";

export type DomainType = "inbound" | "outbound";

// ============================================================================
// MODELS
// ============================================================================

export interface Project {
  id: string;
  organization_id: string;
  name: string;
  slug: string;
  description?: string;
  environment: ProjectEnvironment;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface Endpoint {
  id: string;
  project_id: string;
  organization_id: string;
  name: string;
  url: string;
  description?: string;
  is_active: boolean;
  max_retries: number;
  retry_strategy: RetryStrategy;
  retry_delay_seconds: number;
  retry_max_delay_seconds: number;
  timeout_seconds: number;
  custom_headers: Record<string, string>;
  verify_ssl: boolean;
  rate_limit_per_second: number;
  last_success_at?: string;
  last_failure_at?: string;
  consecutive_failures: number;
  total_deliveries: number;
  total_failures: number;
  circuit_breaker_status: CircuitBreakerStatus;
  circuit_opened_at?: string;
  circuit_failure_threshold: number;
  circuit_reset_timeout_secs: number;
  created_at: string;
  updated_at: string;
}

export interface CustomDomain {
  id: string;
  organization_id: string;
  project_id?: string;
  domain: string;
  domain_type: DomainType;
  verification_token: string;
  verification_status: DomainVerificationStatus;
  verified_at?: string;
  dns_check_attempts: number;
  dns_check_error?: string;
  ssl_status: DomainSSLStatus;
  ssl_expires_at?: string;
  ssl_auto_renew: boolean;
  ssl_error?: string;
  is_active: boolean;
  created_at: string;
}

// ============================================================================
// REQUESTS
// ============================================================================

export interface CreateProjectRequest {
  name: string;
  slug?: string; // auto-generated from name if omitted
  description?: string;
  environment: ProjectEnvironment;
}

export interface UpdateProjectRequest {
  name?: string;
  description?: string;
}

export interface CreateEndpointRequest {
  name: string;
  url: string;
  description?: string;
  max_retries?: number;
  retry_strategy?: RetryStrategy;
  retry_delay_seconds?: number;
  retry_max_delay_seconds?: number;
  timeout_seconds?: number;
  custom_headers?: Record<string, string>;
  verify_ssl?: boolean;
  rate_limit_per_second?: number;
  circuit_failure_threshold?: number;
  circuit_reset_timeout_secs?: number;
}

export interface UpdateEndpointRequest {
  name?: string;
  url?: string;
  description?: string;
  max_retries?: number;
  timeout_seconds?: number;
  custom_headers?: Record<string, string>;
  verify_ssl?: boolean;
  rate_limit_per_second?: number;
}

export interface AddDomainRequest {
  domain: string;
  project_id?: string;
}

// ============================================================================
// RESPONSES
// ============================================================================

export interface CreateProjectResponse {
  project: Project;
}

export interface GetProjectResponse {
  project: Project;
}

export interface ListProjectsResponse {
  projects: Project[];
  total: number;
}

export interface UpdateProjectResponse {
  project: Project;
}

export interface DeleteProjectResponse {
  message: string;
}

export interface ToggleProjectResponse {
  project: Project;
}

export interface CreateEndpointResponse {
  endpoint: Endpoint;
}

export interface GetEndpointResponse {
  endpoint: Endpoint;
}

export interface ListEndpointsResponse {
  endpoints: Endpoint[];
  total: number;
}

export interface UpdateEndpointResponse {
  endpoint: Endpoint;
}

export interface DeleteEndpointResponse {
  message: string;
}

export interface ToggleEndpointResponse {
  endpoint: Endpoint;
}

export interface AddDomainResponse {
  domain: CustomDomain;
}

export interface GetDomainResponse {
  domain: CustomDomain;
  cname_target: string;   // verify-TOKEN.hooks.yourdomain.com
  instruction: string;    // Human-readable DNS setup guide
}

export interface ListDomainsResponse {
  domains: CustomDomain[];
  total: number;
}

export interface DeleteDomainResponse {
  message: string;
}

export interface TriggerVerificationResponse {
  domain: CustomDomain;
  message: string;
}