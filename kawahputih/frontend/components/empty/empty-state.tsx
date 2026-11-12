import { Inbox } from "lucide-react";

export function EmptyState({ title = "Belum ada data", description }: { title?: string; description?: string }) {
  return (
    <div className="flex flex-col items-center justify-center gap-2 rounded-2xl border border-dashed border-slate-300 py-16 text-center dark:border-slate-700">
      <Inbox className="h-10 w-10 text-slate-400" />
      <p className="font-medium text-slate-700 dark:text-slate-300">{title}</p>
      {description && <p className="text-sm text-slate-500">{description}</p>}
    </div>
  );
}
