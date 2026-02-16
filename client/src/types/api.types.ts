// ======================= API TYPES =======================

/**
 * Backend Success Response Format
 * { success: true, data: {...} }
 */
export interface ApiSuccessResponse<T = unknown> {
  success: true;
  data: T;
}

/**
 * Backend Error Response Format
 * { success: false, error: { code, message, fields? } }
 */
export interface ApiErrorResponse {
  success: false;
  error: {
    code: number;
    message: string;
    fields?: Record<string, string>;
  };
}

/**
 * Generic API Response (union of success and error)
 */
export type ApiResponse<T = unknown> = 
  | ApiSuccessResponse<T> 
  | ApiErrorResponse;

/**
 * Simple message response (used by many endpoints)
 */
export interface MessageResponse {
  message: string;
}

/**
 * Custom API Error class for consistent error handling
 */
export class ApiError extends Error {
  code: number;
  status: number;
  fields?: Record<string, string>;

  constructor(
    message: string,
    code: number,
    status: number,
    fields?: Record<string, string>,
  ) {
    super(message);
    this.name = "ApiError";
    this.code = code;
    this.status = status;
    this.fields = fields;

    // Maintains proper stack trace
    if ((Error as any).captureStackTrace) {
      (Error as any).captureStackTrace(this, ApiError);
    }
  }

  /**
   * Check if this is a validation error (has field errors)
   */
  isValidationError(): boolean {
    return !!this.fields && Object.keys(this.fields).length > 0;
  }

  /**
   * Check if this is a specific HTTP status error
   */
  isStatus(status: number): boolean {
    return this.status === status;
  }

  /**
   * Get user-friendly error message with validation details
   */
  getUserMessage(): string {
    if (this.isValidationError()) {
      const fieldErrors = Object.entries(this.fields!)
        .map(([field, error]) => `${field}: ${error}`)
        .join(", ");
      return `${this.message} - ${fieldErrors}`;
    }
    return this.message;
  }

  /**
   * Get validation error for a specific field
   */
  getFieldError(field: string): string | undefined {
    return this.fields?.[field];
  }

  /**
   * Check if error is client-side (4xx)
   */
  isClientError(): boolean {
    return this.status >= 400 && this.status < 500;
  }

  /**
   * Check if error is server-side (5xx)
   */
  isServerError(): boolean {
    return this.status >= 500;
  }
}

/**
 * Type guard to check if response is success
 */
export function isSuccessResponse<T>(
  response: ApiResponse<T>,
): response is ApiSuccessResponse<T> {
  return response.success === true;
}

/**
 * Type guard to check if response is error
 */
export function isErrorResponse(
  response: ApiResponse,
): response is ApiErrorResponse {
  return response.success === false;
}