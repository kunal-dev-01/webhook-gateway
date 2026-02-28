// src/pages/support/Support.tsx
import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useListTickets } from "@/hooks/useSupport"; // Ensure correct path to your hooks
import { Loader } from "@/common/Loader";
import { Error, X } from "@/svg/Common"; // Assume X is your plus/close icon
import SubmitButton from "@/common/SubmitButton";
import { CreateTicketForm } from "./CreateTicketForm";
import type { Ticket, TicketStatus } from "@/types/support.types";

const Support = () => {
  const [activeTab, setActiveTab] = useState<TicketStatus | "all">("all");
  const [showCreateForm, setShowCreateForm] = useState(false);
  
  const { tickets, isLoading, isError } = useListTickets(
    activeTab === "all" ? {} : { status: activeTab }
  );

  if (isLoading) {
    return (
      <div className="flex justify-center items-center min-h-[60vh]">
        <Loader className="w-8 h-8 text-indigo-500" />
      </div>
    );
  }

  if (isError) {
    return (
      <div className="flex flex-col justify-center items-center gap-5 min-h-[80vh]">
        <Error className="w-20 h-20 text-red-500" />
        <h1 className="text-3xl font-bold tracking-tight text-zinc-700 dark:text-white">
          Failed to load tickets.
        </h1>
      </div>
    );
  }

  return (
    <div className="mx-auto w-full max-w-7xl py-8 md:py-12">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-8 md:mb-12">
        <div>
          <h1 className="text-3xl font-bold tracking-tight text-zinc-700 dark:text-white">
            Support Center
          </h1>
          <p className="text-zinc-500 dark:text-zinc-400 mt-2 text-lg">
            Need help? Manage your support tickets or create a new one.
          </p>
        </div>

        <div className="inline-block">
          <SubmitButton 
            onClick={() => setShowCreateForm(!showCreateForm)}
            className="px-6"
          >
            <X className={`w-5 h-5 transition-transform ${showCreateForm ? "" : "rotate-45"}`} />
            {showCreateForm ? "Cancel" : "New Ticket"}
          </SubmitButton>
        </div>
      </div>

      {/* Tabs & Content */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
        <div className={`lg:col-span-${showCreateForm ? '7' : '12'} space-y-6 transition-all`}>
          
          {/* TABS */}
          <div className="flex items-center gap-1 overflow-x-auto border-b border-zinc-200 dark:border-zinc-800 pb-0.5">
            {["all", "open", "in_progress", "resolved", "closed"].map((tab) => (
              <button
                key={tab}
                onClick={() => setActiveTab(tab as any)}
                className={`px-4 py-3 text-sm font-medium border-b-2 transition-all duration-200 whitespace-nowrap capitalize ${
                  activeTab === tab
                    ? "border-indigo-500 text-indigo-600 dark:text-indigo-400"
                    : "border-transparent text-zinc-500 hover:text-zinc-700 dark:hover:text-zinc-300"
                }`}
              >
                {tab.replace("_", " ")}
              </button>
            ))}
          </div>

          {/* TICKET LIST */}
          <div className="space-y-4">
            {tickets.length === 0 ? (
              <div className="bg-zinc-50/50 dark:bg-zinc-900/30 border border-zinc-200 dark:border-zinc-800 rounded-4xl p-10 text-center">
                <p className="text-zinc-500 mt-1">No tickets found for this view.</p>
              </div>
            ) : (
              tickets.map((ticket) => <TicketCard key={ticket.id} ticket={ticket} />)
            )}
          </div>
        </div>

        {/* CREATE FORM SIDEBAR */}
        {showCreateForm && (
          <div className="lg:col-span-5 lg:sticky lg:top-8 animate-in slide-in-from-right-8 duration-300 fade-in">
             <CreateTicketForm onSuccess={() => setShowCreateForm(false)} />
          </div>
        )}
      </div>
    </div>
  );
};

// ============================================================================
// TICKET CARD COMPONENT
// ============================================================================
const TicketCard = ({ ticket }: { ticket: Ticket }) => {
  const navigate = useNavigate();

  const getStatusColor = (status: TicketStatus) => {
    switch (status) {
      case "open": return "bg-blue-100 text-blue-700 dark:bg-blue-500/10 dark:text-blue-400";
      case "resolved": return "bg-green-100 text-green-700 dark:bg-green-500/10 dark:text-green-400";
      case "closed": return "bg-zinc-100 text-zinc-700 dark:bg-zinc-800 dark:text-zinc-400";
      default: return "bg-indigo-100 text-indigo-700 dark:bg-indigo-500/10 dark:text-indigo-400";
    }
  };

  const getPriorityColor = (priority: string) => {
    return priority === "high" || priority === "urgent" 
      ? "border-red-200 text-red-600 dark:border-red-500/30 dark:text-red-400" 
      : "border-zinc-200 text-zinc-600 dark:border-zinc-700 dark:text-zinc-400";
  };

  return (
    <div 
      onClick={() => navigate(`/support/${ticket.id}`)}
      className="bg-white dark:bg-zinc-900/70 border border-zinc-200 dark:border-zinc-800 rounded-3xl p-5 hover:border-indigo-400 dark:hover:border-indigo-500 transition-colors cursor-pointer group"
    >
      <div className="flex justify-between items-start mb-3">
        <div className="flex gap-3 items-center">
          <span className={`px-2.5 py-1 text-[10px] font-bold uppercase tracking-wider rounded-full ${getStatusColor(ticket.status)}`}>
            {ticket.status.replace("_", " ")}
          </span>
          <span className="text-sm font-medium text-zinc-500 dark:text-zinc-400">
            {ticket.ticket_code}
          </span>
        </div>
        <span className={`px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider rounded-md border ${getPriorityColor(ticket.priority)}`}>
          {ticket.priority}
        </span>
      </div>
      
      <h3 className="text-lg font-semibold text-zinc-900 dark:text-white group-hover:text-indigo-500 transition-colors">
        {ticket.subject}
      </h3>
      
      <div className="mt-4 flex items-center justify-between text-xs text-zinc-500">
        <span className="capitalize border border-zinc-200 dark:border-zinc-800 px-2 py-1 rounded-lg">
          Category: {ticket.category.replace("_", " ")}
        </span>
        <span>Updated: {new Date(ticket.updated_at).toLocaleDateString()}</span>
      </div>
    </div>
  );
};

export default Support;