<template>
  <div class="container mt-5">
    <div
      class="border border-danger text-dark rounded-sm p-3 my-4"
      role="alert"
      v-if="this.error"
    >
      <h5 class="font-weight-bold">Error:</h5>
      {{ this.error }}
    </div>
    test page thingy
  </div>
</template>

<script>
export default {
  name: "CreateProjectIssue",
};
</script>
