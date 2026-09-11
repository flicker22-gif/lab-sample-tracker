import http from './http'
import type {
  CreateResultPayload,
  CreateSamplePayload,
  CreateTransferPayload,
  Location,
  PageResult,
  Sample,
  SampleDetail,
  User,
} from '@/types'

export interface SampleQuery {
  keyword?: string
  status?: string
  page?: number
  page_size?: number
}

export const sampleApi = {
  list(params: SampleQuery) {
    return http.get<unknown, PageResult<Sample>>('/samples', { params })
  },
  detail(id: number) {
    return http.get<unknown, SampleDetail>(`/samples/${id}`)
  },
  create(payload: CreateSamplePayload) {
    return http.post<unknown, { id: number; code: string }>('/samples', payload)
  },
  createTransfer(id: number, payload: CreateTransferPayload) {
    return http.post<unknown, unknown>(`/samples/${id}/transfers`, payload)
  },
  createResult(id: number, payload: CreateResultPayload) {
    return http.post<unknown, unknown>(`/samples/${id}/results`, payload)
  },
}

export const metaApi = {
  users() {
    return http.get<unknown, User[]>('/users')
  },
  locations(type?: string) {
    return http.get<unknown, Location[]>('/locations', { params: type ? { type } : {} })
  },
}
