<script setup lang="ts">
import { Progress } from "@/components/ui/progress";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Trash2 } from "lucide-vue-next";

const props = defineProps({
  download: {
    type: Object,
    required: true,
  },
});

defineEmits(["delete"]);

const statusText = computed(() => {
  switch (props.download.status) {
    case "pending":
      return "Pending";
    case "downloading":
      return "Downloading";
    case "completed":
      return "Completed";
    case "failed":
      return "Failed";
    default:
      return "Unknown";
  }
});

const statusBadgeClass = computed(() => {
  switch (props.download.status) {
    case "pending":
      return "bg-muted text-foreground";
    case "downloading":
      return "bg-info";
    case "completed":
      return "bg-success";
    case "failed":
      return "bg-destructive";
    default:
      return "bg-muted text-foreground";
  }
});

const formatBytes = (bytesPerSecond: number) => {
  const units = ["B", "KB", "MB", "GB"];
  let size = bytesPerSecond;
  let unitIndex = 0;

  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024;
    unitIndex++;
  }

  return `${size.toFixed(1)} ${units[unitIndex]}`;
};

// Methods
const formatDate = (dateString: string) => {
  const date = new Date(dateString);
  return date.toLocaleString();
};
</script>

<template>
  <div class="p-6 bg-secondary rounded-xl w-full sm:max-w-4/5">
    <div class="flex items-center gap-3">
      <!-- Download Info -->
      <div class="flex-1 shrink min-w-0">
        <div class="flex items-center gap-3 lg:mb-2">
          <!-- Filename -->
          <h3 class="text-lg font-medium truncate">
            {{ download.filename }}
          </h3>

          <!-- Status Badge -->
          <Badge :class="statusBadgeClass">
            {{ statusText }}
          </Badge>
        </div>

        <div class="flex flex-col gap-1 lg:flex-row lg:items-center lg:gap-8">
          <!-- URL -->
          <p class="text-sm text-foreground/70 truncate flex-1 grow-3">
            {{ download.url }}
          </p>

          <!-- Progress Bar -->
          <div
            v-if="download.progress > 0 && download.downloaded > 0"
            class="flex items-center gap-4 lg:flex-1 lg:grow-5"
          >
            <Progress
              class="lg:flex-1"
              :model-value="download.progress"
              :title="`${download.progress}% - ${formatBytes(
                download.downloaded,
              )}`"
            />
            <span class="text-foreground/70 text-sm">
              {{ formatBytes(download.speed) }}/s
            </span>
          </div>
        </div>

        <!-- Error Message -->
        <div
          v-if="download.error"
          class="mt-2 p-2 bg-red-200 rounded text-sm text-destructive"
        >
          {{ download.error }}
        </div>

        <!-- Timestamps -->
        <div
          class="flex flex-col md:flex-row gap-1 md:gap-4 text-xs text-foreground/70 mt-2"
        >
          <span>Created: {{ formatDate(download.created_at) }}</span>
          <span v-if="download.updated_at !== download.created_at">
            Updated: {{ formatDate(download.updated_at) }}
          </span>
        </div>
      </div>

      <!-- Actions -->
      <div class="flex-shrink-0">
        <!-- TODO: Add Pause/Resume buttons -->
        <Button
          variant="ghost"
          size="icon"
          class="cursor-pointer"
          title="Delete download"
          @click="$emit('delete', download.id)"
        >
          <Trash2 class="size-5 stroke-foreground" />
        </Button>
      </div>
    </div>
  </div>
</template>
