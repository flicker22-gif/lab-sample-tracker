import type { LotStatus, MachineStatusValue, ProcessResultValue } from '@/types/production'

export const MACHINE_STATUS_LABELS: Record<MachineStatusValue, string> = {
  idle: '空闲',
  running: '加工中',
  maintenance: '维护',
  fault: '故障',
}

export const MACHINE_STATUS_TAG: Record<MachineStatusValue, 'success' | 'warning' | 'info' | 'danger'> = {
  idle: 'success',
  running: 'warning',
  maintenance: 'info',
  fault: 'danger',
}

export const LOT_STATUS_LABELS: Record<LotStatus, string> = {
  pending: '待加工',
  in_progress: '加工中',
  completed: '已完工',
  scrapped: '已报废',
}

export const LOT_STATUS_TAG: Record<LotStatus, 'info' | 'warning' | 'success' | 'danger'> = {
  pending: 'info',
  in_progress: 'warning',
  completed: 'success',
  scrapped: 'danger',
}

export const RESULT_LABELS: Record<ProcessResultValue, string> = {
  pending: '加工中',
  ok: '合格',
  ng: '不合格',
  rework: '返工',
}

export const RESULT_TAG: Record<ProcessResultValue, 'info' | 'success' | 'danger' | 'warning'> = {
  pending: 'info',
  ok: 'success',
  ng: 'danger',
  rework: 'warning',
}

// 秒 → 可读时长
export function formatDuration(sec: number): string {
  if (!sec || sec < 0) return '—'
  const h = Math.floor(sec / 3600)
  const m = Math.floor((sec % 3600) / 60)
  const s = sec % 60
  if (h > 0) return `${h}小时${m}分`
  if (m > 0) return `${m}分${s}秒`
  return `${s}秒`
}
