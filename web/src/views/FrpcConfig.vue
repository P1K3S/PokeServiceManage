<template>
  <div style="display: flex; gap: 16px">
    <el-card shadow="hover" style="flex: 1">
      <template #header>
        <span>内网穿透配置生成器</span>
      </template>

      <div v-loading="loading" style="min-height: 200px">
        <el-empty v-if="!loading && egressGroups.length === 0" description="暂无可用出站方式，请先配置非直连的出站方式" />

        <template v-else-if="!loading">
          <el-form label-width="100px">
            <el-form-item label="出站服务">
              <el-select v-model="selectedEgressId" placeholder="请选择出站服务" style="width: 400px" @change="onEgressChange" clearable>
                <el-option
                  v-for="g in egressGroups"
                  :key="g.egressServiceId"
                  :label="g.egressName + ' (' + g.items.length + ' 个服务)'"
                  :value="g.egressServiceId"
                />
              </el-select>
            </el-form-item>

            <el-form-item v-if="selectedEgressId && availableMachines.length > 0" label="选择主机">
              <el-select v-model="selectedMachineId" placeholder="全部主机" style="width: 300px" @change="onMachineChange" clearable>
                <el-option
                  v-for="m in availableMachines"
                  :key="m.id"
                  :label="m.name + ' (' + m.count + ' 个)'"
                  :value="m.id"
                />
              </el-select>
            </el-form-item>

            <el-form-item v-if="filteredItems.length > 0" label="选择服务">
              <el-checkbox-group v-model="selectedIds" style="display: flex; flex-wrap: wrap; gap: 12px">
                <el-checkbox v-for="item in filteredItems" :key="item.id" :value="item.id" :label="item.id" border style="margin-right: 0; padding: 8px 14px">
                  {{ item.label }}
                </el-checkbox>
              </el-checkbox-group>
            </el-form-item>

            <el-form-item v-if="filteredItems.length > 0">
              <el-button type="primary" :disabled="selectedIds.length === 0" @click="handleGenerate" :loading="generating">
                生成配置
              </el-button>
            </el-form-item>
          </el-form>

          <div v-if="configText" style="margin-top: 16px">
            <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px">
              <span style="font-weight: bold">生成的配置文件 (frpc.toml)</span>
              <el-button type="success" size="small" @click="handleCopy">复制到剪贴板</el-button>
            </div>
            <pre style="background: #f5f7fa; border: 1px solid #dcdfe6; border-radius: 4px; padding: 16px; overflow-x: auto; font-size: 13px; line-height: 1.6; white-space: pre-wrap; word-break: break-all"><code>{{ configText }}</code></pre>
          </div>
        </template>
      </div>
    </el-card>

    <el-card shadow="hover" style="width: 320px; flex-shrink: 0; align-self: flex-start">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span>通知公告</span>
          <el-button v-if="isSuperAdmin" type="primary" link size="small" @click="openEditNotice">编辑</el-button>
        </div>
      </template>
      <div v-loading="noticeLoading" style="min-height: 80px">
        <el-empty v-if="!noticeLoading && !notice" description="暂无通知" :image-size="60" />
        <div v-else-if="notice">
          <h4 style="margin: 0 0 12px 0; color: #303133">{{ notice.title || '通知' }}</h4>
          <div style="color: #606266; white-space: pre-wrap; word-break: break-all; line-height: 1.6; font-size: 14px">
            {{ notice.content }}
          </div>
          <div style="margin-top: 12px; color: #909399; font-size: 12px">
            更新于: {{ formatDate(notice.updatedAt) }}
          </div>
        </div>
      </div>
    </el-card>

    <el-dialog v-model="editNoticeVisible" title="编辑通知" width="500px" :close-on-click-modal="false" :lock-scroll="false">
      <el-form label-width="60px">
        <el-form-item label="标题">
          <el-input v-model="editNoticeForm.title" placeholder="请输入标题" />
        </el-form-item>
        <el-form-item label="内容">
          <el-input v-model="editNoticeForm.content" type="textarea" :rows="8" placeholder="请输入通知内容" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editNoticeVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSaveNotice" :loading="savingNotice">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { getEgressMethods, generateFrpc } from '../api/egress'
import { getServices } from '../api/service'
import { getOtherServices } from '../api/otherService'
import { getMachines } from '../api/machine'
import { getNotice, updateNotice } from '../api/notice'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '../stores/auth'

const loading = ref(false)
const generating = ref(false)
const egressGroups = ref([])
const selectedEgressId = ref('')
const selectedMachineId = ref('')
const selectedIds = ref([])
const configText = ref('')

const currentItems = ref([])
const availableMachines = ref([])

const filteredItems = computed(() => {
  if (!selectedMachineId.value) return currentItems.value
  return currentItems.value.filter(item => item.machineId === selectedMachineId.value)
})

