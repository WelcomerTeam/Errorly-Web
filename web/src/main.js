import Vue from "vue";
import router from "./router";

import axios from "axios";
import JSONBig from "json-bigint";
var jsonBig = JSONBig({ storeAsString: true });

import "bootstrap";
import "bootstrap/dist/css/bootstrap.min.css";

import App from "./App.vue";
import "./registerServiceWorker";

Vue.config.productionTip = false;

new Vue({
  router,
  render: (h) => h(App),
  data() {
    return {
      error: "",

      userLoading: true,
      userAuthenticated: false,
      userProjects: [],
      user: {},

      projectFilter: "",
    };
  },
  mounted() {
    this.fetchMe();
  },
  methods: {
    fetchMe() {
      axios
        .get("/api/me", { transformResponse: [(data) => jsonBig.parse(data)] })
        .then((result) => {
          var data = result.data;
          if (data.success) {
            this.userAuthenticated = data.data.authenticated;
            this.userProjects = data.data.projects;
            this.user = data.data.user;
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
        })
        .finally(() => {
          this.userLoading = false;
        });
    },
  },
}).$mount("#app");
