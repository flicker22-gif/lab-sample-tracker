export interface PageResult<T> {
  list: T[]
  total: number
  page: number
  page_size: number
}

export type SampleStatus =
  | 'received'
  | 'in_transit'
  | 'testing'
  | 'stored'
  | 'reported'
  | 'discarded'

export interface Sample {
  id: number
  code: string
  name: string
  category: string
  source: string
  status: SampleStatus
  remark: string
  receiver_id: number
  receiver_name: string
  current_location_id: number | null
  current_location_name: string
  current_location_type: string
  received_at: string
  created_at: string
}

export type TransferAction = 'receive' | 'transfer' | 'load' | 'unload' | 'store' | 'discard'

export interface Transfer {
  id: number
  action: TransferAction
  from_location_id: number | null
  from_location_name: string
  from_location_type: string
  to_location_id: number
  to_location_name: string
  to_location_type: string
  operator_id: number
  operator_name: string
  occurred_at: string
  note: string
}

export type Conclusion = 'qualified' | 'unqualified' | 'na' | ''

export interface TestResult {
  id: number
  item_name: string
  method: string
  instrument: string
  result_value: string
  unit: string
  conclusion: Conclusion
  report_no: string
  analyst_id: number
  analyst_name: string
  tested_at: string | null
  created_at: string
}

export interface SampleDetail extends Sample {
  transfers: Transfer[]
  results: TestResult[]
}

export interface User {
  id: number
  username: string
  full_name: string
}

export type LocationType = 'device' | 'fridge' | 'bench' | 'discard'

export interface Location {
  id: number
  name: string
  type: LocationType
  building: string
  room: string
  temperature: number
}

export interface CreateSamplePayload {
  name: string
  category?: string
  source?: string
  receiver_id: number
  location_id: number
  remark?: string
}

export interface CreateTransferPayload {
  from_location_id?: number | null
  to_location_id: number
  operator_id: number
  action: Exclude<TransferAction, 'receive'>
  occurred_at?: string
  note?: string
}

export interface CreateResultPayload {
  item_name: string
  method?: string
  instrument?: string
  result_value?: string
  unit?: string
  conclusion?: Conclusion
  report_no?: string
  analyst_id: number
  tested_at?: string
}
