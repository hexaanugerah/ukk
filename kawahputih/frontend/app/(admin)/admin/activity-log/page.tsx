import { redirect } from "next/navigation";

// Aktivitas admin kini ditampilkan di /admin/audit-log.
export default function Page() {
  redirect("/admin/audit-log");
}
