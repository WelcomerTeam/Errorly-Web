<template>
  <div class="container mt-2 p-4">
    <a
      class="text-dark mb-2 text-decoration-none h6 d-flex"
      @click="$router.go(-1)"
      href="#"
    >
      <svg-icon class="mr-3" type="mdi" :path="mdiChevronLeft" />
      Back to projects
    </a>

    <div class="pb mb-4 border-bottom border-muted">
      <h3 class="pb-1">Create Issue</h3>
    </div>

    <div class="text-center py-5" v-if="this.error">
      <error :message="this.error" />
    </div>
    <div v-else>
      <form-input
        v-model="issue.error"
        type="text"
        placeholder="Title (required)"
        class="mb-2"
      />
      <form-input
        v-model="issue.description"
        type="area"
        placeholder="Description"
        class="mb-4"
      />

      <form-input
        v-model="issue.function"
        type="text"
        placeholder="createIssue()"
        label="Function (required)"
      />
      <p class="text-muted">
        The function is a place holder for identifying the function the error is
        in
      </p>

      <form-input
        v-model="issue.checkpoint"
        type="text"
        placeholder="internal/api.go:53"
        label="Checkpoint"
      />
      <p class="text-muted">
        The checkpoint allows you to identify the exact file and line number the
        error occurred on. If you don't know the line number, you do not have to
        include it.
      </p>

      <form-input
        v-model="issue.traceback"
        type="area"
        placeholder="Traceback"
        label="Traceback"
      />
      <p class="text-muted">
        Enter a detailed issue traceback if you have one.
      </p>

      <form-input
        v-model="issue.lock_comments"
        type="checkbox"
        label="Lock comments"
      />
      <p class="text-muted">If enabled, comments are locked when created</p>

      <form-input
        v-model="issue.assigned"
        type="select"
        label="Assign to"
        :values="contributors"
      />
      <p class="text-muted">
        If necessary, you can assign yourself or someone else to the issue when
        making
      </p>

      <p class="text-muted">
        When creating an issue, if it finds an issue with the same title and
        function, it will increment the occurrences of the already made issue
        instead of creating a new one.
      </p>

      <button
        type="button"
        class="btn btn-success"
        :disabled="$parent.project.settings.archived || !validRequest()"
        v-on:click="createIssue(issue)"
      >
        Create Issue
      </button>
    </div>
  </div>
</template>

<script>
import { mdiChevronLeft } from "@mdi/js";
import axios from "axios";
import JSONBig from "json-bigint";
import qs from "qs";

import Error from "@/components/Error.vue";
import FormInput from "@/components/FormInput.vue";
import SvgIcon from "@jamescoyle/vue-icon";

var jsonBig = JSONBig({ storeAsString: true });

export default {
  components: { FormInput, SvgIcon, Error },
  name: "CreateProjectIssue",
  data() {
    return {
      error: "",
      issue: {
        error: "",
        description: "",
        function: "",
        checkpoint: "",
        traceback: "",
        lock_comments: false,
        assigned: 0,
      },
      contributors: {},
      mdiChevronLeft: mdiChevronLeft,
    };
  },
  beforeRouteEnter(to, from, next) {
    next((vm) => vm.fetchContributors());
  },
  methods: {
    fetchContributors() {
      this.contributors = {
        0: "Nobody",
      };
      axios
        .get("/api/project/" + this.$route.params.id + "/contributors", {
          transformResponse: [(data) => jsonBig.parse(data)],
        })
        .then((result) => {
          var data = result.data;
          if (data.success) {
            Object.entries(data.data.users).forEach(([id, user]) => {
              this.$set(this.contributors, id, user.name);
            });
          } else {
            this.error = data.error;
          }
        })
        .catch((error) => {
          if (error.response?.data) {
            this.error = error.response.data.error || error.response.data;
          } else {
            this.error = error.text || error.toString();
          }
        });
    },
    validRequest() {
      if (this.issue.error.trim() == "") {
        return false;
      }
      if (this.issue.function.trim() == "") {
        return false;
      }
      return true;
    },
    createIssue(issue) {
      axios
        .post(
          "/api/project/" + this.$route.params.id + "/issues",
          qs.stringify(issue),
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
            this.$set(this.$parent.issues, data.data.issue.id, data.data.issue);

            if (data.data.new) {
              if (data.data.issue.type == 0) {
                this.$parent.project.active_issues++;
              }
              if (data.data.issue.type == 1) {
                this.$parent.project.open_issues++;
              }
              if (data.data.issue.type == 3) {
                this.$parent.project.resolved_issues++;
              }
            }

            this.$router.push(
              "/project/" +
                this.$route.params.id +
                "/issue/" +
                data.data.issue.id
            );
          } else {
            this.error = data.error;
          }
        })
        .catch((error) => {
          if (error.response?.data) {
            this.error = error.response.data.error || error.response.data;
          } else {
            this.error = error.text || error.toString();
          }
        });
    },
  },
};
</script>
