<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import type { BinDef } from '@/types/wafer'

const props = withDefaults(
  defineProps<{
    data: number[][]
    binDefs: BinDef[]
    notch?: 'up' | 'down' | 'left' | 'right'
    hiddenBins?: Set<number>
    height?: number
  }>(),
  { notch: 'down', hiddenBins: () => new Set<number>(), height: 460 },
)

const wrapRef = ref<HTMLDivElement>()
const canvasRef = ref<HTMLCanvasElement>()
const tooltip = ref({ show: false, x: 0, y: 0, text: '' })

interface HoverCell {
  row: number
  col: number
  bin: number
}
const emit = defineEmits<{ (e: 'hover', cell: HoverCell | null): void }>()

// bin 号 → 颜色/定义（未在 binDefs 中声明的用兜底调色板）
const FALLBACK = ['#95a5a6', '#16a085', '#2980b9', '#8e44ad', '#d35400', '#c0392b']
const colorMap = computed(() => {
  const m = new Map<number, BinDef>()
  for (const d of props.binDefs) m.set(d.number, d)
  return m
})
function binColor(bin: number): string {
  return colorMap.value.get(bin)?.color || FALLBACK[bin % FALLBACK.length]
}
function binName(bin: number): string {
  return colorMap.value.get(bin)?.name || `Bin${bin}`
}

const rows = computed(() => props.data.length)
const cols = computed(() => props.data.reduce((m, r) => Math.max(m, r.length), 0))

let cell = 10
let offsetX = 0
let offsetY = 0
let hover: HoverCell | null = null

function draw() {
  const canvas = canvasRef.value
  const wrap = wrapRef.value
  if (!canvas || !wrap || rows.value === 0) return

  const dpr = window.devicePixelRatio || 1
  const cssW = wrap.clientWidth
  const cssH = props.height
  canvas.width = cssW * dpr
  canvas.height = cssH * dpr
  canvas.style.width = `${cssW}px`
  canvas.style.height = `${cssH}px`
  const ctx = canvas.getContext('2d')!
  ctx.setTransform(dpr, 0, 0, dpr, 0, 0)
  ctx.clearRect(0, 0, cssW, cssH)

  const pad = 26
  cell = Math.max(2, Math.floor(Math.min((cssW - pad * 2) / cols.value, (cssH - pad * 2) / rows.value)))
  const gridW = cell * cols.value
  const gridH = cell * rows.value
  offsetX = (cssW - gridW) / 2
  offsetY = (cssH - gridH) / 2

  // 晶圆圆底
  const cx = cssW / 2
  const cy = cssH / 2
  const radius = Math.min(gridW, gridH) / 2 + cell * 0.9
  ctx.beginPath()
  ctx.arc(cx, cy, radius, 0, Math.PI * 2)
  ctx.fillStyle = '#f0f2f5'
  ctx.fill()
  ctx.strokeStyle = '#c0c4cc'
  ctx.lineWidth = 1.5
  ctx.stroke()

  // 缺口（notch）
  drawNotch(ctx, cx, cy, radius)

  // 晶粒
  for (let r = 0; r < rows.value; r++) {
    const line = props.data[r]
    for (let c = 0; c < line.length; c++) {
      const bin = line[c]
      if (bin < 0) continue
      const x = offsetX + c * cell
      const y = offsetY + r * cell
      const isHover = hover && hover.row === r && hover.col === c
      if (props.hiddenBins.has(bin)) {
        ctx.fillStyle = '#fbfcfd'
        ctx.fillRect(x + 0.5, y + 0.5, cell - 1, cell - 1)
        ctx.strokeStyle = '#eef0f3'
        ctx.strokeRect(x + 0.5, y + 0.5, cell - 1, cell - 1)
      } else {
        ctx.fillStyle = binColor(bin)
        ctx.globalAlpha = isHover ? 0.75 : 1
        ctx.fillRect(x + 0.5, y + 0.5, cell - 1, cell - 1)
        ctx.globalAlpha = 1
      }
      if (isHover) {
        ctx.strokeStyle = '#303133'
        ctx.lineWidth = 2
        ctx.strokeRect(x + 1, y + 1, cell - 2, cell - 2)
        ctx.lineWidth = 1
      }
    }
  }
}

function drawNotch(ctx: CanvasRenderingContext2D, cx: number, cy: number, radius: number) {
  const size = cell * 1.6
  ctx.save()
  ctx.fillStyle = '#fff'
  ctx.strokeStyle = '#c0c4cc'
  ctx.beginPath()
  let nx = cx
  let ny = cy
  switch (props.notch) {
    case 'down':
      nx = cx
      ny = cy + radius
      ctx.arc(nx, ny, size / 2, 0, Math.PI * 2)
      break
    case 'up':
      ctx.arc(cx, cy - radius, size / 2, 0, Math.PI * 2)
      break
    case 'left':
      ctx.arc(cx - radius, cy, size / 2, 0, Math.PI * 2)
      break
    case 'right':
      ctx.arc(cx + radius, cy, size / 2, 0, Math.PI * 2)
      break
  }
  ctx.fill()
  ctx.stroke()
  ctx.restore()
}

function onMove(e: MouseEvent) {
  const canvas = canvasRef.value
  if (!canvas) return
  const rect = canvas.getBoundingClientRect()
  const x = e.clientX - rect.left
  const y = e.clientY - rect.top
  const c = Math.floor((x - offsetX) / cell)
  const r = Math.floor((y - offsetY) / cell)
  if (r >= 0 && r < rows.value && c >= 0 && c < (props.data[r]?.length ?? 0)) {
    const bin = props.data[r][c]
    if (bin >= 0) {
      hover = { row: r, col: c, bin }
      tooltip.value = {
        show: true,
        x: e.clientX - rect.left + 12,
        y: e.clientY - rect.top + 12,
        text: `(${r + 1}, ${c + 1}) · ${binName(bin)}`,
      }
      emit('hover', hover)
      draw()
      return
    }
  }
  if (hover) {
    hover = null
    tooltip.value.show = false
    emit('hover', null)
    draw()
  }
}

function onLeave() {
  hover = null
  tooltip.value.show = false
  emit('hover', null)
  draw()
}

let ro: ResizeObserver | null = null
onMounted(() => {
  draw()
  ro = new ResizeObserver(() => draw())
  if (wrapRef.value) ro.observe(wrapRef.value)
})
onBeforeUnmount(() => ro?.disconnect())
watch(() => [props.data, props.binDefs, props.hiddenBins, props.notch], draw, { deep: true })
</script>

<template>
  <div ref="wrapRef" class="wafer-wrap">
    <canvas ref="canvasRef" @mousemove="onMove" @mouseleave="onLeave" />
    <div v-if="tooltip.show" class="wafer-tip" :style="{ left: tooltip.x + 'px', top: tooltip.y + 'px' }">
      {{ tooltip.text }}
    </div>
  </div>
</template>

<style scoped>
.wafer-wrap {
  position: relative;
  width: 100%;
}
canvas {
  display: block;
  cursor: crosshair;
}
.wafer-tip {
  position: absolute;
  pointer-events: none;
  background: rgb(48 49 51 / 92%);
  color: #fff;
  font-size: 12px;
  padding: 4px 8px;
  border-radius: 4px;
  white-space: nowrap;
  z-index: 10;
}
</style>
