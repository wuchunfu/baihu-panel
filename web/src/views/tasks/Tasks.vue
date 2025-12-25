<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle } from '@/components/ui/alert-dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Badge } from '@/components/ui/badge'
import { Checkbox } from '@/components/ui/checkbox'
import Pagination from '@/components/Pagination.vue'
import DirTreeSelect from '@/components/DirTreeSelect.vue'
import { Plus, Play, Pencil, Trash2, Search, ScrollText, ChevronDown, X, GitBranch, Terminal } from 'lucide-vue-next'
import { api, type Task, type EnvVar, type RepoConfig } from '@/api'
import { toast } from 'vue-sonner'
import { useSiteSettings } from '@/composables/useSiteSettings'
import { useRouter } from 'vue-router'
import TextOverflow from '@/components/TextOverflow.vue'

const router = useRouter()
const { pageSize } = useSiteSettings()

const tasks = ref<Task[]>([])
const showDialog = ref(false)
const showRepoDialog = ref(false)
const editingTask = ref<Partial<Task>>({})
const isEdit = ref(false)
const showDeleteDialog = ref(false)
const deleteTaskId = ref<number | null>(null)

// 清理配置
const cleanType = ref('')
const cleanKeep = ref(30)

// 仓库同步配置
const repoConfig = ref<RepoConfig>({
  source_type: 'git',
  source_url: '',
  target_path: '',
  branch: '',
  sparse_path: '',
  single_file: false,
  proxy: 'none',
  proxy_url: '',
  auth_token: ''
})

// 环境变量
const allEnvVars = ref<EnvVar[]>([])
const selectedEnvIds = ref<number[]>([])
const envSearchQuery = ref('')

const filterName = ref('')
const currentPage = ref(1)
const total = ref(0)
let searchTimer: ReturnType<typeof setTimeout> | null = null

const cronPresets = [
  { label: '每5秒', value: '*/5 * * * * *' },
  { label: '每30秒', value: '*/30 * * * * *' },
  { label: '每分钟', value: '0 * * * * *' },
  { label: '每5分钟', value: '0 */5 * * * *' },
  { label: '每小时', value: '0 0 * * * *' },
  { label: '每天0点', value: '0 0 0 * * *' },
  { label: '每天8点', value: '0 0 8 * * *' },
  { label: '每周一', value: '0 0 0 * * 1' },
  { label: '每月1号', value: '0 0 0 1 * *' },
]

const proxyOptions = [
  { label: '不使用代理', value: 'none' },
  { label: 'ghproxy.com', value: 'ghproxy' },
  { label: 'mirror.ghproxy.com', value: 'mirror' },
  { label: '自定义代理', value: 'custom' },
]

// 计算清理配置 JSON
const cleanConfig = computed(() => {
  if (!cleanType.value || cleanType.value === 'none' || cleanKeep.value <= 0) return ''
  return JSON.stringify({ type: cleanType.value, keep: cleanKeep.value })
})

// 过滤后的环境变量列表（排除已选中的）
const filteredEnvVars = computed(() => {
  return allEnvVars.value.filter(env => {
    const matchSearch = !envSearchQuery.value || env.name.toLowerCase().includes(envSearchQuery.value.toLowerCase())
    const notSelected = !selectedEnvIds.value.includes(env.id)
    return matchSearch && notSelected
  })
})

// 已选中的环境变量对象列表
const selectedEnvs = computed(() => {
  return selectedEnvIds.value
    .map(id => allEnvVars.value.find(e => e.id === id))
    .filter((e): e is EnvVar => e !== undefined)
})

// 计算 envs 字符串
const envsString = computed(() => selectedEnvIds.value.join(','))

async function loadEnvVars() {
  try {
    allEnvVars.value = await api.env.all()
  } catch { /* ignore */ }
}

function addEnv(id: number) {
  if (!selectedEnvIds.value.includes(id)) {
    selectedEnvIds.value.push(id)
  }
  envSearchQuery.value = ''
}

function removeEnv(id: number) {
  const idx = selectedEnvIds.value.indexOf(id)
  if (idx !== -1) {
    selectedEnvIds.value.splice(idx, 1)
  }
}

