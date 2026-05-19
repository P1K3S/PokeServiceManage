<template>
  <div class="file-manager">
    <div class="fm-toolbar">
      <el-button size="small" @click="goBack" :disabled="currentPath === '/'">上级</el-button>
      <el-button size="small" @click="goHome">根目录</el-button>
      <el-button size="small" @click="refreshFiles" :loading="fileLoading">刷新</el-button>
      <el-button size="small" type="primary" @click="showMkdirDialog = true">新建文件夹</el-button>
      <el-button size="small" type="success" @click="triggerUpload">上传</el-button>
      <input ref="uploadInput" type="file" style="display: none" @change="handleUpload" multiple />
    </div>

    <div class="fm-breadcrumb">
      <span v-for="(seg, idx) in pathSegments" :key="idx" class="breadcrumb-item">
        <span v-if="idx > 0"> / </span>
        <a @click="navigateTo(seg.path)">{{ seg.name }}</a>
      </span>
    </div>

    <div class="fm-list" v-loading="fileLoading">
      <div v-if="currentPath !== '/'" class="fm-item fm-item-dir" @click="goBack">
        <span class="fm-icon">📁</span>
        <span class="fm-name">..</span>
      </div>
      <div
        v-for="f in files"
        :key="f.path"
        class="fm-item"
        :class="{ 'fm-item-dir': f.isDir, 'fm-item-file': !f.isDir, 'fm-selected': selectedFile && selectedFile.path === f.path }"
        @click="selectFile(f)"
        @dblclick="openFile(f)"
        @contextmenu.prevent="showContextMenu($event, f)"
      >
        <span class="fm-icon">{{ f.isDir ? '📁' : fileIcon(f.name) }}</span>
        <span class="fm-name" :title="f.name">{{ f.name }}</span>
        <span class="fm-size" v-if="!f.isDir">{{ formatSize(f.size) }}</span>
        <span class="fm-time">{{ f.modTime }}</span>
      </div>
      <div v-if="!fileLoading && files.length === 0" class="fm-empty">空目录</div>
    </div>

    <div class="fm-status">
      <span>{{ files.length }} 项</span>
      <span v-if="selectedFile">选中: {{ selectedFile.name }}</span>
    </div>

    <el-dialog v-model="showMkdirDialog" title="新建文件夹" width="400px" :lock-scroll="false">
      <el-input v-model="newDirName" placeholder="文件夹名称" @keyup.enter="handleMkdir" />
      <template #footer>
        <el-button @click="showMkdirDialog = false">取消</el-button>
        <el-button type="primary" @click="handleMkdir" :loading="mkdirLoading">创建</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showRenameDialog" title="重命名" width="400px" :lock-scroll="false">
      <el-input v-model="renameName" placeholder="新名称" @keyup.enter="handleRename" />
      <template #footer>
        <el-button @click="showRenameDialog = false">取消</el-button>
        <el-button type="primary" @click="handleRename" :loading="renameLoading">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showFileEditor" :title="'编辑: ' + editingFileName" width="700px" :lock-scroll="false">
      <el-input v-model="editingContent" type="textarea" :rows="20" style="font-family: monospace" />
      <template #footer>
        <el-button @click="showFileEditor = false">取消</el-button>
        <el-button type="primary" @click="handleSaveFile" :loading="savingFile">保存</el-button>
      </template>
    </el-dialog>

    <ul v-if="contextMenu.visible" class="context-menu" :style="{ left: contextMenu.x + 'px', top: contextMenu.y + 'px' }">
      <li v-if="contextMenu.file && contextMenu.file.isDir" @click="openFile(contextMenu.file)">打开</li>
      <li v-if="contextMenu.file && !contextMenu.file.isDir" @click="downloadFile(contextMenu.file)">下载</li>
      <li v-if="contextMenu.file && contextMenu.file.isDir" @click="downloadDir(contextMenu.file)">下载(压缩)</li>
      <li v-if="contextMenu.file && isTextFile(contextMenu.file.name)" @click="editFile(contextMenu.file)">编辑</li>
      <li @click="renameFile(contextMenu.file)">重命名</li>
      <li class="context-menu-danger" @click="deleteFile(contextMenu.file)">删除</li>
    </ul>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { sftpList, sftpDownload, sftpDownloadDir, sftpUpload, sftpMkdir, sftpRemove, sftpRename, sftpReadFile, sftpWriteFile } from '../api/sftp'
import { ElMessage, ElMessageBox } from 'element-plus'

defineOptions({ name: 'FileManager' })

const props = defineProps({
  machineId: {
    type: [String, Number],
    required: true
  }
})

const emit = defineEmits(['path-changed'])

const fileLoading = ref(false)
const currentPath = ref('/')
const files = ref([])
const selectedFile = ref(null)

const showMkdirDialog = ref(false)
const newDirName = ref('')
const mkdirLoading = ref(false)

const showRenameDialog = ref(false)
const renameName = ref('')
const renameOldPath = ref('')
const renameLoading = ref(false)

