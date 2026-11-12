package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/KawahPutih/backend/pkg/apperror"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service { return &Service{repo: repo} }

type Range struct {
	From     string // YYYY-MM-DD inclusive
	To       string // YYYY-MM-DD inclusive
	Label    string
	PrevFrom string
	PrevTo   string
}

// ResolveRange memetakan preset rentang (dan rentang kustom opsional) ke tanggal from/to serta rentang sebelumnya untuk kartu persentase perubahan.
func ResolveRange(rangePreset, customJSON, fromQ, toQ string) (*Range, error) {
	today := time.Now().Format("2006-01-02")
	day := func(iso string) time.Time {
		d, _ := time.Parse("2006-01-02", iso)
		return d
	}
	before := func(iso string, n int) string {
		return day(iso).AddDate(0, 0, -n).Format("2006-01-02")
	}

	switch rangePreset {
	case "today":
		return &Range{From: today, To: today, Label: "Hari ini", PrevFrom: before(today, 1), PrevTo: before(today, 1)}, nil
	case "30d":
		return &Range{From: before(today, 29), To: today, Label: "30 hari terakhir", PrevFrom: before(today, 59), PrevTo: before(today, 30)}, nil
	case "month":
		d := day(today)
		first := time.Date(d.Year(), d.Month(), 1, 0, 0, 0, 0, time.Local)
		prevFirst := first.AddDate(0, -1, 0)
		prevLast := first.AddDate(0, 0, -1)
		return &Range{
			From: first.Format("2006-01-02"), To: today, Label: "Bulan ini",
			PrevFrom: prevFirst.Format("2006-01-02"), PrevTo: prevLast.Format("2006-01-02"),
		}, nil
	case "custom":
		from, to := fromQ, toQ
		if from == "" || to == "" {
			// frontend mengirim rentang kustom sebagai JSON di ?custom=
			var c struct {
				From string `json:"from"`
				To   string `json:"to"`
			}
			if err := json.Unmarshal([]byte(customJSON), &c); err == nil {
				from, to = c.From, c.To
			}
		}
		if from == "" || to == "" {
			return nil, apperror.Validation("Rentang tanggal kustom memerlukan tanggal mulai dan selesai.")
		}
		n := int(day(to).Sub(day(from)).Hours()/24) + 1
		if n < 1 {
			return nil, apperror.Validation("Tanggal selesai tidak boleh sebelum tanggal mulai.")
		}
		return &Range{From: from, To: to, Label: fmt.Sprintf("%s s/d %s", from, to), PrevFrom: before(from, n), PrevTo: before(from, 1)}, nil
	default: // "7d"
		return &Range{From: before(today, 6), To: today, Label: "7 hari terakhir", PrevFrom: before(today, 13), PrevTo: before(today, 7)}, nil
	}
}

func (s *Service) AdminDashboard(ctx context.Context) (*AdminDashboardStats, error) {
	today := time.Now().Format("2006-01-02")
	totalBookings, err := s.repo.TotalBookings(ctx)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	ticketsSold, err := s.repo.TotalTicketsSold(ctx)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	revenue, err := s.repo.TotalRevenue(ctx)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	visitors, err := s.repo.VisitorsCount(ctx)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	bookingsToday, err := s.repo.BookingsCreatedOn(ctx, today)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	revenueToday, err := s.repo.TotalRevenueBetweenOn(ctx, today, today)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	available, err := s.repo.TicketsAvailable(ctx)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	pending, err := s.repo.PendingPayments(ctx)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	staff, err := s.repo.ActiveStaffCount(ctx)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	visitorsToday, err := s.repo.VisitorsOnVisitDate(ctx, today)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return &AdminDashboardStats{
		TotalBookings:    totalBookings,
		TotalTicketsSold: ticketsSold,
		Revenue:          revenue,
		TotalVisitors:    visitors,
		VisitorsToday:    visitorsToday,
		BookingsToday:    bookingsToday,
		RevenueToday:     revenueToday,
		TicketsAvailable: available,
		PendingPayments:  pending,
		ActiveStaff:      staff,
	}, nil
}

