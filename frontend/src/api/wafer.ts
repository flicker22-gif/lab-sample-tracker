import http from './http'
import type { MapVersion, Wafer, WaferMap } from '@/types/wafer'

export const waferApi = {
  listByLot(lotId: number) {
    return http.get<unknown, Wafer[]>(`/lots/${lotId}/wafers`)
  },
  versions(waferId: number) {
    return http.get<unknown, MapVersion[]>(`/wafers/${waferId}/maps`)
  },
  getMap(id: number) {
    return http.get<unknown, WaferMap>(`/maps/${id}`)
  },
  upload(waferId: number, file: File, operatorId: number, remark = '') {
    const fd = new FormData()
    fd.append('file', file)
    fd.append('operator_id', String(operatorId))
    if (remark) fd.append('remark', remark)
    return http.post<unknown, { id: number; version: number; total_dies: number; pass_dies: number; yield: number }>(
      `/wafers/${waferId}/maps`,
      fd,
      { headers: { 'Content-Type': 'multipart/form-data' } },
    )
  },
  downloadUrl(id: number) {
    return `/api/v1/maps/${id}/download`
  },
}
