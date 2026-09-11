<script setup lang="ts">
import { reactive, ref } from 'vue'
import type { FormInstance, FormRules, UploadRawFile } from 'element-plus'
import { ElMessage } from 'element-plus'
import { waferApi } from '@/api/wafer'
import { useMetaStore } from '@/stores/meta'
import type { Wafer } from '@/types/wafer'

const props = defineProps<{ wafer: Wafer | null }>()
const emit = defineEmits<{ (e: 'uploaded'): void }>()

const meta = useMetaStore()
const visible = ref(false)
const submitting = ref(false)
const formRef = ref<FormInstance>()
const file = ref<UploadRawFile>()

const form = reactive({ operator_id: 0, remark: '' })
const rules: FormRules = {
  operator_id: [{ required: true, message: '请选择操作员', trigger: 'change' }],
}

async function open() {
  if (!props.wafer) return
  await meta.load()
  form.operator_id = 0
  form.remark = ''
  file.value = undefined
  formRef.value?.clearValidate()
  visible.value = true
}
defineExpose({ open })

function onFileChange(f: { raw: UploadRawFile }) {
  file.value = f.raw
}

async function submit() {
  if (!formRef.value || !props.wafer) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return
  if (!file.value) {
    ElMessage.warning('请选择 bin map 文件')
    return
  }
  submitting.value = true
  try {
    const res = await waferApi.upload(props.wafer.id, file.value, form.operator_id, form.remark)
    ElMessage.success(`v${res.version} 已解析：${res.total_dies} 颗晶粒，良率 ${(res.yield * 100).toFixed(2)}%`)
    visible.value = false
    emit('uploaded')
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <el-dialog v-model="visible" title="上传 Bin Map" width="480px" destroy-on-close>
    <el-form ref="formRef" :model="form" :rules="rules" label-width="88px">
      <el-form-item label="晶圆">
        <el-input :model-value="wafer ? `槽位 ${wafer.slot_no}${wafer.wafer_id ? ' · ' + wafer.wafer_id : ''}` : ''" disabled />
      </el-form-item>
      <el-form-item label="Map 文件" required>
        <el-upload
          :auto-upload="false"
          :limit="1"
          accept=".txt,.map,.csv,.log"
          :on-change="onFileChange"
        >
          <el-button :icon="'Upload'">选择文件</el-button>
          <template #tip>
            <div class="tip">文本格式：整数为 bin 号，. / - 为空晶粒；支持 Notch、BinDef 头部</div>
          </template>
        </el-upload>
      </el-form-item>
      <el-form-item label="操作员" prop="operator_id">
        <el-select v-model="form.operator_id" placeholder="请选择" style="width: 100%">
          <el-option v-for="u in meta.users" :key="u.id" :label="u.full_name" :value="u.id" />
        </el-select>
      </el-form-item>
      <el-form-item label="备注">
        <el-input v-model="form.remark" type="textarea" :rows="2" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="submit">上传并解析</el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.tip {
  color: #909399;
  font-size: 12px;
  margin-top: 4px;
}
</style>
