export interface Facility {
  id: string;
  name: string;
  description: string;
  icon: string;
  image_url: string;
  location: string;
  operating_hours: string;
  status: "active" | "maintenance" | "inactive";
  is_published: boolean;
  sort_order: number;
  created_at: string;
  updated_at: string;
}

export interface CreateFacilityPayload {
  name: string;
  description?: string;
  icon?: string;
  image_url?: string;
  location?: string;
  operating_hours?: string;
  status?: "active" | "maintenance" | "inactive";
  is_published?: boolean;
  sort_order?: number;
}

export type UpdateFacilityPayload = CreateFacilityPayload;

export const FACILITY_STATUSES = ["active", "maintenance", "inactive"] as const;
export type FacilityStatus = (typeof FACILITY_STATUSES)[number];