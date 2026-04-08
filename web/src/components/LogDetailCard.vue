<script setup lang="ts">
import { ref } from 'vue'
import { 
  X, Trash2, Maximize2, CheckCircle2, XCircle, AlertCircle, Clock, Ban, 
  Zap as ZapIcon, Search
} from 'lucide-vue-next'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import LogContent from './LogContent.vue'
import { TASK_STATUS } from '@/constants'
import type { TaskLog } from '@/api'

interface Props {
  log: TaskLog | null
  content: string
  title?: string
  loading?: boolean
  isStopping?: boolean
  showClose?: boolean
  variant?: 'full' | 'simple'
}

const props = withDefaults(defineProps<Props>(), {
  log: null,
  content: '',
  title: '日志详情',
  loading: false,
  isStopping: false,
  showClose: true,
  variant: 'full'
})

defineEmits<{
  'close': []
  'stop': []
  'delete': [id: string]
  'maximize': []
}>()

const searchKeyword = ref('')

function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}毫秒`
  if (ms < 60000) return `${(ms / 1000).toFixed(1)}秒`
  return `${(ms / 60000).toFixed(1)}分钟`
}

function getStatusBadgeClass(status: string) {
  switch (status) {
    case TASK_STATUS.SUCCESS:
      return 'bg-green-500/10 text-green-600 border-green-500/20 dark:bg-green-500/20 dark:text-green-400 dark:border-green-500/30 shadow-[0_0_8px_-2px_rgba(34,197,94,0.15)]'
    case TASK_STATUS.FAILED:
      return 'bg-red-500/10 text-red-600 border-red-500/20 dark:bg-red-500/20 dark:text-red-400 dark:border-red-500/30'
    case TASK_STATUS.RUNNING:
      return 'bg-blue-500/10 text-blue-600 border-blue-500/20 dark:bg-blue-500/20 dark:text-blue-400 dark:border-blue-500/30'
    case TASK_STATUS.PENDING:
      return 'bg-amber-500/10 text-amber-600 border-amber-500/20 dark:bg-amber-500/20 dark:text-amber-400 dark:border-amber-500/30'
    case TASK_STATUS.TIMEOUT:
      return 'bg-orange-500/10 text-orange-600 border-orange-500/20 dark:bg-orange-500/20 dark:text-orange-400 dark:border-orange-500/30'
    case TASK_STATUS.CANCELLED:
      return 'bg-muted/50 text-muted-foreground border-muted-foreground/10'
    default:
      return 'bg-secondary text-secondary-foreground border-transparent'
  }
}
</script>

<template>
  <div v-if="log" class="w-full h-full flex flex-col overflow-hidden bg-card">
    <!-- 头部菜单 -->
    <div class="flex items-center justify-between px-4 h-11 border-b bg-muted/20 shrink-0 gap-4">
      <div class="flex items-center gap-3 min-w-0">
        <span class="text-sm font-normal text-muted-foreground whitespace-nowrap">{{ title }}</span>
        
        <!-- Simple 模式下的状态显示 -->
        <Badge v-if="variant === 'simple'" variant="outline" :class="[
          'capitalize px-2 py-0.5 font-normal rounded-full border text-[10px] hidden sm:flex',
          getStatusBadgeClass(log.status)
        ]">
          {{ log.status }}
        </Badge>

        <Button v-if="log.status === TASK_STATUS.RUNNING" variant="destructive" size="sm"
          class="h-6 px-2 text-[10px]" :disabled="isStopping" @click="$emit('stop')">
          {{ isStopping ? '停止中...' : '停止任务' }}
        </Button>
      </div>

      <div class="flex items-center gap-2">
        <!-- Searchbox for Simple mode -->
        <div v-if="variant === 'simple'" class="relative flex-1 sm:flex-none">
          <Search class="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" />
          <Input v-model="searchKeyword" placeholder="搜索内容..." class="h-7 pl-8 w-full sm:w-48 text-xs bg-muted/30" />
        </div>

        <Button v-if="variant === 'simple'" variant="ghost" size="icon" class="h-7 w-7 text-muted-foreground"
          title="全屏切换" @click="$emit('maximize')">
          <Maximize2 class="h-3.5 w-3.5" />
        </Button>

        <Button variant="ghost" size="icon" class="h-7 w-7 text-muted-foreground hover:text-destructive"
          title="删除该日志" @click="$emit('delete', log.id)">
          <Trash2 class="h-3.5 w-3.5" />
        </Button>
        <Button v-if="showClose" variant="ghost" size="icon" class="h-7 w-7" @click="$emit('close')" title="关闭">
          <X class="h-3.5 w-3.5" />
        </Button>
      </div>
    </div>

    <!-- 任务元数据 (仅在 Full 模式下展示) -->
    <div v-if="variant === 'full'"
      class="px-4 py-3 border-b space-y-2 text-sm text-foreground/80 shrink-0 overflow-y-auto max-h-[40vh]">
      <div class="flex justify-between items-center h-6">
        <span class="text-sm font-normal text-muted-foreground">任务名称</span>
        <span class="text-sm font-normal text-muted-foreground">{{ log.task_name }}</span>
      </div>
      <div class="flex justify-between items-center h-8">
        <span class="text-sm font-normal text-muted-foreground">状态</span>
        <Badge variant="outline" :class="[
          'capitalize px-3 py-1 font-normal rounded-full border shadow-sm transition-all duration-300 ring-4 ring-transparent hover:ring-primary/5',
          getStatusBadgeClass(log.status)
        ]">
          <div class="flex items-center gap-1.5">
            <CheckCircle2 v-if="log.status === TASK_STATUS.SUCCESS" class="h-3.5 w-3.5 fill-green-500/20" />
            <XCircle v-else-if="log.status === TASK_STATUS.FAILED" class="h-3.5 w-3.5 fill-red-500/20" />
            <ZapIcon v-else-if="log.status === TASK_STATUS.RUNNING"
              class="h-3.5 w-3.5 fill-current animate-pulse text-blue-500" />
            <Clock v-else-if="log.status === TASK_STATUS.PENDING" class="h-3.5 w-3.5 fill-amber-500/20" />
            <AlertCircle v-else-if="log.status === TASK_STATUS.TIMEOUT" class="h-3.5 w-3.5 fill-orange-500/20" />
            <Ban v-else-if="log.status === TASK_STATUS.CANCELLED" class="h-3.5 w-3.5" />
            <span class="text-[10px] font-normal uppercase">{{ log.status === TASK_STATUS.SUCCESS ? 'SUCCESS' : log.status }}</span>
          </div>
        </Badge>
      </div>
      <div class="flex justify-between items-center h-6">
        <span class="text-sm font-normal text-muted-foreground">耗时</span>
        <span class="text-sm font-normal text-muted-foreground">{{ formatDuration(log.duration) }}</span>
      </div>
      <div class="flex justify-between items-center h-6">
        <span class="text-sm font-normal text-muted-foreground">开始时间</span>
        <span class="text-sm font-normal text-muted-foreground">{{ log.start_time || '-' }}</span>
      </div>
      <div class="flex justify-between items-center h-6">
        <span class="text-sm font-normal text-muted-foreground">结束时间</span>
        <span class="text-sm font-normal text-muted-foreground">{{ log.end_time || '-' }}</span>
      </div>
      <div class="pt-1.5 pb-1">
        <span class="text-sm font-normal text-muted-foreground block mb-1">执行命令</span>
        <code
          class="block font-mono bg-muted/40 px-3 py-2 rounded text-xs break-all border border-muted-foreground/10 leading-relaxed overflow-y-auto max-h-24 font-normal">
          {{ log.command }}
        </code>
      </div>
    </div>

    <!-- 日志输出容器 -->
    <div class="flex-1 flex flex-col overflow-hidden"
      :class="variant === 'simple' ? 'bg-black/5 dark:bg-white/5' : 'bg-black/[0.02] dark:bg-white/[0.02]'">
      <!-- 错误信息提示 -->
      <div v-if="log.error" class="px-4 py-3 border-b bg-red-500/5 space-y-2 text-sm shrink-0">
        <div class="flex items-center gap-2 text-red-500 font-medium">
          <XCircle class="h-4 w-4" />
          <span class="font-normal">系统错误</span>
        </div>
        <code class="block font-mono bg-red-500/10 text-red-600 px-2 py-1 rounded text-xs break-all">
          {{ log.error }}
        </code>
      </div>
      
      <!-- 日志工具栏 (仅在 Full 模式展示，Simple 模式直接显示内容) -->
      <div v-if="variant === 'full'" 
        class="px-4 py-2.5 text-sm text-muted-foreground border-b bg-muted/20 flex items-center justify-between shrink-0">
        <span class="text-sm font-normal text-muted-foreground text-[12px]">日志输出</span>
        <Button variant="ghost" size="icon" class="h-6 w-6" @click="$emit('maximize')" title="全屏查看">
          <Maximize2 class="h-3.5 w-3.5" />
        </Button>
      </div>

      <!-- 日志列表 -->
      <div class="flex-1 overflow-auto">
        <LogContent 
          class="h-full"
          :content="content" 
          :loading="loading" 
          empty-description="此任务执行期间未产生标准输出日志" 
        />
      </div>
    </div>
  </div>
</template>

<style scoped>
:deep(code) {
  display: block;
  padding: 0 !important;
  margin: 0 !important;
  background: transparent !important;
}

:deep(span) {
  vertical-align: top;
}
</style>
