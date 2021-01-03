<template>
  <div>
    <div class="d-flex mb-3">
      <div
        class="input-group input-group-sm border border-secondary rounded d-flex flex-fill"
        style="width: fit-content"
      >
        <button
          class="btn btn-sm border-right"
          type="button"
          data-toggle="dropdown"
          aria-expanded="false"
          aria-label="Search filter"
        >
          <span style="font-size: smaller">Filters:</span>
        </button>
        <ul class="dropdown-menu">
          <li v-for="(v, k, index) in filterTemplates" v-bind:key="index">
            <a
              class="dropdown-item"
              @click="
                $parent.issue_query = v;
                $parent.page = 0;
                $parent.fetchIssues();
              "
              >{{ k }}
            </a>
          </li>
        </ul>
        <button
          class="btn"
          type="button"
          id="search-submit"
          @click="$parent.fetchIssues()"
          aria-label="Search issues"
        >
          <svg-icon
            class="text-secondary"
            type="mdi"
            :height="21"
            :width="21"
            :path="mdiMagnify"
          />
        </button>
        <input
          type="text"
          class="form-control border-white"
          placeholder="Filter by..."
          v-model="$parent.issue_query"
          aria-label="Search"
          aria-describedby="search-submit"
        />
        <button
          class="btn"
          type="button"
          id="search-clear"
          @click="$parent.issue_query = ''"
          aria-label="Clear issue search"
        >
          <svg-icon
            class="text-secondary"
            type="mdi"
            :height="21"
            :width="21"
            :path="mdiCloseCircleOutline"
          />
        </button>
      </div>
      <router-link
        v-if="!this.$parent.project.settings.archived && this.$parent.elevated"
        :to="'/project/' + this.$route.params.id + '/issue/create'"
        class="ml-3"
      >
        <button class="btn btn-success" type="button">New Issue</button>
      </router-link>
      <div class="ml-3" v-else>
        <button class="btn btn-success" type="button" :disabled="true">
          New Issue
        </button>
      </div>
    </div>
    <div class="table-responsive">
      <div class="text-center my-5" v-if="$parent.issue_error">
        <alert :message="$parent.issue_error" />
      </div>
      <div v-else>
        <table class="table table-borderless table-hover card d-table">
          <thead class="card-header">
            <tr class="d-table-row">
              <th class="p-0 align-middle"></th>
              <th
                colspan="0"
                scope="col"
                class="align-middle"
                style="width: 1px"
              >
                <input
                  class="form-check-input"
                  type="checkbox"
                  value=""
                  id="flexCheckDefault"
                  aria-label="Select all issue"
                  @click="
                    (vm) => {
                      $parent.selectAllIssues(vm.target.checked);
                    }
                  "
                />
              </th>
              <th colspan="6" scope="col" class="settings col-6 align-middle">
                <div
                  v-if="
                    !this.$parent.project.settings.archived &&
                    this.$parent.elevated
                  "
                >
                  <div class="btn-group dropright">
                    <button
                      class="btn btn-secondary btn-sm dropdown-toggle"
                      type="button"
                      data-toggle="dropdown"
                      aria-expanded="false"
                      aria-label="Modify issue"
                    >
                      <svg-icon
                        type="mdi"
                        :height="16"
                        :path="icons[this.marked]"
                      />
                      {{ text[this.marked] }}
                    </button>
                    <ul class="dropdown-menu">
                      <li>
                        <a
                          class="dropdown-item"
                          v-on:click.prevent="marked = 'none'"
                        >
                          <svg-icon
                            type="mdi"
                            :height="16"
                            :path="mdiDotsHorizontal"
                          />
                          Select Action</a
                        >
                      </li>
                      <li>
                        <a
                          class="dropdown-item"
                          v-on:click.prevent="marked = 'assign'"
                        >
                          <svg-icon
                            type="mdi"
                            :height="16"
                            :path="mdiAccountPlus"
                          />
                          Assign user</a
                        >
                      </li>
                      <li>
                        <a
                          class="dropdown-item"
                          v-on:click.prevent="marked = 'deassign'"
                        >
                          <svg-icon
                            type="mdi"
                            :height="16"
                            :path="mdiAccountRemove"
                          />
                          Deassign user</a
                        >
                      </li>
                      <li>
                        <hr class="dropdown-divider" />
                      </li>
                      <li>
                        <a
                          class="dropdown-item"
                          v-on:click.prevent="marked = 'resolved'"
                        >
                          <svg-icon type="mdi" :height="16" :path="mdiTray" />
                          Mark Resolved</a
                        >
                      </li>
                      <li>
                        <a
                          class="dropdown-item"
                          v-on:click.prevent="marked = 'active'"
                        >
                          <svg-icon
                            type="mdi"
                            :height="16"
                            :path="mdiTrayFull"
                          />
                          Mark Active</a
                        >
                      </li>
                      <li>
                        <a
                          class="dropdown-item"
                          v-on:click.prevent="marked = 'open'"
                        >
                          <svg-icon
                            type="mdi"
                            :height="16"
                            :path="mdiTrayAlert"
                          />
                          Mark Open</a
                        >
                      </li>
                      <li>
                        <a
                          class="dropdown-item"
                          v-on:click.prevent="marked = 'invalid'"
                        >
                          <svg-icon
                            type="mdi"
                            :height="16"
                            :path="mdiTrayRemove"
                          />
                          Mark Invalid</a
                        >
                      </li>
                      <li>
                        <hr class="dropdown-divider" />
                      </li>
                      <li>
                        <a
                          class="dropdown-item"
                          v-on:click.prevent="marked = 'lock'"
                        >
                          <svg-icon type="mdi" :height="16" :path="mdiLock" />
                          Lock Comments</a
                        >
                      </li>
                      <li>
                        <a
                          class="dropdown-item"
                          v-on:click.prevent="marked = 'unlock'"
                        >
                          <svg-icon
                            type="mdi"
                            :height="16"
                            :path="mdiLockOpenVariant"
                          />
                          Unlock Comments</a
                        >
                      </li>
                    </ul>
                  </div>
                  <div
                    class="btn-group dropright"
                    v-if="marked == 'assign' || marked == 'deassign'"
                  >
                    <button
                      class="btn btn-secondary btn-sm dropdown-toggle"
                      type="button"
                      data-toggle="dropdown"
                      aria-expanded="false"
                      aria-label="Select user to assign/deassign"
                    >
                      {{ $parent.getUsername(this.assigned, "Nobody") }}
                    </button>
                    <ul class="dropdown-menu">
                      <li>
                        <a
                          class="dropdown-item user-select"
                          v-on:click.prevent="assigned = undefined"
                          >Nobody</a
                        >
                      </li>
                      <li>
                        <input
                          class="m-2"
                          type="text"
                          placeholder="Filter users..."
                          v-model="assigneeFilter"
                        />
                      </li>
                      <li><hr class="dropdown-divider" /></li>
                      <li v-for="user in filterAssignee()" v-bind:key="user.id">
                        <a
                          class="dropdown-item user-select"
                          v-on:click.prevent="assigned = user.id"
                          v-if="!user.integration"
                        >
                          <img
                            v-if="user.avatar"
                            :width="24"
                            :height="24"
                            :src="user.avatar"
                            alt="User profile picture"
                          />
                          {{ user.name }}
                        </a>
                      </li>
                    </ul>
                  </div>
                  <button
                    class="btn btn-outline-secondary btn-sm"
                    aria-label="Execute actions"
                    @click="
                      $parent.execute(
                        marked,
                        $parent.getCheckedIssues(),
                        assigned
                      )
                    "
                    :disabled="
                      marked == 'none' ||
                      ((marked == 'assign' || marked == 'deassign') &&
                        !assigned) ||
                      $parent.getCheckedIssues().length == 0
                    "
                  >
                    <div
                      v-if="$parent.executing"
                      style="width: 0.6rem; height: 0.6rem; vertical-align: sub"
                      class="spinner-border"
                      role="status"
                    >
                      <span class="visually-hidden">Loading...</span>
                    </div>
                    <svg-icon
                      v-else
                      type="mdi"
                      style="width: 1rem; height: 1rem; vertical-align: sub"
                      :path="mdiPlay"
                    />
                  </button>
                </div>
              </th>
              <th colspan="1" scope="col" class="text-center align-middle">
                Status
              </th>
              <th colspan="1" scope="col" class="text-center align-middle">
                Occurrences
              </th>
              <th colspan="1" scope="col" class="text-center align-middle">
                Assignee
              </th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="issue in $parent.issues"
              v-bind:key="issue.id"
              class="list-group-item d-table-row ticket"
              style="height: 70px"
            >
              <th class="p-0">
                <div
                  :style="{ background: statusBackground[issue.type] }"
                ></div>
              </th>
              <th colspan="0" class="align-middle">
                <input
                  class="form-check-input"
                  type="checkbox"
                  v-model="issue.checked"
                  aria-label="Select issue"
                  :id="'checkIssue' + issue.id"
                />
              </th>

              <th class="ticket-status" colspan="6">
                <button
                  class="btn text-dark"
                  style="padding: 0"
                  @click="
                    $parent.elevated
                      ? $parent.starIssue(issue.id, !issue.starred)
                      : 0
                  "
                  aria-label="Star issue"
                >
                  <svg-icon
                    type="mdi"
                    :path="issue.starred ? mdiStar : mdiStarOutline"
                  />
                </button>
                <router-link
                  class="issue-error text-decoration-none"
                  :to="'/project/' + $route.params.id + '/issue/' + issue.id"
                >
                  {{ issue.error }}
                </router-link>
                <span class="issue-function text-secondary">{{
                  issue.function
                }}</span>
                <span class="issue-checkpoint text-secondary">{{
                  issue.checkpoint
                }}</span>
                <span class="issue-description">{{ issue.description }}</span>
                <div class="ticket-footer align-middle">
                  <!-- <span>Welcomer</span> -->
                  Last modified
                  <span>
                    <timeago
                      :datetime="issue.last_modified"
                      :auto-update="60"
                      :includeSeconds="true"
                    />
                  </span>
                  Created
                  <span>
                    <timeago
                      :datetime="issue.created_at"
                      :auto-update="60"
                      :includeSeconds="true"
                    />
                    by
                    {{ $parent.getUsername(issue.created_by_id, "ghost") }}
                    <span
                      class="badge rounded-pill bg-primary"
                      v-if="$parent.getIntegration(issue.created_by_id)"
                      >Integration</span
                    >
                  </span>
                  <svg-icon
                    type="mdi"
                    :height="20"
                    :path="
                      issue.comments_locked
                        ? mdiMessageTextLock
                        : mdiMessageText
                    "
                  />
                  <span>{{ issue.comment_count }}</span>
                  <svg-icon
                    type="mdi"
                    :height="20"
                    :path="mdiXml"
                    v-if="issue.traceback"
                  />
                </div>
              </th>
              <td colspan="1" class="text-center align-middle">
                {{ statusKey[issue.type] }}
              </td>
              <td colspan="1" class="text-center align-middle">
                {{ issue.occurrences }}
              </td>
              <td colspan="1" class="text-center align-middle">
                {{ $parent.getUsername(issue.assignee_id, "Unassigned") }}
              </td>
            </tr>
          </tbody>
        </table>
        <div class="text-center">
          <div
            class="btn"
            @click="
              if ($parent.page > 0) {
                $parent.page = Math.min(
                  Math.max($parent.page - 1, 0),
                  Math.ceil($parent.total_issues / $parent.page_limit) - 1
                );
                $parent.fetchIssues();
              }
            "
          >
            <svg-icon type="mdi" :path="mdiChevronLeft" />
          </div>
          <span
            >Page <b>{{ $parent.page + 1 }}</b> of
            <b>{{
              Math.ceil($parent.total_issues / $parent.page_limit)
            }}</b></span
          >
          <div
            class="btn"
            @click="
              if (
                $parent.page <
                Math.ceil($parent.total_issues / $parent.page_limit) - 1
              ) {
                $parent.page = Math.min(
                  Math.max($parent.page + 1, 0),
                  Math.ceil($parent.total_issues / $parent.page_limit) - 1
                );
                $parent.fetchIssues();
              }
            "
          >
            <svg-icon type="mdi" :path="mdiChevronRight" />
          </div>
        </div>
      </div>
      <div v-if="this.$parent.issues_loading">
        <div class="text-center my-5">
          <div class="spinner-border text-muted mb-2" role="status">
            <span class="visually-hidden">Searching for issues...</span>
          </div>
        </div>
      </div>
      <div v-else>
        <div
          class="text-center my-5"
          v-if="
            $parent.project.active_issues +
              $parent.project.open_issues +
              $parent.project.resolved_issues ==
            0
          "
        >
          <svg-icon
            class="mb-2 text-muted"
            type="mdi"
            :width="64"
            :height="64"
            :path="mdiTrayAlert"
          />
          <h3>Welcome to your issues page.</h3>
          <p class="m-auto my-2 col-10 text-muted">
            Welcome to the issues tab, where you can track current and previous
            issues that have been made. Start off your issues tab by creating
            your first issue (which we hope you should never need to do).
          </p>
        </div>
        <div
          class="text-center my-5"
          v-else-if="$parent.issues.length == 0 && !$parent.issue_error"
        >
          <svg-icon
            class="mb-2 text-muted"
            type="mdi"
            :width="64"
            :height="64"
            :path="mdiSelectSearch"
          />
          <h3>Uh oh, we could not find anything.</h3>
          <p class="m-auto my-2 col-10 text-muted">
            Try broadening your search term or creating a new issue
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import SvgIcon from "@jamescoyle/vue-icon";
import {
  mdiTray,
  mdiTrayFull,
  mdiTrayAlert,
  mdiTrayRemove,
  mdiLock,
  mdiLockOpenVariant,
  mdiClockOutline,
  mdiMessageText,
  mdiXml,
  mdiStar,
  mdiStarOutline,
  mdiPlay,
  mdiSelectSearch,
  mdiAlertCircle,
  mdiMagnify,
  mdiCloseCircleOutline,
  mdiAccountPlus,
  mdiAccountRemove,
  mdiDotsHorizontal,
  mdiMessageTextLock,
  mdiChevronLeft,
  mdiChevronRight,
} from "@mdi/js";

