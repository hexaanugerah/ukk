"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { Calendar, ChevronLeft, ChevronRight } from "lucide-react";
import {
  addMonths,
  eachDayOfInterval,
  endOfMonth,
  endOfWeek,
  format,
  isAfter,
  isBefore,
  isSameDay,
  isSameMonth,
  parseISO,
  startOfDay,
  startOfMonth,
  startOfWeek,
  subMonths,
} from "date-fns";
import { id as localeId } from "date-fns/locale";
import { cn } from "@/lib/utils";

// Date picker — input tanggal dengan kalender bergaya situs (menggantikan input date bawaan browser). Nilai memakai format ISO yyyy-MM-dd.

const WEEKDAYS = ["Sen", "Sel", "Rab", "Kam", "Jum", "Sab", "Min"];

interface DatePickerProps {
  value: string;
  onChange: (iso: string) => void;
  min?: string;
  max?: string;
  placeholder?: string;
  id?: string;
  className?: string;
  disabled?: boolean;
  "aria-invalid"?: boolean;
  "aria-describedby"?: string;
}

export function DatePicker({
  value,
  onChange,
  min,
  max,
  placeholder = "Pilih tanggal",
  id,
  className,
  disabled = false,
  ...aria
}: DatePickerProps) {
  const [open, setOpen] = useState(false);
  const selected = value ? parseISO(value) : null;
  const [viewMonth, setViewMonth] = useState<Date>(() => startOfMonth(selected ?? new Date()));
  const wrapRef = useRef<HTMLDivElement>(null);

  const minDate = min ? startOfDay(parseISO(min)) : null;
  const maxDate = max ? startOfDay(parseISO(max)) : null;

  // Tutup saat klik di luar atau menekan Escape.
  useEffect(() => {
    if (!open) return;
    function onPointerDown(e: MouseEvent) {
      if (wrapRef.current && !wrapRef.current.contains(e.target as Node)) setOpen(false);
    }
    function onKeyDown(e: KeyboardEvent) {
      if (e.key === "Escape") setOpen(false);
    }
    document.addEventListener("mousedown", onPointerDown);
    document.addEventListener("keydown", onKeyDown);
    return () => {
      document.removeEventListener("mousedown", onPointerDown);
      document.removeEventListener("keydown", onKeyDown);
    };
  }, [open]);

  // Saat dibuka, tampilkan bulan dari nilai terpilih (atau bulan ini).
  function openCalendar() {
    setViewMonth(startOfMonth(selected ?? new Date()));
    setOpen((v) => !v);
  }

  const days = useMemo(() => {
    const start = startOfWeek(startOfMonth(viewMonth), { weekStartsOn: 1 });
    const end = endOfWeek(endOfMonth(viewMonth), { weekStartsOn: 1 });
    return eachDayOfInterval({ start, end });
  }, [viewMonth]);

  const canPrev = !minDate || isAfter(startOfMonth(viewMonth), startOfMonth(minDate));
  const canNext = !maxDate || isBefore(startOfMonth(viewMonth), startOfMonth(maxDate));

  function isDisabled(d: Date) {
    const day = startOfDay(d);
    if (minDate && isBefore(day, minDate)) return true;
    if (maxDate && isAfter(day, maxDate)) return true;
    return false;
  }

  return (
    <div ref={wrapRef} className="relative">
      <button
        type="button"
        id={id}
        onClick={openCalendar}
        disabled={disabled}
        aria-haspopup="dialog"
        aria-expanded={open}
        {...aria}
        className={cn(
          "flex h-10 w-full items-center gap-2 rounded-md border border-[#d7e3ed] bg-white px-3 text-left text-sm outline-none transition-colors focus:border-[#1768ad] disabled:cursor-not-allowed disabled:bg-[#f7fbff]",
          aria["aria-invalid"] ? "border-[#c53030]" : "",
          className
        )}
      >
        <Calendar className="h-4 w-4 shrink-0 text-[#94A3B8]" aria-hidden="true" />
        <span className={cn("truncate", selected ? "font-medium text-[#17324d]" : "text-[#94A3B8]")}>
          {selected ? format(selected, "d MMMM yyyy", { locale: localeId }) : placeholder}
        </span>
      </button>

      {open && (
        <div
          role="dialog"
          aria-label="Pilih tanggal"
          className="absolute left-0 z-30 mt-2 w-[304px] rounded-md border border-[#d7e3ed] bg-white p-3 shadow-lg"
        >
          <div className="mb-2 flex items-center justify-between">
            <button
              type="button"
              onClick={() => canPrev && setViewMonth((m) => subMonths(m, 1))}
              disabled={!canPrev}
              aria-label="Bulan sebelumnya"
              className="grid h-8 w-8 place-items-center rounded-md border border-[#d7e3ed] text-[#29445e] hover:bg-[#f2f7fb] disabled:cursor-not-allowed disabled:opacity-40"
            >
              <ChevronLeft className="h-4 w-4" />
            </button>
            <p className="text-sm font-bold text-[#17324d]">{format(viewMonth, "MMMM yyyy", { locale: localeId })}</p>
            <button
              type="button"
              onClick={() => canNext && setViewMonth((m) => addMonths(m, 1))}
              disabled={!canNext}
              aria-label="Bulan berikutnya"
              className="grid h-8 w-8 place-items-center rounded-md border border-[#d7e3ed] text-[#29445e] hover:bg-[#f2f7fb] disabled:cursor-not-allowed disabled:opacity-40"
            >
              <ChevronRight className="h-4 w-4" />
            </button>
          </div>

          <div className="grid grid-cols-7 gap-1">
            {WEEKDAYS.map((d) => (
              <span key={d} className="grid h-7 place-items-center text-[10px] font-bold uppercase tracking-wide text-[#94A3B8]">
                {d}
              </span>
            ))}
            {days.map((d) => {
              const isSelected = selected ? isSameDay(d, selected) : false;
              const isToday = isSameDay(d, new Date());
              return (
                <button
                  key={d.toISOString()}
                  type="button"
                  disabled={isDisabled(d)}
                  onClick={() => {
                    onChange(format(d, "yyyy-MM-dd"));
                    setOpen(false);
                  }}
                  aria-current={isSelected ? "date" : undefined}
                  className={cn(
                    "grid h-9 w-9 place-items-center rounded-md text-sm tabular-nums transition-colors",
                    !isSameMonth(d, viewMonth) && "text-[#c3cfda]",
                    isDisabled(d)
                      ? "cursor-not-allowed text-[#c3cfda] line-through"
                      : "hover:bg-[#eaf4fd]",
                    isToday && !isSelected && "ring-1 ring-inset ring-[#bcd8ee]",
                    isSelected && "bg-[#1768ad] font-bold text-white hover:bg-[#14548f]"
                  )}
                >
                  {d.getDate()}
                </button>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
}
