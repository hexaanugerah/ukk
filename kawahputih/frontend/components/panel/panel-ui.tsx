import type { ReactNode } from "react";
import { cn } from "@/lib/utils";

/* Komponen visual bersama untuk panel Admin / Manager / Petugas. */

export function PageHeader({
  title,
  subtitle,
  kicker,
  action,
}: {
  title: string;
  subtitle?: string;
  kicker?: string;
  action?: ReactNode;
}) {
  return (
    <div className="mb-4 flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
      <div>
        {kicker && <p className="text-[10px] font-semibold uppercase tracking-wide text-[#698097]">{kicker}</p>}
        <h1 className="text-[19px] font-extrabold leading-tight text-[#14548f]">{title}</h1>
        {subtitle && <p className="mt-1.5 text-[12px] text-[#6f8bad]">{subtitle}</p>}
      </div>
      {action}
    </div>
  );
}

export function SectionCard({
  title,
  subtitle,
  action,
  className,
  bodyClassName,
  children,
}: {
  title?: string;
  subtitle?: string;
  action?: ReactNode;
  className?: string;
  bodyClassName?: string;
  children: ReactNode;
}) {
  return (
    <section className={cn("min-w-0 max-w-full rounded-lg border border-[#c8dded] bg-white", className)}>
      {(title || action) && (
        <header className="flex flex-wrap items-center justify-between gap-2 border-b border-[#dce8f2] px-4 py-2.5">
          <div>
            {title && <h2 className="text-[13px] font-extrabold text-[#14548f]">{title}</h2>}
            {subtitle && <p className="mt-0.5 text-[10px] text-[#698097]">{subtitle}</p>}
          </div>
          {action}
        </header>
      )}
      <div className={cn("min-w-0 p-4", bodyClassName)}>{children}</div>
    </section>
  );
}

export function StatCard({
  label,
  value,
  delta,
  trend = "up",
  icon,
  accent = false,
}: {
  label: string;
  value: string;
  delta?: string;
  trend?: "up" | "down";
  icon?: ReactNode;
  accent?: boolean;
}) {
  return (
    <div
      className={cn(
        "min-h-[78px] rounded-lg border px-4 py-2.5",
        accent ? "border-[#1768ad] bg-[#1768ad] text-white" : "border-[#d7e3ed] bg-white"
      )}
    >
      <div className="flex items-start justify-between gap-2">
        <p className={cn("text-[12px] font-medium", accent ? "text-blue-100" : "text-[#58779a]")}>{label}</p>
        {icon && <span className={cn("shrink-0 [&>svg]:h-3 [&>svg]:w-3", accent ? "text-blue-100" : "text-[#1768ad]")}>{icon}</span>}
      </div>
      <p className={cn("mt-2.5 text-[19px] font-extrabold leading-none tabular-nums", accent ? "text-white" : "text-[#174f8d]")}>{value}</p>
      {delta && (
        <p className={cn("mt-1.5 text-[10px]", accent ? "text-blue-100" : trend === "up" ? "text-[#0f9d64]" : "text-red-600")}>
          {trend === "up" ? "^" : "v"} {delta}
        </p>
      )}
    </div>
  );
}

export function StatGrid({ children, cols = 4 }: { children: ReactNode; cols?: 2 | 3 | 4 }) {
  return (
    <div
      className={cn(
        "grid gap-2.5",
        cols === 2 && "grid-cols-1 sm:grid-cols-2",
        cols === 3 && "grid-cols-1 sm:grid-cols-3",
        cols === 4 && "grid-cols-1 sm:grid-cols-2 xl:grid-cols-4"
      )}
    >
      {children}
    </div>
  );
}

const STATUS_TONE: Record<string, string> = {
  berhasil: "bg-[#e6f7f0] text-[#16876a]",
  valid: "bg-[#e6f7f0] text-[#16876a]",
  aktif: "bg-[#e6f7f0] text-[#16876a]",
  selesai: "bg-[#e9f3fb] text-[#1768ad]",
  menunggu: "bg-[#fdf3e3] text-[#b7791f]",
  "menunggu pembayaran": "bg-[#fdf3e3] text-[#b7791f]",
  dibatalkan: "bg-[#fdeaea] text-[#c53030]",
  gagal: "bg-[#fdeaea] text-[#c53030]",
  nonaktif: "bg-[#eef2f6] text-[#698097]",
  bermasalah: "bg-[#fdeaea] text-[#c53030]",
};

