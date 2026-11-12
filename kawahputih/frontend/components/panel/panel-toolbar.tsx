"use client";

import { useMemo, useState, type ReactNode } from "react";
import { Search } from "lucide-react";
import { cn } from "@/lib/utils";

/* Komponen toolbar interaktif (client) untuk panel. */

export function SearchInput({
  value,
  onChange,
  placeholder = "Cari…",
  className,
}: {
  value: string;
  onChange: (v: string) => void;
  placeholder?: string;
  className?: string;
}) {
  return (
    <div className={cn("relative", className)}>
      <Search className="pointer-events-none absolute left-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-[#94A3B8]" />
      <input
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        className="h-7 w-full rounded-md border border-[#d7e3ed] bg-white pl-8 pr-2.5 text-[11px] text-[#17324d] outline-none placeholder:text-[#94A3B8] focus:border-[#1768ad]"
      />
    </div>
  );
}

export function FilterSelect({
  value,
  onChange,
  options,
  label = "Semua status",
}: {
  value: string;
  onChange: (v: string) => void;
  options: string[];
  label?: string;
}) {
  return (
    <select
      value={value}
      onChange={(e) => onChange(e.target.value)}
      className="h-7 rounded-md border border-[#d7e3ed] bg-white px-2.5 text-[11px] text-[#29445e] outline-none focus:border-[#1768ad]"
    >
      <option value="">{label}</option>
      {options.map((o) => (
        <option key={o} value={o}>
          {o}
        </option>
      ))}
    </select>
  );
}

export function ToggleGroup({
  options,
  value,
  onChange,
}: {
  options: string[];
  value: string;
  onChange: (v: string) => void;
}) {
  return (
    <div className="inline-flex flex-wrap gap-1 rounded-md border border-[#d7e3ed] bg-white p-1">
      {options.map((o) => (
        <button
          key={o}
          type="button"
          onClick={() => onChange(o)}
          className={cn(
            "rounded px-2.5 py-0.5 text-[10px] font-semibold transition-colors",
            value === o ? "bg-[#1768ad] text-white" : "text-[#29445e] hover:bg-[#eef5fb]"
          )}
        >
          {o}
        </button>
      ))}
    </div>
  );
}

export function DateField({ label = "Pilih tanggal" }: { label?: string }) {
  const [date, setDate] = useState("");
  return (
    <label className="inline-flex h-7 items-center gap-2 rounded-md border border-[#d7e3ed] bg-white px-2.5 text-[11px] text-[#29445e]">
      <input
        type="date"
        value={date}
        onChange={(e) => setDate(e.target.value)}
        className="bg-transparent outline-none"
      />
      {!date && <span className="pointer-events-none -ml-[6.5rem] text-[#94A3B8]">{label}</span>}
    </label>
  );
}

export interface FColumn<T> {
  header: string;
  render: (row: T) => ReactNode;
  className?: string;
}

export function FilterableTable<T extends { id: string }>({
  columns,
  rows,
  searchPlaceholder = "Cari…",
  searchKeys,
  statusKey,
  statusOptions,
  toolbarExtra,
}: {
  columns: FColumn<T>[];
  rows: T[];
  searchPlaceholder?: string;
  searchKeys: (keyof T)[];
  statusKey?: keyof T;
  statusOptions?: string[];
  toolbarExtra?: ReactNode;
}) {
  const [q, setQ] = useState("");
  const [status, setStatus] = useState("");

  const filtered = useMemo(() => {
    const needle = q.trim().toLowerCase();
    return rows.filter((row) => {
      const matchesQ =
        !needle ||
        searchKeys.some((k) => String(row[k] ?? "").toLowerCase().includes(needle));
      const matchesStatus = !status || !statusKey || String(row[statusKey]) === status;
      return matchesQ && matchesStatus;
    });
  }, [q, status, rows, searchKeys, statusKey]);

  return (
    <div>
      <div className="mb-2.5 flex flex-wrap items-center gap-2">
        <SearchInput value={q} onChange={setQ} placeholder={searchPlaceholder} className="w-full sm:max-w-xs" />
        {statusOptions && statusKey && (
          <FilterSelect value={status} onChange={setStatus} options={statusOptions} />
        )}
        <div className="ml-auto flex items-center gap-2">{toolbarExtra}</div>
      </div>
      <div className="panel-scroll rounded-md border border-[#e5edf3]">
        <table className="panel-table">
          <thead>
            <tr>
              {columns.map((c, i) => (
                <th key={i} className={c.className}>
                  {c.header}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {filtered.length === 0 ? (
              <tr>
                <td colSpan={columns.length} className="py-8 text-center text-[11px] text-[#94A3B8]">
                  Tidak ada data yang cocok.
                </td>
              </tr>
            ) : (
              filtered.map((row) => (
                <tr key={row.id}>
                  {columns.map((c, i) => (
                    <td key={i} className={c.className}>
                      {c.render(row)}
                    </td>
                  ))}
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
