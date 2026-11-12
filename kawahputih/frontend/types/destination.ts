export interface TourismPackage {
  id: string;
  name: string;
  slug: string;
  description: string;
  cover_image: string;
  price_cents: number;
  currency: string;
  duration_hours: number;
  max_capacity: number;
  is_active: boolean;
}

export interface CreatePackagePayload {
  name: string;
  description: string;
  cover_image?: string;
  price_cents: number;
  duration_hours: number;
  max_capacity: number;
}

export type UpdatePackagePayload = CreatePackagePayload & { is_active: boolean };
