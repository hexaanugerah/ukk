package tourismpackage

func toResponse(p *Package) Response {
	return Response{
		ID: p.ID, Name: p.Name, Slug: p.Slug, Description: p.Description, CoverImage: p.CoverImage,
		PriceCents: p.PriceCents, Currency: p.Currency, DurationHours: p.DurationHours,
		MaxCapacity: p.MaxCapacity, IsActive: p.IsActive,
	}
}

func toResponseList(packages []Package) []Response {
	out := make([]Response, 0, len(packages))
	for i := range packages {
		out = append(out, toResponse(&packages[i]))
	}
	return out
}
