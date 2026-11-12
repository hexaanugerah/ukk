import { z } from "zod";

export const loginSchema = z.object({
  email: z
    .string()
    .min(1, "Email wajib diisi")
    .email("Format email tidak valid")
    .max(150, "Email maksimal 150 karakter"),
  password: z
    .string()
    .min(1, "Password wajib diisi")
    .min(8, "Password minimal 8 karakter")
    .max(72, "Password maksimal 72 karakter"),
});
export type LoginFormValues = z.infer<typeof loginSchema>;

export const registerSchema = z
  .object({
    name: z
      .string()
      .min(3, "Nama minimal 3 karakter")
      .max(100, "Nama maksimal 100 karakter")
      .regex(/^[^<>]*$/, "Nama tidak boleh mengandung karakter khusus"),
    email: z
      .string()
      .min(1, "Email wajib diisi")
      .email("Format email tidak valid")
      .max(150, "Email maksimal 150 karakter"),
    phone: z
      .string()
      .optional()
      .refine((val) => !val || /^(\+62|62|0)8[0-9]{8,11}$/.test(val), {
        message: "Format nomor telepon Indonesia tidak valid",
      }),
    password: z
      .string()
      .min(8, "Password minimal 8 karakter")
      .max(72, "Password maksimal 72 karakter")
      .regex(/[A-Z]/, "Password harus mengandung huruf besar")
      .regex(/[a-z]/, "Password harus mengandung huruf kecil")
      .regex(/[0-9]/, "Password harus mengandung angka"),
    confirmPassword: z.string().min(1, "Konfirmasi password wajib diisi"),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: "Konfirmasi password tidak cocok",
    path: ["confirmPassword"],
  });
export type RegisterFormValues = z.infer<typeof registerSchema>;

export const forgotPasswordSchema = z.object({
  email: z
    .string()
    .min(1, "Email wajib diisi")
    .email("Format email tidak valid")
    .max(150, "Email maksimal 150 karakter"),
});
export type ForgotPasswordFormValues = z.infer<typeof forgotPasswordSchema>;

export const resetPasswordSchema = z
  .object({
    token: z.string().min(1, "Token reset wajib diisi"),
    password: z
      .string()
      .min(8, "Password minimal 8 karakter")
      .max(72, "Password maksimal 72 karakter")
      .regex(/[A-Z]/, "Password harus mengandung huruf besar")
      .regex(/[a-z]/, "Password harus mengandung huruf kecil")
      .regex(/[0-9]/, "Password harus mengandung angka"),
    confirmPassword: z.string().min(1, "Konfirmasi password wajib diisi"),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: "Konfirmasi password tidak cocok",
    path: ["confirmPassword"],
  });
export type ResetPasswordFormValues = z.infer<typeof resetPasswordSchema>;

export const changePasswordSchema = z
  .object({
    currentPassword: z
      .string()
      .min(1, "Password saat ini wajib diisi")
      .min(8, "Password minimal 8 karakter")
      .max(72, "Password maksimal 72 karakter"),
    newPassword: z
      .string()
      .min(8, "Password baru minimal 8 karakter")
      .max(72, "Password baru maksimal 72 karakter")
      .regex(/[A-Z]/, "Password harus mengandung huruf besar")
      .regex(/[a-z]/, "Password harus mengandung huruf kecil")
      .regex(/[0-9]/, "Password harus mengandung angka"),
    confirmPassword: z.string().min(1, "Konfirmasi password wajib diisi"),
  })
  .refine((data) => data.newPassword === data.confirmPassword, {
    message: "Konfirmasi password tidak cocok",
    path: ["confirmPassword"],
  })
  .refine((data) => data.currentPassword !== data.newPassword, {
    message: "Password baru harus berbeda dari password saat ini",
    path: ["newPassword"],
  });
export type ChangePasswordFormValues = z.infer<typeof changePasswordSchema>;

// Helper validasi
export function sanitizeInput(input: string): string {
  return input.replace(/\0/g, "").trim().slice(0, 10000);
}

export function sanitizeEmail(email: string): string {
  return email.toLowerCase().trim();
}

export function validatePhone(phone: string): { valid: boolean; message?: string } {
  const cleaned = phone.replace(/[\s-]/g, "");
  let normalized = cleaned;
  
  if (cleaned.startsWith("+62")) {
    normalized = "0" + cleaned.slice(3);
  } else if (cleaned.startsWith("62")) {
    normalized = "0" + cleaned.slice(2);
  }
  
  if (!normalized.startsWith("08")) {
    return { valid: false, message: "Nomor telepon harus diawali 08" };
  }
  
  if (normalized.length < 10 || normalized.length > 13) {
    return { valid: false, message: "Nomor telepon harus 10-13 digit" };
  }
  
  return { valid: true };
}