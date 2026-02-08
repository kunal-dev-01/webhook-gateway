import axios, {
  AxiosError,
  type AxiosInstance,
  type InternalAxiosRequestConfig,
} from "axios";
import { useAuthStore } from "@/store/auth.store";
import { ApiError, type ApiErrorResponse } from "@/types/api.types";

const BASE_URL = import.meta.env.VITE_API_URL || "http://localhost:8080/api/v1";

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
      withCredentials: true,
    });

    this.setupInterceptors();
  }

  private setupInterceptors() {
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

    this.instance.interceptors.response.use(
      (response) => response,
      async (error: AxiosError<ApiErrorResponse>) => {
        const originalRequest = error.config as InternalAxiosRequestConfig & {
          _retry?: boolean;
        };

        if (
          error.response?.status === 401 &&
          !originalRequest._retry &&
          originalRequest.url !== "/auth/refresh" &&
          originalRequest.url !== "/auth/login"
        ) {
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
            const { data } = await this.instance.post<{
              access_token: string;
            }>("/auth/refresh");

            const newToken = data.access_token;
            useAuthStore.getState().setAccessToken(newToken);

            this.refreshSubscribers.forEach((callback) => callback(newToken));
            this.refreshSubscribers = [];

            originalRequest.headers.Authorization = `Bearer ${newToken}`;
            return this.instance(originalRequest);
          } catch (refreshError) {
            useAuthStore.getState().clearAuth();
            window.location.href = "/auth/login";
            return Promise.reject(refreshError);
          } finally {
            this.isRefreshing = false;
          }
        }

        const apiError = new ApiError(
          error.response?.data?.error || error.message || "An error occurred",
          error.response?.data?.code,
          error.response?.status,
        );

        return Promise.reject(apiError);
      },
    );
  }

  public getAxios() {
    return this.instance;
  }
}

export const axiosClient = new AxiosClient().getAxios();