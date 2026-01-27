<script setup lang="ts">
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { toast } from "vue-sonner";

const id = useRoute().params.id;
const { data: archive } = useNuxtData("archive");

const props = defineProps(["source"]);
const sourceKey = computed(() => `newMeta-${props.source.value}`);

const {
  data: newMeta,
  execute,
  status,
} = useLazyAsyncData(
  sourceKey,
  () =>
    $fetch(`/api/a/${id}/meta/search`, {
      method: "post",
      body: {
        source: props.source.value,
        title: archive.value.title,
      },
      onResponseError({ response }) {
        toast.error(h("pre", "Error: " + response._data.message));
        open.value = false;
      },
    }),
  {
    immediate: false,
  },
);

const emit = defineEmits(["save"]);

const open = ref(false);

function continueEmit(values: any) {
  emit("save", values);
}
</script>

<template>
  <Dialog v-model:open="open">
    <DialogTrigger as-child>
      <Button variant="ghost" class="w-full" @click="execute">{{
        source.label
      }}</Button>
    </DialogTrigger>
    <DialogContent class="sm:max-w-3xl max-h-[90dvh] overflow-y-auto">
      <DialogHeader>
        <DialogTitle>Update Metadata</DialogTitle>
        <DialogDescription />
      </DialogHeader>

      <div v-if="status !== 'success'" class="mx-auto">
        <LoadingDots />
      </div>
      <UpdateMetadataForm
        v-else
        :new="newMeta"
        @save="continueEmit"
        @close="open = false"
      />
    </DialogContent>
  </Dialog>
</template>
