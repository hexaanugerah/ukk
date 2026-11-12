"use client";

import { useState } from "react";
import { Plus, Minus } from "lucide-react";
import type { FAQ } from "@/types/faq";

interface FaqAccordionProps {
  items?: FAQ[];
}

export function FaqAccordion({ items = [] }: FaqAccordionProps) {
  const [open, setOpen] = useState<number | null>(0);
  return (
    <div className="divide-y divide-[#e5edf3] rounded-md border border-[#d7e3ed] bg-white">
      {items.length === 0 ? (
        <p className="px-5 py-4 text-sm text-[#698097]">Belum ada FAQ yang dipublikasikan.</p>
      ) : (
        items.map((item, i) => {
          const isOpen = open === i;
          return (
            <div key={item.id}>
              <button
                onClick={() => setOpen(isOpen ? null : i)}
                className="flex w-full items-center justify-between gap-4 px-5 py-4 text-left"
                aria-expanded={isOpen}
              >
                <span className="text-sm font-semibold text-[#17324d]">{item.question}</span>
                <span className="grid h-6 w-6 shrink-0 place-items-center rounded-full bg-[#e9f3fb] text-[#1768ad]">
                  {isOpen ? <Minus className="h-3.5 w-3.5" /> : <Plus className="h-3.5 w-3.5" />}
                </span>
              </button>
              {isOpen && <p className="px-5 pb-4 text-sm leading-relaxed text-[#698097]">{item.answer}</p>}
            </div>
          );
        })
      )}
    </div>
  );
}