<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { waferApi } from '@/api/wafer'
import type { MapVersion, Wafer, WaferMap } from '@/types/wafer'
import { formatDateTime } from '@/utils/datetime'
import WaferMapCanvas from './WaferMapCanvas.vue'
import UploadMapDialog from './UploadMapDialog.vue'

const props = defineProps<{ lotId: number | null; modelValue: boolean }>()
const emit = defineEmits<{
  (e: 'update:modelValue', v: boolean): void
}>()

const visible = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v),
})

const wafers = ref<Wafer[]>([])
const loadingWafers = ref(false)
const currentWafer = ref<Wafer | null>(null)
const map = ref<WaferMap | null>(null)
const mapLoading = ref(false)
const versions = ref<MapVersion[]>([])
const selectedVersionId = ref<number>(0)
const hiddenBins = ref<Set<number>>(new Set())
const uploadDialog = ref<InstanceType<typeof UploadMapDialog>>()

async function loadWafers() {
  if (props.lotId == null) return
  loadingWafers.value = true
  try {
    wafers.value = await waferApi.listByLot(props.lotId)
    // 默认选中第一个有 map 的晶圆，否则第一片
    currentWafer.value = wafers.value.find((w) => w.has_map) ?? wafers.value[0] ?? null
    if (currentWafer.value) await loadWafer(currentWafer.value)
  } finally {
    loadingWafers.value = false
  }
}

watch(
  () => [props.modelValue, props.lotId] as const,
  ([open]) => {
    if (open) loadWafers()
  },
)

async function selectWafer(w: Wafer) {
  currentWafer.value = w
  await loadWafer(w)
}

async function loadWafer(w: Wafer, preferMapId?: number) {
  hiddenBins.value = new Set()
  map.value = null
  versions.value = []
  selectedVersionId.value = 0
  const [vs] = await Promise.all([waferApi.versions(w.id)])
  versions.value = vs
  const mapId = preferMapId ?? w.current_map_id ?? vs[0]?.id
  if (mapId) {
    selectedVersionId.value = mapId
    await loadMap(mapId)
  }
}

async function loadMap(id: number) {
  mapLoading.value = true
  try {
    map.value = await waferApi.getMap(id)
  } finally {
    mapLoading.value = false
  }
}

async function onVersionChange(id: number) {
  await loadMap(id)
}

async function onUploaded() {
  const waferId = currentWafer.value?.id
  wafers.value = await waferApi.listByLot(props.lotId!)
  const fresh = wafers.value.find((x) => x.id === waferId)
  if (fresh) await loadWafer(fresh)
}

function toggleBin(bin: number) {
  const next = new Set(hiddenBins.value)
  if (next.has(bin)) next.delete(bin)
  else next.add(bin)
  hiddenBins.value = next
}

const yieldPct = computed(() => (map.value ? (map.value.yield * 100).toFixed(2) : '—'))
const yieldType = computed(() => {
  if (!map.value) return 'info'
  const y = map.value.yield
  if (y >= 0.95) return 'success'
  if (y >= 0.8) return 'warning'
  return 'danger'
})

function fmtSize(n: number): string {
  if (n < 1024) return `${n} B`
  return `${(n / 1024).toFixed(1)} KB`
}

function download(id: number) {
  window.open(waferApi.downloadUrl(id), '_blank')
}
</script>

