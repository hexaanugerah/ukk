package article

func toResponse(a *Article) Response {
	return Response{
		ID: a.ID, Title: a.Title, Slug: a.Slug, Excerpt: a.Excerpt, Content: a.Content,
		CoverImage: a.CoverImage, Status: a.Status, AuthorID: a.AuthorID, PublishedAt: a.PublishedAt,
	}
}

func toResponseList(articles []Article) []Response {
	out := make([]Response, 0, len(articles))
	for i := range articles {
		out = append(out, toResponse(&articles[i]))
	}
	return out
}
