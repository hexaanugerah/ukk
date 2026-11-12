export interface GalleryItem {
  id: string;
  title: string;
  image_url: string;
  category: string;
  sort_order: number;
  uploaded_by: string;
}

export interface CreateGalleryItemPayload {
  title: string;
  image_url: string;
  category?: string;
  sort_order?: number;
}
