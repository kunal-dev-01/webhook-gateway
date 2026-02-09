import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { QRCodeSVG } from "qrcode.react";
import { toast } from "sonner";

// Store & Hooks
import { useAuthStore } from "@/store/auth.store";
import { useSetup2FA, useEnable2FA, useDisable2FA } from "@/hooks/useAuth";

// Components
import AuthInput from "@/common/AuthInput";
import SubmitButton from "@/common/SubmitButton";
import DeleteButton from "@/common/DeleteButton";

// Icons
import { Copy, Shield } from "@/svg/Common";
import { Smartphone } from "@/svg/Auth";

// Types
import { type Setup2FAResponse } from "@/types/auth.types";
import {
  disable2FASchema,
  verify2FASchema,
  type Disable2FAForm,
  type Verify2FAForm,
} from "@/schemas/user.schema";

// ─────────────────────────── Helper ─────────────────────────────
const copyToClipboard = async (text: string) => {
  try {
    await navigator.clipboard.writeText(text);
    toast.success("Copied to clipboard");
  } catch (err) {
    toast.error("Failed to copy");
  }
};

// ─────────────────────────── Component ─────────────────────────────
export const TwoFactorAuth = () => {
  const { user, setUser } = useAuthStore();

  // Hooks
  const setupMutation = useSetup2FA();
  const enableMutation = useEnable2FA();
  const disableMutation = useDisable2FA();

  // Local State
  const [step, setStep] = useState<
    "idle" | "setup" | "verify" | "backup_codes" | "disable"
  >("idle");

  // Store Setup Response temporarily (Secret, QR, Codes)
  const [setupData, setSetupData] = useState<Setup2FAResponse | null>(null);

  // Forms
  const verifyForm = useForm<Verify2FAForm>({
    resolver: zodResolver(verify2FASchema),
    defaultValues: { code: "" },
  });

  const disableForm = useForm<Disable2FAForm>({
    resolver: zodResolver(disable2FASchema),
    defaultValues: { password: "" },
  });

  // ─────────────────────────── Handlers ─────────────────────────────

  // 1. Start Setup (Get QR Code)
  const handleStartSetup = () => {
    setupMutation.setup2FA(undefined, {
      onSuccess: (data: Setup2FAResponse) => {
        setSetupData(data);
        setStep("setup");
        toast.dismiss(); // Clear any previous toasts
      },
      onError: (error: any) => {
        const msg = error?.response?.data?.error || "Failed to initialize 2FA.";
        toast.error(msg);
      },
    });
  };

  // 2. Verify & Enable 2FA
  const onVerifySubmit = (data: Verify2FAForm) => {
    enableMutation.enable2FA(
      { code: data.code },
      {
        onSuccess: () => {
          // Update Global Store
          if (user) {
            setUser({ ...user, two_factor_enabled: true });
          }

          toast.success("2FA Enabled Successfully.");

          // Move to Backup Codes step if we have them from the setup phase
          if (setupData?.recovery_codes) {
            setStep("backup_codes");
          } else {
            setStep("idle");
          }

          verifyForm.reset();
        },
        onError: (error: any) => {
          const msg =
            error?.response?.data?.error || "Invalid code. Please try again.";
          toast.error(msg);
        },
      },
    );
  };

  // 3. Disable 2FA
  const onDisableSubmit = (data: Disable2FAForm) => {
    disableMutation.disable2FA(
      { password: data.password },
      {
        onSuccess: () => {
          // Update Global Store
          if (user) {
            setUser({ ...user, two_factor_enabled: false });
          }
          toast.success("2FA has been disabled successfully.");
          setStep("idle");
          disableForm.reset();
          setSetupData(null);
        },
        onError: (error: any) => {
          const msg = error?.response?.data?.error || "Incorrect password.";
          toast.error(msg);
        },
      },
    );
  };

  const is2FAEnabled = user?.two_factor_enabled;

  return (
    <div className="max-w-3xl relative">
      <div className="flex items-start gap-4 mb-6">
        <div
          className={`p-3 rounded-xl hidden sm:block ${
            is2FAEnabled
              ? "bg-emerald-50 text-emerald-600"
              : "bg-zinc-100 text-zinc-500"
          }`}
        >
          <Shield className="w-6 h-6" />
        </div>
        <div>
          <h2 className="text-xl font-bold text-zinc-900 dark:text-white flex items-center gap-2">
            Two-Factor Authentication
            {is2FAEnabled && (
              <span className="text-xs bg-emerald-100 text-emerald-700 px-2 py-0.5 rounded-full border border-emerald-200">
                Enabled
              </span>
            )}
          </h2>
          <p className="text-sm text-zinc-500 mt-1">
            Add an extra layer of security to your account by requiring a code
            from your phone.
          </p>
        </div>
      </div>

      {/* ─────────────────────────── IDLE STATE ─────────────────────────── */}
      {step === "idle" && (
        <div className="bg-zinc-50 dark:bg-zinc-800/50 rounded-2xl p-6 border border-zinc-100 dark:border-zinc-800 flex flex-col md:flex-row items-center justify-between gap-6">
          <div className="flex items-center gap-4">
            <div className="w-12 h-12 bg-white dark:bg-zinc-800 rounded-full flex items-center justify-center border border-zinc-200 dark:border-zinc-700 shrink-0">
              <Smartphone className="w-6 h-6 text-indigo-500" />
            </div>
            <div>
              <h3 className="font-semibold text-zinc-900 dark:text-white">
                Authenticator App
              </h3>
              <p className="text-xs text-zinc-500">
                Use apps like Google Authenticator or Authy.
              </p>
            </div>
          </div>

          {!is2FAEnabled ? (
            <SubmitButton
              onClick={handleStartSetup}
              loading={setupMutation.isPending}
              className="w-full md:w-auto px-6"
            >
              {setupMutation.isPending ? "Starting..." : "Enable 2FA"}
            </SubmitButton>
          ) : (
            <div className="flex gap-3 w-full md:w-auto">
              <button
                onClick={handleStartSetup}
                className="px-4 py-2.5 text-sm font-medium text-zinc-700 dark:text-zinc-300 bg-white dark:bg-zinc-800 border border-zinc-200 dark:border-zinc-700 rounded-xl hover:bg-zinc-50 dark:hover:bg-zinc-700 transition-colors w-full md:w-auto"
              >
                Reset / Regenerate
              </button>
              <button
                onClick={() => setStep("disable")}
                className="px-4 py-2.5 text-sm font-medium text-red-600 dark:text-red-400 bg-white dark:bg-zinc-800 border border-red-200 dark:border-red-900/30 rounded-xl hover:bg-red-50 dark:hover:bg-red-900/10 transition-colors w-full md:w-auto"
              >
                Disable 2FA
              </button>
            </div>
          )}
        </div>
      )}

      {/* ─────────────────────────── SETUP: QR Code ─────────────────────────── */}
      {step === "setup" && setupData && (
        <div className="space-y-6 border-t border-zinc-100 dark:border-zinc-800 pt-6 animate-in fade-in slide-in-from-top-4 duration-300">
          <div className="flex flex-col md:flex-row gap-8">
            {/* QR Image */}
            <div className="shrink-0 flex items-center justify-center">
              <div className="p-2 bg-white rounded-xl border border-zinc-200">
                {setupData.qr_code_url ? (
                  <QRCodeSVG
                    value={setupData.qr_code_url}
                    size={192}
                    level="M"
                    includeMargin={true}
                  />
                ) : (
                  <div className="w-48 h-48 flex items-center justify-center text-zinc-400">
                    Loading QR...
                  </div>
                )}
              </div>
            </div>

            {/* Instructions & Input */}
            <div className="flex-1 space-y-4">
              <div>
                <h3 className="font-semibold text-zinc-900 dark:text-white">
                  Scan QR Code
                </h3>
                <p className="text-sm text-zinc-500 mb-2">
                  Open your authenticator app and scan the image.
                </p>
                <p className="text-xs text-zinc-400 break-all font-mono bg-zinc-100 dark:bg-zinc-800 p-2 rounded-lg">
                  Secret: {setupData.secret}
                </p>
              </div>

              <form
                onSubmit={verifyForm.handleSubmit(onVerifySubmit)}
                className="space-y-4"
              >
                <AuthInput
                  label="Enter 6-digit Code"
                  type="text"
                  name="code"
                  placeholder="e.g. 123456"
                  register={verifyForm.register}
                  error={verifyForm.formState.errors.code?.message}
                  disabled={enableMutation.isPending}
                />
                <div className="flex gap-3">
                  <SubmitButton loading={enableMutation.isPending}>
                    {enableMutation.isPending
                      ? "Verifying..."
                      : "Verify & Activate"}
                  </SubmitButton>
                  <button
                    type="button"
                    onClick={() => {
                      setStep("idle");
                      verifyForm.reset();
                    }}
                    className="px-4 py-2 text-zinc-500 hover:text-zinc-700 dark:hover:text-zinc-300 font-medium text-sm transition-colors"
                  >
                    Cancel
                  </button>
                </div>
              </form>
            </div>
          </div>
        </div>
      )}

      {/* ─────────────────────────── BACKUP CODES ─────────────────────────── */}
      {step === "backup_codes" && setupData?.recovery_codes && (
        <div className="animate-in zoom-in-95 duration-300 space-y-4 bg-zinc-50 dark:bg-zinc-800/30 p-6 rounded-xl border border-zinc-200 dark:border-zinc-800 mt-6">
          <div className="flex justify-between items-start">
            <div>
              <h3 className="font-bold text-red-600 dark:text-red-400">
                Save your Backup Codes!
              </h3>
              <p className="text-sm text-zinc-500">
                If you lose your phone, these codes are the only way to access
                your account.
              </p>
            </div>
            <button
              onClick={() =>
                copyToClipboard(setupData.recovery_codes.join("\n"))
              }
              className="p-2 text-zinc-400 hover:text-indigo-600 transition-colors"
              title="Copy All"
            >
              <Copy className="w-5 h-5" />
            </button>
          </div>

          <div className="grid grid-cols-2 gap-3 font-mono text-sm">
            {setupData.recovery_codes.map((code, idx) => (
              <div
                key={idx}
                className="bg-white dark:bg-zinc-900 p-2 rounded border border-zinc-200 dark:border-zinc-800 text-center tracking-wider text-zinc-700 dark:text-zinc-300 select-all"
              >
                {code}
              </div>
            ))}
          </div>

          <button
            onClick={() => setStep("idle")}
            className="w-full py-2.5 bg-zinc-900 dark:bg-white text-white dark:text-black font-semibold rounded-xl hover:opacity-90 transition-opacity"
          >
            I have saved these codes
          </button>
        </div>
      )}

      {/* ─────────────────────────── DISABLE 2FA ─────────────────────────── */}
      {step === "disable" && (
        <form
          onSubmit={disableForm.handleSubmit(onDisableSubmit)}
          className="animate-in fade-in slide-in-from-top-2 duration-300 space-y-4 p-4 bg-red-50 dark:bg-red-900/10 border border-red-200 dark:border-red-900/30 rounded-xl mt-6"
        >
          <h3 className="font-semibold text-red-800 dark:text-red-200">
            Disable Two-Factor Authentication
          </h3>
          <p className="text-sm text-red-700 dark:text-red-300">
            Your account will be less secure without 2FA. Please enter your{" "}
            <b>password</b> to confirm.
          </p>

          <AuthInput
            label="Confirm Password"
            type="password"
            name="password"
            placeholder="Enter your password"
            register={disableForm.register}
            error={disableForm.formState.errors.password?.message}
            disabled={disableMutation.isPending}
          />

          <div className="flex gap-3">
            <DeleteButton loading={disableMutation.isPending}>
              {disableMutation.isPending ? "Disabling..." : "Disable 2FA"}
            </DeleteButton>
            <button
              type="button"
              onClick={() => {
                setStep("idle");
                disableForm.reset();
              }}
              className="px-4 py-2 text-zinc-500 font-medium text-sm hover:text-zinc-700 dark:hover:text-zinc-300 transition-colors"
            >
              Cancel
            </button>
          </div>
        </form>
      )}
    </div>
  );
};
