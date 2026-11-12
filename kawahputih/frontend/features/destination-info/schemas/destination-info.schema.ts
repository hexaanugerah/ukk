import { z } from "zod";

export const destinationInfoSchema = z.object({
  title: z.string().min(3, "Judul minimal 3 karakter").max(200),
  description: z.string().min(10, "Deskripsi minimal 10 karakter"),
  history: z.string().optional().default(""),
  destination_desc: z.string().optional().default(""),
  location: z.string().max(500).optional().default(""),
  operating_hours: z.string().max(200).optional().default(""),
  contact: z.string().max(500).optional().default(""),
  additional_info: z.string().optional().default(""),
});

export type DestinationInfoFormValues = z.infer<typeof destinationInfoSchema>;