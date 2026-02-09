import z from "zod";

// Zod Schema Definition
export const profileSchema = z.object({
  name: z.string().min(2, "Name must be at least 2 characters long"),
  email: z.string().email("Invalid email address").readonly(), 
});

// Type Definition
export type ProfileFormData = z.infer<typeof profileSchema>;