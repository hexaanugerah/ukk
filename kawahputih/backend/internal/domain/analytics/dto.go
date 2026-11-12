package analytics

// AdminDashboardStats cocok dengan tipe AdminDashboardStats frontend.
type AdminDashboardStats struct {
	TotalBookings    int64 `json:"total_bookings"`
	TotalTicketsSold int64 `json:"total_tickets_sold"`
	Revenue          int64 `json:"revenue"`
	TotalVisitors    int64 `json:"total_visitors"`
	VisitorsToday    int64 `json:"visitors_today"`
	BookingsToday    int64 `json:"bookings_today"`
	RevenueToday     int64 `json:"revenue_today"`
	TicketsAvailable int64 `json:"tickets_available"`
	PendingPayments  int64 `json:"pending_payments"`
	ActiveStaff      int64 `json:"active_staff"`
}

// ManagerAnalytics cocok dengan tipe ManagerAnalytics frontend.
type ManagerAnalytics struct {
	Summary               ManagerSummary `json:"summary"`
	Changes               ManagerChanges `json:"changes"`
	RevenueTrend          []TrendPoint   `json:"revenue_trend"`
	BookingTrend          []TrendPoint   `json:"booking_trend"`
	WeekdayVsWeekend      []TrendPoint   `json:"weekday_vs_weekend"`
	TicketTypePerformance []TrendPoint   `json:"ticket_type_performance"`
	PaymentStatus         []TrendPoint   `json:"payment_status"`
	DailyRows             []DailyRow     `json:"daily_rows"`
	RangeLabel            string         `json:"range_label"`
}

type ManagerSummary struct {
	TotalRevenue    int64   `json:"total_revenue"`
	RevenueToday    int64   `json:"revenue_today"`
	TotalBookings   int64   `json:"total_bookings"`
	ActiveStaff     int64   `json:"active_staff"`
	TotalTickets    int64   `json:"total_tickets"`
	TotalVisitors   int64   `json:"total_visitors"`
	AvgBookingValue int64   `json:"avg_booking_value"`
	ConversionRate  float64 `json:"conversion_rate"`
}

type ManagerChanges struct {
	Revenue  float64 `json:"revenue"`
	Bookings float64 `json:"bookings"`
	Visitors float64 `json:"visitors"`
}

type TrendPoint struct {
	Label string `json:"label"`
	Value int64  `json:"value"`
}

type DailyRow struct {
	Date          string `json:"date"`
	Bookings      int64  `json:"bookings"`
	Tickets       int64  `json:"tickets"`
	Revenue       int64  `json:"revenue"`
	Visitors      int64  `json:"visitors"`
	PaymentStatus string `json:"payment_status"`
}