const fetchData = async () => {
  loading.value = true
  try {
    const [egressRes, dockerRes, otherRes, machinesRes] = await Promise.all([
      getEgressMethods({ page: 1, pageSize: 500, isDirect: false }),
      getServices({ page: 1, pageSize: 200 }),
      getOtherServices({ page: 1, pageSize: 200 }),
      getMachines({ page: 1, pageSize: 200 })
    ])

    const methods = egressRes.data.list || []
    const dockerServices = dockerRes.data.list || []
    const otherServices = otherRes.data.list || []
    const machines = machinesRes.data.list || []

    const dockerServiceMap = {}
    for (const s of dockerServices) {
      dockerServiceMap[s.id] = s
    }
    const otherServiceMap = {}
    for (const s of otherServices) {
      otherServiceMap[s.id] = s
    }
    const machineMap = {}
    for (const m of machines) {
      machineMap[m.id] = m
    }

    const groups = {}
    for (const m of methods) {
      if (!m.egressServiceId || m.egressServiceId === 0) continue
      const key = m.egressServiceId
      if (!groups[key]) {
        const es = dockerServiceMap[m.egressServiceId]
        groups[key] = {
          egressServiceId: m.egressServiceId,
          egressName: es ? `${es.name}-${es.machineName || ''}` : `出站服务#${m.egressServiceId}`,
          items: []
        }
      }

      let machineId = 0
      let machineName = ''
      if (m.serviceType === 'other') {
        const os = otherServiceMap[m.serviceId]
        if (os) { machineId = os.machineId; machineName = machineMap[os.machineId]?.name || '' }
      } else {
        const ds = dockerServiceMap[m.serviceId]
        if (ds) { machineId = ds.machineId; machineName = machineMap[ds.machineId]?.name || '' }
      }

      const label = `${m.serviceName || '服务#' + m.serviceId} (${m.publicIp}:${m.publicPort} → ${m.internalIp}:${m.internalPort})`
      groups[key].items.push({
        id: m.id,
        label,
        machineId,
        machineName
      })
    }

    egressGroups.value = Object.values(groups)
  } catch {
  } finally {
    loading.value = false
  }
}

const onEgressChange = () => {
  selectedIds.value = []
  selectedMachineId.value = ''
  configText.value = ''
  const group = egressGroups.value.find(g => g.egressServiceId === selectedEgressId.value)
  currentItems.value = group ? group.items : []
  availableMachines.value = []

  if (group) {
    const machineCounts = {}
    for (const item of group.items) {
      if (item.machineId) {
        machineCounts[item.machineId] = (machineCounts[item.machineId] || 0) + 1
      }
    }
    availableMachines.value = Object.entries(machineCounts).map(([id, count]) => {
      const item = group.items.find(i => i.machineId === Number(id))
      return { id: Number(id), name: item?.machineName || '未知', count }
    })
  }
}

const onMachineChange = () => {
  selectedIds.value = []
  configText.value = ''
}

const handleGenerate = async () => {
  if (selectedIds.value.length === 0) return
  generating.value = true
  try {
    const res = await generateFrpc(selectedIds.value)
    configText.value = res.data.config
    ElMessage.success('配置生成成功')
  } catch {
    ElMessage.error('配置生成失败')
  } finally {
    generating.value = false
  }
}

const handleCopy = () => {
  if (!configText.value) return
  if (navigator.clipboard && window.isSecureContext) {
    navigator.clipboard.writeText(configText.value).then(() => {
      ElMessage.success('已复制到剪贴板')
    }).catch(() => {
      fallbackCopy(configText.value)
    })
  } else {
    fallbackCopy(configText.value)
  }
}

const fallbackCopy = (text) => {
  const textarea = document.createElement('textarea')
  textarea.value = text
  textarea.style.position = 'fixed'
  textarea.style.left = '-9999px'
  document.body.appendChild(textarea)
  textarea.select()
  try {
    document.execCommand('copy')
    ElMessage.success('已复制到剪贴板')
  } catch {
    ElMessage.warning('复制失败，请手动复制')
  }
  document.body.removeChild(textarea)
}

const authStore = useAuthStore()
const isSuperAdmin = computed(() => authStore.role === 'super_admin')
const noticeLoading = ref(false)
const notice = ref(null)
const editNoticeVisible = ref(false)
const savingNotice = ref(false)
const editNoticeForm = ref({ title: '', content: '' })

const fetchNotice = async () => {
  noticeLoading.value = true
  try {
    const res = await getNotice()
    notice.value = res.data
  } catch {
  } finally {
    noticeLoading.value = false
  }
}

const openEditNotice = () => {
  editNoticeForm.value = {
    title: notice.value?.title || '',
    content: notice.value?.content || ''
  }
  editNoticeVisible.value = true
}

const handleSaveNotice = async () => {
  savingNotice.value = true
  try {
    const res = await updateNotice(editNoticeForm.value)
    notice.value = res.data
    ElMessage.success('保存成功')
    editNoticeVisible.value = false
  } catch {
  } finally {
    savingNotice.value = false
  }
}

const formatDate = (dateStr) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN')
}

onMounted(() => {
  fetchData()
  fetchNotice()
})
</script>