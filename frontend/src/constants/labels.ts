import type { LocationType, SampleStatus, TransferAction } from '@/types'

export const STATUS_LABELS: Record<SampleStatus, string> = {
  received: '已收样',
  in_transit: '流转中',
  testing: '检测中',
  stored: '存储中',
  reported: '已出报告',
  discarded: '已销毁',
}

export const STATUS_TAG_TYPE: Record<SampleStatus, 'info' | 'warning' | 'primary' | 'success' | 'danger'> = {
  received: 'info',
  in_transit: 'warning',
  testing: 'warning',
  stored: 'primary',
  reported: 'success',
  discarded: 'danger',
}

export const ACTION_LABELS: Record<TransferAction, string> = {
  receive: '收样',
  transfer: '移交',
  load: '上机',
  unload: '下机',
  store: '入库',
  discard: '销毁',
}

// 时间线节点类型，映射 el-timeline-item type
export const ACTION_TIMELINE_TYPE: Record<TransferAction, 'primary' | 'success' | 'warning' | 'danger'> = {
  receive: 'success',
  transfer: 'primary',
  load: 'warning',
  unload: 'primary',
  store: 'primary',
  discard: 'danger',
}

export const LOCATION_TYPE_LABELS: Record<LocationType, string> = {
  device: '检测设备',
  fridge: '冰箱',
  bench: '实验台',
  discard: '销毁点',
}

export const CONCLUSION_LABELS: Record<string, string> = {
  qualified: '合格',
  unqualified: '不合格',
  na: '待定',
  '': '—',
}
