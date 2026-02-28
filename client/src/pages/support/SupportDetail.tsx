// src/pages/support/SupportDetail.tsx
import { useParams, useNavigate } from "react-router-dom";
import { useGetTicket, useListMessages, useAddMessage, useReopenTicket } from "@/hooks/useSupport";
import { Loader } from "@/common/Loader";
import { Error } from "@/svg/Common";
import { useState } from "react";
import SubmitButton from "@/common/SubmitButton";
import { toast } from "sonner";

const SupportDetail = () => {
  const { ticketId } = useParams<{ ticketId: string }>();
  const navigate = useNavigate();
  const [replyText, setReplyText] = useState("");

  const { ticket, isLoading: isLoadingTicket, isError: isErrorTicket } = useGetTicket(ticketId || "");
  const { messages, isLoading: isLoadingMsgs } = useListMessages(ticketId || "");
  const { addMessageAsync, isPending: isSending } = useAddMessage(ticketId || "");
  const { reopenTicketAsync, isPending: isReopening } = useReopenTicket(ticketId || "");

  if (!ticketId) return <div className="text-center py-10 text-red-500">Invalid Ticket ID</div>;
  if (isLoadingTicket || isLoadingMsgs) return <div className="flex justify-center items-center min-h-[60vh]"><Loader className="w-8 h-8 text-indigo-500" /></div>;
  if (isErrorTicket || !ticket) return <div className="flex justify-center items-center min-h-[60vh]"><Error className="w-20 h-20 text-red-500" /></div>;

  const handleReply = async () => {
    if (!replyText.trim()) return;
    try {
      await addMessageAsync({ body: replyText });
      setReplyText("");
    } catch (err) {
      toast.error("Failed to send reply");
    }
  };

  const handleReopen = async () => {
    try {
      await reopenTicketAsync();
      toast.success("Ticket reopened");
    } catch (err) {
      toast.error("Failed to reopen");
    }
  };

  const isClosed = ticket.status === "closed" || ticket.status === "resolved";

  return (
    <div className="mx-auto w-full max-w-5xl py-8 md:py-12">
      <button onClick={() => navigate("/support")} className="text-sm text-zinc-500 hover:text-indigo-500 mb-6 flex items-center gap-1">
        ← Back to Tickets
      </button>

      {/* TICKET HEADER */}
      <div className="bg-white dark:bg-zinc-900/70 border border-zinc-200 dark:border-zinc-800 rounded-3xl p-6 md:p-8 mb-6 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
        <div>
          <div className="flex items-center gap-3 mb-2">
            <span className="text-sm font-medium text-zinc-500">{ticket.ticket_code}</span>
            <span className="px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider bg-zinc-100 text-zinc-700 dark:bg-zinc-800 dark:text-zinc-400 rounded-md">
              {ticket.status.replace("_", " ")}
            </span>
          </div>
          <h1 className="text-2xl font-bold text-zinc-900 dark:text-white">{ticket.subject}</h1>
        </div>
        
        {isClosed && (
          <SubmitButton onClick={handleReopen} loading={isReopening} className="w-auto px-6 bg-zinc-800 hover:bg-zinc-700">
            Reopen Ticket
          </SubmitButton>
        )}
      </div>

      {/* CHAT THREAD */}
      <div className="bg-zinc-50 dark:bg-zinc-900/30 border border-zinc-200 dark:border-zinc-800 rounded-3xl p-6 mb-6 min-h-100 max-h-[60vh] overflow-y-auto custom-scrollbar space-y-6">
        {messages.map((msg) => {
          const isMe = msg.sender_type === "user";
          return (
            <div key={msg.id} className={`flex ${isMe ? "justify-end" : "justify-start"}`}>
              <div className={`max-w-[80%] rounded-2xl p-4 ${
                isMe 
                  ? "bg-indigo-600 text-white rounded-br-none" 
                  : "bg-white dark:bg-zinc-800 border border-zinc-200 dark:border-zinc-700 text-zinc-800 dark:text-zinc-200 rounded-bl-none"
              }`}>
                <div className="text-sm whitespace-pre-wrap">{msg.body}</div>
                <div className={`text-[10px] mt-2 ${isMe ? "text-indigo-200" : "text-zinc-400"}`}>
                  {new Date(msg.created_at).toLocaleString()} {msg.sender_type === "agent" && "• Agent"}
                </div>
              </div>
            </div>
          );
        })}
      </div>

      {/* REPLY BOX */}
      {!isClosed ? (
        <div className="bg-white dark:bg-zinc-900/70 border border-zinc-200 dark:border-zinc-800 rounded-3xl p-4 flex gap-3">
          <textarea
            value={replyText}
            onChange={(e) => setReplyText(e.target.value)}
            disabled={isSending}
            placeholder="Type your reply here..."
            className="flex-1 bg-transparent border-none outline-none resize-none p-2 text-sm text-zinc-900 dark:text-white"
            rows={2}
          />
          <div className="flex items-end">
            <SubmitButton onClick={handleReply} loading={isSending} disabled={!replyText.trim() || isSending} className="px-6 h-10">
              Send
            </SubmitButton>
          </div>
        </div>
      ) : (
        <div className="text-center p-4 text-sm text-zinc-500 dark:text-zinc-400 border border-dashed border-zinc-300 dark:border-zinc-700 rounded-3xl">
          This ticket is closed. Reopen it to send a new message.
        </div>
      )}
    </div>
  );
};

export default SupportDetail;