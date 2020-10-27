<template>
  <div class="text-center py-5" v-if="this.error">
    <error :message="this.error" />
  </div>
  <div v-else-if="this.project">
    <div class="banner-image">
      <a
        href="https://www.pexels.com/photo/snowcap-mountains-during-sunset-808031/"
      >
        <span class="image-credit">
          <svg-icon type="mdi" :height="16" :path="mdiCamera" />
          <span>Michal Pech</span>
          <svg-icon type="mdi" :height="16" :path="mdiMapMarker" />
          <span>Breitenwang, Austria</span>
        </span>
      </a>
      <div>
        <div>
          <div>
            <span>{{ this.project.settings.display_name }}</span>
          </div>
          <div></div>
          <div>
            <div>
              <pie-chart
                :height="100"
                :width="100"
                :options="{
                  animation: { animateScale: true },
                  tooltips: { display: false },
                  legend: { display: false },
                  title: { display: false },
                  cutoutPercentage: 50,
                  reactive: false,
                  aspectRatio: 1,
                }"
                :chart-data="{
                  datasets: [
                    {
                      data: [
                        this.project.open_issues,
                        this.project.active_issues,
                        this.project.resolved_issues,
                      ],
                      backgroundColor: ['#0d6efd', '#6c757d', '#28a745'],
                      borderColor: ['#0256d4', '#555c62', '#1f7f35'],
                    },
                  ],
                  labels: ['Open', 'Active', 'Resolved'],
                }"
              >
              </pie-chart>
            </div>
            <div>
              <span>
                <svg-icon type="mdi" :height="20" :path="mdiTrayAlert" />
                {{ this.project.open_issues }} Open</span
              >
              <span>
                <svg-icon type="mdi" :height="20" :path="mdiTrayFull" />
                {{ this.project.active_issues }} Active</span
              >
              <span>
                <svg-icon type="mdi" :height="20" :path="mdiTray" />
                {{ this.project.resolved_issues }} Resolved</span
              >
            </div>
          </div>
        </div>
      </div>
    </div>
    <div class="container-xl p-4">
      <ul class="nav nav-tabs">
        <li class="nav-item">
          <router-link
            :to="'/project/' + this.$route.params.id + '/'"
            class="nav-link text-body"
            :class="{
              active:
                this.$route.path.endsWith(this.project.id + '/') ||
                this.$route.path.endsWith(this.project.id),
            }"
            aria-current="page"
            href="#"
          >
            <svg-icon type="mdi" :height="20" :path="mdiInformationOutline" />
            Overview
          </router-link>
        </li>
        <li class="nav-item">
          <router-link
            :to="'/project/' + this.$route.params.id + '/issues'"
            class="nav-link text-body"
            :class="{ active: this.$route.path.endsWith('/issues') }"
            href="#"
          >
            <svg-icon type="mdi" :height="20" :path="mdiTrayAlert" />
            Issues
            <span class="badge rounded-pill bg-secondary">
              {{ this.project.open_issues + this.project.active_issues }}
            </span>
          </router-link>
        </li>
        <li class="nav-item">
          <router-link
            :to="'/project/' + this.$route.params.id + '/settings'"
            class="nav-link text-body"
            :class="{ active: this.$route.path.endsWith('/settings') }"
            href="#"
          >
            <svg-icon type="mdi" :height="20" :path="mdiCogOutline" />
            Settings
          </router-link>
        </li>
      </ul>
      <!-- <pre>{{ this.project }}</pre> -->
      <router-view class="p-4" />
    </div>
  </div>
</template>

<script>
import axios from "axios";
import qs from "qs";
import PieChart from "@/components/PieChart.vue";
import SvgIcon from "@jamescoyle/vue-icon";
import {
  mdiCamera,
  mdiMapMarker,
  mdiTray,
  mdiTrayFull,
  mdiTrayAlert,
  mdiCogOutline,
  mdiInformationOutline,
} from "@mdi/js";
import Error from "@/components/Error.vue";

