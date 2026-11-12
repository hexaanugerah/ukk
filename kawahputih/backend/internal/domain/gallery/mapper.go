package gallery

func toResponse(i *Item) Response {
	return Response{ID: i.ID, Title: i.Title, ImageURL: i.ImageURL, Category: i.Category, SortOrder: i.SortOrder, UploadedBy: i.UploadedBy}
}

func toResponseList(items []Item) []Response {
	out := make([]Response, 0, len(items))
	for i := range items {
		out = append(out, toResponse(&items[i]))
	}
	return out
}
