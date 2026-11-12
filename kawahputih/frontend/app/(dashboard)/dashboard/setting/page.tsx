"use client";

import { PageHeader } from "@/components/panel/panel-ui";
import { Card } from "@/components/ui/card";
import { ChangePasswordForm } from "@/features/authentication/components/change-password-form";

// Keamanan akun pengunjung — ganti password akun yang sedang login.

export default function SettingPage() {
  return (
    <div>
      <PageHeader title="Keamanan Akun" subtitle="Kelola sandi akun kamu" />
      <Card className="max-w-md p-6">
        <h2 className="text-base font-semibold text-[#14548f]">Ubah Password</h2>
        <p className="mb-4 mt-1 text-xs text-slate-500">
          Masukkan password lama untuk memastikan perubahan dilakukan oleh pemilik akun.
        </p>
        <ChangePasswordForm />
      </Card>
    </div>
  );
}
