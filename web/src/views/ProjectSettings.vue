<template>
  <div v-if="this.error">
    <error :message="this.error" />
  </div>
  <div v-else-if="!this.$parent.elevated" class="text-center my-5">
    <error :message="'You do not have permission to do this'" />
  </div>
  <div v-else>
    <toast-stack ref="toastStack" />

    <pre>{{ JSON.stringify(project, null, 4) }}</pre>

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
        <form-input
          class="mb-2"
          type="text"
          v-model="project.settings.display_name"
          label="Project name"
          placeholder="Enter a project name"
        />
        <form-input
          class="mb-2"
          type="area"
          v-model="project.settings.description"
          label="Description"
          placeholder="Enter a description for your project"
        />

        <form-input
          type="text"
          v-model="project.settings.url"
          label="Project URL"
          placeholder="Enter a website for your project"
        />
        <p class="text-muted mb-4">
          Enter a URL for your project if you have a related project. Don't have
          a website? Why not include a link to the repository if you have one
        </p>

        <form-input
          type="checkbox"
          v-model="project.settings.private"
          label="Private Project"
        />
        <p class="text-muted">
          If your project is private, the project is only viewable by
          contributors.
        </p>
        <form-input
          type="checkbox"
          v-model="project.settings.limited"
          label="Limited Project"
        />
        <p class="text-muted mb-4">
          If your project is limited, only contributors are able to interact
          with issues or create new ones.
        </p>

        <form-input
          type="checkbox"
          v-model="project.settings.archived"
          label="Archived Project"
        />
        <p class="text-muted">
          If your project is archived, new issues cannot be made.
        </p>

        <form-submit class="mt-2 mb-5" @click="saveProjectSettings(project)" />

        <button
          class="btn btn-outline-dark"
          @click="
            $refs.toastStack.addToast({
              title: 'Hello there',
              body: 'Hello :D',
            })
          "
        >
          Create toast
        </button>
        <button
          class="btn btn-outline-dark"
          @click="
            $refs.toastStack.addToast({
              body: 'Hello :D',
              onlyBody: true,
              class: 'text-white bg-primary border-0',
              closeClass: 'btn-close-white',
            })
          "
        >
          Create coloured toast
        </button>

        <button
          class="btn btn-outline-danger w-100"
          @click="showDeleteProjectModal()"
        >
          Delete Project
        </button>
        <div
          class="modal fade"
          id="deleteProjectModal"
          tabindex="-1"
          aria-labelledby="deleteProjectModalLabel"
          aria-hidden="true"
        >
          <div class="modal-dialog">
            <div class="modal-content">
              <div class="modal-header">
                <h5 class="modal-title" id="deleteProjectModal">
                  Delete Project?
                </h5>
                <button
                  type="button"
                  class="btn-close"
                  data-dismiss="modal"
                  aria-label="Close"
                ></button>
              </div>
              <div class="modal-body">
                <p>
                  Are you sure you want to delete the project? You will lose
                  everything and this cannot be undone.
                </p>
                <p>
                  Confirm by entering <b>{{ deleteProjectModal.target }}</b>
                </p>
                <form-input
                  type="text"
                  v-model="deleteProjectModal.confirmation"
                  :placeholder="'Type: ' + deleteProjectModal.target"
                />
              </div>
              <div class="modal-footer">
                <button
                  type="button"
                  class="btn btn-secondary"
                  data-dismiss="modal"
                >
                  Close
                </button>
                <button
                  type="button"
                  class="btn btn-danger"
                  @click="deleteProject()"
                  :disabled="
                    deleteProjectModal.target != deleteProjectModal.confirmation
                  "
                >
                  Delete Project
                </button>
              </div>
            </div>
          </div>
        </div>
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
import { Modal } from "bootstrap";
import SvgIcon from "@jamescoyle/vue-icon";

import Error from "@/components/Error.vue";
import FormInput from "@/components/FormInput.vue";
import FormSubmit from "@/components/FormSubmit.vue";
import ToastStack from "@/components/ToastStack.vue";

export default {
  components: { Error, FormInput, FormSubmit, SvgIcon, ToastStack },
  name: "ProjectSettings",
  data() {
    var data = {
      error: undefined,
      project: JSON.parse(JSON.stringify(this.$parent.project)),
      mdiCogOutline: mdiCogOutline,
      mdiAccountDetails: mdiAccountDetails,
      mdiWrench: mdiWrench,
      mdiWebhook: mdiWebhook,

      deleteProjectModal: {
        _modal: undefined,
        target: this.$parent.project.settings.display_name,
        confirm: "",
      },
    };
    return data;
  },
  methods: {
    showDeleteProjectModal() {
      this.deleteProjectModal._modal = new Modal(
        document.getElementById("deleteProjectModal")
      );
      this.deleteProjectModal._modal.show();
    },
    deleteProject() {},
    saveProjectSettings() {},
  },
};
</script>
