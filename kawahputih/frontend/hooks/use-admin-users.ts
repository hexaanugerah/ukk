import { useQuery } from "@tanstack/react-query";
import { staffService } from "@/services/staff.api.service";
import { visitorService } from "@/services/visitor.api.service";
import type { Staff, Visitor } from "@/types/api";

// Hook data petugas/pengunjung untuk panel admin (backend API).
export function useStaff() {
  return useQuery<Staff[]>({
    queryKey: ["admin", "staff"],
    queryFn: () => staffService.list(),
  });
}

export function useVisitors() {
  return useQuery<Visitor[]>({
    queryKey: ["admin", "visitors"],
    queryFn: () => visitorService.list(),
  });
}
