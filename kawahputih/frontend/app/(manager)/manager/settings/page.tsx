"use client";

import { PageHeader, SectionCard } from "@/components/panel/panel-ui";
import { ChangePasswordForm } from "@/features/authentication/components/change-password-form";

// Keamanan akun manager — ganti password akun yang sedang login.

export default function ManagerSettingsPage() {
  return (
    <div>
      <PageHeader title="Keamanan Akun" subtitle="Kelola sandi akun manager" />
      <SectionCard title="Ubah Password" subtitle="Password lama wajib diverifikasi" className="max-w-xl">
        <ChangePasswordForm className="max-w-sm" />
      </SectionCard>
    </div>
  );
}