const showFileEditor = ref(false)
const editingFileName = ref('')
const editingFilePath = ref('')
const editingContent = ref('')
const savingFile = ref(false)

const uploadInput = ref(null)

const contextMenu = ref({ visible: false, x: 0, y: 0, file: null })

const pathSegments = computed(() => {
  const segs = [{ name: '/', path: '/' }]
  if (currentPath.value === '/') return segs
  const parts = currentPath.value.split('/').filter(Boolean)
  let p = ''
  for (const part of parts) {
    p += '/' + part
    segs.push({ name: part, path: p })
  }
  return segs
})

const loadFiles = async (dirPath) => {
  if (!props.machineId) return
  fileLoading.value = true
  try {
    const res = await sftpList(props.machineId, dirPath)
    currentPath.value = res.data.path
    files.value = res.data.files || []
    selectedFile.value = null
    emit('path-changed', currentPath.value)
  } catch (e) {
    ElMessage.error('读取目录失败')
  } finally {
    fileLoading.value = false
  }
}

const refreshFiles = () => loadFiles(currentPath.value)
const goHome = () => loadFiles('/')
const goBack = () => {
  if (currentPath.value === '/') return
  const parent = currentPath.value.split('/').slice(0, -1).join('/') || '/'
  loadFiles(parent)
}
const navigateTo = (path) => loadFiles(path)

const selectFile = (f) => {
  selectedFile.value = f
}

const openFile = (f) => {
  if (f.isDir) {
    loadFiles(f.path)
  } else if (isTextFile(f.name)) {
    editFile(f)
  } else {
    downloadFile(f)
  }
}

const fileIcon = (name) => {
  const ext = name.split('.').pop().toLowerCase()
  const icons = {
    jpg: '🖼️', jpeg: '🖼️', png: '🖼️', gif: '🖼️', svg: '🖼️', webp: '🖼️',
    mp4: '🎬', avi: '🎬', mkv: '🎬', mov: '🎬',
    mp3: '🎵', wav: '🎵', flac: '🎵',
    zip: '📦', tar: '📦', gz: '📦', rar: '📦', '7z': '📦',
    pdf: '📄', doc: '📝', docx: '📝', xls: '📊', xlsx: '📊',
    go: '🔷', py: '🐍', js: '📜', ts: '📜', vue: '💚',
    sh: '⚙️', yaml: '⚙️', yml: '⚙️', json: '⚙️', toml: '⚙️', conf: '⚙️', cfg: '⚙️',
  }
  return icons[ext] || '📄'
}

const isTextFile = (name) => {
  const ext = name.split('.').pop().toLowerCase()
  const textExts = ['txt', 'log', 'conf', 'cfg', 'ini', 'yaml', 'yml', 'json', 'toml', 'xml', 'md',
    'sh', 'bash', 'zsh', 'env', 'gitignore', 'dockerignore',
    'js', 'ts', 'jsx', 'tsx', 'vue', 'html', 'css', 'scss', 'less',
    'go', 'py', 'java', 'c', 'cpp', 'h', 'rs', 'rb', 'php', 'pl',
    'sql', 'proto', 'graphql',
    'dockerfile', 'makefile', 'cmake']
  return textExts.includes(ext) || name.toLowerCase() === 'dockerfile' || name.toLowerCase() === 'makefile'
}

