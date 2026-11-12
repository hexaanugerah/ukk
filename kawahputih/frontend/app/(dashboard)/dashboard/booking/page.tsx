import type { Metadata } from "next";
import { MyBookingsList } from "@/features/booking/components/my-bookings-list";
import { PageHeader } from "@/components/panel/panel-ui";

export const metadata: Metadata = { title: "Booking Saya" };

export default function MyBookingsPage() {
  return (
    <div>
      <PageHeader title="Booking Saya" subtitle="Riwayat & status pemesanan tiket kamu" />
      <MyBookingsList />
    </div>
  );
}
