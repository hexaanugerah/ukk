export interface Article {
  id: string;
  title: string;
  slug: string;
  excerpt: string;
  content: string;
  cover_image: string;
  status: "draft" | "published" | "archived";
  author_id: string;
  published_at?: number;
}

export interface CreateArticlePayload {
  title: string;
  excerpt: string;
  content: string;
  cover_image?: string;
}
