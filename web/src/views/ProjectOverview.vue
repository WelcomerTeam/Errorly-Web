<template>
  <div>
    <div class="p-3 border-bottom border-muted">
      <h5 class="font-weight-bold">About</h5>
      <p v-if="this.$parent.project.settings.description">
        {{ this.$parent.project.settings.description }}
      </p>
      <p class="text-secondary" v-else>No description provided</p>
      <div v-if="this.$parent.project.settings.url">
        <a
          :href="this.$parent.project.settings.url"
          class="text-decoration-none"
        >
          <svg-icon type="mdi" :height="16" :path="mdiLinkVariant" />
          {{ getHost(this.$parent.project.settings.url) }}
        </a>
      </div>
    </div>
    <!-- <pre>
      :errors line graph
      :created by ... on ...
      :contributors
    </pre> -->
  </div>
</template>

<script>
import SvgIcon from "@jamescoyle/vue-icon";
import { mdiLinkVariant } from "@mdi/js";

export default {
  components: {
    SvgIcon,
  },
  name: "ProjectOverview",
  data() {
    return {
      mdiLinkVariant: mdiLinkVariant,
    };
  },
  methods: {
    getHost(url) {
      var _url = new URL(url);
      return _url.host;
    },
  },
};
</script>
