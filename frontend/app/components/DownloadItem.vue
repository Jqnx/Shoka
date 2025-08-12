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
        return "bg-gray-700 text-slate-300";
      case "downloading":
        return "bg-chart-2";
      case "completed":
        return "bg-primary";
      case "failed":
        return "bg-destructive";
      default:
        return "bg-gray-700 text-slate-300";
    }
  });

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
        <div class="flex items-center gap-3 md:mb-2">
          <!-- Filename -->
          <h3 class="text-lg font-medium truncate">
            {{ download.filename }}
          </h3>

          <!-- Status Badge -->
          <Badge :class="statusBadgeClass">
            {{ statusText }}
          </Badge>
        </div>

        <div class="flex flex-col gap-1 md:flex-row md:items-center md:gap-0">
          <!-- URL -->
          <p class="text-sm text-gray-500 truncate flex-1">
            {{ download.url }}
          </p>

          <!-- Progress Bar -->
          <Progress
            v-if="download.progress != 0"
            class="md:flex-1"
            :model-value="download.progress"
            :title="`${download.progress}%`" />
        </div>

        <!-- Error Message -->
        <div
          v-if="download.error"
          class="mt-2 p-2 bg-red-200 rounded text-sm text-destructive">
          {{ download.error }}
        </div>

        <!-- Timestamps -->
        <div
          class="flex flex-col md:flex-row gap-1 md:gap-4 text-xs text-gray-400 mt-2">
          <span>Created: {{ formatDate(download.created_at) }}</span>
          <span v-if="download.updated_at !== download.created_at">
            Updated: {{ formatDate(download.updated_at) }}
          </span>
        </div>
      </div>

      <!-- Actions -->
      <div class="flex-shrink-0">
        <Button
          variant="default"
          size="icon"
          class="bg-transparent cursor-pointer hover:bg-gray-700"
          title="Delete download"
          @click="$emit('delete', download.id)">
          <Trash2 class="size-5" />
        </Button>
      </div>
    </div>
  </div>
</template>