async function loadTasks() {
  try {
    const res = await api.tasks.list({ page: currentPage.value, page_size: pageSize.value, name: filterName.value || undefined })
    tasks.value = res.data
    total.value = res.total
  } catch { toast.error('加载任务失败') }
}

function handleSearch() {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    currentPage.value = 1
    loadTasks()
  }, 300)
}

function handlePageChange(page: number) {
  currentPage.value = page
  loadTasks()
}

function openCreate() {
  editingTask.value = { name: '', command: '', type: 'task', schedule: '0 * * * * *', timeout: 30, work_dir: '', enabled: true, clean_config: '', envs: '' }
  cleanType.value = 'none'
  cleanKeep.value = 30
  selectedEnvIds.value = []
  envSearchQuery.value = ''
  isEdit.value = false
  showDialog.value = true
}

function openCreateRepo() {
  editingTask.value = { name: '', type: 'repo', schedule: '0 0 0 * * *', timeout: 30, enabled: true, clean_config: '', envs: '' }
  repoConfig.value = { source_type: 'git', source_url: '', target_path: '', branch: '', sparse_path: '', single_file: false, proxy: 'none', proxy_url: '', auth_token: '' }
  cleanType.value = 'none'
  cleanKeep.value = 30
  isEdit.value = false
  showRepoDialog.value = true
}

function openEdit(task: Task) {
  editingTask.value = { ...task }
  // 解析清理配置
  if (task.clean_config) {
    try {
      const config = JSON.parse(task.clean_config)
      cleanType.value = config.type || 'none'
      cleanKeep.value = config.keep || 30
    } catch {
      cleanType.value = 'none'
      cleanKeep.value = 30
    }
  } else {
    cleanType.value = 'none'
    cleanKeep.value = 30
  }
  // 解析环境变量
  if (task.envs) {
    selectedEnvIds.value = task.envs.split(',').map(s => parseInt(s.trim())).filter(n => !isNaN(n))
  } else {
    selectedEnvIds.value = []
  }
  envSearchQuery.value = ''
  isEdit.value = true

  // 根据任务类型打开不同弹窗
  if (task.type === 'repo') {
    if (task.config) {
      try {
        repoConfig.value = JSON.parse(task.config)
      } catch {
        repoConfig.value = { source_type: 'git', source_url: '', target_path: '', branch: '', sparse_path: '', single_file: false, proxy: 'none', proxy_url: '', auth_token: '' }
      }
    }
    showRepoDialog.value = true
  } else {
    showDialog.value = true
  }
}

async function saveTask() {
  try {
    editingTask.value.clean_config = cleanConfig.value
    editingTask.value.envs = envsString.value
    editingTask.value.type = 'task'
    if (isEdit.value && editingTask.value.id) {
      await api.tasks.update(editingTask.value.id, editingTask.value)
      toast.success('任务已更新')
    } else {
      await api.tasks.create(editingTask.value)
      toast.success('任务已创建')
    }
    showDialog.value = false
    loadTasks()
  } catch { toast.error('保存失败') }
}

async function saveRepoTask() {
  try {
    editingTask.value.clean_config = cleanConfig.value
    editingTask.value.type = 'repo'
    editingTask.value.config = JSON.stringify(repoConfig.value)
    editingTask.value.command = `[${repoConfig.value.source_type}] ${repoConfig.value.source_url}`
    if (isEdit.value && editingTask.value.id) {
      await api.tasks.update(editingTask.value.id, editingTask.value)
      toast.success('同步任务已更新')
    } else {
      await api.tasks.create(editingTask.value)
      toast.success('同步任务已创建')
    }
    showRepoDialog.value = false
    loadTasks()
  } catch { toast.error('保存失败') }
}

function confirmDelete(id: number) {
  deleteTaskId.value = id
  showDeleteDialog.value = true
}

async function deleteTask() {
  if (!deleteTaskId.value) return
  try {
    await api.tasks.delete(deleteTaskId.value)
    toast.success('任务已删除')
    loadTasks()
  } catch { toast.error('删除失败') }
  showDeleteDialog.value = false
  deleteTaskId.value = null
}

