<script setup>
import { motion } from "motion-v";

const dotVariants = {
  jump: {
    y: -30,
    transition: {
      duration: 0.8,
      repeat: Infinity,
      repeatType: "mirror",
      ease: "easeInOut",
    },
  },
};
</script>

<template>
  <motion.div
    class="flex justify-center items-center gap-[10px]"
    animate="jump"
    :transition="{ staggerChildren: -0.2, staggerDirection: -1 }"
  >
    <motion.div
      class="rounded-[50%] will-change-transform size-2.5 bg-primary"
      :variants="dotVariants"
    />
    <motion.div
      class="rounded-[50%] will-change-transform size-2.5 bg-primary"
      :variants="dotVariants"
    />
    <motion.div
      class="rounded-[50%] will-change-transform size-2.5 bg-primary"
      :variants="dotVariants"
    />
  </motion.div>
</template>