const formatSize = (bytes) => {
  if (bytes === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  return (bytes / Math.pow(1024, i)).toFixed(i > 0 ? 1 : 0) + ' ' + units[i]
}

const downloadFile = (f) => {
  if (!props.machineId) return
  const url = sftpDownload(props.machineId, f.path)
  const a = document.createElement('a')
  a.href = url
  a.download = f.name
  a.click()
}

const downloadDir = (f) => {
  if (!props.machineId) return
  const url = sftpDownloadDir(props.machineId, f.path)
  const a = document.createElement('a')
  a.href = url
  a.download = f.name + '.zip'
  a.click()
}

const triggerUpload = () => {
  uploadInput.value?.click()
}

const handleUpload = async (e) => {
  const fileList = e.target.files
  if (!fileList || fileList.length === 0) return
  for (const file of fileList) {
    try {
      await sftpUpload(props.machineId, currentPath.value, file)
      ElMessage.success(`${file.name} 上传成功`)
    } catch {
      ElMessage.error(`${file.name} 上传失败`)
    }
  }
  refreshFiles()
  e.target.value = ''
}

const handleMkdir = async () => {
  if (!newDirName.value.trim()) return
  mkdirLoading.value = true
  try {
    const dirPath = currentPath.value === '/' ? '/' + newDirName.value.trim() : currentPath.value + '/' + newDirName.value.trim()
    await sftpMkdir(props.machineId, dirPath)
    ElMessage.success('创建成功')
    showMkdirDialog.value = false
    newDirName.value = ''
    refreshFiles()
  } catch {
    ElMessage.error('创建失败')
  } finally {
    mkdirLoading.value = false
  }
}

const editFile = async (f) => {
  try {
    const res = await sftpReadFile(props.machineId, f.path)
    editingFileName.value = f.name
    editingFilePath.value = f.path
    editingContent.value = res.data.content
    showFileEditor.value = true
  } catch (e) {
    ElMessage.error('读取文件失败')
  }
}

const handleSaveFile = async () => {
  savingFile.value = true
  try {
    await sftpWriteFile(props.machineId, editingFilePath.value, editingContent.value)
    ElMessage.success('保存成功')
    showFileEditor.value = false
    refreshFiles()
  } catch {
    ElMessage.error('保存失败')
  } finally {
    savingFile.value = false
  }
}

const renameFile = (f) => {
  if (!f) return
  renameOldPath.value = f.path
  renameName.value = f.name
  showRenameDialog.value = true
  hideContextMenu()
}

const handleRename = async () => {
  if (!renameName.value.trim()) return
  renameLoading.value = true
  try {
    const newPath = currentPath.value === '/' ? '/' + renameName.value.trim() : currentPath.value + '/' + renameName.value.trim()
    await sftpRename(props.machineId, renameOldPath.value, newPath)
    ElMessage.success('重命名成功')
    showRenameDialog.value = false
    refreshFiles()
  } catch {
    ElMessage.error('重命名失败')
  } finally {
    renameLoading.value = false
  }
}

const deleteFile = async (f) => {
  if (!f) return
  hideContextMenu()
  try {
    await ElMessageBox.confirm(`确定删除 ${f.isDir ? '文件夹' : '文件'} "${f.name}" 吗？`, '删除确认', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await sftpRemove(props.machineId, f.path, f.isDir)
    ElMessage.success('删除成功')
    if (selectedFile.value && selectedFile.value.path === f.path) {
      selectedFile.value = null
    }
    refreshFiles()
  } catch {
  }
}

const showContextMenu = (event, f) => {
  selectedFile.value = f
  contextMenu.value = { visible: true, x: event.clientX, y: event.clientY, file: f }
}

const hideContextMenu = () => {
  contextMenu.value.visible = false
}

watch(() => props.machineId, (newId) => {
  if (newId) {
    currentPath.value = '/'
    files.value = []
    selectedFile.value = null
    loadFiles('/')
  }
})

defineExpose({
  refreshFiles,
  loadFiles
})

onMounted(() => {
  if (props.machineId) {
    loadFiles('/')
  }
  document.addEventListener('click', hideContextMenu)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', hideContextMenu)
})
</script>

<style scoped>
.file-manager {
  flex: 1 1 40%;
  min-width: 320px;
  display: flex;
  flex-direction: column;
  background: #fff;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  overflow: hidden;
}

.fm-toolbar {
  display: flex;
  gap: 4px;
  padding: 8px;
  border-bottom: 1px solid #ebeef5;
  flex-wrap: wrap;
}

.fm-breadcrumb {
  padding: 6px 12px;
  font-size: 13px;
  background: #f5f7fa;
  border-bottom: 1px solid #ebeef5;
  overflow-x: auto;
  white-space: nowrap;
}

.breadcrumb-item a {
  color: #409eff;
  cursor: pointer;
  text-decoration: none;
}

.breadcrumb-item a:hover {
  text-decoration: underline;
}

.fm-list {
  flex: 1;
  overflow-y: auto;
  padding: 4px 0;
  min-height: 300px;
  max-height: calc(100vh - 360px);
}

.fm-item {
  display: flex;
  align-items: center;
  padding: 6px 12px;
  cursor: pointer;
  font-size: 13px;
  border-bottom: 1px solid #f0f0f0;
}

.fm-item:hover {
  background: #ecf5ff;
}

.fm-selected {
  background: #d9ecff !important;
}

.fm-icon {
  width: 24px;
  text-align: center;
  flex-shrink: 0;
}

.fm-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-right: 8px;
}

.fm-size {
  width: 70px;
  text-align: right;
  color: #909399;
  font-size: 12px;
  flex-shrink: 0;
}

.fm-time {
  width: 140px;
  text-align: right;
  color: #909399;
  font-size: 12px;
  flex-shrink: 0;
}

.fm-empty {
  text-align: center;
  color: #b0b0b0;
  padding: 40px 0;
}

.fm-status {
  padding: 6px 12px;
  font-size: 12px;
  color: #909399;
  border-top: 1px solid #ebeef5;
  display: flex;
  gap: 16px;
}

.context-menu {
  position: fixed;
  background: #fff;
  border: 1px solid #ddd;
  border-radius: 6px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.15);
  list-style: none;
  padding: 4px 0;
  margin: 0;
  z-index: 9999;
  min-width: 120px;
}

.context-menu li {
  padding: 8px 16px;
  cursor: pointer;
  font-size: 13px;
}

.context-menu li:hover {
  background: #ecf5ff;
  color: #409eff;
}

.context-menu-danger:hover {
  background: #fef0f0 !important;
  color: #f56c6c !important;
}
</style>
