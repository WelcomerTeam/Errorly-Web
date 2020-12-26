<template>
  <div v-if="this.error">
    <error :message="this.error" />
  </div>
  <div v-else-if="!this.$parent.elevated" class="text-center my-5">
    <error :message="'You do not have permission to do this'" />
  </div>
  <div v-else>
    <div
      class="modal fade"
      id="transferOwnershipModal"
      tabindex="-1"
      aria-labelledby="transferOwnershipModalLabel"
      aria-hidden="true"
    >
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title" id="transferOwnershipModal">
              Transfer Project Ownership?
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
              Are you sure you want to transfer the project ownership to
              <b>{{ transferOwnershipModal.target }}</b
              >? This cannot be undone.
            </p>
            <p>
              Confirm by entering <b>{{ transferOwnershipModal.target }}</b>
            </p>
            <form-input
              type="text"
              v-model="transferOwnershipModal.confirm"
              :placeholder="'Type: ' + transferOwnershipModal.target"
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
              @click="transferOwnershipTo(transferOwnershipModal.contributor)"
              :disabled="
                transferOwnershipModal.target != transferOwnershipModal.confirm
              "
            >
              Transfer Ownership
            </button>
          </div>
        </div>
      </div>
    </div>

    <div
      class="modal fade"
      id="removeContributorModal"
      tabindex="-1"
      aria-labelledby="removeContributorModalLabel"
      aria-hidden="true"
    >
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title" id="removeContributorModalModal">
              Remove Contributor?
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
              Are you sure you want to remove
              <b>{{ removeContributorModal.target }}</b> from contributors?
            </p>
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
              @click="removeContributor(removeContributorModal.contributor)"
            >
              Remove Contributor
            </button>
          </div>
        </div>
      </div>
    </div>

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
          id="v-pills-invites-tab"
          data-toggle="pill"
          href="#v-pills-invites"
          role="tab"
          aria-controls="v-pills-invites"
          aria-selected="false"
        >
          <svg-icon type="mdi" :height="20" :path="mdiAccountMultiplePlus" />
          Invites
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
        <div class="d-flex pb-3 border-bottom border-muted">
          <div
            class="input-group input-group-sm border border-secondary rounded mr-3"
          >
            <!-- <button class=" btn" type="button" id="search-submit" aria-label="Search">
              <svg-icon type="mdi" :width="21" :height="21" :path="mdiMagnify" />
            </button> -->
            <input
              type="text"
              class="form-control border-white"
              placeholder="Filter contributors by..."
              aria-label="Search"
              aria-describedby="search-submit"
              v-model="contributorFilter"
            />
            <button
              class="btn"
              type="button"
              id="search-clear"
              aria-label="Empty query"
              @click="contributorFilter = ''"
            >
              <svg-icon
                type="mdi"
                :width="21"
                :height="21"
                :path="mdiCloseCircleOutline"
              />
            </button>
          </div>
        </div>

        <table class="table table-borderless table-hover d-table">
          <tbody>
            <tr
              v-for="(contributor, index) in this.filterContributors"
              v-bind:key="index"
              class="list-group-item d-table-row card text-left py-3 border-bottom border-muted border-top-0 border-left-0 border-right-0"
            >
              <th class="ticket-status" colspan="8">
                <img
                  :src="$parent.getAvatar(contributor)"
                  class="rounded-circle"
                  width="32"
                  height="32"
                  alt="Profile picture"
                />
                <span class="contributor-name ml-2 align-middle">
                  <svg-icon
                    v-if="contributor == project.created_by_id"
                    class="text-warning"
                    type="mdi"
                    width="18"
                    height="18"
                    :path="mdiCrown"
                  />
                  {{ $parent.getUsername(contributor) }}
                </span>
              </th>
              <th class="text-right align-middle" colspan="4">
                <div class="dropdown">
                  <button
                    class="btn text-dark"
                    type="button"
                    id="dropdownMenuButton"
                    data-toggle="dropdown"
                    aria-expanded="false"
                  >
                    <svg-icon
                      type="mdi"
                      width="20"
                      height="20"
                      :path="mdiDotsVertical"
                    />
                  </button>
                  <ul
                    class="dropdown-menu"
                    aria-labelledby="dropdownMenuButton"
                  >
                    <li v-if="!(contributor == $root.user.id)">
                      <a
                        class="dropdown-item user-select-none pe-auto"
                        @click="showRemoveContributorModal(contributor)"
                        >Remove Contributor</a
                      >
                    </li>
                    <li
                      v-if="
                        $root.user.id == project.created_by_id &&
                        !(contributor == project.created_by_id)
                      "
                    >
                      <hr class="dropdown-divider" />
                    </li>
                    <li
                      v-if="
                        $root.user.id == project.created_by_id &&
                        !(contributor == project.created_by_id)
                      "
                    >
                      <a
                        class="dropdown-item text-danger user-select-none pe-auto"
                        @click="showTransferOwnershipModal(contributor)"
                        >Transfer Ownership</a
                      >
                    </li>
                  </ul>
                </div>
              </th>
            </tr>
          </tbody>
        </table>

        <div class="text-center">
          <span class="text-muted">Want to invite new contributors?</span>
          <p>Create an invite code</p>
        </div>
      </div>
      <div
        class="tab-pane fade"
        id="v-pills-invites"
        role="tabpanel"
        aria-labelledby="v-pills-invites-tab"
      >
        <div class="d-flex pb-3 border-bottom border-muted">
          <button
            class="btn btn-success w-100"
            @click="showCreateInviteModal()"
          >
            Create Invite
          </button>
        </div>

        <table class="table table-borderless table-hover d-table">
          <tbody>
            <tr
              v-for="(invite, index) in project.invite_codes"
              v-bind:key="index"
              class="list-group-item d-table-row card text-left py-3 border-bottom border-muted border-top-0 border-left-0 border-right-0"
            >
              <th class="invite" colspan="7">
                <span class="invite-code">{{ invite.code }}</span>
                <span class="invite-footer align-middle">
                  Created by
                  <b>ImRock</b>
                  <timeago
                    :datetime="invite.created_at"
                    :auto-update="60"
                    :includeSeconds="true"
                  />
                  <div
                    v-if="new Date(invite.expires_by) > new Date()"
                    class="d-inline"
                  >
                    expires
                    <timeago
                      :datetime="invite.expires_by"
                      :auto-update="60"
                      :includeSeconds="true"
                    />
                  </div>
                </span>
              </th>
              <th class="text-right align-middle" colspan="1">
                {{ invite.uses }} /
                {{ invite.max_uses == 0 ? "∞" : invite.max_uses }}
              </th>
              <th class="text-right align-middle" colspan="4">
                <button
                  class="btn text-dark"
                  type="button"
                  @click="removeInvite(invite.id)"
                >
                  <svg-icon
                    type="mdi"
                    width="20"
                    height="20"
                    :path="mdiClose"
                  />
                </button>
              </th>
            </tr>
          </tbody>
        </table>
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

