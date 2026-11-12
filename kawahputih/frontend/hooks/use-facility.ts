import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { facilityService } from "@/services/facility.service";
import type { CreateFacilityPayload, UpdateFacilityPayload } from "@/types/facility";

export function useFacilityList(params?: { page?: number; limit?: number; status?: string; published_only?: boolean }) {
  return useQuery({
    queryKey: ["facility", params],
    queryFn: () => facilityService.list(params),
  });
}

export function useFacility(id: string) {
  return useQuery({
    queryKey: ["facility", id],
    queryFn: () => facilityService.get(id),
    enabled: !!id,
  });
}

export function usePublishedFacilityList() {
  return useQuery({
    queryKey: ["facility", "published"],
    queryFn: () => facilityService.listPublished(),
  });
}

export function useCreateFacility() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (payload: CreateFacilityPayload) => facilityService.create(payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["facility"] }),
  });
}

export function useUpdateFacility() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateFacilityPayload }) => facilityService.update(id, payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["facility"] }),
  });
}

export function useDeleteFacility() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => facilityService.remove(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["facility"] }),
  });
}