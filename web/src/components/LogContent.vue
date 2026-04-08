<script setup lang="ts">
import { AlertCircle, Loader2 } from 'lucide-vue-next'
import Ansi from 'ansi-to-vue3'

interface Props {
  content?: string
  loading?: boolean
  loadingText?: string
  emptyTitle?: string
  emptyDescription?: string
}

withDefaults(defineProps<Props>(), {
  content: '',
  loading: false,
  loadingText: '正在获取日志内容',
  emptyTitle: '未检测到输出内容',
  emptyDescription: '此任务执行期间未产生标准输出（Stdout）或错误输出（Stderr）日志。'
})
</script>

<template>
  <div class="flex-1 flex flex-col h-full">
    <!-- 加载状态 -->
    <template v-if="loading">
      <div class="flex-1 flex flex-col items-center justify-center p-4 select-none text-center">
        <Loader2 class="h-10 w-10 animate-spin text-primary/30 mb-4" />
        <span class="text-sm text-muted-foreground font-medium animate-pulse">{{ loadingText }}</span>
      </div>
    </template>

    <!-- 空状态 -->
    <template v-else-if="!content || !content.trim()">
      <div class="flex-1 flex flex-col items-center justify-center p-4 select-none text-center">
        <div class="w-14 h-14 rounded-3xl bg-muted/20 flex items-center justify-center mb-4 border border-muted-foreground/10 mx-auto">
          <AlertCircle class="h-7 w-8 text-muted-foreground/20" />
        </div>
        <span class="text-sm text-muted-foreground font-medium">{{ emptyTitle }}</span>
        <p class="text-[11px] text-muted-foreground/40 mt-1.5 max-w-[280px] leading-relaxed mx-auto">
          {{ emptyDescription }}
        </p>
      </div>
    </template>

    <!-- 正常内容 -->
    <template v-else>
      <div class="p-4 text-xs font-mono whitespace-pre-wrap break-all leading-relaxed">
        <Ansi>{{ content }}</Ansi>
      </div>
    </template>
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
