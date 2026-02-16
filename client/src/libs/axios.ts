import axios, {
  AxiosError,
  type AxiosInstance,
  type InternalAxiosRequestConfig,
} from "axios";
import { useAuthStore } from "@/store/auth.store";
import { ApiError, type ApiErrorResponse } from "@/types/api.types";

const BASE_URL = import.meta.env.VITE_API_BASE_URL!;

/**
 * Backend success response for token refresh
 */
interface RefreshTokenSuccessResponse {
  success: true;
  data: {
    access_token: string;
    refresh_token: string;
    expires_in: number;
  };
}

class AxiosClient {
  private instance: AxiosInstance;
  private isRefreshing = false;
  private refreshSubscribers: ((token: string) => void)[] = [];

  constructor() {
    this.instance = axios.create({
      baseURL: BASE_URL,
      timeout: 30000,
      headers: {
        "Content-Type": "application/json",
      },
      withCredentials: true, // Required for httpOnly cookies
    });

    this.setupInterceptors();
  }

  private setupInterceptors() {
    // ============================================================================
    // REQUEST INTERCEPTOR - Add Bearer token
    // ============================================================================
    this.instance.interceptors.request.use(
      (config: InternalAxiosRequestConfig) => {
        const token = useAuthStore.getState().accessToken;
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
      },
      (error) => Promise.reject(error),
    );

    // ============================================================================
    // RESPONSE INTERCEPTOR - Handle errors and token refresh
    // ============================================================================
    this.instance.interceptors.response.use(
      (response) => response,
      async (error: AxiosError<ApiErrorResponse>) => {
        const originalRequest = error.config as InternalAxiosRequestConfig & {
          _retry?: boolean;
        };

        // ========================================================================
        // HANDLE 401 UNAUTHORIZED - Token Refresh Logic
        // ========================================================================
        if (
          error.response?.status === 401 &&
          !originalRequest._retry &&
          originalRequest.url !== "/auth/refresh" &&
          originalRequest.url !== "/auth/login"
        ) {
          // Queue requests if already refreshing
          if (this.isRefreshing) {
            return new Promise((resolve) => {
              this.refreshSubscribers.push((token: string) => {
                originalRequest.headers.Authorization = `Bearer ${token}`;
                resolve(this.instance(originalRequest));
              });
            });
          }

          originalRequest._retry = true;
          this.isRefreshing = true;

          try {
            // Attempt token refresh
            const { data } = await this.instance.post<RefreshTokenSuccessResponse>(
              "/auth/refresh",
            );

            const newToken = data.data.access_token;
            useAuthStore.getState().setAccessToken(newToken);

            // Retry all queued requests with new token
            this.refreshSubscribers.forEach((callback) => callback(newToken));
            this.refreshSubscribers = [];

            // Retry original request
            originalRequest.headers.Authorization = `Bearer ${newToken}`;
            return this.instance(originalRequest);
          } catch (refreshError) {
            // Refresh failed - logout user
            useAuthStore.getState().clearAuth();
            window.location.href = "/auth/login";
            return Promise.reject(refreshError);
          } finally {
            this.isRefreshing = false;
          }
        }

        // ========================================================================
        // CONVERT TO ApiError
        // ========================================================================
        const apiError = this.createApiError(error);
        return Promise.reject(apiError);
      },
    );
  }

  /**
   * Create ApiError from backend error response
   * Backend format: { success: false, error: { code, message, fields? } }
   */
  private createApiError(error: AxiosError<ApiErrorResponse>): ApiError {
    const status = error.response?.status || 500;
    const responseData = error.response?.data;

    // Network error (no response)
    if (!responseData) {
      return new ApiError(
        error.message || "Network error occurred",
        0,
        status,
      );
    }

    // Backend error format
    if (responseData.success === false && responseData.error) {
      return new ApiError(
        responseData.error.message,
        responseData.error.code,
        status,
        responseData.error.fields,
      );
    }

    // Fallback
    return new ApiError(
      "An unexpected error occurred",
      500,
      status,
    );
  }

  public getAxios(): AxiosInstance {
    return this.instance;
  }
}

// Export singleton instance
export const axiosClient = new AxiosClient().getAxios();