export default {
  components: {
    SvgIcon,
  },
  name: "ProjectIssues",
  data() {
    return {
      marked: "none",
      assigned: undefined,
      assigneeFilter: "",
      icons: {
        none: mdiDotsHorizontal,
        assign: mdiAccountPlus,
        deassign: mdiAccountRemove,
        resolved: mdiTray,
        active: mdiTrayFull,
        open: mdiTrayAlert,
        invalid: mdiTrayRemove,
        lock: mdiLock,
        unlock: mdiLockOpenVariant,
      },

      filterTemplates: {
        "No Assignee": "assignee:no",
        "Assigned to me": "assignee:@me",
        "Active Issues": "is:active",
        "Open Issues": "is:open",
        "Starred Issues": "is:starred",
        Newest: "sort:created_at-desc",
        Oldest: "sort:created_at-asc",
        "Recently updated": "sort:last_modified-desc",
        "Least recently updated": "sort:last_modified-asc",
        "Most occurred": "sort:occurrences-desc",
        "Least occurred": "sort:occurrences-asc",
        "Error Type": "sort:error-asc",
        Checkpoint: "sort:checkpoint-asc",
      },

      statusKey: {
        0: "Active",
        1: "Open",
        2: "Invalid",
        3: "Resolved",
      },

      statusBackground: {
        0: "#0d6efd",
        1: "#dc3545",
        2: "#6c757d",
        3: "#28a745",
      },

      text: {
        none: "Select Action",
        assign: "Assign user",
        deassign: "Deassign user",
        resolved: "Mark resolved",
        active: "Mark active",
        open: "Mark open",
        invalid: "Mark invalid",
        lock: "Lock comments",
        unlock: "Unlock comments",
      },

      mdiAccountPlus: mdiAccountPlus,
      mdiAccountRemove: mdiAccountRemove,
      mdiTray: mdiTray,
      mdiTrayFull: mdiTrayFull,
      mdiTrayAlert: mdiTrayAlert,
      mdiTrayRemove: mdiTrayRemove,
      mdiLock: mdiLock,
      mdiLockOpenVariant: mdiLockOpenVariant,
      mdiClockOutline: mdiClockOutline,
      mdiMessageText: mdiMessageText,
      mdiXml: mdiXml,
      mdiStar: mdiStar,
      mdiStarOutline: mdiStarOutline,
      mdiPlay: mdiPlay,
      mdiSelectSearch: mdiSelectSearch,
      mdiAlertCircle: mdiAlertCircle,
      mdiMagnify: mdiMagnify,
      mdiCloseCircleOutline: mdiCloseCircleOutline,
      mdiDotsHorizontal: mdiDotsHorizontal,
      mdiMessageTextLock: mdiMessageTextLock,
      mdiChevronLeft: mdiChevronLeft,
      mdiChevronRight: mdiChevronRight,
    };
  },
  beforeRouteEnter(to, from, next) {
    next((vm) => {
      // // If it is equal to 1, then we entered from an issue page.
      // if (Object.keys(vm.$parent.issues).length <= 1) {
      //   // vm.$parent.fetchIssues();
      vm.$parent.loadIssues({
        page: vm.$parent.page,
        q: vm.$parent.issue_query,
      });
      //   }
    });
  },
  methods: {
    filterAssignee() {
      var contributors = Object.values(this.$parent.contributors);
      if (this.projectFilter == "") {
        return contributors;
      }
      var filter = this.assigneeFilter.toLowerCase();
      return contributors.filter((object) => {
        return object.name.toLowerCase().includes(filter);
      });
    },
  },
};
</script>

