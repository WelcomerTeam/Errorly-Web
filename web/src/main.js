import axios from "axios";
import router from "./router";
import Vue from "vue";

import App from "./App.vue";

import "bootstrap";
import "bootstrap/dist/css/bootstrap.min.css";
import "./registerServiceWorker";

Vue.config.productionTip = false;

new Vue({
  router,
  render: h => h(App),
  data() {
    return {
      error: "",

      userLoading: true,
      userAuthenticated: false,
      userProjects: [],
      user: {},

      projectFilter: ""
    };
  },
  mounted() {
    this.fetchMe();
  },
  methods: {
    fetchMe() {
      axios
        .get("/api/me")
        .then(result => {
          var data = result.data;
          if (data.success) {
            this.userAuthenticated = data.data.authenticated;
            this.userProjects = data.data.projects;
            this.user = data.data.user;
          } else {
            this.error = data.error;
          }
        })
        .catch(error => {
          if (error.response?.data) {
            this.error = error.response.data.error || error.response.data;
          } else {
            this.error = error.toString();
          }
        })
        .finally(() => {
          this.userLoading = false;
        });
    }
  }
}).$mount("#app");