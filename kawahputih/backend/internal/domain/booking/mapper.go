package booking

import (
	"context"
	"time"
)

func toItemResponse(items []BookingItem) []ItemResponse {
	out := make([]ItemResponse, 0, len(items))
	for _, item := range items {
		out = append(out, ItemResponse{
			TicketID: item.TicketID, TicketName: item.TicketName,
			Price: item.Price, Quantity: item.Quantity,
		})
	}
	return out
}

func timestamps(b *Booking) (string, string) {
	return b.CreatedAt.Format(time.RFC3339), b.UpdatedAt.Format(time.RFC3339)
}

func toBookingResponse(b *Booking, payment *PaymentView) BookingResponse {
	return *toBookingResponsePtr(b, payment)
}

func toBookingResponsePtr(b *Booking, payment *PaymentView) *BookingResponse {
	createdAt, updatedAt := timestamps(b)
	resp := &BookingResponse{
		ID: b.ID, UserID: b.UserID, BookingCode: b.BookingCode,
		VisitorName: b.VisitorName, Email: b.Email, Phone: b.Phone,
		VisitDate: b.VisitDate, Items: toItemResponse(b.Items),
		Subtotal: b.Subtotal, Discount: b.Discount, ServiceFee: b.ServiceFee,
		Tax: b.Tax, Total: b.Total, TotalTickets: b.TotalTickets,
		Status: b.Status, Note: b.Note, CreatedAt: createdAt, UpdatedAt: updatedAt,
		Payment: payment,
	}
	if b.CheckedInAt != nil {
		resp.CheckedInAt = b.CheckedInAt.Format(time.RFC3339)
		resp.CheckedInGate = b.CheckedInGate
	}
	return resp
}

// toBookingResponseList menambahkan snapshot pembayaran terbaru tiap booking lewat callback paymentReader (satu query terbatch agar endpoint daftar hanya dua query, bukan N+1).
func toBookingResponseList(ctx context.Context, bookings []Booking, paymentReader PaymentReader) []BookingResponse {
	out := make([]BookingResponse, 0, len(bookings))
	for i := range bookings {
		var pv *PaymentView
		if paymentReader != nil {
			pv = paymentReader(ctx, bookings[i].ID)
		}
		out = append(out, *toBookingResponsePtr(&bookings[i], pv))
	}
	return out
}
