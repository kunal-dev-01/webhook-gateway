export interface ApiSuccessResponse<T = unknown> {
  success: true;
  data?: T;
  message?: string;
}

export interface ApiErrorResponse {
  success: false;
  error: string;
  code?: string;
}

export type ApiResponse<T = unknown> = ApiSuccessResponse<T> | ApiErrorResponse;

export interface MessageResponse {
  success: boolean;
  message: string;
}

export class ApiError extends Error {
  code?: string;
  status?: number;

  constructor(
    message: string,
    code?: string,
    status?: number,
  ) {
    super(message);
    this.name = "ApiError";
    this.code = code;
    this.status = status;
  }
}