import { useState, useRef, useEffect } from "react";
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
import SelectInput from "@/common/SelectInput";

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

  // Avatar Profile Popover State (Hover & Click Handling)
  const [activeAvatarId, setActiveAvatarId] = useState<string | null>(null);
  const listRef = useRef<HTMLDivElement>(null);

  // Close popover when clicked outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (listRef.current && !listRef.current.contains(event.target as Node)) {
        setActiveAvatarId(null);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

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

  // Base Options for SelectInput
  const ROLE_OPTIONS = [
    { label: "Admin", value: "admin" },
    { label: "Member", value: "member" },
    { label: "Viewer", value: "viewer" },
  ];

  return (
    <section className="bg-white dark:bg-zinc-900/70 border border-zinc-200 dark:border-zinc-800 rounded-4xl p-6 md:p-8 space-y-6">
      {/* HEADER */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-6">
        <div className="flex items-center gap-4">
          <div className="p-3 rounded-xl bg-zinc-100 dark:bg-zinc-800 shrink-0">
            <Users className="w-6 h-6 text-zinc-500" />
          </div>
          <div className="flex flex-col">
            <h2 className="text-xl font-bold text-zinc-900 dark:text-white">
              Team Members
            </h2>
            <p className="text-sm text-zinc-500 mt-0.5">
              Manage your team members, view profiles, and update roles.
            </p>
          </div>
        </div>
        <span className="lg:inline-flex hidden items-center justify-center px-4 py-1.5 text-sm font-bold bg-zinc-100 dark:bg-zinc-800 text-zinc-600 dark:text-zinc-300 rounded-full border border-zinc-200 dark:border-zinc-700">
          {members?.length || 0} <Users className="w-4 h-4 ml-1" />
        </span>
      </div>

      <div className="w-full h-px bg-zinc-100 dark:bg-zinc-800/80 mb-6"></div>

      {/* LIST CONTENT */}
      {isLoading ? (
        <div className="flex justify-center py-12">
          <Loader className="w-8 h-8 text-indigo-500" />
        </div>
      ) : isError ? (
        <div className="flex flex-col items-center text-center py-10 bg-red-50 dark:bg-red-500/10 rounded-2xl border border-red-100 dark:border-red-900/30">
          <Error className="w-10 h-10 text-red-500 mb-3" />
          <h3 className="text-red-800 dark:text-red-400 font-semibold">
            Failed to load members
          </h3>
          <p className="text-sm text-red-600/80 dark:text-red-400/80 mt-1">
            Please try refreshing the page.
          </p>
        </div>
      ) : (
        <div ref={listRef} className="space-y-3 max-h-125 pr-2">
          {members.map((member) => {
            const isOwner = member.role === "owner";
            const isDisabled = !canManage || isOwner || isUpdating;

            // Dynamic Options (If user is owner, they see "Workspace Owner" in dropdown)
            const memberOptions = isOwner
              ? [{ label: "Workspace Owner", value: "owner" }]
              : ROLE_OPTIONS;

            return (
              <div
                key={member.id}
                className="group w-full flex flex-col p-4 sm:p-5 rounded-2xl border border-zinc-200 dark:border-zinc-800/60 bg-white dark:bg-zinc-900/40 hover:bg-zinc-50 dark:hover:bg-zinc-800/80 hover:border-zinc-300 dark:hover:border-zinc-700 transition-all duration-300 gap-4"
              >
                {/* ✅ LEFT SIDE: Avatar + Name + Email */}
                <div className="flex items-center gap-4 relative">
                  {/* Avatar Wrapper with State for Hover/Click Popover */}
                  <div
                    className="relative shrink-0"
                    onMouseEnter={() => setActiveAvatarId(member.id)}
                    onMouseLeave={() => setActiveAvatarId(null)}
                  >
                    <button
                      onClick={() =>
                        setActiveAvatarId(
                          activeAvatarId === member.id ? null : member.id,
                        )
                      }
                      className="relative block w-14 h-14 rounded-full overflow-hidden border-2 border-transparent hover:border-indigo-500/50 dark:hover:border-indigo-400/50 outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 transition-all duration-200 cursor-pointer"
                    >
                      {member.avatar_url ? (
                        <img
                          src={member.avatar_url}
                          alt={member.full_name}
                          className="w-full h-full object-cover"
                        />
                      ) : (
                        <div className="w-full h-full bg-linear-to-br from-indigo-100 to-indigo-50 dark:from-indigo-900/60 dark:to-indigo-800/40 flex items-center justify-center">
                          <span className="text-sm font-bold text-indigo-600 dark:text-indigo-400">
                            {member.full_name?.charAt(0).toUpperCase()}
                          </span>
                        </div>
                      )}
                    </button>

                    {/* ✨ PRO AVATAR POPOVER (Glassmorphic Card) */}
                    <div
                      className={`absolute z-50 left-0 sm:left-1/2 sm:-translate-x-1/2 top-full mt-3 w-64 p-1 transition-all duration-300 transform origin-top sm:origin-top-left ${
                        activeAvatarId === member.id
                          ? "opacity-100 visible translate-y-0 scale-100"
                          : "opacity-0 invisible -translate-y-2 scale-95"
                      }`}
                    >
                      <div className="bg-white/95 dark:bg-zinc-900/95 backdrop-blur-xl border border-zinc-200/80 dark:border-zinc-700/80 rounded-3xl p-6 text-center">
                        {/* Big Avatar */}
                        <div className="w-20 h-20 rounded-full overflow-hidden border-4 border-white dark:border-zinc-800 mx-auto mb-3">
                          {member.avatar_url ? (
                            <img
                              src={member.avatar_url}
                              alt={member.full_name}
                              className="w-full h-full object-cover"
                            />
                          ) : (
                            <div className="w-full h-full bg-linear-to-br from-indigo-100 to-indigo-50 dark:from-indigo-900/60 dark:to-indigo-800/40 flex items-center justify-center">
                              <span className="text-2xl font-bold text-indigo-600 dark:text-indigo-400">
                                {member.full_name?.charAt(0).toUpperCase()}
                              </span>
                            </div>
                          )}
                        </div>

                        <h4 className="font-bold text-zinc-900 dark:text-white text-base truncate px-2">
                          {member.full_name}
                        </h4>
                        <p className="text-xs text-zinc-500 dark:text-zinc-400 mt-0.5 truncate px-2">
                          {member.email}
                        </p>

                        <div className="mt-4 pt-4 border-t border-zinc-100 dark:border-zinc-800/80">
                          <span className="inline-flex items-center px-2.5 py-1 rounded-md bg-zinc-100 dark:bg-zinc-800 text-[10px] font-semibold text-zinc-600 dark:text-zinc-300 uppercase tracking-wider">
                            Joined{" "}
                            {new Date(member.joined_at).toLocaleDateString(
                              "en-US",
                              { month: "short", year: "numeric" },
                            )}
                          </span>
                        </div>
                      </div>
                    </div>
                  </div>

                  {/* Info Text (List Row) */}
                  <div className="flex flex-col min-w-0">
                    <p className="text-[15px] font-semibold text-zinc-900 dark:text-white truncate">
                      {member.full_name}
                    </p>
                    <p className="text-[12px] text-zinc-500 dark:text-zinc-400 truncate mt-0.5 font-mono">
                      {member.email}
                    </p>
                  </div>
                </div>

                {/* ✅ RIGHT SIDE: Custom SelectInput + DeleteButton */}
                <div className="w-full space-y-4 ">
                  <div className="w-full">
                    {/* HINT: SelectInput ab label="Role" ke sath clean dikhega */}
                    <SelectInput
                      label="Role"
                      options={memberOptions}
                      value={member.role}
                      onChange={(val) =>
                        handleRoleChange(member.user_id, val as OrgMemberRole)
                      }
                      disabled={isDisabled}
                      className="w-full p-3"
                    />
                  </div>

                  {canManage && !isOwner && (
                    <DeleteButton
                      loading={isRemoving}
                      onClick={() => handleRemove(member.user_id)}
                    >
                      Remove
                    </DeleteButton>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      )}
    </section>
  );
};
