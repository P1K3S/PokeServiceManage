<template>
  <div>
    <el-card shadow="hover">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span>Docker服务列表</span>
          <el-button type="primary" @click="openForm('create')">新增服务</el-button>
        </div>
      </template>

      <el-form :inline="true" :model="search" style="margin-bottom: 10px">
        <el-form-item label="名称">
          <el-input v-model="search.keyword" placeholder="模糊搜索" clearable @clear="fetchData" @keyup.enter="fetchData" />
        </el-form-item>
        <el-form-item label="所属主机">
          <el-select v-model="search.machineId" placeholder="全部" clearable @change="fetchData" style="width: 160px" filterable>
            <el-option v-for="m in machineOptions" :key="m.id" :label="m.name" :value="m.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="search.status" placeholder="全部" clearable @change="fetchData" style="width: 120px">
            <el-option label="运行中" :value="1" />
            <el-option label="已停止" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="fetchData">查询</el-button>
          <el-button @click="handleCheckAllStatus" :loading="allChecking" style="margin-left: 8px">检测所有状态</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="list" stripe border v-loading="loading" style="width: 100%">
        <el-table-column prop="name" label="服务名称" min-width="140" show-overflow-tooltip align="center" />
        <el-table-column prop="machineName" label="所属主机" width="140" show-overflow-tooltip align="center" />
        <el-table-column label="源IP" width="130" align="center">
          <template #default="{ row }">
            {{ row.dockerSourceIp || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="源端口" width="110" align="center">
          <template #default="{ row }">
            {{ row.dockerSourcePort || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="port" label="宿主机端口" width="130" align="center" sortable>
          <template #default="{ row }">
            {{ row.port || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="protocol" label="协议" width="70" align="center" />
        <el-table-column label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'primary' : 'info'" size="small">
              {{ row.status === 1 ? '运行中' : '已停止' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="egressCount" label="出站" width="65" align="center" />
        <el-table-column prop="remark" label="备注" min-width="140" show-overflow-tooltip align="center" />
        <el-table-column label="操作" width="220" align="center">
          <template #default="{ row }">
            <el-button type="info" link size="small" @click="viewDetail(row)">查看</el-button>
            <el-button type="primary" link size="small" @click="openForm('edit', row)">编辑</el-button>
            <el-switch
              :model-value="row.locked"
              size="small"
              :loading="row._locking"
              style="margin-left: 4px"
              @click.prevent.stop
              @change="(v) => handleToggleLock(row, v)"
              inline-prompt
              active-text="锁"
              inactive-text=""
            />
            <el-button type="danger" link size="small" @click="handleDelete(row)" style="margin-left: 4px">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        style="margin-top: 16px; justify-content: flex-end"
        @change="fetchData"
      />
    </el-card>

    <el-dialog v-model="formVisible" :title="formMode === 'create' ? '新增服务' : '编辑服务'" width="580px" :close-on-click-modal="false" :lock-scroll="false">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="所属主机" prop="machineId">
          <el-select v-model="form.machineId" placeholder="请选择" style="width: 100%" filterable :teleported="false">
            <el-option v-for="m in machineOptions" :key="m.id" :label="m.name" :value="m.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="服务名称" prop="name">
          <el-input v-model="form.name" placeholder="例如：Nginx" />
        </el-form-item>
        <el-form-item label="源IP" prop="dockerSourceIp">
          <el-input v-model="form.dockerSourceIp" placeholder="例如：172.17.0.2" />
        </el-form-item>
        <el-divider content-position="left">端口映射</el-divider>
        <el-form-item label="端口1">
          <div style="display: flex; align-items: center; gap: 6px; width: 100%">
            <el-input-number v-model="form.port" :min="0" :max="65535" style="width: 130px" placeholder="宿主机" />
            <span>→</span>
            <el-input-number v-model="form.dockerSourcePort" :min="0" :max="65535" style="width: 130px" placeholder="容器" />
            <el-select v-model="form.protocol" style="width: 90px">
              <el-option label="TCP" value="TCP" />
              <el-option label="UDP" value="UDP" />
            </el-select>
          </div>
        </el-form-item>
        <template v-for="(pair, idx) in portPairs" :key="idx">
          <el-form-item :label="'端口' + (idx + 2)">
            <div style="display: flex; align-items: center; gap: 6px; width: 100%">
              <el-input-number v-model="pair.hostPort" :min="0" :max="65535" style="width: 130px" placeholder="宿主机" />
              <span>→</span>
              <el-input-number v-model="pair.containerPort" :min="0" :max="65535" style="width: 130px" placeholder="容器" />
              <el-select v-model="pair.protocol" style="width: 90px">
                <el-option label="TCP" value="TCP" />
                <el-option label="UDP" value="UDP" />
              </el-select>
              <el-button type="danger" :icon="Delete" circle size="small" @click="removePortPair(idx)" />
            </div>
          </el-form-item>
        </template>
        <el-form-item label=" ">
          <el-button type="primary" link @click="addPortPair">+ 新增端口</el-button>
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" active-text="运行中" inactive-text="已停止" />
        </el-form-item>
        <el-form-item label="出站服务">
          <el-switch v-model="form.isEgress" active-text="是" inactive-text="否" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailVisible" title="服务详情" width="480px" :close-on-click-modal="false" :lock-scroll="false">
      <template v-if="detailData">
        <el-descriptions :column="1" border>
          <el-descriptions-item label="服务名称">{{ detailData.name }}</el-descriptions-item>
          <el-descriptions-item label="所属主机">{{ detailData.machineName }}</el-descriptions-item>
          <el-descriptions-item label="端口">
            <template v-if="detailData.portMappings && JSON.parse(detailData.portMappings).length > 0">
              <div v-for="(m, i) in JSON.parse(detailData.portMappings)" :key="i" style="padding: 2px 0">
                宿主机 {{ m.hostPort || '-' }} → 容器 {{ m.containerPort || '-' }}/{{ m.protocol }}
              </div>
            </template>
            <template v-else>{{ detailData.port || '-' }}</template>
          </el-descriptions-item>
          <el-descriptions-item label="源IP">{{ detailData.dockerSourceIp || '-' }}</el-descriptions-item>
          <el-descriptions-item label="协议">{{ detailData.protocol || 'TCP' }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="detailData.status === 1 ? 'primary' : 'info'" size="small">
              {{ detailData.status === 1 ? '运行中' : '已停止' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="出站服务">
            <el-tag :type="detailData.isEgress ? 'success' : 'info'" size="small">
              {{ detailData.isEgress ? '是' : '否' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="备注">{{ detailData.remark || '-' }}</el-descriptions-item>
        </el-descriptions>
      </template>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted } from 'vue'
import { getServices, createService, updateService, deleteService, checkService } from '../api/service'
import { getMachines } from '../api/machine'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete } from '@element-plus/icons-vue'

const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const loading = ref(false)
const machineOptions = ref([])
const search = reactive({ keyword: '', machineId: '', status: '' })
const formVisible = ref(false)
const formMode = ref('create')
const formRef = ref(null)
const submitting = ref(false)
const allChecking = ref(false)
const editId = ref(null)

const detailVisible = ref(false)
const detailData = ref(null)

const portPairs = ref([])

const form = reactive({
  machineId: null, name: '', port: 80, protocol: 'TCP',
  dockerSourceIp: '', dockerSourcePort: null, status: 1, isEgress: false, remark: ''
})

const rules = {
  machineId: [{ required: true, message: '请选择所属主机', trigger: 'change' }],
  name: [{ required: true, message: '请输入服务名称', trigger: 'blur' }],
  port: [{ required: true, message: '请输入端口号', trigger: 'blur' }]
}

const fetchData = async () => {
  loading.value = true
  try {
    const res = await getServices({ page: page.value, pageSize: pageSize.value, ...search })
    list.value = res.data.list
    total.value = res.data.total
  } catch {
  } finally {
    loading.value = false
  }
}

const fetchMachines = async () => {
  try {
    const res = await getMachines({ page: 1, pageSize: 200 })
    machineOptions.value = res.data.list
  } catch {}
}

const openForm = (mode, row) => {
  formMode.value = mode
  formVisible.value = true
  portPairs.value = []
  if (mode === 'edit' && row) {
    editId.value = row.id
    Object.assign(form, {
      machineId: row.machineId, name: row.name,
      port: row.port, protocol: row.protocol || 'TCP',
      dockerSourceIp: row.dockerSourceIp || '', dockerSourcePort: row.dockerSourcePort || null,
      status: row.status, isEgress: row.isEgress || false, remark: row.remark || ''
    })
    if (row.portMappings) {
      try {
        const allMappings = JSON.parse(row.portMappings)
        if (allMappings.length > 1) {
          portPairs.value = allMappings.slice(1).map(m => ({
            hostPort: parseInt(m.hostPort) || null,
            containerPort: parseInt(m.containerPort) || null,
            protocol: m.protocol || 'TCP'
          }))
        }
      } catch {}
    }
  } else {
    editId.value = null
    Object.assign(form, { machineId: null, name: '', port: 80, protocol: 'TCP', dockerSourceIp: '', dockerSourcePort: null, status: 1, isEgress: false, remark: '' })
  }
}

const addPortPair = () => {
  portPairs.value.push({ hostPort: null, containerPort: null, protocol: 'TCP' })
}

const removePortPair = (idx) => {
  portPairs.value.splice(idx, 1)
}

const handleSubmit = async () => {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return
  submitting.value = true
  try {
    const allMappings = []
    const firstHost = parseInt(form.port)
    const firstContainer = parseInt(form.dockerSourcePort)
    if (firstHost || firstContainer) {
      allMappings.push({
        hostPort: firstHost ? String(firstHost) : '',
        containerPort: firstContainer ? String(firstContainer) : '',
        protocol: form.protocol || 'TCP'
      })
    }
    for (const p of portPairs.value) {
      const hp = parseInt(p.hostPort)
      const cp = parseInt(p.containerPort)
      if (hp || cp) {
        allMappings.push({
          hostPort: hp ? String(hp) : '',
          containerPort: cp ? String(cp) : '',
          protocol: p.protocol || 'TCP'
        })
      }
    }
    const payload = { ...form, portMappings: JSON.stringify(allMappings) }
    if (formMode.value === 'create') {
      await createService(payload)
      ElMessage.success('创建成功')
    } else {
      await updateService(editId.value, payload)
      ElMessage.success('更新成功')
    }
    formVisible.value = false
    fetchData()
  } catch {
  } finally {
    submitting.value = false
  }
}

const handleDelete = (row) => {
  ElMessageBox.confirm(`确定删除服务「${row.name}」吗？关联的出站方式也将被删除。`, '确认删除', {
    confirmButtonText: '确定删除',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    await deleteService(row.id)
    ElMessage.success('删除成功')
    fetchData()
  }).catch(() => {})
}

const handleToggleLock = async (row, val) => {
  row._locking = true
  try {
    await updateService(row.id, { locked: val })
    row.locked = val
    ElMessage.success(val ? '已锁定，Docker检测将不会覆盖' : '已解锁')
  } catch {
    ElMessage.error('操作失败')
  } finally {
    row._locking = false
  }
}

const handleCheckAllStatus = async () => {
  if (list.value.length === 0) {
    ElMessage.info('没有Docker服务需要检测')
    return
  }
  allChecking.value = true
  let running = 0
  let stopped = 0
  for (const s of list.value) {
    try {
      const res = await checkService(s.id)
      s.status = res.data.status
      if (res.data.status === 1) {
        running++
      } else {
        stopped++
      }
    } catch {
      stopped++
    }
  }
  ElMessage.success(`检测完成：运行中 ${running} 个，已停止 ${stopped} 个`)
  fetchData()
  allChecking.value = false
}

const viewDetail = (row) => {
  detailData.value = row
  detailVisible.value = true
}

onMounted(() => {
  fetchData()
  fetchMachines()
})
</script>