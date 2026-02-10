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

// Types
import { 
  type Setup2FAResponse, 
  type Enable2FAResponse 
} from "@/types/auth.types";

import {
  disable2FASchema,
  verify2FASchema,
  type Disable2FAForm,
  type Verify2FAForm,
} from "@/schemas/user.schema";

// Icons (Adjust paths if needed)
import { Copy, Shield, TriangleAlert } from "@/svg/Common"; 
import { Smartphone } from "@/svg/Auth";
import DeleteButton from "@/common/DeleteButton";

// ─────────────────────────── Component ─────────────────────────────
export const TwoFactorAuth = () => {
  const { user, setUser } = useAuthStore();

  // Hooks
  const setupMutation = useSetup2FA();
  const enableMutation = useEnable2FA();
  const disableMutation = useDisable2FA();

  // Local State
  const [step, setStep] = useState<"idle" | "setup" | "backup_codes" | "disable">("idle");
  
  // Data States
  const [setupData, setSetupData] = useState<Setup2FAResponse | null>(null);
  const [recoveryCodes, setRecoveryCodes] = useState<string[]>([]);

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

  // 1. Start Setup (Get Secret & QR)
  const handleStartSetup = () => {
    setupMutation.setup2FA(undefined, {
      onSuccess: (data: Setup2FAResponse) => {
        setSetupData(data);
        setStep("setup");
        toast.dismiss();
      },
      onError: (error: any) => {
        const msg = error?.response?.data?.error || "Failed to initialize 2FA.";
        toast.error(msg);
      },
    });
  };

  // 2. Verify Code & Enable (Get Recovery Codes)
  const onVerifySubmit = (data: Verify2FAForm) => {
    enableMutation.enable2FA(
      { code: data.code },
      {
        onSuccess: (response: any) => { 
          // Check if response has data property or is direct response
          const result = (response.data || response) as Enable2FAResponse; 

          // Update Global Store
          if (user) {
            setUser({ ...user, two_factor_enabled: true });
          }

          toast.success("2FA Enabled Successfully!");

          // Get codes from response
          if (result.recovery_codes && result.recovery_codes.length > 0) {
            setRecoveryCodes(result.recovery_codes);
            setStep("backup_codes");
          } else {
            setStep("idle");
          }

          verifyForm.reset();
        },
        onError: (error: any) => {
          const msg = error?.response?.data?.error || "Invalid verification code.";
          toast.error(msg);
        },
      }
    );
  };

  // 3. Disable 2FA
  const onDisableSubmit = (data: Disable2FAForm) => {
    disableMutation.disable2FA(
      { password: data.password },
      {
        onSuccess: () => {
          if (user) {
            setUser({ ...user, two_factor_enabled: false });
          }
          toast.success("2FA disabled successfully.");
          setStep("idle");
          disableForm.reset();
          setSetupData(null);
          setRecoveryCodes([]);
        },
        onError: (error: any) => {
          const msg = error?.response?.data?.error || "Incorrect password.";
          toast.error(msg);
        },
      }
    );
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    toast.success("Copied to clipboard");
  };

  const is2FAEnabled = user?.two_factor_enabled;

  return (
    <div className="max-w-3xl relative">
      <div className="flex items-start gap-4 mb-6">
        <div className={`p-3 rounded-xl hidden sm:block ${is2FAEnabled ? "bg-emerald-50 text-emerald-600" : "bg-zinc-100 text-zinc-500"}`}>
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
            Secure your account with TOTP (Google Authenticator, Authy).
          </p>
        </div>
      </div>

      {/* ─────────────────────────── MAIN CARD ─────────────────────────── */}
      <div className="bg-zinc-50 dark:bg-zinc-800/50 rounded-2xl p-6 border border-zinc-100 dark:border-zinc-800 mb-6">
        <div className="flex flex-col md:flex-row items-center justify-between gap-6">
            <div className="flex items-center gap-4">
                <div className="w-12 h-12 bg-white dark:bg-zinc-800 rounded-full flex items-center justify-center border border-zinc-200 dark:border-zinc-700 shrink-0">
                <Smartphone className="w-6 h-6 text-indigo-500" />
                </div>
                <div>
                <h3 className="font-semibold text-zinc-900 dark:text-white">Authenticator App</h3>
                <p className="text-xs text-zinc-500">Use apps like Google Authenticator or Authy.</p>
                </div>
            </div>

            {/* IF NOT ENABLED: Show Main Action Button */}
            {!is2FAEnabled && step === "idle" && (
                <SubmitButton onClick={handleStartSetup} loading={setupMutation.isPending} className="w-full md:w-auto px-6">
                {setupMutation.isPending ? "Starting..." : "Enable 2FA"}
                </SubmitButton>
            )}
        </div>

        {/* IF ENABLED: Show Management Options Below */}
        {is2FAEnabled && step === "idle" && (
            <div className="flex flex-wrap gap-3 mt-6 pt-6 border-t border-zinc-200 dark:border-zinc-700">
                <button
                    onClick={handleStartSetup}
                    className="px-4 py-2 text-sm font-medium text-zinc-700 dark:text-zinc-300 bg-white dark:bg-zinc-800 border border-zinc-200 dark:border-zinc-700 rounded-xl hover:bg-zinc-50 dark:hover:bg-zinc-700 transition-colors"
                >
                    Regenerate Keys
                </button>
                <button
                    onClick={() => setStep("disable")}
                    className="px-4 py-2 text-sm font-medium text-red-600 bg-white dark:bg-zinc-800 border border-red-200 dark:border-red-900/30 rounded-xl hover:bg-red-50 dark:hover:bg-red-900/10 transition-colors"
                >
                    Disable 2FA
                </button>
            </div>
        )}
      </div>

      {/* ─────────────────────────── STEP 1: SCAN QR ─────────────────────────── */}
      {step === "setup" && setupData && (
        <div className="bg-zinc-50 dark:bg-zinc-800/30 rounded-2xl p-6 border border-zinc-200 dark:border-zinc-800 animate-in fade-in slide-in-from-top-4 duration-300">
          <div className="flex flex-col md:flex-row gap-8">
            <div className="shrink-0 flex items-center justify-center">
              <div className="p-3 bg-white rounded-xl border border-zinc-200 shadow-sm">
                <QRCodeSVG value={setupData.qr_code_url} size={180} level="M" />
              </div>
            </div>

            <div className="flex-1 space-y-4">
              <div>
                <h3 className="font-semibold text-zinc-900 dark:text-white">1. Scan QR Code</h3>
                <p className="text-sm text-zinc-500 mb-2">Open your authenticator app and scan the image.</p>
                <div className="flex items-center gap-2 bg-white dark:bg-zinc-900 p-2 rounded-lg border border-zinc-200 dark:border-zinc-700">
                  <code className="text-xs text-zinc-600 dark:text-zinc-400 font-mono flex-1 truncate">{setupData.secret}</code>
                  <button onClick={() => copyToClipboard(setupData.secret)} className="text-zinc-400 hover:text-indigo-500 p-1">
                    <Copy className="w-4 h-4" />
                  </button>
                </div>
              </div>

              <form onSubmit={verifyForm.handleSubmit(onVerifySubmit)} className="space-y-4">
                <AuthInput
                  label="2. Enter 6-digit Code"
                  type="text"
                  name="code"
                  placeholder="000 000"
                  register={verifyForm.register}
                  error={verifyForm.formState.errors.code?.message}
                  disabled={enableMutation.isPending}
                />
                <div className="flex gap-3">
                  <SubmitButton loading={enableMutation.isPending} className="w-full">
                    {enableMutation.isPending ? "Verifying..." : "Verify & Enable"}
                  </SubmitButton>
                  <button
                    type="button"
                    onClick={() => {
                      setStep("idle");
                      verifyForm.reset();
                    }}
                    className="px-4 py-2 text-zinc-500 hover:text-zinc-700 font-medium text-sm"
                  >
                    Cancel
                  </button>
                </div>
              </form>
            </div>
          </div>
        </div>
      )}

      {/* ─────────────────────────── STEP 2: BACKUP CODES ─────────────────────────── */}
      {step === "backup_codes" && recoveryCodes.length > 0 && (
        <div className="animate-in zoom-in-95 duration-300 space-y-5 bg-amber-50/50 dark:bg-amber-900/10 p-6 rounded-2xl border border-amber-200/50 dark:border-amber-800/30">
          <div className="flex items-start gap-3">
            <TriangleAlert className="w-6 h-6 text-amber-600 shrink-0 mt-1" />
            <div className="flex-1">
              <h3 className="font-bold text-amber-800 dark:text-amber-500 text-lg">Save these Backup Codes</h3>
              <p className="text-sm text-amber-700/80 dark:text-amber-400/80 mt-1">
                If you lose access to your phone, these codes are the <b>only way</b> to recover your account. Store them somewhere safe.
              </p>
            </div>
          </div>

          <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
            {recoveryCodes.map((code, idx) => (
              <div key={idx} className="bg-white dark:bg-zinc-900 p-2.5 rounded-lg border border-amber-200 dark:border-zinc-700 text-center font-mono text-sm text-zinc-700 dark:text-zinc-300 select-all shadow-sm">
                {code}
              </div>
            ))}
          </div>

          <div className="flex flex-col sm:flex-row gap-3 pt-2">
            <button
              onClick={() => copyToClipboard(recoveryCodes.join("\n"))}
              className="flex items-center justify-center gap-2 px-4 py-2.5 bg-white dark:bg-zinc-800 border border-amber-300 dark:border-amber-800/50 text-amber-800 dark:text-amber-500 rounded-xl hover:bg-amber-50 dark:hover:bg-zinc-700 font-medium transition-colors"
            >
              <Copy className="w-4 h-4" /> Copy All Codes
            </button>
            <button
              onClick={() => setStep("idle")}
              className="flex-1 px-4 py-2.5 bg-amber-600 text-white rounded-xl hover:bg-amber-700 font-medium shadow-sm transition-colors"
            >
              I have saved these codes
            </button>
          </div>
        </div>
      )}

      {/* ─────────────────────────── DISABLE FORM ─────────────────────────── */}
      {step === "disable" && (
        <div className="animate-in fade-in slide-in-from-top-2 duration-300 bg-red-50 dark:bg-red-900/10 border border-red-200 dark:border-red-900/30 rounded-2xl p-6">
          <form onSubmit={disableForm.handleSubmit(onDisableSubmit)} className="space-y-4">
            <div>
                <h3 className="font-semibold text-red-900 dark:text-red-200 flex items-center gap-2 text-lg">
                    Disable Two-Factor Authentication
                </h3>
                <p className="text-sm text-red-700 dark:text-red-300 mt-1">
                    Enter your password to confirm. This will significantly lower your account security.
                </p>
            </div>

            <div className="flex flex-col sm:flex-row gap-4 items-start sm:items-end">
                <div className="flex-1 w-full">
                <AuthInput
                    label="Current Password"
                    type="password"
                    name="password"
                    placeholder="••••••••"
                    register={disableForm.register}
                    error={disableForm.formState.errors.password?.message}
                    disabled={disableMutation.isPending}
                />
                </div>
                <div className="flex gap-3 w-full sm:w-auto">
                    <DeleteButton 
                        loading={disableMutation.isPending} 
                        className="px-6"
                    >
                        {disableMutation.isPending ? "Disabling..." : "Confirm Disable"}
                    </DeleteButton>
                    <button
                        type="button"
                        onClick={() => {
                        setStep("idle");
                        disableForm.reset();
                        }}
                        className="px-4 py-2.5 bg-white dark:bg-zinc-800 border border-red-200 dark:border-red-900/30 text-red-700 dark:text-red-400 font-medium rounded-xl dark:hover:bg-zinc-700 hover:bg-red-50 transition-colors"
                    >
                        Cancel
                    </button>
                </div>
            </div>
          </form>
        </div>
      )}
    </div>
  );
};