function getProject(projectID, callback) {
  axios
    .get("/api/project/" + projectID)
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
// test: {{ this.$route.query.name }}

export default {
  components: {
    PieChart,
    SvgIcon,
    Error,
  },
  name: "ProjectOverview",
  data() {
    return {
      total_issues: 0,
      page: this.$route.query.page || 0,
      issue_query: this.$route.query.q || "",
      issues: {},
      contributors: {},
      contributors_loaded: false,

      executing: false,

      issue_error: undefined,
      issues_loading: true,

      page_limit: 25,
      error: undefined,
      project: undefined,
      mdiCamera: mdiCamera,
      mdiMapMarker: mdiMapMarker,
      mdiTray: mdiTray,
      mdiTrayFull: mdiTrayFull,
      mdiTrayAlert: mdiTrayAlert,
      mdiCogOutline: mdiCogOutline,
      mdiInformationOutline: mdiInformationOutline,
    };
  },
  beforeRouteEnter(to, from, next) {
    var projectID = to.params?.id;
    if (from.params?.id != projectID && projectID != undefined) {
      getProject(to.params.id, (err, project) => {
        next((vm) => vm.setData(err, project));
      });
    } else {
      next();
    }
  },
  beforeRouteUpdate(to, from, next) {
    var projectID = to.params?.id;
    if (from.params?.id != projectID && projectID != undefined) {
      getProject(to.params.id, (err, project) => {
        // Reset Project State
        this.total_issues = 0;
        this.page = this.$route.query.page || 0;
        this.issue_query = this.$route.query.q || "";
        this.issues = {};
        // preserve the contributors cache,
        // we still still keep loaded false in the event there are members that need to be cached which regardless will be set to true in the future by loadIssues
        this.contributors_loaded = false;
        this.executing = false;
        this.issue_error = undefined;
        this.issues_loading = true;
        this.setData(err, project);
      });
    }
    next();
  },
  methods: {
    selectAllIssues(toggle) {
      for (var issueID in this.issues) {
        var issue = this.issues[issueID];
        issue.checked = toggle;
        this.$set(this.issues, issueID, issue);
      }
    },
    getCheckedIssues() {
      return Object.values(this.issues)
        .filter((issue) => {
          return issue.checked;
        })
        .map((issue) => {
          return issue.id;
        });
    },
    executeTask(query) {
      this.executing = true;
      this.issue_error = undefined;
      axios
        .post(
          "/api/project/" + this.$route.params.id + "/execute",
          qs.stringify(query),
          {
            headers: {
              "content-type": "application/x-www-form-urlencoded;charset=utf-8",
            },
          }
        )
        .then((result) => {
          var data = result.data;
          if (data.success) {
            data.data.issues.forEach((issue) => {
              issue["checked"] = this.issues[issue.id]
                ? this.issues[issue.id]["checked"]
                : false;
              this.$set(this.issues, issue.id, issue);
            });
            this.project = data.data.project;
          }
        })
        .catch((error) => {
          if (error.response?.data) {
            this.issue_error = error.response.data.error || error.response.data;
          } else {
            this.issue_error = error.toString();
          }
        })
        .finally(() => {
          this.executing = false;
        });
    },
    starIssue(id, star) {
      this.issues[id].starred = star;
      axios
        .post(
          "/api/project/" + this.$route.params.id + "/execute",
          qs.stringify({
            action: "star",
            issues: qs.stringify([id]),
            starring: star,
          }),
          {
            headers: {
              "content-type": "application/x-www-form-urlencoded;charset=utf-8",
            },
          }
        )
        .then((result) => {
          var data = result.data;
          if (data.success) {
            data.data.issues.forEach((issue) => {
              issue["checked"] = this.issues[issue.id]
                ? this.issues[issue.id]["checked"]
                : false;
              this.$set(this.issues, issue.id, issue);
            });
          }
        })
        .catch((error) => {
          if (error.response?.data) {
            this.issue_error = error.response.data.error || error.response.data;
          } else {
            this.issue_error = error.toString();
          }
        });
    },
    fetchIssues() {
      var query = {
        page: this.page,
      };
      if (this.issue_query != "") {
        query["q"] = this.issue_query;
      }
      var path = "/project/" + this.$route.params.id + "/issues";
      if (
        path != this.$route.path ||
        this.$route.query.page != query["page"] ||
        this.$route.query.q != query["q"]
      ) {
        // Persists
        this.$router.push(
          {
            path: path,
            query: query,
          },
          () => {
            this.loadIssues(query);
          }
        );
      }
      //  else {
      //   this.loadIssues(query);
      // }
    },
    loadIssues(query) {
      this.issues_loading = true;
      this.issues = {};
      this.issue_error = undefined;
      axios
        .get("/api/project/" + this.$route.params.id + "/issues", {
          params: query,
        })
        .then((result) => {
          var data = result.data;
          if (data.success) {
            var userQuery = [];
            this.total_issues = data.data.total_issues;
            this.page = data.data.page;
            data.data.issues.forEach((issue) => {
              issue.checked = false;
              if (
                issue.assignee_id != 0 &&
                !(issue.assignee_id in this.contributors) &&
                !userQuery.includes(issue.assignee_id)
              ) {
                userQuery.push(issue.assignee_id);
              }
              if (
                issue.created_by_id != 0 &&
                !(issue.created_by_id in this.contributors) &&
                !userQuery.includes(issue.created_by_id)
              ) {
                userQuery.push(issue.created_by_id);
              }
              this.$set(this.issues, issue.id, issue);
            });
            if (userQuery.length > 0) {
              // fetch users
              axios
                .get("/api/project/" + this.$route.params.id + "/lazy", {
                  params: {
                    q: qs.stringify(userQuery),
                  },
                })
                .then((result) => {
                  var data = result.data;
                  if (data.success) {
                    Object.values(data.data.users).forEach((user) => {
                      // We will use $set as this overcomes a Vue limitation
                      // where adding new properties to an object will not
                      // trigger changes.
                      this.$set(this.contributors, user.id, user);
                      // this.contributors[user.id] = user;
                    });
                  } else {
                    this.issue_error = data.error;
                  }
                })
                .catch((error) => {
                  if (error.response?.data) {
                    this.issue_error =
                      error.response.data.error || error.response.data;
                  } else {
                    this.issue_error = error.toString();
                  }
                })
                .finally(() => {
                  this.contributors_loaded = true;
                });
            } else {
              this.contributors_loaded = true;
            }
          } else {
            this.issue_error = data.error;
          }
        })
        .catch((error) => {
          if (error.response?.data) {
            this.issue_error = error.response.data.error || error.response.data;
          } else {
            this.issue_error = error.toString();
          }
        })
        .finally(() => {
          this.issues_loading = false;
        });
    },
    getUsername(id) {
      if (!this.contributors_loaded) {
        return "...";
      }
      return this.contributors[id] ? this.contributors[id].name : undefined;
    },
    getIntegration(id) {
      return this.contributors[id] ? this.contributors[id].integration : false;
    },
    setData(err, response) {
      if (err && err != response) {
        this.error = err.toString();
      } else {
        this.error = undefined;
        this.project = response.project;
        this.contributors = response.contributors || {};

        // this.issues = response.issues.map(issue => {
        //   issue.checked = false;
        //   return issue;
        // });
        // this.page = response.page;
        // this.total_issues = response.total_issues;
      }
    },
  },
};
</script>

<style scoped>
.image-credit {
  margin: 4px;
  border-radius: 5px;
  font-size: 10px;
  position: absolute;
  background: rgba(0, 0, 0, 0.85);
  color: #cccccc;
  padding: 3px 8px;
}
.banner-image {
  background-image: url(/img/banner.png);
  background-color: rgba(0, 0, 0, 0.8);
  background-repeat: no-repeat;
  background-size: cover;
  background-position: center;
}
.banner-image > div {
  width: 100%;
  height: 100%;
  background: linear-gradient(
    270deg,
    rgba(0, 0, 0, 0) 0%,
    rgba(0, 0, 0, 0.7) 33.3%,
    rgba(0, 0, 0, 0.7) 66.6%,
    rgba(0, 0, 0, 0) 100%
  );
  display: inline-flex;
  padding: 3rem 1rem;
}
.banner-image > div > div {
  margin: auto;
  display: flex;
  align-items: center;
  width: 100%;
}
.banner-image > div > div > div:nth-of-type(1) {
  width: 45%;
  text-align: right;
  color: var(--bs-light);
  font-size: min(50px, 6vw);
  font-weight: 600;
}
.banner-image > div > div > div:nth-of-type(1) > span {
  text-overflow: ellipsis;
  overflow: hidden;
  display: block;
}
.banner-image > div > div > div:nth-of-type(2) {
  background: #a3a3a3;
  width: 1px;
  height: 100%;
  min-height: 12rem;
  margin: 0 5%;
}
.banner-image > div > div > div:nth-of-type(3) {
  width: 45%;
  display: flex;
}
.banner-image > div > div > div:nth-of-type(3) > div:nth-of-type(1) {
  width: 100px;
}
.banner-image > div > div > div:nth-of-type(3) > div:nth-of-type(2) {
  color: white;
  margin: auto 0 auto 1rem;
  font-size: 14px;
  flex-direction: column;
  display: flex;
}
</style>
