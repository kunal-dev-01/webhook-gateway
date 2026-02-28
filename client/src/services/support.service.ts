import { axiosClient } from "@/libs/axios";
import { unwrapResponse } from "@/utils/api";
import type {
  CreateTicketRequest,
  CreateTicketResponse,
  GetTicketResponse,
  ListTicketsResponse,
  ListTicketsParams,
  AddMessageRequest,
  AddMessageResponse,
  ListMessagesResponse,
  UploadAttachmentResponse,
  GetDownloadURLResponse,
  SubmitSatisfactionRequest,
  SubmitSatisfactionResponse,
  ReopenTicketResponse,
} from "@/types/support.types";

export const supportService = {
  // ============================================================================
  // TICKETS
  // ============================================================================

  createTicket: async (data: CreateTicketRequest): Promise<CreateTicketResponse> => {
    const response = await axiosClient.post("/support/tickets", data);
    return unwrapResponse<CreateTicketResponse>(response);
  },

  listTickets: async (params?: ListTicketsParams): Promise<ListTicketsResponse> => {
    const response = await axiosClient.get("/support/tickets", { params });
    return unwrapResponse<ListTicketsResponse>(response);
  },

  getTicket: async (ticketId: string): Promise<GetTicketResponse> => {
    const response = await axiosClient.get(`/support/tickets/${ticketId}`);
    return unwrapResponse<GetTicketResponse>(response);
  },

  reopenTicket: async (ticketId: string): Promise<ReopenTicketResponse> => {
    const response = await axiosClient.post(`/support/tickets/${ticketId}/reopen`);
    return unwrapResponse<ReopenTicketResponse>(response);
  },

  submitSatisfaction: async (
    ticketId: string,
    data: SubmitSatisfactionRequest,
  ): Promise<SubmitSatisfactionResponse> => {
    const response = await axiosClient.post(`/support/tickets/${ticketId}/satisfaction`, data);
    return unwrapResponse<SubmitSatisfactionResponse>(response);
  },

  // ============================================================================
  // MESSAGES
  // ============================================================================

  addMessage: async (ticketId: string, data: AddMessageRequest): Promise<AddMessageResponse> => {
    const response = await axiosClient.post(`/support/tickets/${ticketId}/messages`, data);
    return unwrapResponse<AddMessageResponse>(response);
  },

  listMessages: async (ticketId: string): Promise<ListMessagesResponse> => {
    const response = await axiosClient.get(`/support/tickets/${ticketId}/messages`);
    return unwrapResponse<ListMessagesResponse>(response);
  },

  // ============================================================================
  // ATTACHMENTS
  // ============================================================================

  uploadAttachment: async (
    ticketId: string,
    messageId: string,
    file: File,
  ): Promise<UploadAttachmentResponse> => {
    const formData = new FormData();
    formData.append("file", file);

    const response = await axiosClient.post(
      `/support/tickets/${ticketId}/messages/${messageId}/attachments`,
      formData,
      { headers: { "Content-Type": "multipart/form-data" } },
    );
    return unwrapResponse<UploadAttachmentResponse>(response);
  },

  getDownloadURL: async (
    ticketId: string,
    attachmentId: string,
  ): Promise<GetDownloadURLResponse> => {
    const response = await axiosClient.get(
      `/support/tickets/${ticketId}/attachments/${attachmentId}/download`,
    );
    return unwrapResponse<GetDownloadURLResponse>(response);
  },
};