export function StatusPill({ status }: { status: string }) {
  const tone = STATUS_TONE[status.toLowerCase()] ?? "bg-[#e9f3fb] text-[#1768ad]";
  return <span className={cn("inline-block rounded-full px-2 py-0.5 text-[10px] font-bold", tone)}>{status}</span>;
}

export function PanelTable({
  head,
  children,
  className,
}: {
  head: ReactNode[];
  children: ReactNode;
  className?: string;
}) {
  return (
    <div className={cn("panel-scroll -mx-4 -mb-4 mt-1", className)}>
      <table className="panel-table">
        <thead>
          <tr>
            {head.map((h, i) => (
              <th key={i}>{h}</th>
            ))}
          </tr>
        </thead>
        <tbody>{children}</tbody>
      </table>
    </div>
  );
}

export function LineChart({
  data,
  height = 175,
  color = "#1768ad",
}: {
  data: { label: string; value: number }[];
  height?: number;
  color?: string;
}) {
  const width = 960;
  const padX = 28;
  const padTop = 24;
  const padBottom = 24;
  const values = data.map((d) => d.value);
  const max = Math.max(...values, 1);
  const min = Math.min(...values, 0);
  const range = Math.max(max - min, 1);
  const innerW = width - padX * 2;
  const innerH = height - padTop - padBottom;
  const points = data.map((d, i) => {
    const x = padX + (data.length === 1 ? innerW / 2 : (i / (data.length - 1)) * innerW);
    const y = padTop + innerH - ((d.value - min) / range) * innerH;
    return { x, y, label: d.label };
  });
  const path = points.map((p, i) => `${i === 0 ? "M" : "L"} ${p.x} ${p.y}`).join(" ");

  return (
    <div className="w-full">
      <svg viewBox={`0 0 ${width} ${height}`} className="h-auto w-full" role="img" aria-label="Grafik garis">
        {[0, 1, 2, 3, 4].map((i) => {
          const y = padTop + (innerH / 4) * i;
          return <line key={i} x1={padX} x2={width - padX} y1={y} y2={y} stroke="#dfe7ef" strokeWidth="1" />;
        })}
        <line x1={padX} x2={padX} y1={padTop} y2={height - padBottom} stroke="#b7d0e7" strokeWidth="1.5" />
        <line x1={padX} x2={width - padX} y1={height - padBottom} y2={height - padBottom} stroke="#b7d0e7" strokeWidth="1.5" />
        <path d={path} fill="none" stroke={color} strokeLinecap="round" strokeLinejoin="round" strokeWidth="5" />
        {points.map((p) => (
          <circle key={`${p.x}-${p.y}`} cx={p.x} cy={p.y} r="8" fill="white" stroke={color} strokeWidth="5" />
        ))}
      </svg>
    </div>
  );
}

export function BarChart({
  data,
  height = 145,
  color = "#1768ad",
  valueSuffix = "",
}: {
  data: { label: string; value: number }[];
  height?: number;
  color?: string;
  valueSuffix?: string;
}) {
  const max = Math.max(...data.map((d) => d.value), 1);
  return (
    <div className="flex items-end gap-5 sm:gap-8" style={{ height }}>
      {data.map((d) => (
        <div key={d.label} className="flex h-full min-w-0 flex-1 flex-col items-center justify-end gap-1.5">
          <div
            className="w-full max-w-[110px] rounded-t-lg transition-all"
            style={{ height: `${Math.max(18, (d.value / max) * 100)}%`, background: color, opacity: 0.95 }}
            role="img"
            aria-label={`${d.label}: ${d.value}`}
          />
          <span className="truncate text-[11px] text-[#698097]">{d.label}</span>
        </div>
      ))}
    </div>
  );
}