<style scoped>
.dropdown-menu a {
  cursor: pointer;
}
.ticket {
  height: 0;
}
.ticket th:first-child {
  height: inherit;
  vertical-align: middle;
}
.ticket th:first-child div {
  width: 8px;
  height: 90%;
  max-height: 80px;
  border-radius: 0 4px 4px 0;
}

.ticket .ticket-status > * {
  vertical-align: middle;
}
.ticket .ticket-status .issue-error {
  font-weight: normal;
  font-size: x-large;
  color: black;
}
.ticket .ticket-status .issue-error:hover {
  color: var(--bs-primary);
}

.ticket .ticket-status .issue-function {
  display: block;
  font-weight: normal;
  font-size: small;
}
.ticket .ticket-status .issue-checkpoint {
  display: block;
  font-weight: normal;
}
.ticket .ticket-status .issue-description {
  font-weight: 500;
  font-size: larger;
}
.ticket .ticket-status > div > span:not(:last-child)::after {
  content: "•";
  padding: 0 0 0 5px;
}

.ticket .ticket-footer {
  font-weight: 400;
  font-size: smaller;
}
.ticket .ticket-footer > svg {
  margin: 0 2px;
}
.ticket td {
  font-weight: 600;
  font-size: large;
}

.assign-select {
  display: inline-table;
}
.assign-select label:nth-child(2) {
  border-radius: 0.2rem 0 0 0.2rem;
}
.assign-select label:nth-child(4) {
  border-radius: 0 0.2rem 0.2rem 0;
}

.settings div {
  margin: 0.2rem;
}
.settings button {
  margin: 0.2rem 0;
}

.user-select {
  cursor: pointer;
}
</style>
