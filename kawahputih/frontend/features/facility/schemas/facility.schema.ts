import { z } from "zod";
import { FACILITY_STATUSES } from "@/types/facility";

export const facilitySchema = z.object({
  name: z.string().min(3, "Nama fasilitas minimal 3 karakter").max(150),
  description: z.string().max(500).optional().default(""),
  icon: z.string().max(200).optional().default(""),
  image_url: z.string().url("URL gambar tidak valid").optional().or(z.literal("")),
  location: z.string().max(200).optional().default(""),
  operating_hours: z.string().max(200).optional().default(""),
  status: z.enum(FACILITY_STATUSES).optional().default("active"),
  is_published: z.boolean().optional().default(false),
  sort_order: z.number().int().min(0).optional().default(0),
});

export type FacilityFormValues = z.infer<typeof facilitySchema>;