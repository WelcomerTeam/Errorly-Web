<template>
  <div v-if="this.error">
    <error :message="this.error" />
  </div>
  <div v-else-if="!this.$parent.elevated" class="text-center my-5">
    <error :message="'You do not have permission to do this'" />
  </div>
  <div v-else>
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
        <!-- Settings Tab -->
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

        <form-submit class="mt-2 mb-5" @click="saveProjectSettings()" />

        <button
          v-if="$root.user.id == project.created_by_id"
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
                  v-model="deleteProjectModal.confirm"
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
                    deleteProjectModal.target != deleteProjectModal.confirm
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
          errorly.com/project/{project_id}/join/{join_code}

          {
            code: "...","
            project_id: X,
            created_by_id: ...,
            created_at: ...,
            expires_by: ...,
          }

          /api/project/{project_id}/join/{join_code}  -- join project through code
          /api/project/{project_id}/join/regenerate   -- creates new join link

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
          /api/project/{project_id}/integrations
            - GET  - get integration
            - POST - create integrations
          /api/project/{project_id}/integration/{integration_id} POST   - update integrations
          /api/project/{project_id}/integration/{integration_id}/reset  - resets integration token
          /api/project/{project_id}/integration/{integration_id}/delete

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
          /api/project/{project_id}/webhooks
          - GET  - get webhooks
          - POST - create webhooks
          /api/project/{project_id}/webhooks/{webhook_id} POST   - update webhook
          /api/project/{project_id}/webhooks/{webhook_id}/test   - tests webhook
          /api/project/{project_id}/webhooks/{webhook_id}/delete

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
import SvgIcon from "@jamescoyle/vue-icon";
import {
  mdiCogOutline,
  mdiAccountDetails,
  mdiWrench,
  mdiWebhook,
} from "@mdi/js";
import axios from "axios";
import { Modal } from "bootstrap";
import JSONBig from "json-bigint";
import qs from "qs";

import Error from "@/components/Error.vue";
import FormInput from "@/components/FormInput.vue";
import FormSubmit from "@/components/FormSubmit.vue";

var jsonBig = JSONBig({ storeAsString: true });

export default {
  components: { Error, FormInput, FormSubmit, SvgIcon },
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
    deleteProject() {
      axios
        .post(
          "/api/project/" + this.$route.params.id + "/delete",
          qs.stringify({
            confirm: this.deleteProjectModal.confirm,
          }),
          {
            transformResponse: [(data) => jsonBig.parse(data)],
            headers: {
              "content-type": "application/x-www-form-urlencoded;charset=utf-8",
            },
          }
        )
        .then((result) => {
          var data = result.data;
          if (data.success) {
            this.$bvToast.toast(
              `Project was deleted. Redirecting you to projects...`,
              {
                title: "Successfully Deleted",
                appendToast: true,
              }
            );
            this.$root.fetchMe();
            setTimeout(() => {
              this.$router.push("/projects");
            }, 3000);
          }
        })
        .catch((error) => {
          if (error.response?.data) {
            this.$bvToast.toast(
              error.response.data.error || error.response.data,
              {
                title: "Failed to delete project",
                appendToast: true,
              }
            );
          } else {
            this.$bvToast.toast(error.text || error.toString(), {
              title: "Failed to delete project",
              appendToast: true,
            });
          }
        })
        .finally(() => {
          this.deleteProjectModal._modal.hide();
        });
    },
    saveProjectSettings() {
      axios
        .post(
          "/api/project/" + this.$route.params.id,
          qs.stringify(this.project.settings),
          {
            transformResponse: [(data) => jsonBig.parse(data)],
            headers: {
              "content-type": "application/x-www-form-urlencoded;charset=utf-8",
            },
          }
        )
        .then((result) => {
          var data = result.data;
          if (data.success) {
            var settings = data.data.settings;
            this.project.settings = settings;
            this.$parent.project.settings = settings;
            this.$bvToast.toast(`Project settings have been saved`, {
              title: "Successfully Saved",
              appendToast: true,
            });
            Object.entries(this.$root.userProjects).forEach(([id, project]) => {
              if (project.id == this.project.id) {
                project.name = settings.display_name;
                project.description = settings.description;
                project.private = settings.private;
                project.archived = settings.archived;
                this.$set(this.$root.userProjects, id, project);
              }
            });
          } else {
            this.$bvToast.toast(data.error, {
              title: "Failed to save project settings",
              appendToast: true,
            });
          }
        })
        .catch((error) => {
          if (error.response?.data) {
            this.$bvToast.toast(
              error.response.data.error || error.response.data,
              {
                title: "Failed to save project settings",
                appendToast: true,
              }
            );
          } else {
            this.$bvToast.toast(error.text || error.toString(), {
              title: "Failed to save project settings",
              appendToast: true,
            });
          }
        });
    },
  },
};
</script>
