<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { TASK_STATUS } from '@/constants'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import Pagination from '@/components/Pagination.vue'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import {
  RefreshCw, X, Search, Trash2, Maximize2
} from 'lucide-vue-next'
import LogViewer from './LogViewer.vue'
import { api, type TaskLog } from '@/api'
import { Badge } from '@/components/ui/badge'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { toast } from 'vue-sonner'
import { useSiteSettings } from '@/composables/useSiteSettings'

const route = useRoute()
const { pageSize } = useSiteSettings()

const logs = ref<TaskLog[]>([])
const selectedLog = ref<TaskLog | null>(null)
const filterKeyword = ref('')
const filterTaskId = ref<string | undefined>(undefined)
const filterStatus = ref<string | undefined>(undefined)
const currentPage = ref(1)
const total = ref(0)

const showDeleteDialog = ref(false)
const deleteLogId = ref<string | null>(null)
const showFullscreen = ref(false)
const showClearDialog = ref(false)

let searchTimer: ReturnType<typeof setTimeout> | null = null

async function loadLogs() {
  try {
    const params: { page: number; page_size: number; task_id?: string; task_name?: string; status?: string } = {
      page: currentPage.value,
      page_size: pageSize.value
    }
    if (filterTaskId.value) {
      params.task_id = filterTaskId.value
    }
    if (filterKeyword.value.trim()) {
      params.task_name = filterKeyword.value.trim()
    }
    if (filterStatus.value && filterStatus.value !== 'all') {
      params.status = filterStatus.value
    }
    const response = await api.logs.list(params)
    logs.value = response.data
    total.value = response.total
  } catch {
    toast.error('加载日志失败')
  }
}

function handleSearch() {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    currentPage.value = 1
    loadLogs()
  }, 300)
}

function handleStatusChange() {
  currentPage.value = 1
  loadLogs()
}

function handlePageChange(page: number) {
  currentPage.value = page
  loadLogs()
}

async function selectLog(log: TaskLog) {
  selectedLog.value = log
}

function closeDetail() {
  selectedLog.value = null
}

