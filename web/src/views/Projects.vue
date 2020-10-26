<template>
  <div class="container mt-5 text-center">
    <h3 class="pb-3 font-weight-bold">Select a project</h3>

    <div class="d-flex pb-3 border-bottom border-muted">
      <div
        class="input-group input-group-sm border border-secondary rounded mr-3"
      >
        <button class=" btn" type="button" id="search-submit">
          <svg
            class="text-secondary"
            style="width:21px;height:21px"
            viewBox="0 0 24 24"
          >
            <path
              fill="currentColor"
              d="M9.5,3A6.5,6.5 0 0,1 16,9.5C16,11.11 15.41,12.59 14.44,13.73L14.71,14H15.5L20.5,19L19,20.5L14,15.5V14.71L13.73,14.44C12.59,15.41 11.11,16 9.5,16A6.5,6.5 0 0,1 3,9.5A6.5,6.5 0 0,1 9.5,3M9.5,5C7,5 5,7 5,9.5C5,12 7,14 9.5,14C12,14 14,12 14,9.5C14,7 12,5 9.5,5Z"
            ></path>
          </svg>
        </button>
        <input
          type="text"
          class="form-control border-white"
          placeholder="Filter by..."
          aria-label="Search"
          aria-describedby="search-submit"
          @change="
            v => {
              $root.projectFilter = v.target.value;
            }
          "
        />
        <button class="btn" type="button" id="search-clear">
          <svg
            class="text-secondary"
            style="width:21px;height:21px"
            viewBox="0 0 24 24"
          >
            <path
              fill="currentColor"
              d="M12,20C7.59,20 4,16.41 4,12C4,7.59 7.59,4 12,4C16.41,4 20,7.59 20,12C20,16.41 16.41,20 12,20M12,2C6.47,2 2,6.47 2,12C2,17.53 6.47,22 12,22C17.53,22 22,17.53 22,12C22,6.47 17.53,2 12,2M14.59,8L12,10.59L9.41,8L8,9.41L10.59,12L8,14.59L9.41,16L12,13.41L14.59,16L16,14.59L13.41,12L16,9.41L14.59,8Z"
            ></path>
          </svg>
        </button>
      </div>
      <router-link to="/project/create" class="text-decoration-none">
        <button
          class="d-flex btn btn-outline-success"
          type="button"
          id="create-project"
        >
          <svg style="width:24px;height:24px" viewBox="0 0 24 24">
            <path
              fill="currentColor"
              d="M17,14H19V17H22V19H19V22H17V19H14V17H17V14M5,3H19C20.11,3 21,3.89 21,5V12.8C20.39,12.45 19.72,12.2 19,12.08V5H5V19H12.08C12.2,19.72 12.45,20.39 12.8,21H5C3.89,21 3,20.11 3,19V5C3,3.89 3.89,3 5,3M7,7H17V9H7V7M7,11H17V12.08C16.15,12.22 15.37,12.54 14.68,13H7V11M7,15H12V17H7V15Z"
            />
          </svg>
          New
        </button>
      </router-link>
    </div>

    <div
      v-if="this.$root.userProjects == 0"
      class="card text-left py-3 border-bottom border-muted border-top-0 border-left-0 border-right-0"
    >
      <h3>You do not have any projects</h3>
      <p>Join a project with a link from the owner or create a new project</p>
      <router-link to="/project/create" class="text-decoration-none">
        <button
          class="d-flex btn btn-outline-success m-auto"
          type="button"
          id="create-project"
        >
          <svg style="width:24px;height:24px" viewBox="0 0 24 24">
            <path
              fill="currentColor"
              d="M17,14H19V17H22V19H19V22H17V19H14V17H17V14M5,3H19C20.11,3 21,3.89 21,5V12.8C20.39,12.45 19.72,12.2 19,12.08V5H5V19H12.08C12.2,19.72 12.45,20.39 12.8,21H5C3.89,21 3,20.11 3,19V5C3,3.89 3.89,3 5,3M7,7H17V9H7V7M7,11H17V12.08C16.15,12.22 15.37,12.54 14.68,13H7V11M7,15H12V17H7V15Z"
            />
          </svg>
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
          <svg style="width:16px;height:16px" viewBox="0 0 24 24">
            <path
              fill="currentColor"
              d="M2 12H4V17H20V12H22V17A2 2 0 0 1 20 19H4A2 2 0 0 1 2 17M13 12H11V14H13M13 4H11V10H13Z"
            />
          </svg>
          {{ project.open_issues }}
        </span>
        <span class="card-text text-muted">
          <svg style="width:16px;height:16px" viewBox="0 0 24 24">
            <path
              fill="currentColor"
              d="M18 5H6V7H18M6 9H18V11H6M2 12H4V17H20V12H22V17A2 2 0 0 1 20 19H4A2 2 0 0 1 2 17M18 13H6V15H18Z"
            />
          </svg>
          {{ project.active_issues }}
        </span>
        <span class="card-text text-muted">
          <svg style="width:16px;height:16px" viewBox="0 0 24 24">
            <path
              fill="currentColor"
              d="M2 12H4V17H20V12H22V17A2 2 0 0 1 20 19H4A2 2 0 0 1 2 17Z"
            />
          </svg>
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
export default {
  name: "Projects",
  computed: {
    filterProjects() {
      if (this.projectFilter == "") {
        return this.userProjects;
      }
      var filter = this.$root.projectFilter.toLowerCase();
      return this.$root.userProjects.filter(object => {
        return object.name.toLowerCase().includes(filter);
      });
    }
  }
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