func (s *Service) ManagerAnalytics(ctx context.Context, r *Range) (*ManagerAnalytics, error) {
	rows, err := s.repo.BookingsByDateRange(ctx, r.From, r.To)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	var totalRevenue, totalTickets, totalBookings int64
	byDate := make(map[string]dailyRow, len(rows))
	for _, row := range rows {
		byDate[row.VisitDate] = row
		totalRevenue += row.Revenue
		totalTickets += row.Tickets
		totalBookings += row.Bookings
	}

	prevRows, err := s.repo.BookingsByDateRange(ctx, r.PrevFrom, r.PrevTo)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	var prevRevenue, prevBookings, prevTickets int64
	for _, row := range prevRows {
		prevRevenue += row.Revenue
		prevBookings += row.Bookings
		prevTickets += row.Tickets
	}

	revenueToday, err := s.repo.TotalRevenueBetweenOn(ctx, today2(), today2())
	if err != nil {
		return nil, apperror.Internal(err)
	}
	staff, err := s.repo.ActiveStaffCount(ctx)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	allVisitors, err := s.repo.VisitorsCount(ctx)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	visitorsInRange, err := s.repo.VisitorsInRange(ctx, r.From, r.To)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	prevVisitors, err := s.repo.VisitorsInRange(ctx, r.PrevFrom, r.PrevTo)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	// Bucket harian (dibatasi 31 titik seperti tampilan laporan frontend).
	days := int(dayOf(r.To).Sub(dayOf(r.From)).Hours()/24) + 1
	buckets := days
	if buckets > 31 {
		buckets = 31
	}
	step := int(math.Ceil(float64(days) / float64(buckets)))
	if step < 1 {
		step = 1
	}
	revenueTrend := make([]TrendPoint, 0, buckets)
	bookingTrend := make([]TrendPoint, 0, buckets)
	dailyRows := make([]DailyRow, 0, buckets)
	var weekday, weekend int64
	for i := 0; i < days; i += step {
		bucketDate := dayOf(r.From).AddDate(0, 0, i).Format("2006-01-02")
		row := byDate[bucketDate]
		label := bucketDate[5:]
		if days == 1 {
			label = "Hari ini"
		}
		revenueTrend = append(revenueTrend, TrendPoint{Label: label, Value: row.Revenue})
		bookingTrend = append(bookingTrend, TrendPoint{Label: label, Value: row.Bookings})
		dailyRows = append(dailyRows, DailyRow{
			Date: bucketDate, Bookings: row.Bookings, Tickets: row.Tickets,
			Revenue: row.Revenue, Visitors: row.Tickets,
			PaymentStatus: ternary(row.Bookings > 0, "PAID", "-"),
		})
	}
	for d, row := range byDate {
		wd := dayOf(d).Weekday()
		if wd == time.Saturday || wd == time.Sunday {
			weekend += row.Tickets
		} else {
			weekday += row.Tickets
		}
	}

	items, err := s.repo.TicketPerformance(ctx, r.From, r.To, 6)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	performance := make([]TrendPoint, 0, len(items))
	for _, it := range items {
		performance = append(performance, TrendPoint{Label: it.Name, Value: it.Quant})
	}

	payStatus, err := s.repo.PaymentsByStatus(ctx, r.From, r.To)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	paymentStatus := make([]TrendPoint, 0, 5)
	for _, st := range []string{"PAID", "PENDING", "FAILED", "EXPIRED", "REFUNDED"} {
		paymentStatus = append(paymentStatus, TrendPoint{Label: st, Value: payStatus[st]})
	}

	summary := ManagerSummary{
		TotalRevenue:    totalRevenue,
		RevenueToday:    revenueToday,
		TotalBookings:   totalBookings,
		ActiveStaff:     staff,
		TotalTickets:    totalTickets,
		TotalVisitors:   visitorsInRange,
		AvgBookingValue: 0,
		ConversionRate:  0,
	}
	if totalBookings > 0 {
		summary.AvgBookingValue = totalRevenue / totalBookings
	}
	if allVisitors > 0 {
		summary.ConversionRate = math.Round(float64(totalBookings)/float64(allVisitors)*1000) / 10
	}
	return &ManagerAnalytics{
		Summary:               summary,
		Changes:               ManagerChanges{Revenue: pct(totalRevenue, prevRevenue), Bookings: pct(totalBookings, prevBookings), Visitors: pct(visitorsInRange, prevVisitors)},
		RevenueTrend:          revenueTrend,
		BookingTrend:          bookingTrend,
		WeekdayVsWeekend:      []TrendPoint{{Label: "Weekday", Value: weekday}, {Label: "Weekend", Value: weekend}},
		TicketTypePerformance: performance,
		PaymentStatus:         paymentStatus,
		DailyRows:             dailyRows,
		RangeLabel:            r.Label,
	}, nil
}

func today2() string { return time.Now().Format("2006-01-02") }
func dayOf(iso string) time.Time {
	d, _ := time.Parse("2006-01-02", iso)
	return d
}
func ternary(cond bool, a, b string) string {
	if cond {
		return a
	}
	return b
}
func pct(current, previous int64) float64 {
	if previous == 0 {
		if current > 0 {
			return 100
		}
		return 0
	}
	return math.Round(float64(current-previous)/float64(previous)*1000) / 10
}
