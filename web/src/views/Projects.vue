<template>
  <div class="container mt-5 text-center">
    <h3 class="pb-3 font-weight-bold">Select a project</h3>

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
          placeholder="Filter by..."
          aria-label="Search"
          aria-describedby="search-submit"
          v-model="$root.projectFilter"
        />
        <button
          class="btn"
          type="button"
          id="search-clear"
          aria-label="Empty query"
          @click="$root.projectFilter = ''"
        >
          <svg-icon
            type="mdi"
            :width="21"
            :height="21"
            :path="mdiCloseCircleOutline"
          />
        </button>
      </div>
      <router-link to="/project/create" class="text-decoration-none">
        <button
          class="d-flex btn btn-outline-success"
          type="button"
          id="create-project"
        >
          <svg-icon
            type="mdi"
            :width="24"
            :heights="24"
            :path="mdiTextBoxCheckOutline"
          />
          New
        </button>
      </router-link>
    </div>

    <div v-if="!this.$root.userAuthenticated" class="text-center py-3">
      <div class="card-body">
        <h3>You are not logged in!</h3>
        <p class="text-muted">
          Create an account to create your own projects and contribute to
          others. Its free
        </p>
        <a href="/login">
          <button
            class="btn btn-outline-dark mr-2"
            type="button"
            aria-label="Login with Discord"
          >
            <svg-icon type="mdi" :path="mdiDiscord" />
            Login with Discord
          </button>
        </a>
      </div>
    </div>
    <div
      v-else-if="this.$root.userProjects == 0"
      class="card text-left text-center py-3 border-bottom border-muted border-top-0 border-left-0 border-right-0"
    >
      <h3>You do not have any projects</h3>
      <p>Join a project with a link from the owner or create a new project</p>
      <router-link to="/project/create" class="text-decoration-none">
        <button
          class="d-flex btn btn-outline-success m-auto"
          type="button"
          id="create-project"
        >
          <svg-icon
            type="mdi"
            :width="24"
            :height="24"
            :path="mdiTextBoxCheckOutline"
          />
          New
        </button>
      </router-link>
    </div>
    <div
      v-else
      v-for="(project, index) in this.filterProjects"
      v-bind:key="index"
      class="card text-left py-3 border-bottom border-muted border-top-0 border-left-0 border-right-0"
    >
      <div class="card-body">
        <router-link
          :to="'/project/' + project.id"
          class="card-title font-weight-bold h3 text-decoration-none"
        >
          {{ project.name }}</router-link
        >
        <p class="card-title text-dark">
          {{ project.description || "No description set" }}
        </p>
        <span class="card-text text-muted">
          <svg-icon type="mdi" :width="16" :height="16" :path="mdiTrayAlert" />
          {{ project.open_issues }}
        </span>
        <span class="card-text text-muted">
          <svg-icon type="mdi" :width="16" :height="16" :path="mdiTrayFull" />
          {{ project.active_issues }}
        </span>
        <span class="card-text text-muted">
          <svg-icon type="mdi" :width="16" :height="16" :path="mdiTray" />
          {{ project.resolved_issues }}
        </span>
        <span v-if="project.private" class="badge rounded-pill bg-secondary"
          >Private</span
        >
        <span
          v-if="project.archived"
          class="badge rounded-pill bg-warning text-dark"
          >Archived</span
        >
      </div>
    </div>
  </div>
</template>

<script>
import {
  mdiDiscord,
  mdiMagnify,
  mdiCloseCircleOutline,
  mdiTray,
  mdiTrayFull,
  mdiTrayAlert,
  mdiTextBoxCheckOutline,
} from "@mdi/js";
import SvgIcon from "@jamescoyle/vue-icon";
export default {
  components: {
    SvgIcon,
  },
  name: "Projects",
  data() {
    return {
      mdiDiscord: mdiDiscord,
      mdiMagnify: mdiMagnify,
      mdiCloseCircleOutline: mdiCloseCircleOutline,
      mdiTray: mdiTray,
      mdiTrayFull: mdiTrayFull,
      mdiTrayAlert: mdiTrayAlert,
      mdiTextBoxCheckOutline: mdiTextBoxCheckOutline,
    };
  },
  computed: {
    filterProjects() {
      if (this.projectFilter == "") {
        return this.userProjects;
      }
      var filter = this.$root.projectFilter.toLowerCase();
      return this.$root.userProjects.filter((object) => {
        return object.name.toLowerCase().includes(filter);
      });
    },
  },
};
</script>

<style>
.card-text {
  margin: 0 1rem 0 0;
}
.card-text > svg {
  margin-right: 2px;
}
</style>
