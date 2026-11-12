"use client";

import { useEffect, useRef, useState } from "react";
import jsQR from "jsqr";
import { Camera, CameraOff, CheckCircle2, Loader2, ScanLine, XCircle } from "lucide-react";
import { PageHeader, SectionCard } from "@/components/panel/panel-ui";
import { useAuthStore } from "@/store/auth.store";
import { scannerService } from "@/services/scanner.api.service";
import { toast } from "sonner";

// Scan QR — kamera membaca QR tiket, lalu hasilnya divalidasi ke backend.
// Payload QR yang dipakai e-tiket adalah `qr_data` dari backend, sehingga
// pemindaian langsung terhubung ke data tiket yang sama dengan yang dilihat pengunjung.

type ScanKind = "VALID" | "USED" | "INVALID" | "EXPIRED" | "CANCELLED";

interface ScanResult {
  kind: ScanKind;
  message: string;
  ticket?: {
    id: string;
    booking_code: string;
    visitor_name: string;
    visit_date: string;
    ticket_name: string;
    quantity: number;
    status: string;
    checked_in_at?: string;
    checked_in_gate?: string;
  };
}

const RESULT_STYLE: Record<ScanKind, string> = {
  VALID: "border-[#16876a] bg-[#f0fbf7] text-[#16876a]",
  USED: "border-[#b7791f] bg-[#fdf3e3] text-[#8a6420]",
  INVALID: "border-[#c53030] bg-[#fdeaea] text-[#c53030]",
  EXPIRED: "border-[#b7791f] bg-[#fdf3e3] text-[#8a6420]",
  CANCELLED: "border-[#c53030] bg-[#fdeaea] text-[#c53030]",
};

function defaultMessage(status: string): string {
  switch (status) {
    case "VALID":
      return "Tiket valid — pengunjung dipersilakan masuk";
    case "USED":
      return "Tiket sudah digunakan";
    case "EXPIRED":
      return "Tiket kadaluarsa";
    case "CANCELLED":
      return "Tiket dibatalkan";
    default:
      return "Tiket tidak dapat digunakan";
  }
}

