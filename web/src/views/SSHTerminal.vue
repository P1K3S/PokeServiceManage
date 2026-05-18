<template>
  <div>
    <el-card shadow="hover">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span>SSH 终端</span>
          <div style="display: flex; gap: 8px; align-items: center">
            <el-select v-model="selectedMachineId" placeholder="选择主机" style="width: 220px" @change="handleMachineChange">
              <el-option v-for="m in machines" :key="m.id" :label="m.name + ' (' + m.ip + ')'" :value="m.id">
                <span>{{ m.name }}</span>
                <span style="float: right; color: #999; font-size: 12px">{{ m.ip }}</span>
              </el-option>
            </el-select>
            <el-button type="primary" @click="connect" :disabled="!selectedMachineId || connected">连接</el-button>
            <el-button type="danger" @click="disconnect" :disabled="!connected">断开</el-button>
            <el-button :type="showFileManager ? 'warning' : 'success'" @click="toggleFileManager" :disabled="!selectedMachineId">
              {{ showFileManager ? '关闭文件管理' : '文件管理' }}
            </el-button>
          </div>
        </div>
      </template>

      <div :style="{ display: showFileManager ? 'flex' : 'block', gap: '12px' }">
        <div ref="terminalContainer" class="terminal-container" :style="{ flex: showFileManager ? '1 1 60%' : '1', minWidth: 0 }"></div>

        <div v-if="showFileManager" class="file-manager">
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
        </div>
      </div>
    </el-card>

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
import { ref, computed, onMounted, onBeforeUnmount, onActivated, nextTick } from 'vue'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'
import { getMachines } from '../api/machine'
import { sftpList, sftpDownload, sftpDownloadDir, sftpUpload, sftpMkdir, sftpRemove, sftpRename, sftpReadFile, sftpWriteFile } from '../api/sftp'
import { ElMessage, ElMessageBox } from 'element-plus'

defineOptions({ name: 'SSHTerminal' })

const machines = ref([])
const selectedMachineId = ref('')
const connected = ref(false)
const terminalContainer = ref(null)
const showFileManager = ref(false)

let term = null
let fitAddon = null
let ws = null
let connectionId = 0

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
const renameIsDir = ref(false)
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

const fetchMachines = async () => {
  try {
    const res = await getMachines({ pageSize: 999 })
    machines.value = (res.data.list || []).filter(m => m.sshEnabled)
  } catch {}
}

const initTerminal = () => {
  if (term) {
    term.dispose()
    term = null
  }
  term = new Terminal({
    cursorBlink: true,
    fontSize: 14,
    fontFamily: '"Cascadia Code", "Fira Code", Menlo, Monaco, "Courier New", monospace',
    theme: {
      background: '#1e1e2e',
      foreground: '#cdd6f4',
      cursor: '#f5e0dc',
      selectionBackground: '#585b70',
      black: '#45475a',
      red: '#f38ba8',
      green: '#a6e3a1',
      yellow: '#f9e2af',
      blue: '#89b4fa',
      magenta: '#f5c2e7',
      cyan: '#94e2d5',
      white: '#bac2de',
      brightBlack: '#585b70',
      brightRed: '#f38ba8',
      brightGreen: '#a6e3a1',
      brightYellow: '#f9e2af',
      brightBlue: '#89b4fa',
      brightMagenta: '#f5c2e7',
      brightCyan: '#94e2d5',
      brightWhite: '#a6adc8'
    }
  })
  fitAddon = new FitAddon()
  term.loadAddon(fitAddon)
  term.open(terminalContainer.value)
  fitAddon.fit()

  term.onData((data) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(new Uint8Array([0, ...new TextEncoder().encode(data)]))
    }
  })

  term.onResize(({ cols, rows }) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(new Uint8Array([1, rows, cols]))
    }
  })

  window.addEventListener('resize', handleResize)
}

const handleResize = () => {
  if (fitAddon) fitAddon.fit()
}

const closeWs = () => {
  if (ws) {
    ws.onopen = null
    ws.onmessage = null
    ws.onclose = null
    ws.onerror = null
    if (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING) {
      ws.close()
    }
    ws = null
  }
  connected.value = false
}

