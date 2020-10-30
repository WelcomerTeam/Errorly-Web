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

    <div v-if="this.issue_loading">
      Issue
    </div>
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
      {{ issue_error }}
      {{ JSON.stringify(issue) }}
    </div>
  </div>
</template>

<script>
import axios from "axios";
import SvgIcon from "@jamescoyle/vue-icon";
import { mdiChevronLeft } from "@mdi/js";
import JSONBig from "json-bigint";
import { mdiAlertCircle } from "@mdi/js";
var jsonBig = JSONBig({ storeAsString: true });

function getIssue(projectID, issueID, callback) {
  axios
    .get("/api/project/" + projectID + "/issues?id=" + issueID, {
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
  components: { SvgIcon },
  name: "ViewProjectIssue",
  data() {
    return {
      issue: undefined,
      issue_error: undefined,
      issue_loading: true,

      mdiChevronLeft: mdiChevronLeft,
      mdiAlertCircle: mdiAlertCircle
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
  },
};
</script>
