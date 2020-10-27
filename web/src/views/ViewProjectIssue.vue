<template>
  <div class="container mt-2 p-4">
    <a
      class="text-dark mb-2 text-decoration-none h6 d-flex"
      @click="$router.go(-1)"
      href="#"
    >
      <svg-icon class="mr-3" type="mdi" :path="mdiChevronLeft" />
      Back to issues
    </a>
    issue: {{ this.$route.params.id }}
  </div>
</template>

<script>
import SvgIcon from "@jamescoyle/vue-icon";
import { mdiChevronLeft } from "@mdi/js";
export default {
  components: { SvgIcon },
  name: "ViewProjectIssue",
  data() {
    return {
      mdiChevronLeft: mdiChevronLeft,
    };
  },
};
</script>
