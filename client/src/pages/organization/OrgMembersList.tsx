import { toast } from "sonner";
import {
  useListMembers,
  useUpdateMemberRole,
  useRemoveMember,
} from "@/hooks/useOrg";
import { type OrgMemberRole } from "@/types/org.types";
import { Error } from "@/svg/Common";
import { Loader } from "@/common/Loader";
import DeleteButton from "@/common/DeleteButton";
import { Users } from "@/svg/Auth";

export const OrgMembersList = ({
  orgId,
  currentUserRole,
}: {
  orgId: string;
  currentUserRole: OrgMemberRole;
}) => {
  const { members, isLoading, isError } = useListMembers(orgId);
  const { updateRole, isPending: isUpdating } = useUpdateMemberRole(orgId);
  const { removeMember, isPending: isRemoving } = useRemoveMember(orgId);

  const handleRoleChange = (userId: string, newRole: OrgMemberRole) => {
    if (newRole === "owner") return; // API type restricts this
    updateRole(
      { userId, data: { role: newRole as Exclude<OrgMemberRole, "owner"> } },
      {
        onSuccess: () => toast.success("Role updated successfully"),
        onError: () => toast.error("Failed to update role"),
      },
    );
  };

  const handleRemove = (userId: string) => {
    removeMember(userId, {
      onSuccess: () => toast.success("Member removed from workspace"),
      onError: () => toast.error("Failed to remove member"),
    });
  };

  const canManage = currentUserRole === "owner" || currentUserRole === "admin";

  return (
    <section className="bg-white dark:bg-zinc-900/70 border border-zinc-200 dark:border-zinc-800 rounded-4xl p-6 md:p-8 space-y-6">
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-3">
          <div className="p-3 bg-zinc-100 dark:bg-zinc-800 rounded-xl hidden sm:block">
            <Users className="w-6 h-6 text-zinc-500" />
          </div>
          <div className="flex flex-col">
            <h2 className="text-xl font-bold text-zinc-900 dark:text-white">
            Team Members
          </h2>
            <p className="text-sm text-zinc-500">
              Manage your team members and their roles.
            </p>
          </div>
        </div>
        <span className="px-3 py-1 text-xs font-semibold bg-zinc-100 dark:bg-zinc-800 text-zinc-600 dark:text-zinc-300 rounded-full">
          {members.length} Users
        </span>
      </div>

      <div className="w-full h-px bg-zinc-100 dark:bg-zinc-800/80 mb-6"></div>

      {isLoading ? (
        <div className="flex justify-center py-8">
          <Loader className="w-6 h-6 text-indigo-500" />
        </div>
      ) : isError ? (
        <div className="text-center py-8 text-red-500">
          <Error className="w-8 h-8 mx-auto mb-2" /> Failed to load
        </div>
      ) : (
        <div className="space-y-3 max-h-100 overflow-y-auto pr-2 custom-scrollbar">
          {members.map((member) => (
            <div
              key={member.id}
              className="flex items-center justify-between p-4 rounded-2xl border border-zinc-200 dark:border-zinc-800 bg-zinc-50/50 dark:bg-zinc-900/30"
            >
              <div>
                <p className="text-sm font-semibold text-zinc-900 dark:text-white">
                  User ID: {member.user_id.slice(0, 8)}...
                </p>
                <p className="text-[11px] text-zinc-500 mt-0.5">
                  Joined {new Date(member.joined_at).toLocaleDateString()}
                </p>
              </div>

              <div className="flex items-center gap-3">
                {/* Role Selector */}
                <select
                  disabled={!canManage || member.role === "owner" || isUpdating}
                  value={member.role}
                  onChange={(e) =>
                    handleRoleChange(
                      member.user_id,
                      e.target.value as OrgMemberRole,
                    )
                  }
                  className="bg-transparent border border-zinc-200 dark:border-zinc-700 rounded-lg text-xs px-2 py-1.5 text-zinc-700 dark:text-zinc-300 outline-none disabled:opacity-50 cursor-pointer"
                >
                  <option value="owner" disabled>
                    Owner
                  </option>
                  <option value="admin">Admin</option>
                  <option value="member">Member</option>
                  <option value="viewer">Viewer</option>
                </select>

                {/* Remove Button */}
                {canManage && member.role !== "owner" && (
                  <DeleteButton
                    loading={isRemoving}
                    onClick={() => handleRemove(member.user_id)}
                    className="p-1.5 px-3 text-xs w-auto rounded-lg"
                  >
                    Remove
                  </DeleteButton>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </section>
  );
};
