<template>
  <div class="container mt-5">
    <a
      class="text-dark mb-2 text-decoration-none h6 d-flex"
      @click="$router.go(-1)"
      href="#"
    >
      <svg-icon class="mr-3" type="mdi" :path="mdiChevronLeft" />
      Back to projects
    </a>

    <div class="pb-4 mb-4 border-bottom border-muted">
      <h3 class="pb-1">Create Project</h3>
      <span class="text-secondary"
        >A project allows you to contain all your application errors in a
        centralized place and have the ability to track who is doing what at any
        time.</span
      >
    </div>

    <div class="text-center py-5" v-if="this.error">
      <error :message="this.error" />
    </div>
    <div v-else>
      <form-input
        v-model="project.display_name"
        :type="'text'"
        :label="'Project Name'"
      />
      <p class="text-muted">
        Having a short, memorable and relevant project name is always important.
        Project names require at least 3 characters.
      </p>

      <form-input
        v-model="project.description"
        :type="'text'"
        :label="'Description (optional)'"
      />
      <p class="text-muted">
        Describing what you do is important when people stumble upon your
        project.
      </p>

      <form-input
        v-model="project.url"
        :type="'text'"
        :label="'URL (optional)'"
      />
      <p class="text-muted">
        Have a website? Link it so people can view it when they visit your
        project. Valid URLs only, please.
      </p>

      <form-input
        v-model="project.private"
        :type="'checkbox'"
        :label="'Private'"
      />
      <p class="text-muted">
        If your project is private, only contributors are able to view it the
        project.
      </p>

      <form-input
        v-model="project.limited"
        :type="'checkbox'"
        :label="'Limited'"
      />
      <p class="text-muted">
        If your project is limited, only contributors and integrations are able
        to create new issues.
      </p>

      <button
        type="button"
        class="btn btn-success"
        :disabled="!validRequest()"
        v-on:click="createProject(project)"
      >
        Create Project
      </button>
    </div>
  </div>
</template>

<script>
import axios from "axios";
import qs from "qs";
import Error from "@/components/Error.vue";
import FormInput from "@/components/FormInput.vue";
import SvgIcon from "@jamescoyle/vue-icon";
import JSONBig from "json-bigint";
import { mdiChevronLeft } from "@mdi/js";
var jsonBig = JSONBig({ storeAsString: true });

export default {
  components: {
    FormInput,
    SvgIcon,
    Error,
  },
  name: "CreateProject",
  data: function () {
    return {
      error: "",
      project: {
        display_name: "",
        background: "",
        description: "",
        url: "",
        private: false,
        limited: false,
      },
      mdiChevronLeft: mdiChevronLeft,
    };
  },
  methods: {
    validRequest() {
      if (this.project.display_name.trim().length < 3) {
        return false;
      }

      // Validate URL
      if (this.project.url) {
        let url;
        try {
          url = new URL(this.project.url);
        } catch (_) {
          return false;
        }
        if (!(url.protocol === "http:" || url.protocol === "https:")) {
          return false;
        }
      }

      return true;
    },
    createProject(project) {
      axios
        .post("/api/projects", qs.stringify(project), {
          transformResponse: [(data) => jsonBig.parse(data)],
          headers: {
            "content-type": "application/x-www-form-urlencoded;charset=utf-8",
          },
        })
        .then((result) => {
          var data = result.data;
          if (data.success) {
            this.$root.userProjects.unshift(data.data);
            this.error = "";
            this.$router.push("/project/" + data.data.id);
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
