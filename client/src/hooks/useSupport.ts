import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { supportService } from "@/services/support.service";
import type {
  CreateTicketRequest,
  AddMessageRequest,
  SubmitSatisfactionRequest,
  ListTicketsParams,
} from "@/types/support.types";
import { ApiError } from "@/types/api.types";

// ============================================================================
// QUERY KEYS
// ============================================================================

export const supportKeys = {
  all: ["support"] as const,
  tickets: () => [...supportKeys.all, "tickets"] as const,
  ticketList: (params?: ListTicketsParams) => [...supportKeys.tickets(), "list", params] as const,
  ticketDetail: (ticketId: string) => [...supportKeys.tickets(), ticketId] as const,
  messages: (ticketId: string) => [...supportKeys.ticketDetail(ticketId), "messages"] as const,
};

// ============================================================================
// LIST TICKETS
// ============================================================================

export const useListTickets = (params?: ListTicketsParams) => {
  const query = useQuery({
    queryKey: supportKeys.ticketList(params),
    queryFn: () => supportService.listTickets(params),
    staleTime: 2 * 60 * 1000,
  });

  return {
    tickets: query.data?.tickets ?? [],
    total: query.data?.total ?? 0,
    page: query.data?.page ?? 1,
    limit: query.data?.limit ?? 20,
    totalPages: query.data?.total_pages ?? 0,
    isLoading: query.isLoading,
    isError: query.isError,
    error: query.error as ApiError | null,
    refetch: query.refetch,
  };
};

// ============================================================================
// GET TICKET
// ============================================================================

export const useGetTicket = (ticketId: string) => {
  const query = useQuery({
    queryKey: supportKeys.ticketDetail(ticketId),
    queryFn: () => supportService.getTicket(ticketId),
    staleTime: 1 * 60 * 1000,
    enabled: !!ticketId,
  });

  return {
    ticket: query.data?.ticket,
    isLoading: query.isLoading,
    isError: query.isError,
    error: query.error as ApiError | null,
    refetch: query.refetch,
  };
};

// ============================================================================
// CREATE TICKET
// ============================================================================

export const useCreateTicket = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (data: CreateTicketRequest) => supportService.createTicket(data),
    onSuccess: () => {
      // Invalidate list so new ticket shows up
      queryClient.invalidateQueries({ queryKey: supportKeys.tickets() });
    },
    onError: (error: ApiError) => {
      console.error("Create ticket failed:", error);
    },
  });

  return {
    createTicket: mutation.mutate,
    createTicketAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    data: mutation.data,
    reset: mutation.reset,
  };
};

// ============================================================================
// REOPEN TICKET
// ============================================================================

export const useReopenTicket = (ticketId: string) => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: () => supportService.reopenTicket(ticketId),
    onSuccess: (res) => {
      // Update detail cache directly — no extra refetch needed
      queryClient.setQueryData(supportKeys.ticketDetail(ticketId), res);
      queryClient.invalidateQueries({ queryKey: supportKeys.ticketList() });
    },
    onError: (error: ApiError) => {
      console.error("Reopen ticket failed:", error);
    },
  });

  return {
    reopenTicket: mutation.mutate,
    reopenTicketAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    reset: mutation.reset,
  };
};

// ============================================================================
// SUBMIT SATISFACTION
// ============================================================================

export const useSubmitSatisfaction = (ticketId: string) => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (data: SubmitSatisfactionRequest) =>
      supportService.submitSatisfaction(ticketId, data),
    onSuccess: () => {
      // Refetch ticket so updated rating shows up
      queryClient.invalidateQueries({ queryKey: supportKeys.ticketDetail(ticketId) });
    },
    onError: (error: ApiError) => {
      console.error("Submit satisfaction failed:", error);
    },
  });

  return {
    submitSatisfaction: mutation.mutate,
    submitSatisfactionAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    data: mutation.data,
    reset: mutation.reset,
  };
};

// ============================================================================
// LIST MESSAGES
// ============================================================================

export const useListMessages = (ticketId: string) => {
  const query = useQuery({
    queryKey: supportKeys.messages(ticketId),
    queryFn: () => supportService.listMessages(ticketId),
    staleTime: 30 * 1000, // 30s — messages change more often
    enabled: !!ticketId,
  });

  return {
    messages: query.data?.messages ?? [],
    total: query.data?.total ?? 0,
    isLoading: query.isLoading,
    isError: query.isError,
    error: query.error as ApiError | null,
    refetch: query.refetch,
  };
};

// ============================================================================
// ADD MESSAGE
// ============================================================================

export const useAddMessage = (ticketId: string) => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (data: AddMessageRequest) => supportService.addMessage(ticketId, data),
    onSuccess: () => {
      // Refresh conversation + ticket (status may have changed e.g. resolved → open)
      queryClient.invalidateQueries({ queryKey: supportKeys.messages(ticketId) });
      queryClient.invalidateQueries({ queryKey: supportKeys.ticketDetail(ticketId) });
    },
    onError: (error: ApiError) => {
      console.error("Add message failed:", error);
    },
  });

  return {
    addMessage: mutation.mutate,
    addMessageAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    data: mutation.data,
    reset: mutation.reset,
  };
};

// ============================================================================
// UPLOAD ATTACHMENT
// ============================================================================

export const useUploadAttachment = (ticketId: string, messageId: string) => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (file: File) => supportService.uploadAttachment(ticketId, messageId, file),
    onSuccess: () => {
      // Refresh messages so attachment shows in the conversation
      queryClient.invalidateQueries({ queryKey: supportKeys.messages(ticketId) });
    },
    onError: (error: ApiError) => {
      console.error("Upload attachment failed:", error);
    },
  });

  return {
    uploadAttachment: mutation.mutate,
    uploadAttachmentAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error as ApiError | null,
    data: mutation.data,
    reset: mutation.reset,
  };
};

// ============================================================================
// GET DOWNLOAD URL  (on-demand — not a background query)
// ============================================================================

export const useGetDownloadURL = (ticketId: string, attachmentId: string) => {
  // enabled: false — only fetch when user explicitly clicks download
  const query = useQuery({
    queryKey: [...supportKeys.ticketDetail(ticketId), "attachment", attachmentId, "download"],
    queryFn: () => supportService.getDownloadURL(ticketId, attachmentId),
    enabled: false,
    staleTime: 10 * 60 * 1000, // presigned URL valid for 15min, cache for 10min
    gcTime: 15 * 60 * 1000,
    retry: false,
  });

  return {
    // Call this when user clicks the download button
    fetchDownloadURL: query.refetch,
    downloadURL: query.data?.download_url,
    expiresAt: query.data?.expires_at,
    isLoading: query.isFetching,
    isError: query.isError,
    error: query.error as ApiError | null,
  };
};