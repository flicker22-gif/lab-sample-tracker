export interface Wafer {
  id: number
  lot_id: number
  slot_no: number
  wafer_id: string
  product: string
  current_map_id: number | null
  has_map: boolean
  version: number
  rows: number
  cols: number
  total_dies: number
  pass_dies: number
  yield: number
  remark: string
  uploaded_at: string
}

export interface BinDef {
  number: number
  name: string
  color: string
  pass: boolean
}

export interface BinSummary {
  bin: number
  name: string
  count: number
  color: string
  pass: boolean
}

export interface WaferMap {
  id: number
  wafer_id: number
  version: number
  file_name: string
  file_size: number
  map_data: number[][]
  rows: number
  cols: number
  notch: 'up' | 'down' | 'left' | 'right'
  bin_defs: BinDef[]
  total_dies: number
  pass_dies: number
  bin_summary: BinSummary[]
  yield: number
  operator_id: number
  operator_name: string
  remark: string
  created_at: string
}

export interface MapVersion {
  id: number
  version: number
  file_name: string
  file_size: number
  rows: number
  cols: number
  total_dies: number
  pass_dies: number
  yield: number
  operator_name: string
  remark: string
  created_at: string
}
