import { Construction } from "lucide-react";

export function ComingSoon({ feature }: { feature: string }) {
  return (
    <div className="mx-auto flex max-w-2xl flex-col items-center gap-3 px-4 py-24 text-center">
      <Construction className="h-10 w-10 text-slate-400" />
      <h1 className="text-lg font-semibold">{feature}</h1>
      <p className="text-sm text-slate-500">
        Fitur ini belum tersedia — menunggu modul backend terkait dibangun. Halaman ini sengaja ditandai jelas,
        bukan dikosongkan begitu saja, supaya tidak terlihat seperti bug.
      </p>
    </div>
  );
}
