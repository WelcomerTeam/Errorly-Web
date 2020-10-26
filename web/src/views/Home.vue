<template>
  <div class="text-center">
    <div class="bg-light py-5">
      <div class="col-sm-9 col-md-8 col-lg-6 col-xl-5 p-5 m-auto">
        <h1 class="display-4 font-weight-bold">
          Automate error collection and handling
        </h1>
        <p class="text-dark mt-4 mb-5">
          With Errorly, you can easily integrate both error logging and issue
          handling in one platform. Easily sort out issues and segregate what
          needs to be done and what has already been done.
        </p>
        <div v-if="!this.$root.userAuthenticated">
          <a href="/login">
            <button type="button" class="btn btn-outline-dark mx-2">
              <svg-icon type="mdi" :path="mdiDiscord" />
              Login with Discord
            </button>
          </a>
        </div>
        <div v-else>
          <h3>You're already logged in?</h3>
          <p class="text-dark mb-2">You're one step closer!</p>
          <router-link to="/projects">
            <button type="button" class="btn btn-outline-dark mx-2">
              <svg-icon type="mdi" :path="mdiTextBoxCheckOutline" />
              My Projects
            </button>
          </router-link>
        </div>
      </div>
    </div>
    <div style="overflow: hidden;">
      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1440 320">
        <path
          fill="#f8f9fa"
          fill-opacity="1"
          d="M0,224L80,202.7C160,181,320,139,480,138.7C640,139,800,181,960,176C1120,171,1280,117,1360,90.7L1440,64L1440,0L1360,0C1280,0,1120,0,960,0C800,0,640,0,480,0C320,0,160,0,80,0L0,0Z"
        ></path>
      </svg>
    </div>
  </div>
</template>

<script>
import SvgIcon from "@jamescoyle/vue-icon";

import { mdiDiscord, mdiTextBoxCheckOutline } from "@mdi/js";

export default {
  name: "Home",
  components: {
    SvgIcon
  },
  data() {
    return {
      mdiDiscord: mdiDiscord,
      mdiTextBoxCheckOutline: mdiTextBoxCheckOutline
    };
  }
};
</script>
