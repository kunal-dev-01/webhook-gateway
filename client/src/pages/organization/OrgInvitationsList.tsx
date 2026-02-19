import { z } from "zod";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { toast } from "sonner";
import { useListInvitations, useInviteMember, useRevokeInvitation } from "@/hooks/useInvitation";
import { type OrgMemberRole } from "@/types/org.types";
import { Loader } from "@/common/Loader";
import { Error, Mail, ChevronDown, Trash } from "@/svg/Common"; 
import AuthInput from "@/common/AuthInput";
import SubmitButton from "@/common/SubmitButton";
import DeleteButton from "@/common/DeleteButton";

const inviteSchema = z.object({
  email: z.string().email("Please enter a valid email address"),
  role: z.enum(["admin", "member", "viewer"] as const),
});

type InviteFormValues = z.infer<typeof inviteSchema>;

export const OrgInvitationsList = ({
  orgId,
  currentUserRole,
}: {
  orgId: string;
  currentUserRole: OrgMemberRole;
}) => {
  const { invitations, isLoading, isError } = useListInvitations(orgId);
  const { inviteMemberAsync, isPending: isInviting } = useInviteMember(orgId);
  const { revokeInvitation, isPending: isRevoking } = useRevokeInvitation(orgId);

  const canManage = currentUserRole === "owner" || currentUserRole === "admin";

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<InviteFormValues>({
    resolver: zodResolver(inviteSchema),
    defaultValues: {
      role: "member",
    },
  });

  const onSubmit = async (data: InviteFormValues) => {
    try {
      await inviteMemberAsync(data);
      toast.success(`Invitation sent to ${data.email}`);
      reset();
    } catch (err) {
      toast.error("Failed to send invitation");
    }
  };

  const handleRevoke = (invitationId: string) => {
    revokeInvitation(invitationId, {
      onSuccess: () => toast.success("Invitation revoked"),
      onError: () => toast.error("Failed to revoke invitation"),
    });
  };

  return (
    <section className="bg-white dark:bg-zinc-900/70 border border-zinc-200 dark:border-zinc-800 rounded-4xl p-6 md:p-8">
      <div className="flex items-center gap-3 mb-6">
        <div className="p-3 bg-zinc-100 dark:bg-zinc-800 rounded-xl hidden sm:block">
          <Mail className="w-6 h-6 text-zinc-500" />
        </div>
        <div className="flex flex-col">
          <h2 className="text-xl font-bold text-zinc-900 dark:text-white">
            Pending Invitations
          </h2>
          <p className="text-sm text-zinc-500">
            Invite new members or manage pending invites.
          </p>
        </div>
      </div>

      <div className="w-full h-px bg-zinc-100 dark:bg-zinc-800/80 mb-8"></div>

      {/* INVITATION FORM - VERTICAL LAYOUT */}
      {canManage && (
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-5 mb-10">
          {/* 1. Email Input */}
          <div className="w-full">
            <AuthInput
              label="Email Address"
              type="email"
              name="email"
              placeholder="colleague@example.com"
              register={register}
              error={errors.email?.message}
              disabled={isInviting}
            />
          </div>

          {/* 2. Role Dropdown with Chevron */}
          <div className="w-full">
            <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-200 mb-1.5 ml-1">
              Select Role
            </label>
            <div className="relative">
              <select
                {...register("role")}
                disabled={isInviting}
                className="w-full appearance-none px-4 py-3 rounded-xl border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900/50 text-zinc-900 dark:text-white text-sm outline-none cursor-pointer "
              >
                <option value="admin">Admin</option>
                <option value="member">Member</option>
                <option value="viewer">Viewer</option>
              </select>
              <div className="absolute inset-y-0 right-4 flex items-center pointer-events-none">
                <ChevronDown className="w-5 h-5 text-zinc-400" />
              </div>
            </div>
            {errors.role && (
              <p className="text-xs text-red-500 mt-1.5 ml-1">{errors.role.message}</p>
            )}
          </div>

          {/* 3. Submit Button */}
          <div className="pt-2">
            <SubmitButton loading={isInviting} disabled={isInviting} className="w-full py-3">
              {isInviting ? "Sending..." : "Send Invitation"}
            </SubmitButton>
          </div>
        </form>
      )}

      {/* INVITATIONS LIST */}
      <div className="space-y-4">
        <h3 className="text-xs font-semibold uppercase tracking-wider text-zinc-400 ml-1">
          Recent Invitations
        </h3>
        
        {isLoading ? (
          <div className="flex justify-center py-8">
            <Loader className="w-6 h-6 text-indigo-500" />
          </div>
        ) : isError ? (
          <div className="text-center py-8 text-red-500">
            <Error className="w-8 h-8 mx-auto mb-2" /> Failed to load invitations
          </div>
        ) : invitations.length === 0 ? (
          <p className="text-center text-zinc-500 dark:text-zinc-400 py-6 text-sm bg-zinc-50 dark:bg-zinc-900/20 rounded-2xl border border-dashed border-zinc-200 dark:border-zinc-800">
            No pending invitations.
          </p>
        ) : (
          <div className="space-y-3 max-h-80 overflow-y-auto pr-2 custom-scrollbar">
            {invitations.map((invite) => (
              <div
                key={invite.id}
                className="flex items-center justify-between p-4 rounded-2xl border border-zinc-200 dark:border-zinc-800 bg-zinc-50/50 dark:bg-zinc-900/30"
              >
                <div>
                  <p className="text-sm font-semibold text-zinc-900 dark:text-white">
                    {invite.email}
                  </p>
                  <div className="flex items-center gap-2 mt-1">
                    <span className={`text-[10px] px-2 py-0.5 rounded-full font-bold uppercase tracking-tight ${
                      invite.status === "pending" ? "bg-yellow-100 text-yellow-700 dark:bg-yellow-500/10 dark:text-yellow-400" :
                      invite.status === "expired" ? "bg-red-100 text-red-700 dark:bg-red-500/10 dark:text-red-400" :
                      "bg-zinc-100 text-zinc-700 dark:bg-zinc-800 dark:text-zinc-400"
                    }`}>
                      {invite.status}
                    </span>
                    <span className="text-[11px] text-zinc-500 capitalize">
                      • {invite.role}
                    </span>
                  </div>
                </div>

                {canManage && invite.status === "pending" && (
                  <div className="inline-block rounded-full overflow-hidden">
                    <DeleteButton
                    loading={isRevoking}
                    onClick={() => handleRevoke(invite.id)}
                    className="px-6"
                  >
                    <Trash className="w-5 h-5" />
                  </DeleteButton>
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </section>
  );
};