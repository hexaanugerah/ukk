"use client";

import { useState } from "react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useRegister } from "@/hooks/use-auth";
import { PhoneInput } from "@/components/ui/phone-input";
import { validatePhoneFull } from "@/lib/phone";
import { registerSchema, sanitizeInput, sanitizeEmail, type RegisterFormValues } from "../schemas/auth.schema";
import { toast } from "sonner";

// Register form — mendaftarkan akun PENGUNJUNG baru ke backend (/auth/register). Setelah sukses, user otomatis login dan diarahkan ke beranda.

export function RegisterForm() {
  const register = useRegister();
  const [values, setValues] = useState<RegisterFormValues>({
    name: "",
    email: "",
    phone: "",
    password: "",
    confirmPassword: "",
  });
  const [errors, setErrors] = useState<Partial<Record<keyof RegisterFormValues | "root", string>>>({});
  const [submitting, setSubmitting] = useState(false);

  function set<K extends keyof RegisterFormValues>(key: K, val: RegisterFormValues[K]) {
    setValues((v) => ({ ...v, [key]: val }));
  }

  // Validasi real-time untuk nama
  const handleNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = sanitizeInput(e.target.value);
    set("name", value);
    
    if (value && value.length < 3) {
      setErrors((prev) => ({ ...prev, name: "Nama minimal 3 karakter" }));
    } else if (value.length > 100) {
      setErrors((prev) => ({ ...prev, name: "Nama maksimal 100 karakter" }));
    } else if (/[<>]/.test(value)) {
      setErrors((prev) => ({ ...prev, name: "Nama tidak boleh mengandung karakter khusus" }));
    } else {
      setErrors((prev) => {
        const next = { ...prev };
        delete next.name;
        return next;
      });
    }
  };

  // Validasi real-time untuk email
  const handleEmailChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = sanitizeEmail(e.target.value);
    set("email", value);
    
    if (value && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value)) {
      setErrors((prev) => ({ ...prev, email: "Format email tidak valid" }));
    } else {
      setErrors((prev) => {
        const next = { ...prev };
        delete next.email;
        return next;
      });
    }
  };

  // Validasi real-time untuk telepon (nomor lengkap E.164 dari PhoneInput)
  const handlePhoneChange = (value: string) => {
    set("phone", value);

    if (value) {
      const validation = validatePhoneFull(value);
      if (!validation.valid) {
        setErrors((prev) => ({ ...prev, phone: validation.message! }));
      } else {
        setErrors((prev) => {
          const next = { ...prev };
          delete next.phone;
          return next;
        });
      }
    } else {
      setErrors((prev) => {
        const next = { ...prev };
        delete next.phone;
        return next;
      });
    }
  };

  // Validasi real-time untuk password
  const handlePasswordChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    set("password", value);
    
    if (value && value.length < 8) {
      setErrors((prev) => ({ ...prev, password: "Password minimal 8 karakter" }));
    } else if (value && value.length > 72) {
      setErrors((prev) => ({ ...prev, password: "Password maksimal 72 karakter" }));
    } else if (value && !/[A-Z]/.test(value)) {
      setErrors((prev) => ({ ...prev, password: "Password harus mengandung huruf besar" }));
    } else if (value && !/[a-z]/.test(value)) {
      setErrors((prev) => ({ ...prev, password: "Password harus mengandung huruf kecil" }));
    } else if (value && !/[0-9]/.test(value)) {
      setErrors((prev) => ({ ...prev, password: "Password harus mengandung angka" }));
    } else {
      setErrors((prev) => {
        const next = { ...prev };
        delete next.password;
        return next;
      });
    }
  };

  // Validasi real-time untuk konfirmasi password
  const handleConfirmPasswordChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    set("confirmPassword", value);
    
    if (value && value !== values.password) {
      setErrors((prev) => ({ ...prev, confirmPassword: "Konfirmasi password tidak cocok" }));
    } else {
      setErrors((prev) => {
        const next = { ...prev };
        delete next.confirmPassword;
        return next;
      });
    }
  };

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    
    // Bersihkan input sebelum validasi
    const sanitizedData = {
      name: sanitizeInput(values.name),
      email: sanitizeEmail(values.email),
      phone: values.phone,
      password: values.password,
      confirmPassword: values.confirmPassword,
    };
    
    const parsed = registerSchema.safeParse(sanitizedData);
    if (!parsed.success) {
      const fieldErrors: Partial<Record<keyof RegisterFormValues, string>> = {};
      for (const issue of parsed.error.issues) {
        const key = issue.path[0] as keyof RegisterFormValues;
        if (key && !fieldErrors[key]) fieldErrors[key] = issue.message;
      }
      setErrors(fieldErrors);
      return;
    }

    setErrors({});
    setSubmitting(true);
    const res = await register({
      name: parsed.data.name,
      email: parsed.data.email,
      phone: parsed.data.phone ?? "",
      password: parsed.data.password,
    });
    setSubmitting(false);

    if (res.ok) {
      toast.success("Registrasi berhasil", {
        description: "Link verifikasi telah dikirim ke email Anda. Verifikasi sebelum login berikutnya.",
      });
    } else {
      setErrors({ root: res.error });
      toast.error("Registrasi gagal", { description: res.error });
    }
  }

  const label = "mb-1 block text-sm font-medium";

  return (
    <form onSubmit={onSubmit} className="space-y-4" noValidate>
      <div>
        <label className={label}>Nama Lengkap</label>
        <Input 
          value={values.name} 
          onChange={handleNameChange} 
          autoComplete="name"
          aria-invalid={!!errors.name}
          aria-describedby={errors.name ? "name-error" : undefined}
        />
        {errors.name && <p id="name-error" className="mt-1 text-xs text-red-600" role="alert">{errors.name}</p>}
      </div>

      <div>
        <label className={label}>Email</label>
        <Input 
          type="email" 
          autoComplete="email" 
          value={values.email} 
          onChange={handleEmailChange}
          aria-invalid={!!errors.email}
          aria-describedby={errors.email ? "email-error" : undefined}
        />
        {errors.email && <p id="email-error" className="mt-1 text-xs text-red-600" role="alert">{errors.email}</p>}
      </div>

      <div>
        <label className={label}>No. Telepon (opsional)</label>
        <PhoneInput
          value={values.phone ?? ""}
          onChange={handlePhoneChange}
          invalid={!!errors.phone}
        />
        {errors.phone && <p id="phone-error" className="mt-1 text-xs text-red-600" role="alert">{errors.phone}</p>}
      </div>

      <div>
        <label className={label}>Password</label>
        <Input 
          type="password" 
          autoComplete="new-password" 
          value={values.password} 
          onChange={handlePasswordChange}
          aria-invalid={!!errors.password}
          aria-describedby={errors.password ? "password-error" : undefined}
        />
        {errors.password && <p id="password-error" className="mt-1 text-xs text-red-600" role="alert">{errors.password}</p>}
      </div>

      <div>
        <label className={label}>Konfirmasi Password</label>
        <Input 
          type="password" 
          autoComplete="new-password" 
          value={values.confirmPassword} 
          onChange={handleConfirmPasswordChange}
          aria-invalid={!!errors.confirmPassword}
          aria-describedby={errors.confirmPassword ? "confirm-error" : undefined}
        />
        {errors.confirmPassword && <p id="confirm-error" className="mt-1 text-xs text-red-600" role="alert">{errors.confirmPassword}</p>}
      </div>

      {errors.root && (
        <p role="alert" className="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700">
          {String(errors.root)}
        </p>
      )}

      <Button type="submit" className="w-full" loading={submitting}>
        {submitting ? "Memproses..." : "Daftar"}
      </Button>

      <p className="text-center text-sm text-slate-500">
        Sudah punya akun?{" "}
        <Link href="/login" className="font-medium text-brand-600 hover:underline">
          Masuk
        </Link>
      </p>
    </form>
  );
}