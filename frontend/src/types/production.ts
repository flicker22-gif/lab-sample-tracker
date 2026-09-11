export interface Process {
  id: number
  code: string
  name: string
  seq: number
  enabled: boolean
}

export type MachineStatusValue = 'idle' | 'running' | 'maintenance' | 'fault'

export interface Machine {
  id: number
  code: string
  name: string
  process_id: number
  process_code: string
  process_name: string
  status: MachineStatusValue
  current_lot_id: number | null
  current_lot_no: string
  remark: string
}

export type LotStatus = 'pending' | 'in_progress' | 'completed' | 'scrapped'

export interface WaferLot {
  id: number
  lot_no: string
  product: string
  wafer_count: number
  status: LotStatus
  current_process_id: number | null
  current_process_name: string
  received_at: string
  remark: string
  created_at: string
}

export type ProcessResultValue = 'pending' | 'ok' | 'ng' | 'rework'

export interface ProcessRecord {
  id: number
  lot_id: number
  process_id: number
  process_code: string
  process_name: string
  machine_id: number
  machine_code: string
  machine_name: string
  operator_id: number
  operator_name: string
  start_time: string
  end_time: string | null
  duration_sec: number
  result: ProcessResultValue
  wafer_out: number
  remark: string
}

export interface StatusLog {
  id: number
  machine_id: number
  from_status: MachineStatusValue | ''
  to_status: MachineStatusValue
  operator_name: string
  reason: string
  lot_id: number | null
  occurred_at: string
}

export interface LotDetail extends WaferLot {
  records: ProcessRecord[]
}

export interface CreateLotPayload {
  lot_no?: string
  product?: string
  wafer_count?: number
  remark?: string
}

export interface StartProcessPayload {
  machine_id: number
  operator_id: number
  start_time?: string
  remark?: string
}

export interface EndProcessPayload {
  result: 'ok' | 'ng' | 'rework'
  wafer_out?: number
  end_time?: string
  remark?: string
  operator_id?: number
}

export interface MachineStatusPayload {
  status: Exclude<MachineStatusValue, 'running'>
  operator_id: number
  reason?: string
}
