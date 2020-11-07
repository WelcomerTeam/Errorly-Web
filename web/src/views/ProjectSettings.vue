<template>
  <div v-if="this.error">
    <error :message="this.error" />
  </div>
  <div v-else-if="!this.$parent.elevated" class="text-center my-5">
    <error :message="'You do not have permission to do this'" />
  </div>
  <div v-else>
    <pre>{{ JSON.stringify(project, null, 4) }}</pre>
    <form-input type="text" v-model="project.settings.display_name" />
    <nav>
      <div class="nav nav-tabs" id="nav-tab" role="tablist">
        <a
          class="nav-link active"
          id="v-pills-settings-tab"
          data-toggle="pill"
          href="#v-pills-settings"
          role="tab"
          aria-controls="v-pills-settings"
          aria-selected="true"
        >
          <svg-icon type="mdi" :height="20" :path="mdiCogOutline" />
          Settings
        </a>
        <a
          class="nav-link"
          id="v-pills-contributors-tab"
          data-toggle="pill"
          href="#v-pills-contributors"
          role="tab"
          aria-controls="v-pills-contributors"
          aria-selected="false"
        >
          <svg-icon type="mdi" :height="20" :path="mdiAccountDetails" />
          Contributors
        </a>
        <a
          class="nav-link"
          id="v-pills-integrations-tab"
          data-toggle="pill"
          href="#v-pills-integrations"
          role="tab"
          aria-controls="v-pills-integrations"
          aria-selected="false"
        >
          <svg-icon type="mdi" :height="20" :path="mdiWrench" />
          Integrations
        </a>
        <a
          class="nav-link"
          id="v-pills-webhooks-tab"
          data-toggle="pill"
          href="#v-pills-webhooks"
          role="tab"
          aria-controls="v-pills-webhooks"
          aria-selected="false"
        >
          <svg-icon type="mdi" :height="20" :path="mdiWebhook" />
          Webhooks
        </a>
      </div>
    </nav>
    <div class="tab-content px-5 py-4" id="nav-tabContent">
      <div
        class="tab-pane fade show active"
        id="v-pills-settings"
        role="tabpanel"
        aria-labelledby="v-pills-settings-tab"
      >
        <pre>
          - display name
          - description
          - url

          - private
          - limited

          - archive project
          - delete project
        </pre>
      </div>
      <div
        class="tab-pane fade"
        id="v-pills-contributors"
        role="tabpanel"
        aria-labelledby="v-pills-contributors-tab"
      >
        <pre>
          - invite contributors
            - creates link
            - by discord id
          - manage contributors
            - remove
        </pre>
      </div>
      <div
        class="tab-pane fade"
        id="v-pills-integrations"
        role="tabpanel"
        aria-labelledby="v-pills-integrations-tab"
      >
        <pre>
          - create integration
          - manage integrations
            - integration name
            - get token
            - reset token
            - created by [] [] ago
        </pre>
      </div>
      <div
        class="tab-pane fade"
        id="v-pills-webhooks"
        role="tabpanel"
        aria-labelledby="v-pills-webhooks-tab"
      >
        <pre>
          - create webhooks
          - manage webhooks
              - active: Failing (3)
              - created by [] [] ago
            - reactivate
            - url
            - payload type
            - use json content
            - secret
        </pre>
      </div>
    </div>
  </div>
</template>

<script>
import {
  mdiCogOutline,
  mdiAccountDetails,
  mdiWrench,
  mdiWebhook,
} from "@mdi/js";
import SvgIcon from "@jamescoyle/vue-icon";
import Error from "@/components/Error.vue";
import FormInput from "@/components/FormInput.vue";
export default {
  components: { SvgIcon, Error, FormInput },
  name: "ProjectSettings",
  data() {
    var data = {
      error: undefined,
      project: JSON.parse(JSON.stringify(this.$parent.project)),
      mdiCogOutline: mdiCogOutline,
      mdiAccountDetails: mdiAccountDetails,
      mdiWrench: mdiWrench,
      mdiWebhook: mdiWebhook,
    };
    return data;
  },
};
</script>
