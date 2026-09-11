import http from './http'
import type { PageResult } from '@/types'
import type {
  CreateLotPayload,
  EndProcessPayload,
  LotDetail,
  Machine,
  MachineStatusPayload,
  Process,
  StartProcessPayload,
  StatusLog,
  WaferLot,
} from '@/types/production'

export interface LotQuery {
  keyword?: string
  status?: string
  page?: number
  page_size?: number
}

export const productionApi = {
  // 工艺
  processes() {
    return http.get<unknown, Process[]>('/processes')
  },
  // 机台
  machines(processId?: number) {
    return http.get<unknown, Machine[]>('/machines', {
      params: processId ? { process_id: processId } : {},
    })
  },
  updateMachineStatus(id: number, payload: MachineStatusPayload) {
    return http.put<unknown, unknown>(`/machines/${id}/status`, payload)
  },
  statusLogs(id: number) {
    return http.get<unknown, StatusLog[]>(`/machines/${id}/status-logs`)
  },

  // 批次
  listLots(params: LotQuery) {
    return http.get<unknown, PageResult<WaferLot>>('/lots', { params })
  },
  getLot(id: number) {
    return http.get<unknown, LotDetail>(`/lots/${id}`)
  },
  createLot(payload: CreateLotPayload) {
    return http.post<unknown, { id: number; lot_no: string }>('/lots', payload)
  },
  start(id: number, payload: StartProcessPayload) {
    return http.post<unknown, unknown>(`/lots/${id}/start`, payload)
  },
  end(id: number, payload: EndProcessPayload) {
    return http.post<unknown, unknown>(`/lots/${id}/end`, payload)
  },
  complete(id: number) {
    return http.post<unknown, unknown>(`/lots/${id}/complete`, {})
  },
}
