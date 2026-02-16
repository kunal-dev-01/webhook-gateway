import z from "zod";

// Schema for Verifying/Confirming 2FA (Requires 6-digit code)
export const verify2FASchema = z.object({
  code: z
    .string()
    .min(6, "Code must be 6 digits")
    .max(6, "Code must be 6 digits")
    .regex(/^\d+$/, "Code must contain only numbers"),
});

// Schema for Disabling 2FA (Requires Password per Disable2FARequest type)
export const disable2FASchema = z.object({
  password: z
    .string()
    .min(1, "Password is required to disable 2FA"),
  code: z.string().optional(), // Optional TOTP code
  recovery_code: z.string().optional(), // OR recovery code
});

// Schema for Regenerating Backup Codes (Requires Password)
export const regenerateBackupCodesSchema = z.object({
  password: z
    .string()
    .min(1, "Password is required to regenerate backup codes"),
});


export type Verify2FAForm = z.infer<typeof verify2FASchema>;
export type Disable2FAForm = z.infer<typeof disable2FASchema>;
export type RegenerateBackupCodesForm = z.infer<typeof regenerateBackupCodesSchema>;