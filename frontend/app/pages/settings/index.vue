<script lang="ts" setup>
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "@/components/ui/accordion";
useHead({
  title: "Settings",
});
const token = await useAuth().getToken();
</script>

<template>
  <div class="bg-secondary w-9/10 mx-auto px-4 rounded-xl">
    <Accordion type="single" collapsible>
      <AccordionItem value="item-1">
        <AccordionTrigger class="text-md items-center"
          >Source Settings</AccordionTrigger
        >
        <AccordionContent class="grid grid-cols-3">
          <FlaresolverrForm :token="token" />
        </AccordionContent>
      </AccordionItem>
    </Accordion>
  </div>
</template>
