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

    <div v-if="this.issue_loading">Issue</div>
    <div class="text-center my-5" v-else-if="this.issue_error">
      <svg-icon
        class="mb-2 text-muted"
        type="mdi"
        :width="64"
        :height="64"
        :path="mdiAlertCircle"
      />
      <h3>Oops, something happened.</h3>
      <p class="m-auto my-2 col-10 text-muted">
        Something unexpected happened whilst handling your request, sorry.
      </p>
      <kbd>{{ issue_error }}</kbd>
    </div>
    <div v-else>
      <div class="p-4 border-bottom border-muted">
        <h2 class="display-6 d-flex flex-wrap">
          <span class="mr-2">{{ issue.error }}</span>
          <span
            class="text-muted mr-2 h4 font-weight-normal my-auto"
            aria-label="Error function"
            >{{ issue.function }}</span
          >
          <span
            class="text-muted h5 font-weight-normal my-auto"
            aria-label="Error location"
            >{{ issue.checkpoint }}</span
          >
        </h2>
        <span class="badge rounded-pill bg-primary" aria-label="Error type">
          <svg-icon
            type="mdi"
            width="20"
            height="20"
            :path="statusIcon[issue.type]"
          />
          {{ statusText[issue.type] }}
        </span>
        <span class="dot-right pl-2">
          <span class="pl-1">
            <span
              class="badge rounded-pill bg-primary"
              v-if="$parent.getIntegration(issue.created_by_id)"
              >Integration</span
            >
            <b>{{ $parent.getUsername(issue.created_by_id) || "ghost" }}</b>
            opened this issue
            <b>
              <timeago
                :datetime="issue.created_at"
                :auto-update="60"
                :includeSeconds="true"
              />
            </b>
          </span>
          <span class="pl-1">
            Last modified
            <b>
              <timeago
                :datetime="issue.last_modified"
                :auto-update="60"
                :includeSeconds="true"
              />
            </b>
          </span>
          <span class="pl-1">
            <b>{{ issue.occurrences }}</b>
            Occurrences
          </span>
          <span class="pl-1">
            Assigned to
            <b>{{ $parent.getUsername(issue.assignee_id) || "ghost" }}</b>
          </span>
        </span>
      </div>

      <div class="p-4 border-bottom border-muted">
        <div class="d-flex mt-4">
          <!-- <img
            width="40"
            height="40"
            src="https://cdn.discordapp.com/avatars/143090142360371200/a_70444022ea3e5d73dd00d59c5578b07e.png"
          /> -->
          <div class="card ml-3" style="align-items: stretch; width: 100%">
            <div class="card-header text-black-50">

            <span
              class="badge rounded-pill bg-primary"
              v-if="$parent.getIntegration(issue.created_by_id)"
              >Integration</span
            >
            <b class="text-dark">{{ $parent.getUsername(issue.created_by_id) || "ghost" }}</b>
              created this issue
              <timeago
                :datetime="issue.created_at"
                :auto-update="60"
                :includeSeconds="true"
              />
            </div>
            <div class="card-body">
              {{ issue.description }}
              <pre
                class="bg-light mt-3 mb-0 rounded-lg p-3 border border-muted"
                v-if="issue.traceback"
              >{{ issue.traceback }}
              </pre>
            </div>
          </div>
        </div>

        <div class="d-flex mt-4" v-for="(comment, index) in comments" v-bind:key="index">
          <pre>{{ comment }}</pre>
          <!-- <img
            width="40"
            height="40"
            src="https://cdn.discordapp.com/avatars/143090142360371200/a_70444022ea3e5d73dd00d59c5578b07e.png"
          />
          <div class="card ml-3" style="align-items: stretch; width: 100%">
            <div class="card-header text-black-50">
              <b class="text-dark">ImRock</b> commented 3 days ago
            </div>
            <div class="card-body">This is a message</div>
          </div> -->
        </div>

        <!-- <div class="d-flex mt-4">
          <img
            width="40"
            height="40"
            src="https://cdn.discordapp.com/avatars/143090142360371200/a_70444022ea3e5d73dd00d59c5578b07e.png"
          />
          <div class="card ml-3" style="align-items: stretch; width: 100%">
            <div class="card-header text-black-50">
              <b class="text-dark">ImRock</b> commented 3 days ago
            </div>
            <div class="card-body">This is a message</div>
          </div>
        </div>

        <div class="d-flex ml-5 mt-4">
          <svg-icon
            width="40"
            height="40"
            type="mdi"
            :path="mdiTrayRemove"
            class="ml-4 mr-2"
          />
          <div
            class="align-self-middle ml-2 d-flex"
            style="align-items: stretch; width: 100%"
          >
            <div class="text-dark my-auto">
              Issue marked <b>invalid</b> 3 days ago by <b>ImRock</b>
            </div>
          </div>
        </div>

        <div class="d-flex mt-4">
          <img
            width="40"
            height="40"
            src="https://cdn.discordapp.com/avatars/143090142360371200/a_70444022ea3e5d73dd00d59c5578b07e.png"
          />
          <div class="card ml-3" style="align-items: stretch; width: 100%">
            <div class="card-header text-black-50">
              <b class="text-dark">ImRock</b> commented 3 days ago
            </div>
            <div class="card-body">This is a message</div>
          </div>
        </div> -->
      </div>
      <div class="p-4">
        <form-input
          v-model="comment"
          :type="'area'"
          :placeholder="'Leave a comment'"
          class="mb-2"
        />
        <div class="d-flex">
          <button
            type="button"
            class="btn btn-success mr-2"
            @click="sendComment()"
          >
            Comment
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import axios from "axios";
import FormInput from "@/components/FormInput.vue";
import SvgIcon from "@jamescoyle/vue-icon";
import { mdiChevronLeft } from "@mdi/js";
import JSONBig from "json-bigint";
import {
  mdiAlertCircle,
  mdiTray,
  mdiTrayFull,
  mdiTrayAlert,
  mdiTrayRemove,
} from "@mdi/js";
var jsonBig = JSONBig({ storeAsString: true });

