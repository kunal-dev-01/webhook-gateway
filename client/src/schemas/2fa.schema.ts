import z from "zod";

// Schema for Verifying/Confirming 2FA (Requires 6-digit code)
export const verify2FASchema = z.object({
  code: z
    .string()
    .min(6, "Code must be 6 digits")
    .max(6, "Code must be 6 digits")
    .regex(/^\d+$/, "Code must contain only numbers"),
});

// Schema for Disabling 2FA (Updated)
export const disable2FASchema = z.object({
  password: z
    .string()
    .min(1, "Password is required to disable 2FA"),
  code: z.string().optional(),
  recovery_code: z.string().optional(),
}).refine((data) => data.code || data.recovery_code, {
  // Ye check karega ki dono mein se kam se kam ek value jarur ho
  message: "Please enter verification code or recovery code",
  path: ["code"], 
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