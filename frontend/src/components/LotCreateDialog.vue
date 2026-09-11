<script setup lang="ts">
import { reactive, ref } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { ElMessage } from 'element-plus'
import { productionApi } from '@/api/production'
import type { CreateLotPayload } from '@/types/production'

const emit = defineEmits<{ (e: 'created'): void }>()

const visible = ref(false)
const submitting = ref(false)
const formRef = ref<FormInstance>()

const initial = (): CreateLotPayload => ({
  lot_no: '',
  product: '',
  wafer_count: 25,
  remark: '',
})
const form = reactive<CreateLotPayload>(initial())

const rules: FormRules = {
  wafer_count: [{ required: true, message: '请输入晶圆数量', trigger: 'blur' }],
}

function open() {
  Object.assign(form, initial())
  formRef.value?.clearValidate()
  visible.value = true
}
defineExpose({ open })

async function submit() {
  if (!formRef.value) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return
  submitting.value = true
  try {
    const res = await productionApi.createLot({
      ...form,
      lot_no: form.lot_no?.trim() || undefined,
      product: form.product?.trim() || undefined,
    })
    ElMessage.success(`批次已创建：${res.lot_no}`)
    visible.value = false
    emit('created')
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <el-dialog v-model="visible" title="新增晶圆批次" width="480px" destroy-on-close>
    <el-form ref="formRef" :model="form" :rules="rules" label-width="96px">
      <el-form-item label="批次号">
        <el-input v-model="form.lot_no" placeholder="留空自动生成 LOT-YYYYMMDD-NNN" maxlength="32" />
      </el-form-item>
      <el-form-item label="产品型号">
        <el-input v-model="form.product" placeholder="如 LOGIC-28NM" maxlength="128" />
      </el-form-item>
      <el-form-item label="晶圆数量" prop="wafer_count">
        <el-input-number v-model="form.wafer_count" :min="1" :max="1000" />
      </el-form-item>
      <el-form-item label="备注">
        <el-input v-model="form.remark" type="textarea" :rows="2" maxlength="500" show-word-limit />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="submit">创建批次</el-button>
    </template>
  </el-dialog>
</template>
