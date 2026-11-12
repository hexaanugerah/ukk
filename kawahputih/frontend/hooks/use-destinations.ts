import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { destinationService } from "@/services/destination.service";
import type { CreatePackagePayload, UpdatePackagePayload } from "@/types/destination";

export function usePackages(params?: { page?: number; limit?: number; active_only?: boolean }) {
  return useQuery({
    queryKey: ["packages", params],
    queryFn: () => destinationService.list(params),
  });
}

export function usePackage(id: string) {
  return useQuery({
    queryKey: ["packages", id],
    queryFn: () => destinationService.get(id),
    enabled: !!id,
  });
}

export function useCreatePackage() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (payload: CreatePackagePayload) => destinationService.create(payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["packages"] }),
  });
}

export function useUpdatePackage() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdatePackagePayload }) =>
      destinationService.update(id, payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["packages"] }),
  });
}

export function useDeletePackage() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => destinationService.remove(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["packages"] }),
  });
}
