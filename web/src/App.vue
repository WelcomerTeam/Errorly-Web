<template>
  <div id="app">
    <div v-cloak>
      <nav
        class="navbar navbar-expand-sm navbar-dark"
        style="background: #212529"
      >
        <div class="px-3 container-fluid">
          <router-link class="navbar-brand" to="/">
            <img
              width="64"
              height="64"
              alt="brand icon"
              src="./assets/brand.svg"
            />
            <span class="font-weight-bold text-light h4 pl-2">Errorly</span>
          </router-link>
          <button
            class="navbar-toggler"
            type="button"
            data-toggle="collapse"
            data-target="#navbarNavDropdown"
            aria-controls="navbarNavDropdown"
            aria-expanded="false"
            aria-label="Toggle navigation"
          >
            <span class="navbar-toggler-icon"></span>
          </button>
          <div
            class="collapse navbar-collapse justify-content-end"
            id="navbarNavDropdown"
          >
            <div v-if="this.$root.userLoading">
              <div class="spinner-border" role="status"></div>
            </div>
            <div v-else>
              <div v-if="!this.$root.userAuthenticated">
                <ul class="navbar-nav">
                  <li class="nav-item">
                    <a href="/login">
                      <button
                        class="btn btn-outline-light mr-2"
                        type="button"
                        aria-label="Login with Discord"
                      >
                        <svg-icon type="mdi" :path="mdiDiscord" />
                        Login with Discord
                      </button>
                    </a>
                  </li>
                </ul>
              </div>
              <div v-else>
                <ul class="navbar-nav">
                  <li class="nav-item dropdown">
                    <span
                      class="nav-link dropdown-toggle text-light"
                      id="projectDropdown"
                      role="button"
                      data-toggle="dropdown"
                      aria-expanded="false"
                    >
                      Select a Project
                    </span>
                    <ul
                      class="dropdown-menu dropdown-menu-right"
                      aria-labelledby="projectDropdown"
                    >
                      <li>
                        <router-link
                          to="/project/create"
                          class="dropdown-item text-success"
                          href="#"
                        >
                          <svg-icon type="mdi" :path="mdiTextBoxCheckOutline" />
                          New Project</router-link
                        >
                      </li>
                      <li>
                        <router-link
                          to="/projects"
                          class="dropdown-item"
                          href="#"
                        >
                          View All Projects</router-link
                        >
                      </li>
                      <li>
                        <hr class="dropdown-divider" />
                      </li>
                      <li
                        v-for="(project, index) in this.$root.userProjects"
                        v-bind:key="index"
                      >
                        <router-link
                          :to="'/project/' + project.id"
                          class="dropdown-item"
                          href="#"
                        >
                          {{ project.name }}</router-link
                        >
                      </li>
                    </ul>
                  </li>
                  <li class="nav-item dropdown">
                    <span
                      class="nav-link dropdown-toggle text-light"
                      id="userDropdown"
                      role="button"
                      data-toggle="dropdown"
                      aria-expanded="false"
                    >
                      <img
                        width="24"
                        height="24"
                        :src="this.$root.user.avatar"
                        alt="User profile picture"
                      />
                    </span>
                    <ul
                      class="dropdown-menu dropdown-menu-right"
                      aria-label="Profile"
                    >
                      <li>
                        <span class="dropdown-item">
                          Signed in as <b>{{ this.$root.user.name }}</b>
                        </span>
                      </li>
                      <li>
                        <hr class="dropdown-divider" />
                      </li>
                      <li>
                        <a class="dropdown-item" href="/logout">Sign Out</a>
                      </li>
                    </ul>
                  </li>
                </ul>
              </div>
            </div>
          </div>
        </div>
      </nav>
      <div>
        <div class="text-center py-5" v-if="this.$root.error">
          <error :message="this.$root.error" />
        </div>
        <div v-else>
          <router-view />
        </div>
      </div>
    </div>
  </div>
</template>

<style>
html {
  min-height: 100%;
}

#navbar {
  background: #212529;
}

[v-cloak] {
  visibility: hidden;
}
</style>

<script>
import SvgIcon from "@jamescoyle/vue-icon";
import Error from "@/components/Error.vue";
import { mdiDiscord, mdiTextBoxCheckOutline } from "@mdi/js";

export default {
  components: {
    SvgIcon,
    Error,
  },
  data() {
    return {
      mdiDiscord: mdiDiscord,
      mdiTextBoxCheckOutline: mdiTextBoxCheckOutline,
    };
  },
};
</script>
