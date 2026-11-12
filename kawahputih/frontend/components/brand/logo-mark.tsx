// Logo mark — lambang Kawah Putih (satu-satunya sumber SVG brand). Dipakai navbar/panel/login agar bentuk logo konsisten.

export function LogoMark({ className = "h-6 w-[37px]" }: { className?: string }) {
  return (
    <svg viewBox="0 0 72 46" className={className} aria-hidden="true">
      <path d="M0 42 21 11l9 18 8-12 26 25H0Z" fill="#0d5da6" />
      <path d="M28 42 43 0l29 42H28Z" fill="#1768ad" />
      <path d="M43 0 36 19l9-6 8 13 3-9L43 0Z" fill="#ffffff" />
      <path d="M21 11 15 23l7-4 5 10 3-5-9-13Z" fill="#ffffff" />
      <path d="M0 42h72v4H0Z" fill="#14548f" />
    </svg>
  );
}
