import z from "zod";

// Zod Schema Definition
export const profileSchema = z.object({
  name: z.string().min(2, "Name must be at least 2 characters long"),
  email: z.string().email("Invalid email address").readonly(), 
});

// Schema for Verifying/Enabling 2FA (Requires 6-digit code)
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
});

export const changePasswordSchema = z
  .object({
    current_password: z.string().min(1, "Current password is required"),
    new_password: z
      .string()
      .min(8, "Password must be at least 8 characters")
      .regex(/[A-Z]/, "Password must contain at least one uppercase letter")
      .regex(/[a-z]/, "Password must contain at least one lowercase letter")
      .regex(/[0-9]/, "Password must contain at least one number")
      .regex(/[^A-Za-z0-9]/, "Password must contain at least one special character"),
    confirm_password: z.string().min(1, "Please confirm your password"),
  })
  .refine((data) => data.new_password === data.confirm_password, {
    message: "Passwords don't match",
    path: ["confirm_password"],
  });




// Type Definition
export type ChangePasswordFormData = z.infer<typeof changePasswordSchema>;
export type Verify2FAForm = z.infer<typeof verify2FASchema>;
export type Disable2FAForm = z.infer<typeof disable2FASchema>;
export type ProfileFormData = z.infer<typeof profileSchema>;