export default function PetugasScanPage() {
  const user = useAuthStore((s) => s.user);
  const [code, setCode] = useState("");
  const [scanning, setScanning] = useState(false);
  const [result, setResult] = useState<ScanResult | null>(null);
  const [cameraOn, setCameraOn] = useState(false);
  const [cameraError, setCameraError] = useState<string | null>(null);

  const videoRef = useRef<HTMLVideoElement | null>(null);
  const canvasRef = useRef<HTMLCanvasElement | null>(null);
  const streamRef = useRef<MediaStream | null>(null);
  const scanningRef = useRef(false);
  const lastScanRef = useRef<{ value: string; at: number }>({ value: "", at: 0 });

  const gate = user?.gate ?? "Gerbang 1";
  const petugas = user?.name ?? "Petugas";

  async function runScan(input: string) {
    const ref = input.trim();
    if (!ref) {
      toast.error("Masukkan atau pindai kode tiket terlebih dahulu");
      return;
    }
    if (scanningRef.current) return;
    scanningRef.current = true;
    setScanning(true);
    setResult(null);
    try {
      const response = await scannerService.validateTicket({ ticket_id: ref, gate, petugas });
      const kind = (response.status ?? "INVALID") as ScanKind;
      setResult({
        kind,
        message: response.message || defaultMessage(kind),
        ticket: response.ticket,
      });
      if (kind === "VALID") {
        toast.success("Tiket valid — pengunjung dipersilakan masuk");
      } else if (kind === "USED") {
        toast.warning("Tiket sudah digunakan");
      } else {
        toast.error("Tiket tidak dapat digunakan");
      }
    } catch (error: any) {
      const message = error?.message || "Terjadi kesalahan saat memvalidasi tiket";
      toast.error(message);
      setResult({ kind: "INVALID", message });
    } finally {
      scanningRef.current = false;
      setScanning(false);
    }
  }

  async function startCamera() {
    setCameraError(null);
    if (!navigator.mediaDevices?.getUserMedia) {
      setCameraError("Browser ini tidak mendukung akses kamera. Gunakan input kode manual.");
      return;
    }
    try {
      const stream = await navigator.mediaDevices.getUserMedia({
        video: { facingMode: { ideal: "environment" } },
        audio: false,
      });
      streamRef.current = stream;
      setCameraOn(true);
    } catch {
      setCameraError("Izin kamera ditolak atau kamera tidak tersedia. Gunakan input kode manual.");
    }
  }

  function stopCamera() {
    streamRef.current?.getTracks().forEach((track) => track.stop());
    streamRef.current = null;
    setCameraOn(false);
  }

  // Loop pembacaan frame kamera dengan jsQR. Deteksi hasil yang sama ditahan
  // sebentar agar satu QR tidak mengirim validasi berulang.
  useEffect(() => {
    if (!cameraOn) return;
    const video = videoRef.current;
    const canvas = canvasRef.current;
    const stream = streamRef.current;
    if (!video || !canvas || !stream) return;

    video.srcObject = stream;
    void video.play().catch(() => setCameraError("Kamera tidak dapat diputar."));

    let frame = 0;
    let lastTick = 0;
    const tick = (timestamp: number) => {
      frame = requestAnimationFrame(tick);
      if (timestamp - lastTick < 350) return;
      lastTick = timestamp;
      if (scanningRef.current || video.readyState < 2) return;

      const width = video.videoWidth;
      const height = video.videoHeight;
      if (!width || !height) return;

      canvas.width = width;
      canvas.height = height;
      const ctx = canvas.getContext("2d", { willReadFrequently: true });
      if (!ctx) return;
      ctx.drawImage(video, 0, 0, width, height);
      const image = ctx.getImageData(0, 0, width, height);
      const decoded = jsQR(image.data, width, height, { inversionAttempts: "dontInvert" });
      if (!decoded?.data) return;

      const now = Date.now();
      if (decoded.data === lastScanRef.current.value && now - lastScanRef.current.at < 4000) return;
      lastScanRef.current = { value: decoded.data, at: now };
      setCode(decoded.data);
      void runScan(decoded.data);
    };

    frame = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(frame);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [cameraOn]);

  // Bersihkan kamera saat halaman ditinggalkan.
  useEffect(() => stopCamera, []);

  return (
    <div>
      <PageHeader
        title="Scan QR Code"
        subtitle={`Pindai QR tiket pengunjung — ${gate}, Shift ${user?.shift ?? "Pagi"}`}
        action={
          cameraOn ? (
            <button
              onClick={stopCamera}
              className="inline-flex h-9 items-center gap-1.5 rounded-md border border-[#d7e3ed] bg-white px-3 text-xs font-semibold text-[#29445e] hover:bg-[#f2f7fb]"
            >
              <CameraOff className="h-4 w-4" /> Hentikan Kamera
            </button>
          ) : (
            <button
              onClick={startCamera}
              className="inline-flex h-9 items-center gap-1.5 rounded-md bg-[#1768ad] px-3 text-xs font-semibold text-white hover:bg-[#14548f]"
            >
              <Camera className="h-4 w-4" /> Aktifkan Kamera
            </button>
          )
        }
      />

      <SectionCard bodyClassName="p-0">
        <div className="grid min-h-[360px] place-items-center px-6 py-6">
          <div className="w-full max-w-[380px] text-center">
            <div className="mx-auto mb-[-15px] h-9 w-40 rounded-full border-4 border-[#1768ad] bg-white pt-1 text-xs font-extrabold tracking-wide text-[#1768ad]">
              {scanning ? "MEMVALIDASI" : cameraOn ? "ARAHKAN KE QR" : "SCAN HERE"}
            </div>
            <div className="relative grid aspect-square w-full place-items-center overflow-hidden rounded-lg border-4 border-[#1768ad] bg-black">
              {cameraOn ? (
                <video
                  ref={videoRef}
                  playsInline
                  muted
                  className="h-full w-full object-cover"
                  aria-label="Pratinjau kamera scanner"
                />
              ) : (
                <div className="w-[70%] space-y-2 bg-white p-4">
                  {Array.from({ length: 11 }).map((_, i) => (
                    <div
                      key={i}
                      className="h-2 rounded-sm bg-black"
                      style={{
                        width: `${i % 3 === 0 ? 100 : i % 3 === 1 ? 82 : 92}%`,
                        marginLeft: i % 2 ? "auto" : 0,
                      }}
                    />
                  ))}
                </div>
              )}
              {scanning && (
                <div className="absolute inset-0 grid place-items-center bg-black/40">
                  <Loader2 className="h-12 w-12 animate-spin text-white" />
                </div>
              )}
            </div>
            <canvas ref={canvasRef} className="hidden" aria-hidden="true" />
          </div>
        </div>

        <div className="border-t border-[#e5edf3] bg-white px-6 py-4">
          {cameraError && (
            <p className="mb-3 rounded-md bg-[#fdeaea] px-3 py-2 text-xs text-[#c53030]">{cameraError}</p>
          )}
          <p className="mb-3 text-sm text-[#698097]">
            Pastikan QR code sesuai dengan tiket pengunjung sebelum melakukan validasi.
          </p>
          <div className="flex flex-wrap gap-2">
            <input
              value={code}
              onChange={(e) => setCode(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && runScan(code)}
              placeholder="Masukkan kode tiket / QR manual"
              className="h-10 flex-1 rounded-md border border-[#d7e3ed] px-3 text-sm outline-none focus:border-[#1768ad]"
              aria-label="Kode tiket"
            />
            <button
              onClick={() => runScan(code)}
              disabled={scanning}
              className="inline-flex h-10 items-center gap-1.5 rounded-md bg-[#1768ad] px-4 text-sm font-semibold text-white hover:bg-[#14548f] disabled:opacity-60"
            >
              <ScanLine className="h-4 w-4" /> Validasi
            </button>
          </div>

          {result && (
            <div className={`mt-4 rounded-md border px-4 py-3 ${RESULT_STYLE[result.kind] ?? RESULT_STYLE.INVALID}`}>
              <div className="flex items-start gap-2">
                {result.kind === "VALID" ? (
                  <CheckCircle2 className="mt-0.5 h-5 w-5 shrink-0" />
                ) : (
                  <XCircle className="mt-0.5 h-5 w-5 shrink-0" />
                )}
                <div className="min-w-0">
                  <p className="text-sm font-bold">{result.message}</p>
                  {result.ticket && (
                    <p className="mt-1 text-xs">
                      {result.ticket.visitor_name} • {result.ticket.booking_code} • {result.ticket.ticket_name} ×{" "}
                      {result.ticket.quantity} • kunjungan {result.ticket.visit_date}
                    </p>
                  )}
                </div>
              </div>
            </div>
          )}

          <p className="mt-2 text-[11px] text-[#94A3B8]">
            Pemindaian kamera berjalan di perangkat, lalu divalidasi ke backend dan dicatat ke riwayat aktivitas.
          </p>
        </div>
      </SectionCard>
    </div>
  );
}
