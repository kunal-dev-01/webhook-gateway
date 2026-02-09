import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { toast } from "sonner"; // Sonner import
import { useAuthStore } from "@/store/auth.store";
import { useUpdateProfile } from "@/hooks/useUser";
import { useLogout, useResendVerification } from "@/hooks/useAuth";
import AuthInput from "@/common/AuthInput";
import SubmitButton from "@/common/SubmitButton";
import { TriangleAlert, X } from "@/svg/Common";
import { BadgeCheck, Logout, Pen } from "@/svg/Auth";
import { Loader } from "@/common/Loader";
import { profileSchema, type ProfileFormData } from "@/schemas/user.schema";


export const ProfileSidePanel = () => {
  const user = useAuthStore((state) => state.user);
  const {
    updateProfile,
    isPending,
    isSuccess,
    isError,
    reset: resetMutation,
  } = useUpdateProfile();
  const { logout, isPending: isLogoutLoading } = useLogout();
  const { resendVerification, isPending: isResending } =
    useResendVerification();

  const [isEditing, setIsEditing] = useState(false);
  const [cooldown, setCooldown] = useState(0);

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<ProfileFormData>({
    resolver: zodResolver(profileSchema),
    defaultValues: { name: user?.full_name || "", email: user?.email || "" },
  });

  useEffect(() => {
    if (user) {
      reset({ name: user.full_name, email: user.email });
    }
  }, [user, reset]);

  useEffect(() => {
    if (isSuccess) {
      toast.success("Profile updated successfully.");
      setIsEditing(false);
      resetMutation();
    }
    if (isError) {
      toast.error("Failed to update profile.");
    }
  }, [isSuccess, isError, resetMutation]);

  // Cooldown timer effect
  useEffect(() => {
    if (cooldown > 0) {
      const timer = setTimeout(() => setCooldown(cooldown - 1), 1000);
      return () => clearTimeout(timer);
    }
  }, [cooldown]);

  const onSubmit = (data: ProfileFormData) => {
    if (data.name.trim() !== user?.full_name) {
      updateProfile({ full_name: data.name });
    } else {
      setIsEditing(false);
    }
  };

  const handleCancel = () => {
    setIsEditing(false);
    if (user) reset({ name: user.full_name, email: user.email });
  };

  // Handle resend verification
  const handleResendVerification = () => {
    if (!user?.email || cooldown > 0) return;

    resendVerification(
      { email: user.email },
      {
        onSuccess: () => {
          toast.success("Verification email sent! Please check your inbox."); 
          setCooldown(60);
        },
        onError: () => {
          toast.error("Failed to send verification email. Please try again."); 
        },
      },
    );
  };

  if (!user) return null;

  const initials = user.full_name
    ? user.full_name
        .split(" ")
        .map((n) => n[0])
        .join("")
        .toUpperCase()
        .substring(0, 2)
    : "U";

  return (
    <section className="bg-white dark:bg-zinc-900/70 border border-zinc-200 dark:border-zinc-800 rounded-4xl p-6 relative overflow-hidden flex flex-col h-full justify-between">
      <button
        onClick={isEditing ? handleCancel : () => setIsEditing(true)}
        className={`absolute top-4 right-4 px-3 py-1.5 rounded-full transition-all duration-200 group cursor-pointer border border-zinc-200 dark:border-zinc-800 ${
          isEditing
            ? "text-zinc-400 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20"
            : "text-zinc-400 hover:text-indigo-600 hover:bg-indigo-50 dark:hover:bg-indigo-900/20"
        }`}
        title={isEditing ? "Cancel" : "Edit Profile"}
      >
        {isEditing ? <X className="w-5.5 h-5.5" /> : <Pen className="w-5.5 h-5.5" />}
      </button>

      {/* Main Content */}
      <div className="flex flex-col items-center text-center">
        {/* Avatar */}
        <div className="mb-4 relative">
          <div className="w-24 h-24 rounded-full bg-linear-to-br from-indigo-100 to-white dark:from-indigo-900/40 dark:to-zinc-900 flex items-center justify-center border-4 border-zinc-100 dark:border-zinc-800">
            <span className="text-3xl font-bold text-indigo-600 dark:text-indigo-400">
              {initials}
            </span>
          </div>
          {/* Active Dot */}
          <div className="absolute bottom-1 right-1 w-5 h-5 bg-emerald-500 border-4 border-white dark:border-zinc-800 rounded-full"></div>
        </div>

        {!isEditing ? (
          <div className="space-y-2 w-full animate-in fade-in duration-300">
            <h2 className="text-xl font-bold text-zinc-900 dark:text-white">
              {user.full_name}
            </h2>
            <p className="text-sm text-zinc-500 font-medium break-all">
              {user.email}
            </p>

            {/* Enhanced Email Verification Badge */}
            <div className="pt-3 flex justify-center">
              <div
                className={`flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-semibold border transition-colors duration-300 cursor-default ${
                  user.email_verified
                    ? "bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-900/20 dark:text-emerald-400 dark:border-emerald-800/50"
                    : "bg-amber-50 text-amber-700 border-amber-200 dark:bg-amber-900/20 dark:text-amber-400 dark:border-amber-800/50"
                }`}
              >
                {user.email_verified ? (
                  <>
                    <BadgeCheck className="w-4 h-4" />
                    <span>Verified</span>
                  </>
                ) : (
                  <>
                    <TriangleAlert className="w-4 h-4" />
                    <span>Unverified</span>
                  </>
                )}
              </div>
            </div>

            {/* Email Verification Warning & Resend */}
            {!user.email_verified && (
              <div className="mt-4 p-4 bg-amber-50 dark:bg-amber-900/10 border border-amber-200 dark:border-amber-800/30 rounded-xl space-y-3 animate-in fade-in duration-500">
                <div className="flex items-start gap-2">
                  <TriangleAlert className="w-5 h-5 text-amber-600 dark:text-amber-400 shrink-0 mt-0.5" />
                  <div className="flex-1 text-left">
                    <h3 className="text-sm font-semibold text-amber-800 dark:text-amber-200">
                      Email Not Verified
                    </h3>
                    <p className="text-xs text-amber-700 dark:text-amber-300 mt-1">
                      Please verify your email to access all features and secure
                      your account.
                    </p>
                  </div>
                </div>

                <button
                  onClick={handleResendVerification}
                  disabled={isResending || cooldown > 0}
                  className="w-full px-4 py-2 bg-amber-100 hover:bg-amber-200 dark:bg-amber-800/20 dark:hover:bg-amber-800/30 text-amber-800 dark:text-amber-200 rounded-lg text-xs font-medium transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
                >
                  {isResending ? (
                    <>
                      <span className="w-3.5 h-3.5 border-2 border-amber-300 border-t-amber-600 rounded-full animate-spin"></span>
                      <span>Sending...</span>
                    </>
                  ) : cooldown > 0 ? (
                    <span>Resend in {cooldown}s</span>
                  ) : (
                    <span>Resend Verification Email</span>
                  )}
                </button>
              </div>
            )}
          </div>
        ) : (
          /* Edit Form */
          <form
            onSubmit={handleSubmit(onSubmit)}
            className="w-full space-y-4 pt-2"
          >
            <div className="text-left">
              <AuthInput
                label="Full Name"
                type="text"
                name="name"
                placeholder="Enter your name"
                register={register}
                error={errors.name?.message}
              />
            </div>
            {/* Hidden/Disabled Email Input if needed for structure, or removed as per design preference */}
            
            <div className="pt-2">
              <SubmitButton loading={isPending} disabled={isPending}>
                {isPending ? "Saving..." : "Save Changes"}
              </SubmitButton>
            </div>
          </form>
        )}
      </div>

      {/* Footer Area: Logout Button */}
      {!isEditing && (
        <div className="mt-8 w-full">
          <div className="w-full h-px bg-zinc-100 dark:bg-zinc-800 mb-4"></div>
          <button
            onClick={() => logout()}
            disabled={isLogoutLoading}
            className="w-full group flex items-center justify-center gap-2 px-4 py-2.5 rounded-xl text-zinc-600 dark:text-zinc-400 hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/10 dark:hover:text-red-400 transition-all duration-200 font-medium text-sm cursor-pointer"
          >
            {isLogoutLoading ? (
              <Loader />
            ) : (
              <Logout className="w-6 h-6 transition-transform group-hover:-translate-x-1" />
            )}
            <span>{isLogoutLoading ? "Logging out..." : "Sign Out"}</span>
          </button>
        </div>
      )}
    </section>
  );
};