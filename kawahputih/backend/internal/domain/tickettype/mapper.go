package tickettype

import "time"

func toResponse(t *TicketType) Response {
	status := StatusInactive
	if t.IsActive {
		status = StatusActive
	}
	return Response{
		ID: t.ID, Code: t.Code, Name: t.Name, Description: t.Description,
		Category: t.Category, Price: t.Price, Status: status, Capacity: t.Capacity,
		Available: t.Capacity, IsActive: t.IsActive,
		CreatedAt: t.CreatedAt.Format(time.RFC3339), UpdatedAt: t.UpdatedAt.Format(time.RFC3339),
	}
}

func toResponseList(ts []TicketType) []Response {
	out := make([]Response, 0, len(ts))
	for i := range ts {
		out = append(out, toResponse(&ts[i]))
	}
	return out
}