function getIssue(projectID, issueID, callback) {
  axios
    .get("/api/project/" + projectID + "/issue/" + issueID, {
      transformResponse: [(data) => jsonBig.parse(data)],
    })
    .then((result) => {
      var data = result.data;
      if (data.success) {
        callback(undefined, data.data);
      } else {
        callback(data.error, {});
      }
    })
    .catch((error) => {
      if (error.response?.data) {
        callback(error.response.data.error || error.response.data, {});
      } else {
        callback(error, {});
      }
    });
}

export default {
  components: { SvgIcon, FormInput },
  name: "ViewProjectIssue",
  data() {
    return {
      issue: undefined,
      issue_error: undefined,
      issue_loading: true,

      comment: "",
      comments: [],

      statusIcon: {
        0: mdiTray,
        1: mdiTrayFull,
        2: mdiTrayAlert,
        3: mdiTrayRemove,
      },
      statusText: {
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

      mdiChevronLeft: mdiChevronLeft,
      mdiAlertCircle: mdiAlertCircle,
      mdiTray: mdiTray,
      mdiTrayFull: mdiTrayFull,
      mdiTrayAlert: mdiTrayAlert,
      mdiTrayRemove: mdiTrayRemove,
    };
  },
  beforeRouteEnter(to, from, next) {
    var projectID = to.params?.id;
    var issueID = to.params?.issueid;
    if (
      (from.params?.id != projectID || from.params?.issueid != issueID) &&
      projectID != undefined &&
      issueID != undefined
    ) {
      getIssue(projectID, issueID, (err, project) => {
        next((vm) => vm.setData(err, project));
      });
    } else {
      next();
    }
  },
  beforeRouteUpdate(to, from, next) {
    var projectID = to.params?.id;
    var issueID = to.params?.issueid;
    if (
      (from.params?.id != projectID || from.params?.issueid != issueID) &&
      projectID != undefined &&
      issueID != undefined
    ) {
      getIssue(projectID, issueID, (err, project) => {
        this.issue = undefined;
        this.issue_error = undefined;
        this.issue_loading = true;
        this.setData(err, project);
      });
    }
    next();
  },
  methods: {
    fetchIssue() {},
    setData(err, response) {
      if (err && err != response) {
        this.issue_error = err.toString();
      } else {
        this.issue_error = undefined;
        this.issue = response.issue;
      }
      this.issue_loading = false;
    },
    sendComment() {
      axios
        .post("/api/project/" + projectID + "/issue/" + issueID + "/comments", {
          transformResponse: [(data) => jsonBig.parse(data)],
        })
    },
  },
};
</script>

<style scoped>
.circle {
  min-width: 40px;
  max-width: 40px;
  min-height: 40px;
  max-height: 40px;
  border: darkgrey dashed 1px;
  border-radius: 100%;
}

.dot-right > span:not(:last-child)::after {
  content: "•";
  padding: 0 0 0 5px;
}

/* width */
pre::-webkit-scrollbar {
  width: 14px;
  height: 24px;
}

/* Handle */
pre::-webkit-scrollbar-thumb {
  height: 6px;
  border: 8px solid rgba(0, 0, 0, 0);
  background-color: #888;
  background-clip: padding-box;
  border-radius: 20px;
}
/* Handle on hover */
pre::-webkit-scrollbar-thumb:hover {
  background: #555;
  background-clip: padding-box;
}
pre::-webkit-scrollbar-button {
  width: 0;
  height: 0;
  display: none;
}
pre::-webkit-scrollbar-corner {
  background-color: transparent;
}
</style>
