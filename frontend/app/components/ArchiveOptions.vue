<script setup lang="ts">
import {
  EllipsisVertical,
  Trash2,
  SquareArrowRightIcon,
  RefreshCw,
} from "lucide-vue-next";
import { toast } from "vue-sonner";

const props = defineProps({
  id: String,
  token: String,
});

function saveToFile(id: string | undefined) {
  $fetch(`/api/a/${id}/meta/save`, {
    method: "post",
    onResponseError() {
      toast.error("Failed to save ComicInfo.xml file to archive.");
    },
    onResponse({ response }) {
      if (response.ok) {
        toast.success("Successfully saved ComicInfo.xml file to archive.");
      }
    },
  });
}

function resetReadingProgress(id: string | undefined) {
  $fetch(`/api/a/${id}/rp`, {
    method: "delete",
    onRequest({ options }) {
      options.headers.set("Authorization", `Bearer ${props.token}`);
    },
    onResponseError() {
      toast.error("Failed to reset reading progress.");
    },
    onResponse({ response }) {
      if (response.ok) {
        toast.success("Successfully reset reading progress.");
        refreshNuxtData("archive");
      }
    },
  });
}

function deleteArchive(id: string | undefined) {
  $fetch(`/api/a/${id}`, {
    method: "delete",
    onRequest({ options }) {
      options.headers.set("Authorization", `Bearer ${props.token}`);
    },
    onResponseError() {
      toast.error("Failed to delete archive.");
    },
    onResponse({ response }) {
      if (response.ok) {
        toast.success("Successfully deleted archive.");
        navigateTo({ name: "a" });
      }
    },
  });
}
</script>

<template>
  <DropdownMenu>
    <DropdownMenuTrigger>
      <EllipsisVertical class="size-5 stroke-foreground/70 cursor-pointer" />
    </DropdownMenuTrigger>
    <DropdownMenuContent>
      <DropdownMenuItem @click="saveToFile(props.id)">
        <SquareArrowRightIcon class="stroke-foreground" />
        <p>Export metadata</p>
      </DropdownMenuItem>
      <DropdownMenuItem @click="resetReadingProgress(props.id)">
        <RefreshCw class="stroke-foreground" />
        <p>Reset reading progress</p>
      </DropdownMenuItem>
      <DropdownMenuItem @click="deleteArchive(props.id)">
        <Trash2 class="stroke-destructive" />
        <p class="text-destructive">Delete archive</p>
      </DropdownMenuItem>
    </DropdownMenuContent>
  </DropdownMenu>
</template>
