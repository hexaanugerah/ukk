export interface DestinationInfo {
  id: string;
  title: string;
  description: string;
  history: string;
  destination_desc: string;
  location: string;
  operating_hours: string;
  contact: string;
  additional_info: string;
  status: "draft" | "published" | "archived";
  author_id: string;
  published_at?: number;
  created_at: string;
  updated_at: string;
}

export interface CreateDestinationInfoPayload {
  title: string;
  description: string;
  history?: string;
  destination_desc?: string;
  location?: string;
  operating_hours?: string;
  contact?: string;
  additional_info?: string;
}

export type UpdateDestinationInfoPayload = CreateDestinationInfoPayload & { status?: "draft" | "published" | "archived" };