<script setup lang="ts">
import { reactive, ref } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { ElMessage } from 'element-plus'
import { sampleApi } from '@/api'
import { useMetaStore } from '@/stores/meta'
import type { CreateSamplePayload } from '@/types'

const emit = defineEmits<{ (e: 'created'): void }>()

const meta = useMetaStore()
const visible = ref(false)
const submitting = ref(false)
const formRef = ref<FormInstance>()

const initial = (): CreateSamplePayload => ({
  name: '',
  category: '',
  source: '',
  receiver_id: 0,
  location_id: 0,
  remark: '',
})
const form = reactive<CreateSamplePayload>(initial())

const rules: FormRules = {
  name: [{ required: true, message: '请输入样品名称', trigger: 'blur' }],
  receiver_id: [{ required: true, message: '请选择收样人', trigger: 'change' }],
  location_id: [{ required: true, message: '请选择首次存放位置', trigger: 'change' }],
}

async function open() {
  Object.assign(form, initial())
  await meta.load()
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
    const res = await sampleApi.create({ ...form })
    ElMessage.success(`收样成功，样品编号：${res.code}`)
    visible.value = false
    emit('created')
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <el-dialog v-model="visible" title="新增样品（收样登记）" width="520px" destroy-on-close>
    <el-form ref="formRef" :model="form" :rules="rules" label-width="110px">
      <el-form-item label="样品编号">
        <el-input disabled placeholder="保存后由系统自动生成，如 SP-20260910-0001" />
      </el-form-item>
      <el-form-item label="样品名称" prop="name">
        <el-input v-model="form.name" placeholder="如：饮用水-出厂水" maxlength="128" />
      </el-form-item>
      <el-form-item label="样品类型">
        <el-input v-model="form.category" placeholder="如：水质 / 土壤 / 食品" maxlength="64" />
      </el-form-item>
      <el-form-item label="送检单位">
        <el-input v-model="form.source" placeholder="样品来源单位" maxlength="128" />
      </el-form-item>
      <el-form-item label="收样人" prop="receiver_id">
        <el-select v-model="form.receiver_id" placeholder="请选择" style="width: 100%">
          <el-option v-for="u in meta.users" :key="u.id" :label="u.full_name" :value="u.id" />
        </el-select>
      </el-form-item>
      <el-form-item label="首次存放" prop="location_id">
        <el-select v-model="form.location_id" placeholder="选择收样后存放的冰箱或设备" style="width: 100%">
          <el-option-group label="冰箱/冷库">
            <el-option
              v-for="l in meta.locations.filter((x) => x.type === 'fridge')"
              :key="l.id"
              :label="`${l.name}（${l.temperature}℃）`"
              :value="l.id"
            />
          </el-option-group>
          <el-option-group label="检测设备">
            <el-option
              v-for="l in meta.locations.filter((x) => x.type === 'device')"
              :key="l.id"
              :label="l.name"
              :value="l.id"
            />
          </el-option-group>
          <el-option-group label="其他">
            <el-option
              v-for="l in meta.locations.filter((x) => x.type !== 'fridge' && x.type !== 'device')"
              :key="l.id"
              :label="l.name"
              :value="l.id"
            />
          </el-option-group>
        </el-select>
      </el-form-item>
      <el-form-item label="备注">
        <el-input v-model="form.remark" type="textarea" :rows="2" maxlength="500" show-word-limit />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="submit">确认收样</el-button>
    </template>
  </el-dialog>
</template>