<template>
  <el-drawer v-model="visible" title="晶圆 Bin Map" size="72%" destroy-on-close>
    <div v-loading="loadingWafers" class="wafer-drawer">
      <el-empty v-if="!loadingWafers && !wafers.length" description="该批次还没有晶圆槽位（创建批次时按片数自动生成）" />

      <template v-else>
        <!-- 晶圆槽位 -->
        <div class="slot-bar">
          <span class="bar-label">晶圆槽位：</span>
          <el-radio-group v-model="currentWafer" size="small" @change="selectWafer">
            <el-radio-button
              v-for="w in wafers"
              :key="w.id"
              :value="w"
            >
              #{{ w.slot_no }}
              <el-icon v-if="w.has_map" class="slot-dot"><CircleCheckFilled /></el-icon>
            </el-radio-button>
          </el-radio-group>
          <el-button
            class="upload-btn"
            type="primary"
            size="small"
            :icon="'Upload'"
            :disabled="!currentWafer"
            @click="uploadDialog?.open()"
          >
            上传本片 Map
          </el-button>
        </div>

        <el-empty v-if="currentWafer && !currentWafer.has_map && versions.length === 0"
          description="该片还没有 bin map，点击右上角上传" />

        <el-row v-if="map" :gutter="16" v-loading="mapLoading">
          <el-col :md="16" :sm="24">
            <el-card shadow="never">
              <WaferMapCanvas
                :data="map.map_data"
                :bin-defs="map.bin_defs"
                :notch="map.notch"
                :hidden-bins="hiddenBins"
                :height="480"
              />
            </el-card>
          </el-col>
          <el-col :md="8" :sm="24">
            <el-card shadow="never" class="side-card">
              <template #header>
                <div class="side-head">
                  <span>统计</span>
                  <el-select
                    :model-value="selectedVersionId"
                    size="small"
                    style="width: 118px"
                    @change="onVersionChange"
                  >
                    <el-option
                      v-for="v in versions"
                      :key="v.id"
                      :label="`v${v.version}`"
                      :value="v.id"
                    />
                  </el-select>
                </div>
              </template>

              <div class="yield-box">
                <span class="yield-label">良率</span>
                <el-tag :type="yieldType" size="large" effect="dark" class="yield-tag">
                  {{ yieldPct }}%
                </el-tag>
                <span class="muted yield-sub">{{ map.pass_dies }} / {{ map.total_dies }}</span>
              </div>
              <div class="kv">
                <span>晶粒总数</span><strong>{{ map.total_dies }}</strong>
                <span>合格</span><strong class="pass">{{ map.pass_dies }}</strong>
                <span>不合格</span><strong class="fail">{{ map.total_dies - map.pass_dies }}</strong>
                <span>网格</span><strong>{{ map.rows }} × {{ map.cols }}</strong>
                <span>缺口</span><strong>{{ { up: '上', down: '下', left: '左', right: '右' }[map.notch] }}</strong>
              </div>
              <div class="meta-line muted">
                {{ map.file_name }} · {{ fmtSize(map.file_size) }}
              </div>
              <div class="meta-line muted">
                {{ map.operator_name }} · {{ formatDateTime(map.created_at) }}
              </div>
              <el-button
                type="primary"
                link
                :icon="'Download'"
                @click="download(map.id)"
              >
                下载原始文件
              </el-button>

              <el-divider content-position="left">Bin 图例（点击筛选）</el-divider>
              <div class="legend">
                <div
                  v-for="b in map.bin_summary"
                  :key="b.bin"
                  class="legend-item"
                  :class="{ off: hiddenBins.has(b.bin) }"
                  @click="toggleBin(b.bin)"
                >
                  <span class="legend-color" :style="{ background: b.color }" />
                  <span class="legend-name">{{ b.name }}</span>
                  <span class="legend-count">{{ b.count }}</span>
                  <el-tag size="small" :type="b.pass ? 'success' : 'danger'" effect="plain">
                    {{ b.pass ? 'Pass' : 'Fail' }}
                  </el-tag>
                </div>
                <div v-if="!map.bin_summary.length" class="muted">无 bin 数据</div>
              </div>
            </el-card>
          </el-col>
        </el-row>
      </template>
    </div>

    <UploadMapDialog ref="uploadDialog" :wafer="currentWafer" @uploaded="onUploaded" />
  </el-drawer>
</template>

<style scoped>
.wafer-drawer {
  padding: 0 4px;
}
.slot-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
  margin-bottom: 14px;
}
.bar-label {
  color: #606266;
  font-size: 13px;
}
.slot-dot {
  color: var(--el-color-success);
  margin-left: 2px;
}
.upload-btn {
  margin-left: auto;
}
.side-card .yield-box {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 14px;
}
.yield-label {
  color: #909399;
  font-size: 13px;
}
.yield-tag {
  font-size: 16px;
  font-weight: 700;
}
.yield-sub {
  margin-left: auto;
  font-size: 12px;
}
.side-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.kv {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 6px 12px;
  font-size: 13px;
  color: #606266;
  margin-bottom: 12px;
}
.kv strong {
  color: #303133;
  text-align: right;
}
.pass {
  color: var(--el-color-success);
}
.fail {
  color: var(--el-color-danger);
}
.meta-line {
  font-size: 12px;
  margin-bottom: 4px;
}
.muted {
  color: #909399;
}
.legend {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.legend-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 8px;
  border: 1px solid #ebeef5;
  border-radius: 4px;
  cursor: pointer;
  user-select: none;
}
.legend-item.off {
  opacity: 0.4;
}
.legend-color {
  width: 14px;
  height: 14px;
  border-radius: 3px;
  border: 1px solid rgb(0 0 0 / 10%);
}
.legend-name {
  flex: 1;
  font-size: 13px;
}
.legend-count {
  font-weight: 600;
  font-size: 13px;
}
</style>