export function PieChart({
  data,
  size = 175,
  colors = ["#1768ad", "#4b99cf", "#9ec7e5", "#d6eafa"],
}: {
  data: { label: string; value: number }[];
  size?: number;
  colors?: string[];
}) {
  const total = Math.max(data.reduce((sum, d) => sum + d.value, 0), 1);
  const radius = 67;
  const cx = size / 2;
  const cy = size / 2;
  let start = -90;

  function point(angle: number) {
    const rad = (Math.PI / 180) * angle;
    return { x: cx + radius * Math.cos(rad), y: cy + radius * Math.sin(rad) };
  }

  return (
    <div className="grid gap-5 lg:grid-cols-[minmax(180px,240px)_1fr] lg:items-center">
      <svg viewBox={`0 0 ${size} ${size}`} className="mx-auto h-auto w-full max-w-[185px]" role="img" aria-label="Grafik komposisi">
        {data.map((d, i) => {
          const angle = (d.value / total) * 360;
          const end = start + angle;
          const large = angle > 180 ? 1 : 0;
          const a = point(start);
          const b = point(end);
          const path = [`M ${cx} ${cy}`, `L ${a.x} ${a.y}`, `A ${radius} ${radius} 0 ${large} 1 ${b.x} ${b.y}`, "Z"].join(" ");
          start = end;
          return <path key={d.label} d={path} fill={colors[i % colors.length]} />;
        })}
        <circle cx={cx} cy={cy} r="32" fill="white" />
      </svg>
      <ul className="space-y-2">
        {data.map((d, i) => (
          <li key={d.label} className="flex items-center justify-between gap-4 text-[12px]">
            <span className="flex items-center gap-2 text-[#29445e]">
              <span className="h-2.5 w-2.5 rounded-sm" style={{ background: colors[i % colors.length] }} />
              {d.label}
            </span>
            <span className="font-bold tabular-nums text-[#17324d]">{d.value.toLocaleString("id-ID")}</span>
          </li>
        ))}
      </ul>
    </div>
  );
}

export function ProgressList({
  data,
  color = "#1768ad",
}: {
  data: { name: string; value: number }[];
  color?: string;
}) {
  const max = Math.max(...data.map((d) => d.value), 1);
  return (
    <ul className="space-y-2">
      {data.map((d) => (
        <li key={d.name}>
          <div className="mb-1 flex items-center justify-between text-[11px]">
            <span className="font-medium text-[#29445e]">{d.name}</span>
            <span className="font-semibold tabular-nums text-[#698097]">{d.value.toLocaleString("id-ID")}</span>
          </div>
          <div className="h-2 overflow-hidden rounded-full bg-[#e7eff7]">
            <div className="h-full rounded-full" style={{ width: `${(d.value / max) * 100}%`, background: color }} />
          </div>
        </li>
      ))}
    </ul>
  );
}

export function InfoTile({
  title,
  desc,
  href,
  badge,
}: {
  title: string;
  desc: string;
  href: string;
  badge?: string;
}) {
  return (
    <a
      href={href}
      className="group flex min-h-[62px] items-center gap-3 rounded-md border border-[#e0ebf4] bg-white px-4 py-3 transition-colors hover:border-[#1768ad] hover:bg-[#f7fbff]"
    >
      <div className="grid h-8 w-8 shrink-0 place-items-center rounded-md bg-[#eef7fd] text-sm">
        {badge ?? "[]"}
      </div>
      <div className="min-w-0">
        <h3 className="text-[12px] font-extrabold text-[#174f8d]">{title}</h3>
        <p className="mt-0.5 text-[10px] leading-relaxed text-[#698097]">{desc}</p>
      </div>
    </a>
  );
}

export function ListPanel({
  items,
}: {
  items: { name: string; value: string | number }[];
}) {
  return (
    <ul className="divide-y divide-[#eef2f6]">
      {items.map((it) => (
        <li key={it.name} className="flex items-center justify-between py-2 text-[12px]">
          <span className="text-[#29445e]">{it.name}</span>
          <span className="font-semibold tabular-nums text-[#17324d]">{typeof it.value === "number" ? it.value.toLocaleString("id-ID") : it.value}</span>
        </li>
      ))}
    </ul>
  );
}