<style scoped>
.contributor-name {
  font-weight: normal;
  font-size: large;
}

.invite {
  height: 0;
}
.invite > * {
  vertical-align: middle;
}
.invite .invite-code {
  font-weight: normal;
  font-size: x-large;
  color: black;
}
.invite .invite-footer {
  display: block;
  font-weight: 400;
}
.invite .invite-footer * {
  margin-right: 4px;
}
.invite .invite-footer > svg {
  margin: 0 2px;
}
</style>

<script>
import SvgIcon from "@jamescoyle/vue-icon";
import {
  mdiAccountDetails,
  mdiAccountMultiplePlus,
  mdiCogOutline,
  mdiClose,
  mdiCloseCircleOutline,
  mdiCrown,
  mdiDotsVertical,
  mdiWebhook,
  mdiWrench,
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
      mdiAccountDetails,
      mdiAccountMultiplePlus,
      mdiClose,
      mdiCloseCircleOutline,
      mdiCogOutline,
      mdiCrown,
      mdiDotsVertical,
      mdiWebhook,
      mdiWrench,

      error: undefined,
      project: JSON.parse(JSON.stringify(this.$parent.project)),

      contributorFilter: "",

      deleteProjectModal: {
        _modal: undefined,
        target: this.$parent.project.settings.display_name,
        confirm: "",
      },

      removeContributorModal: {
        _modal: undefined,
        target: undefined,
        confirm: "",
      },

      transferOwnershipModal: {
        _modal: undefined,
        target: undefined,
        confirm: "",
      },
    };
    return data;
  },
  computed: {
    filterContributors() {
      var contributors = []
        .concat(this.project.created_by_id)
        .concat(this.project.settings.contributor_ids);

      if (this.contributorFilter == "") {
        return contributors;
      }
      var filter = this.contributorFilter.toLowerCase();

      return contributors.filter((object) => {
        return this.$parent.getUsername(object).toLowerCase().includes(filter);
      });
    },
  },
  methods: {
    showDeleteProjectModal() {
      this.deleteProjectModal._modal = new Modal(
        document.getElementById("deleteProjectModal")
      );
      this.deleteProjectModal._modal.show();
    },

    showRemoveContributorModal(contributor) {
      this.removeContributorModal._modal = new Modal(
        document.getElementById("removeContributorModal")
      );
      this.removeContributorModal.contributor = contributor;
      this.removeContributorModal.target = this.$parent.getUsername(
        contributor
      );
      this.removeContributorModal._modal.show();
    },

    showTransferOwnershipModal(contributor) {
      this.transferOwnershipModal._modal = new Modal(
        document.getElementById("transferOwnershipModal")
      );
      this.transferOwnershipModal.contributor = contributor;
      this.transferOwnershipModal.target = this.$parent.getUsername(
        contributor
      );
      this.transferOwnershipModal._modal.show();
    },

    showCreateInviteModal() {
      this.showCreateInviteModal._modal = new Modal(
        document.getElementById("createInviteModal")
      );
      this.showCreateInviteModal._modal.show();
    },

    transferOwnershipTo(contributor) {
      axios
        .post(
          "/api/project/" + this.$route.params.id + "/transfer",
          qs.stringify({
            confirm: this.transferOwnershipModal.confirm,
            contributor_id: contributor,
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
              `Project was transferred. Redirecting you to projects...`,
              {
                title: "Successfully Transfered",
                appendToast: true,
              }
            );
            this.$root.fetchMe();
            setTimeout(() => {
              this.$router.push("/projects");
            }, 3000);
          } else {
            this.$bvToast.toast(data.error, {
              title: "Failed to transfer project",
              appendToast: true,
            });
          }
        })
        .catch((error) => {
          if (error.response?.data) {
            this.$bvToast.toast(
              error.response.data.error || error.response.data,
              {
                title: "Failed to transfer project",
                appendToast: true,
              }
            );
          } else {
            this.$bvToast.toast(error.text || error.toString(), {
              title: "Failed to transfer project",
              appendToast: true,
            });
          }
        })
        .finally(() => {
          this.transferOwnershipModal._modal.hide();
        });
    },

    removeContributor(contributor) {
      axios
        .delete(
          "/api/project/" +
            this.$route.params.id +
            "/contributor/" +
            contributor,
          {
            transformResponse: [(data) => jsonBig.parse(data)],
          }
        )
        .then((result) => {
          var data = result.data;
          if (data.success) {
            this.project.settings.contributor_ids = data.data.ids;

            Object.values(data.data.users).forEach((user) => {
              // We will use $set as this overcomes a Vue limitation
              // where adding new properties to an object will not
              // trigger changes.
              this.$set(this.$parent.contributors, user.id, user);
            });
          } else {
            this.$bvToast.toast(data.error, {
              title: "Failed to remove contributor",
              appendToast: true,
            });
          }
        })
        .catch((error) => {
          if (error.response?.data) {
            this.$bvToast.toast(
              error.response.data.error || error.response.data,
              {
                title: "Failed to remove contributor",
                appendToast: true,
              }
            );
          } else {
            this.$bvToast.toast(error.text || error.toString(), {
              title: "Failed to remove contributor",
              appendToast: true,
            });
          }
        })
        .finally(() => {
          this.removeContributorModal._modal.hide();
        });
    },

    removeInvite(id) {
      axios
        .delete(
          "/api/project/" + this.$route.params.id + "/invite/" + id + "/delete",
          {
            transformResponse: [(data) => jsonBig.parse(data)],
          }
        )
        .then((result) => {
          var data = result.data;
          if (data.success) {
            this.$bvToast.toast(
              `Removed invite`,
              {
                title: "Removed invite",
                appendToast: true,
              }
            )

            var invites = [];
            this.project.invite_codes.forEach(invite => {
              if (invite.id !== id) {
                invites.push(invite);
              }
            });
            this.project.invite_codes = invites;
          }
        })
        .catch((error) => {
          if (error.response?.data) {
            this.$bvToast.toast(
              error.response.data.error || error.response.data,
              {
                title: "Failed to delete invite",
                appendToast: true,
              }
            );
          } else {
            this.$bvToast.toast(error.text || error.toString(), {
              title: "Failed to delete invite",
              appendToast: true,
            });
          }
        })
    },

    createInvite() {
      axios
        .post(
          "/api/project/" + this.$route.params.id + "/invite",
          qs.stringify({
            uses: this.showCreateInviteModal.uses,
            expiration: this.showCreateInviteModal.expiration,
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
              `Invite a user with the code ${data.data.code}`,
              {
                title: "Successfully created invite",
                appendToast: true,
              }
            );
          }
        })
        .catch((error) => {
          if (error.response?.data) {
            this.$bvToast.toast(
              error.response.data.error || error.response.data,
              {
                title: "Failed to create invite",
                appendToast: true,
              }
            );
          } else {
            this.$bvToast.toast(error.text || error.toString(), {
              title: "Failed to create invite",
              appendToast: true,
            });
          }
        })
        .finally(() => {
          this.showCreateInviteModal._modal.hide();
        });
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
