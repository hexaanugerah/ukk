import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { destinationInfoService } from "@/services/destination-info.service";
import type { CreateDestinationInfoPayload, UpdateDestinationInfoPayload } from "@/types/destination-info";

export function useDestinationInfoList(params?: { page?: number; limit?: number; status?: string }) {
  return useQuery({
    queryKey: ["destination-info", params],
    queryFn: () => destinationInfoService.list(params),
  });
}

export function useDestinationInfo(id: string) {
  return useQuery({
    queryKey: ["destination-info", id],
    queryFn: () => destinationInfoService.get(id),
    enabled: !!id,
  });
}

export function usePublishedDestinationInfo() {
  return useQuery({
    queryKey: ["destination-info", "published"],
    queryFn: () => destinationInfoService.getPublished(),
  });
}

export function useCreateDestinationInfo() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (payload: CreateDestinationInfoPayload) => destinationInfoService.create(payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["destination-info"] }),
  });
}

export function useUpdateDestinationInfo() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateDestinationInfoPayload }) => destinationInfoService.update(id, payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["destination-info"] }),
  });
}

export function usePublishDestinationInfo() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => destinationInfoService.publish(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["destination-info"] }),
  });
}

export function useArchiveDestinationInfo() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => destinationInfoService.archive(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["destination-info"] }),
  });
}

export function useDeleteDestinationInfo() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => destinationInfoService.remove(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["destination-info"] }),
  });
}