async function runTask(id: number) {
  try { await api.tasks.execute(id); toast.success('任务已执行') } catch { toast.error('执行失败') }
}

async function toggleTask(task: Task, enabled: boolean) {
  try {
    await api.tasks.update(task.id, { ...task, enabled })
    toast.success(enabled ? '任务已启用' : '任务已禁用')
    loadTasks()
  } catch { toast.error('操作失败') }
}

function viewLogs(taskId: number) {
  router.push({ path: '/history', query: { task_id: String(taskId) } })
}

function getTaskTypeTitle(type: string) {
  return type === 'repo' ? '仓库同步' : '普通任务'
}

onMounted(() => {
  loadTasks()
  loadEnvVars()
})
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h2 class="text-xl sm:text-2xl font-bold tracking-tight">定时任务</h2>
        <p class="text-muted-foreground text-sm">管理和调度自动化任务</p>
      </div>
      <div class="flex items-center gap-2">
        <div class="relative flex-1 sm:flex-none">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input v-model="filterName" placeholder="搜索任务..." class="h-9 pl-9 w-full sm:w-56 text-sm" @input="handleSearch" />
        </div>
        <Button variant="outline" @click="openCreateRepo" class="shrink-0">
          <GitBranch class="h-4 w-4 sm:mr-2" /> <span class="hidden sm:inline">仓库同步</span>
        </Button>
        <Button @click="openCreate" class="shrink-0">
          <Plus class="h-4 w-4 sm:mr-2" /> <span class="hidden sm:inline">新建任务</span>
        </Button>
      </div>
    </div>

    <div class="rounded-lg border bg-card overflow-x-auto">
      <!-- 表头 -->
      <div class="flex items-center gap-4 px-4 py-2 border-b bg-muted/50 text-sm text-muted-foreground font-medium min-w-[700px]">
        <span class="w-12 shrink-0">ID</span>
        <span class="w-8 shrink-0 text-center">类型</span>
        <span class="w-20 sm:w-28 shrink-0">名称</span>
        <span class="w-32 sm:flex-1 shrink-0 sm:shrink">命令/地址</span>
        <span class="w-32 shrink-0 hidden md:block">定时规则</span>
        <span class="w-40 shrink-0 hidden lg:block">上次执行</span>
        <span class="w-40 shrink-0 hidden lg:block">下次执行</span>
        <span class="w-12 shrink-0 text-center">状态</span>
        <span class="w-36 shrink-0 text-center">操作</span>
      </div>
      <!-- 列表 -->
      <div class="divide-y min-w-[700px]">
        <div v-if="tasks.length === 0" class="text-sm text-muted-foreground text-center py-8">
          暂无任务
        </div>
        <div
          v-for="task in tasks"
          :key="task.id"
          class="flex items-center gap-4 px-4 py-2 hover:bg-muted/50 transition-colors"
        >
          <span class="w-12 shrink-0 text-muted-foreground text-sm">#{{ task.id }}</span>
          <span class="w-8 shrink-0 flex justify-center" :title="getTaskTypeTitle(task.type || 'task')">
            <GitBranch v-if="task.type === 'repo'" class="h-4 w-4 text-primary" />
            <Terminal v-else class="h-4 w-4 text-muted-foreground" />
          </span>
          <span class="w-20 sm:w-28 font-medium truncate shrink-0 text-sm">
            <TextOverflow :text="task.name" title="任务名称" />
          </span>
          <code class="w-32 sm:flex-1 shrink-0 sm:shrink text-muted-foreground truncate text-xs bg-muted px-2 py-1 rounded">
            <TextOverflow :text="task.command" :title="task.type === 'repo' ? '同步地址' : '执行命令'" />
          </code>
          <code class="w-36 shrink-0 text-muted-foreground text-xs bg-muted px-2 py-1 rounded hidden md:block">{{ task.schedule }}</code>
          <span class="w-40 shrink-0 text-muted-foreground text-xs hidden lg:block">{{ task.last_run || '-' }}</span>
          <span class="w-40 shrink-0 text-muted-foreground text-xs hidden lg:block">{{ task.next_run || '-' }}</span>
          <span class="w-12 flex justify-center shrink-0 cursor-pointer" @click="toggleTask(task, !task.enabled)" :title="task.enabled ? '点击禁用' : '点击启用'">
            <span :class="['w-2 h-2 rounded-full', task.enabled ? 'bg-green-500' : 'bg-gray-400']" />
          </span>
          <span class="w-36 shrink-0 flex justify-center gap-1">
            <Button variant="ghost" size="icon" class="h-7 w-7" @click="runTask(task.id)" title="执行">
              <Play class="h-3.5 w-3.5" />
            </Button>
            <Button variant="ghost" size="icon" class="h-7 w-7" @click="viewLogs(task.id)" title="日志">
              <ScrollText class="h-3.5 w-3.5" />
            </Button>
            <Button variant="ghost" size="icon" class="h-7 w-7" @click="openEdit(task)" title="编辑">
              <Pencil class="h-3.5 w-3.5" />
            </Button>
            <Button variant="ghost" size="icon" class="h-7 w-7 text-destructive" @click="confirmDelete(task.id)" title="删除">
              <Trash2 class="h-3.5 w-3.5" />
            </Button>
          </span>
        </div>
      </div>
      <!-- 分页 -->
      <Pagination :total="total" :page="currentPage" @update:page="handlePageChange" />
    </div>

    <!-- 普通任务弹窗 -->
    <Dialog v-model:open="showDialog">
      <DialogContent class="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>{{ isEdit ? '编辑任务' : '新建任务' }}</DialogTitle>
        </DialogHeader>
        <div class="grid gap-4 py-4">
          <div class="grid grid-cols-4 items-center gap-4">
            <Label class="text-right">任务名称</Label>
            <Input v-model="editingTask.name" placeholder="我的任务" class="col-span-3" />
          </div>
          <div class="grid grid-cols-4 items-center gap-4">
            <Label class="text-right">执行命令</Label>
            <Input v-model="editingTask.command" placeholder="node script.js" class="col-span-3 font-mono" />
          </div>
          <div class="grid grid-cols-4 items-center gap-4">
            <Label class="text-right">工作目录</Label>
            <div class="col-span-3">
              <DirTreeSelect :model-value="editingTask.work_dir || ''" @update:model-value="v => editingTask.work_dir = v" />
            </div>
          </div>
          <div class="grid grid-cols-4 items-center gap-4">
            <Label class="text-right">定时规则</Label>
            <Input v-model="editingTask.schedule" placeholder="0 * * * * *" class="col-span-3 font-mono" />
          </div>
          <div class="grid grid-cols-4 items-start gap-4">
            <span></span>
            <div class="col-span-3">
              <p class="text-xs text-muted-foreground mb-2">格式: 秒 分 时 日 月 周</p>
              <div class="flex flex-wrap gap-1.5">
              <span
                v-for="preset in cronPresets"
                :key="preset.value"
                class="px-2 py-0.5 text-xs rounded-md bg-muted hover:bg-accent cursor-pointer transition-colors"
                @click="editingTask.schedule = preset.value"
              >
                {{ preset.label }}
              </span>
              </div>
            </div>
          </div>
          <div class="grid grid-cols-4 items-center gap-4">
            <Label class="text-right">超时(分钟)</Label>
            <Input v-model.number="editingTask.timeout" type="number" placeholder="30" class="col-span-3" />
          </div>
          <div class="grid grid-cols-4 items-center gap-4">
            <Label class="text-right">日志清理</Label>
            <div class="col-span-3 flex gap-2">
              <Select :model-value="cleanType" @update:model-value="(v) => cleanType = String(v || 'none')">
                <SelectTrigger class="w-28">
                  <SelectValue placeholder="不清理" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="none">不清理</SelectItem>
                  <SelectItem value="day">按天数</SelectItem>
                  <SelectItem value="count">按条数</SelectItem>
                </SelectContent>
              </Select>
              <Input v-if="cleanType && cleanType !== 'none'" v-model.number="cleanKeep" type="number" :placeholder="cleanType === 'day' ? '保留天数' : '保留条数'" class="flex-1" />
            </div>
          </div>
          <div class="grid grid-cols-4 items-start gap-4">
            <Label class="text-right pt-2">环境变量</Label>
            <div class="col-span-3 space-y-2">
              <Popover>
                <PopoverTrigger as-child>
                  <Button variant="outline" class="w-full justify-between font-normal">
                    <span class="text-muted-foreground">搜索并添加环境变量...</span>
                    <ChevronDown class="h-4 w-4 shrink-0 opacity-50" />
                  </Button>
                </PopoverTrigger>
                <PopoverContent class="w-[300px] p-2" align="start">
                  <Input v-model="envSearchQuery" placeholder="搜索环境变量..." class="mb-2 h-8" />
                  <div v-if="filteredEnvVars.length === 0" class="text-sm text-muted-foreground text-center py-2">
                    {{ allEnvVars.length === 0 ? '暂无环境变量' : '无匹配结果' }}
                  </div>
                  <div v-else class="max-h-[160px] overflow-y-auto space-y-1">
                    <div
                      v-for="env in filteredEnvVars"
                      :key="env.id"
                      class="flex items-center gap-2 px-2 py-1.5 rounded hover:bg-muted cursor-pointer text-sm"
                      @click="addEnv(env.id)"
                    >
                      <Plus class="h-3.5 w-3.5 text-muted-foreground" />
                      <span class="truncate">{{ env.name }}</span>
                    </div>
                  </div>
                </PopoverContent>
              </Popover>
              <div v-if="selectedEnvs.length > 0" class="flex flex-wrap gap-1.5">
                <Badge
                  v-for="env in selectedEnvs"
                  :key="env.id"
                  variant="secondary"
                  class="gap-1 pr-1"
                >
                  {{ env.name }}
                  <X class="h-3 w-3 cursor-pointer hover:text-destructive" @click="removeEnv(env.id)" />
                </Badge>
              </div>
            </div>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" @click="showDialog = false">取消</Button>
          <Button @click="saveTask">保存</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- 仓库同步任务弹窗 -->
    <Dialog v-model:open="showRepoDialog">
      <DialogContent class="sm:max-w-[500px] max-h-[85vh] flex flex-col">
        <DialogHeader>
          <DialogTitle>{{ isEdit ? '编辑仓库同步' : '新建仓库同步' }}</DialogTitle>
        </DialogHeader>
        <div class="grid gap-4 py-4 overflow-y-auto flex-1 pr-4 custom-scrollbar">
          <div class="grid grid-cols-4 items-center gap-4">
            <Label class="text-right text-sm">任务名称</Label>
            <Input v-model="editingTask.name" placeholder="我的仓库同步" class="col-span-3" />
          </div>
          <div class="grid grid-cols-4 items-center gap-4">
            <Label class="text-right text-sm">源类型</Label>
            <Select :model-value="repoConfig.source_type" @update:model-value="(v) => repoConfig.source_type = String(v || 'git')">
              <SelectTrigger class="col-span-3">
                <SelectValue placeholder="选择源类型" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="git">Git 仓库</SelectItem>
                <SelectItem value="url">URL 下载</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div class="grid grid-cols-4 items-center gap-4">
            <Label class="text-right text-sm">源地址</Label>
            <Input v-model="repoConfig.source_url" :placeholder="repoConfig.source_type === 'git' ? 'https://github.com/user/repo.git' : 'https://example.com/file.js'" class="col-span-3 font-mono text-sm" />
          </div>
          <div class="grid grid-cols-4 items-center gap-4">
            <Label class="text-right text-sm">目标路径</Label>
            <div class="col-span-3">
              <DirTreeSelect :model-value="repoConfig.target_path || ''" @update:model-value="v => repoConfig.target_path = v" />
            </div>
          </div>
          <div v-if="repoConfig.source_type === 'git'" class="grid grid-cols-4 items-center gap-4">
            <Label class="text-right text-sm">分支</Label>
            <Input v-model="repoConfig.branch" placeholder="main (可选)" class="col-span-3" autocomplete="off" />
          </div>
          <div v-if="repoConfig.source_type === 'git'" class="grid grid-cols-4 items-center gap-4">
            <Label class="text-right text-sm">稀疏路径</Label>
            <Input v-model="repoConfig.sparse_path" placeholder="仅拉取指定目录或文件 (可选)" class="col-span-3" autocomplete="off" />
          </div>
          <div v-if="repoConfig.source_type === 'git' && repoConfig.sparse_path" class="grid grid-cols-4 items-center gap-4">
            <Label class="text-right text-sm">单文件模式</Label>
            <div class="col-span-3 flex items-center gap-2">
              <Checkbox :checked="repoConfig.single_file" @update:checked="(v: boolean) => repoConfig.single_file = v" />
              <span class="text-sm text-muted-foreground">直接下载文件（适用于单个文件同步）</span>
            </div>
          </div>
          <div class="grid grid-cols-4 items-center gap-4">
            <Label class="text-right text-sm">代理</Label>
            <Select :model-value="repoConfig.proxy" @update:model-value="(v) => repoConfig.proxy = String(v || 'none')">
              <SelectTrigger class="col-span-3">
                <SelectValue placeholder="选择代理" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem v-for="opt in proxyOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div v-if="repoConfig.proxy === 'custom'" class="grid grid-cols-4 items-center gap-4">
            <Label class="text-right text-sm">代理地址</Label>
            <Input v-model="repoConfig.proxy_url" placeholder="https://your-proxy.com/" class="col-span-3 font-mono text-sm" autocomplete="off" />
          </div>
          <div class="grid grid-cols-4 items-center gap-4">
            <Label class="text-right text-sm">认证Token</Label>
            <Input v-model="repoConfig.auth_token" type="text" placeholder="可选，用于私有仓库" class="col-span-3" autocomplete="new-password" />
          </div>
          <div class="grid grid-cols-4 items-center gap-4">
            <Label class="text-right text-sm">定时规则</Label>
            <Input v-model="editingTask.schedule" placeholder="0 0 0 * * *" class="col-span-3 font-mono text-sm" />
          </div>
          <div class="grid grid-cols-4 items-start gap-4">
            <span></span>
            <div class="col-span-3">
              <p class="text-xs text-muted-foreground mb-2">格式: 秒 分 时 日 月 周</p>
              <div class="flex flex-wrap gap-1.5">
                <span
                  v-for="preset in cronPresets"
                  :key="preset.value"
                  class="px-2 py-0.5 text-xs rounded-md bg-muted hover:bg-accent cursor-pointer transition-colors"
                  @click="editingTask.schedule = preset.value"
                >
                  {{ preset.label }}
                </span>
              </div>
            </div>
          </div>
          <div class="grid grid-cols-4 items-center gap-4">
            <Label class="text-right text-sm">超时(分钟)</Label>
            <Input v-model.number="editingTask.timeout" type="number" placeholder="30" class="col-span-3" />
          </div>
          <div class="grid grid-cols-4 items-center gap-4">
            <Label class="text-right text-sm">日志清理</Label>
            <div class="col-span-3 flex gap-2">
              <Select :model-value="cleanType" @update:model-value="(v) => cleanType = String(v || 'none')">
                <SelectTrigger class="w-28">
                  <SelectValue placeholder="不清理" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="none">不清理</SelectItem>
                  <SelectItem value="day">按天数</SelectItem>
                  <SelectItem value="count">按条数</SelectItem>
                </SelectContent>
              </Select>
              <Input v-if="cleanType && cleanType !== 'none'" v-model.number="cleanKeep" type="number" :placeholder="cleanType === 'day' ? '保留天数' : '保留条数'" class="flex-1" />
            </div>
          </div>
        </div>
        <DialogFooter class="pt-4 border-t">
          <Button variant="outline" @click="showRepoDialog = false">取消</Button>
          <Button @click="saveRepoTask">保存</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <AlertDialog v-model:open="showDeleteDialog">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>确认删除</AlertDialogTitle>
          <AlertDialogDescription>确定要删除此任务吗？此操作无法撤销。</AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>取消</AlertDialogCancel>
          <AlertDialogAction class="bg-destructive text-white hover:bg-destructive/90" @click="deleteTask">删除</AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  </div>
</template>
