// ============================================================================
// ENUMS
// ============================================================================

export type TicketCategory =
  | "general"
  | "billing"
  | "technical"
  | "account"
  | "feature_request"
  | "bug";

export type TicketPriority = "low" | "normal" | "high" | "urgent";

export type TicketStatus =
  | "open"
  | "in_progress"
  | "waiting"
  | "resolved"
  | "closed";

export type MessageSenderType = "user" | "agent" | "system";

export type ScanStatus = "pending" | "clean" | "infected" | "error";

// ============================================================================
// MODELS
// ============================================================================

export interface Ticket {
  id: string;
  ticket_code: string; // TKT-1001
  ticket_number: number;
  subject: string;
  category: TicketCategory;
  priority: TicketPriority;
  status: TicketStatus;
  related_project_id?: string;
  related_endpoint_id?: string;
  first_response_at?: string;
  resolved_at?: string;
  satisfaction_rating?: number;
  satisfaction_comment?: string;
  is_breaching_sla: boolean;
  created_at: string;
  updated_at: string;
}

export interface TicketAttachment {
  id: string;
  ticket_id: string;
  message_id: string;
  file_name: string;
  file_size_bytes: number;
  file_size_mb: number;
  mime_type: string;
  is_image: boolean;
  scan_status: ScanStatus;
  is_downloadable: boolean;
  created_at: string;
}

export interface TicketMessage {
  id: string;
  ticket_id: string;
  sender_id?: string;
  sender_type: MessageSenderType;
  body: string; // Markdown
  is_edited: boolean;
  edited_at?: string;
  attachments?: TicketAttachment[];
  created_at: string;
}

// ============================================================================
// REQUESTS
// ============================================================================

export interface TicketMetadata {
  browser?: string;
  os?: string;
  app_version?: string;
}

export interface CreateTicketRequest {
  subject: string;
  category: TicketCategory;
  priority?: TicketPriority; // default: "normal"
  body: string;
  related_project_id?: string;
  related_endpoint_id?: string;
  metadata?: TicketMetadata;
}

export interface AddMessageRequest {
  body: string;
}

export interface SubmitSatisfactionRequest {
  rating: number; // 1-5
  comment?: string;
}

export interface ListTicketsParams {
  status?: TicketStatus;
  category?: TicketCategory;
  page?: number;
  limit?: number;
}

// ============================================================================
// RESPONSES
// ============================================================================

export interface CreateTicketResponse {
  ticket: Ticket;
  message: TicketMessage;
}

export interface GetTicketResponse {
  ticket: Ticket;
}

export interface ListTicketsResponse {
  tickets: Ticket[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}

export interface AddMessageResponse {
  message: TicketMessage;
}

export interface ListMessagesResponse {
  messages: TicketMessage[];
  total: number;
}

export interface UploadAttachmentResponse {
  attachment: TicketAttachment;
}

export interface GetDownloadURLResponse {
  download_url: string;
  expires_at: string;
}

export interface SubmitSatisfactionResponse {
  rating: number;
  comment?: string;
}

export interface ReopenTicketResponse {
  ticket: Ticket;
}