function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}毫秒`
  if (ms < 60000) return `${(ms / 1000).toFixed(1)}秒`
  return `${(ms / 60000).toFixed(1)}分钟`
}

async function handleClearLogs() {
  try {
    await api.logs.clear(filterTaskId.value)
    toast.success('相关日志已清空')
    showClearDialog.value = false
    loadLogs()
  } catch (error: any) {
    toast.error(error.message || '清空失败')
  }
}

function confirmDeleteLog(id: string) {
  deleteLogId.value = id
  showDeleteDialog.value = true
}

async function handleDeleteLog() {
  if (!deleteLogId.value) return
  try {
    await api.logs.delete(deleteLogId.value)
    toast.success('该日志已删除')

    if (selectedLog.value?.id === deleteLogId.value) {
      closeDetail()
    }

    showDeleteDialog.value = false
    loadLogs()
  } catch (err: any) {
    toast.error(err.message || '删除失败')
  }
}

function getStatusBadgeClass(status: string) {
  switch (status) {
    case TASK_STATUS.SUCCESS:
      return 'bg-green-500/10 text-green-700 border-green-200/50 dark:bg-green-500/20 dark:text-green-400 dark:border-green-900/50'
    case TASK_STATUS.FAILED:
      return 'bg-red-500/10 text-red-700 border-red-200/50 dark:bg-red-500/20 dark:text-red-400 dark:border-red-900/50'
    case TASK_STATUS.RUNNING:
      return 'bg-blue-500/10 text-blue-700 border-blue-200/50 dark:bg-blue-500/20 dark:text-blue-400 dark:border-blue-900/50'
    case TASK_STATUS.PENDING:
      return 'bg-amber-500/10 text-amber-700 border-amber-200/50 dark:bg-amber-500/20 dark:text-amber-400 dark:border-amber-900/50'
    case TASK_STATUS.TIMEOUT:
      return 'bg-orange-500/10 text-orange-700 border-orange-200/50 dark:bg-orange-500/20 dark:text-orange-400 dark:border-orange-900/50'
    case TASK_STATUS.CANCELLED:
      return 'bg-muted text-muted-foreground border-transparent'
    default:
      return 'bg-secondary text-secondary-foreground border-transparent'
  }
}

onMounted(() => {
  const taskIdParam = route.query.task_id
  if (taskIdParam) {
    filterTaskId.value = String(taskIdParam)
  }
  const statusParam = route.query.status
  if (statusParam) {
    filterStatus.value = String(statusParam)
  }
  loadLogs()
})
</script>

<template>
  <div class="flex flex-col gap-4 h-full">
    <!-- 头部工具栏 -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 shrink-0 px-1">
      <div class="flex items-center gap-3">
        <h2 class="text-xl sm:text-2xl font-bold tracking-tight">执行历史</h2>
      </div>

      <div class="flex flex-col sm:flex-row gap-2.5 w-full md:w-auto">
        <div class="flex items-center gap-2 w-full sm:w-auto">
          <div class="relative flex-1 sm:flex-none">
            <Search class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input v-model="filterKeyword" placeholder="搜索任务名称..." class="h-9 pl-9 w-full sm:w-48 text-sm"
              @input="handleSearch" />
          </div>
          <Select v-model="filterStatus" @update:model-value="handleStatusChange">
            <SelectTrigger class="h-9 w-[110px] text-sm shrink-0">
              <SelectValue placeholder="所有状态" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">所有状态</SelectItem>
              <SelectItem :value="TASK_STATUS.SUCCESS">成功</SelectItem>
              <SelectItem :value="TASK_STATUS.FAILED">失败</SelectItem>
              <SelectItem :value="TASK_STATUS.RUNNING">运行中</SelectItem>
              <SelectItem :value="TASK_STATUS.PENDING">排队中</SelectItem>
              <SelectItem :value="TASK_STATUS.TIMEOUT">超时</SelectItem>
              <SelectItem :value="TASK_STATUS.CANCELLED">已取消</SelectItem>
            </SelectContent>
          </Select>
        </div>

        <div class="flex items-center gap-2 shrink-0">
          <Button variant="outline" size="sm" class="h-9 gap-2 shadow-sm text-destructive border-destructive/20 hover:bg-destructive/10" @click="showClearDialog = true">
            <Trash2 class="h-4 w-4" />
            <span>清空日志</span>
          </Button>
          <Button variant="outline" size="icon" class="h-9 w-9 shadow-sm" @click="loadLogs" title="刷新">
            <RefreshCw class="h-4 w-4" />
          </Button>
        </div>
      </div>
    </div>

    <!-- 主体区域 -->
    <div class="flex-1 flex flex-col lg:flex-row gap-4 min-h-0">
      <!-- 日志列表 -->
      <div class="flex-1 min-w-0 rounded-lg border bg-card overflow-hidden flex flex-col">
        <div class="divide-y flex-1 overflow-y-auto">
          <div v-if="logs.length === 0" class="text-sm text-muted-foreground text-center py-8">
            暂无日志
          </div>
          <div v-for="(log, index) in logs" :key="log.id" :class="[
            'cursor-pointer hover:bg-muted/30 transition-colors group',
            selectedLog?.id === log.id && 'bg-accent/50'
          ]" @click="selectLog(log)">
            <div class="flex items-center gap-4 px-4 py-3">
              <span class="w-16 shrink-0 text-muted-foreground text-sm">#{{ total - (currentPage - 1) * pageSize - index }}</span>
              <span class="w-36 shrink-0 font-medium truncate text-sm">{{ log.task_name }}</span>
              <code class="flex-1 min-w-0 text-muted-foreground truncate text-xs bg-muted/40 px-2 py-1 rounded">
                {{ log.command }}
              </code>
              <Badge variant="outline" :class="getStatusBadgeClass(log.status)" class="shrink-0">
                {{ log.status }}
              </Badge>
              <span class="w-16 text-right shrink-0 text-muted-foreground text-xs">{{ formatDuration(log.duration) }}</span>
              <Button variant="ghost" size="icon" class="h-8 w-8 text-muted-foreground hover:text-destructive" @click.stop="confirmDeleteLog(log.id)">
                <Trash2 class="h-4 w-4" />
              </Button>
            </div>
          </div>
        </div>
        <Pagination :total="total" :page="currentPage" @update:page="handlePageChange" class="p-4 border-t" />
      </div>

      <!-- 日志详情侧边栏 -->
      <div v-if="selectedLog"
        class="w-full lg:w-[480px] rounded-lg border bg-card flex flex-col overflow-hidden shrink-0 max-h-[80vh] lg:max-h-none">
        <div class="flex items-center justify-between px-4 h-11 border-b bg-muted/20">
          <span class="text-sm font-medium">日志详情</span>
          <div class="flex items-center gap-1">
            <Button variant="ghost" size="icon" @click="showFullscreen = true" title="全屏查看">
              <Maximize2 class="h-4 w-4" />
            </Button>
            <Button variant="ghost" size="icon" @click="closeDetail">
              <X class="h-4 w-4" />
            </Button>
          </div>
        </div>
        <div class="flex-1 overflow-y-auto p-4 space-y-4">
           <div class="grid grid-cols-2 gap-y-3 text-sm">
             <span class="text-muted-foreground">任务名称</span>
             <span class="text-right font-medium">{{ selectedLog.task_name }}</span>
             <span class="text-muted-foreground">执行状态</span>
             <Badge :class="getStatusBadgeClass(selectedLog.status)" class="ml-auto">{{ selectedLog.status }}</Badge>
             <span class="text-muted-foreground">执行耗时</span>
             <span class="text-right">{{ formatDuration(selectedLog.duration) }}</span>
           </div>
           <div class="border-t pt-4">
             <span class="text-xs font-semibold uppercase text-muted-foreground block mb-2">执行命令</span>
             <code class="block p-2 bg-muted rounded text-xs break-all font-mono">{{ selectedLog.command }}</code>
           </div>
           <div class="border-t pt-4 flex flex-col items-center justify-center py-12 text-muted-foreground text-xs bg-muted/10 rounded">
             <p>侧边栏仅展示详情</p>
             <p class="mt-1 opacity-70">查看完整日志请点击右上角全屏按钮</p>
           </div>
        </div>
      </div>
    </div>

    <!-- 全屏查看日志 -->
    <LogViewer v-model:open="showFullscreen" :task-name="selectedLog?.task_name"
      :log-id="selectedLog?.id" :initial-status="selectedLog?.status" />

    <!-- 弹窗 (清空/删除) -->
    <AlertDialog :open="showClearDialog" @update:open="showClearDialog = $event">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>确认清空日志?</AlertDialogTitle>
          <AlertDialogDescription>此操作不可撤销。</AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>取消</AlertDialogCancel>
          <AlertDialogAction @click="handleClearLogs" variant="destructive">清空</AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>

    <AlertDialog :open="showDeleteDialog" @update:open="showDeleteDialog = $event">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>确认删除日志?</AlertDialogTitle>
          <AlertDialogDescription>数据将永久删除。</AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>取消</AlertDialogCancel>
          <AlertDialogAction @click="handleDeleteLog" variant="destructive">删除</AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  </div>
</template>