const connect = async () => {
  if (!selectedMachineId.value) return

  closeWs()

  connectionId++
  const currentConnId = connectionId

  if (term) term.reset()

  const token = localStorage.getItem('token')
  if (!token) {
    ElMessage.error('未登录，请先登录')
    return
  }

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const host = window.location.host
  const cols = term ? term.cols : 120
  const rows = term ? term.rows : 30
  const url = `${protocol}//${host}/api/ssh-terminal/${selectedMachineId.value}?token=${encodeURIComponent(token)}&cols=${cols}&rows=${rows}`

  term.writeln('\r\n\x1b[33m正在连接...\x1b[0m')

  try {
    ws = new WebSocket(url)
  } catch (e) {
    term.writeln('\r\n\x1b[31mWebSocket创建失败: ' + e.message + '\x1b[0m')
    return
  }
  ws.binaryType = 'arraybuffer'

  ws.onopen = () => {
    if (currentConnId !== connectionId) return
    connected.value = true
    term.writeln('\x1b[32m已连接\x1b[0m\r\n')
  }

  ws.onmessage = (event) => {
    if (currentConnId !== connectionId) return
    if (event.data instanceof ArrayBuffer) {
      term.write(new Uint8Array(event.data))
    } else {
      term.write(event.data)
    }
  }

  ws.onclose = (event) => {
    if (currentConnId !== connectionId) return
    connected.value = false
    if (event.code === 1006) {
      term.writeln('\r\n\x1b[31m连接异常断开（可能是认证失败或服务端未启动）\x1b[0m')
    } else {
      term.writeln('\r\n\x1b[31m连接已断开 (code: ' + event.code + ')\x1b[0m')
    }
  }

  ws.onerror = () => {
    if (currentConnId !== connectionId) return
    connected.value = false
    term.writeln('\r\n\x1b[31m连接出错\x1b[0m')
    ElMessage.error('WebSocket连接失败，请检查后端服务是否启动')
  }
}

const disconnect = () => {
  closeWs()
  if (term) {
    term.writeln('\r\n\x1b[33m已断开连接\x1b[0m')
  }
}

const handleMachineChange = () => {
  closeWs()
  if (term) term.reset()
  currentPath.value = '/'
  files.value = []
  selectedFile.value = null
  if (showFileManager.value && selectedMachineId.value) {
    loadFiles('/')
  }
}

const toggleFileManager = () => {
  showFileManager.value = !showFileManager.value
  if (showFileManager.value && selectedMachineId.value) {
    loadFiles(currentPath.value)
  }
  nextTick(() => {
    if (fitAddon) fitAddon.fit()
  })
}

const loadFiles = async (dirPath) => {
  if (!selectedMachineId.value) return
  fileLoading.value = true
  try {
    const res = await sftpList(selectedMachineId.value, dirPath)
    currentPath.value = res.data.path
    files.value = res.data.files || []
    selectedFile.value = null
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
  if (!selectedMachineId.value) return
  const url = sftpDownload(selectedMachineId.value, f.path)
  const a = document.createElement('a')
  a.href = url
  a.download = f.name
  a.click()
}

const downloadDir = (f) => {
  if (!selectedMachineId.value) return
  const url = sftpDownloadDir(selectedMachineId.value, f.path)
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
      await sftpUpload(selectedMachineId.value, currentPath.value, file)
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
    await sftpMkdir(selectedMachineId.value, dirPath)
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
    const res = await sftpReadFile(selectedMachineId.value, f.path)
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
    await sftpWriteFile(selectedMachineId.value, editingFilePath.value, editingContent.value)
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
  renameIsDir.value = f.isDir
  showRenameDialog.value = true
  hideContextMenu()
}

const handleRename = async () => {
  if (!renameName.value.trim()) return
  renameLoading.value = true
  try {
    const newPath = currentPath.value === '/' ? '/' + renameName.value.trim() : currentPath.value + '/' + renameName.value.trim()
    await sftpRename(selectedMachineId.value, renameOldPath.value, newPath)
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
    await sftpRemove(selectedMachineId.value, f.path, f.isDir)
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

document.addEventListener('click', hideContextMenu)

onMounted(async () => {
  await nextTick()
  initTerminal()
  fetchMachines()
})

onActivated(() => {
  nextTick(() => {
    if (fitAddon && term) {
      fitAddon.fit()
    }
  })
})

onBeforeUnmount(() => {
  closeWs()
  if (term) {
    term.dispose()
    term = null
  }
  window.removeEventListener('resize', handleResize)
  document.removeEventListener('click', hideContextMenu)
})
</script>

<style scoped>
.terminal-container {
  width: 100%;
  height: calc(100vh - 220px);
  min-height: 400px;
  background: #1e1e2e;
  border-radius: 8px;
  padding: 4px;
}

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
