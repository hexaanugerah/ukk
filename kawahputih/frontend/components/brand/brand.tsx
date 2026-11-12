import Link from "next/link";
import { Mountain } from "lucide-react";

export function Brand({ subtitle, light = false }: { subtitle?: string; light?: boolean }) {
  return (
    <Link href="/" className={`inline-flex items-center gap-2.5 ${light ? "text-white" : "text-[#145b98]"}`}>
      <span className="grid h-9 w-9 place-items-center rounded-md bg-[#e7f2fb] text-[#1768ad]">
        <Mountain className="h-5 w-5" strokeWidth={2.4} />
      </span>
      <span className="block leading-tight">
        <span className="block text-[17px] font-bold">Kawah Putih</span>
        {subtitle && <span className={`block text-[11px] ${light ? "text-blue-100" : "text-[#698097]"}`}>{subtitle}</span>}
      </span>
    </Link>
  );
}
