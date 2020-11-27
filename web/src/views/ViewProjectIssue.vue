<template>
  <div>
    <a
      class="text-dark mb-2 text-decoration-none h6 d-flex"
      @click="$router.go(-1)"
      href="#"
    >
      <svg-icon class="mr-3" type="mdi" :path="mdiChevronLeft" />
      Back to issues
    </a>

    <div
      v-if="this.issue_loading"
      class="spinner-border mx-auto d-flex my-4"
      role="status"
    >
      <span class="visually-hidden">Loading...</span>
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
      <div class="p-4 border-bottom border-muted">
        <h2 class="display-6 d-flex flex-wrap">
          <h5 class="my-auto mr-1">
            <span
              class="badge rounded-pill"
              :style="{
                background: statusBackground[$parent.issues[issue_id].type],
              }"
              aria-label="Error type"
            >
              <svg-icon
                type="mdi"
                width="20"
                height="20"
                :path="statusIcon[$parent.issues[issue_id].type]"
              />
              {{ statusText[$parent.issues[issue_id].type] }}
            </span>
          </h5>
          <span class="mr-2">{{ $parent.issues[issue_id].error }}</span>
          <span
            class="text-muted mr-2 h4 font-weight-normal my-auto"
            aria-label="Error function"
            >{{ $parent.issues[issue_id].function }}</span
          >
          <span
            class="text-muted h5 font-weight-normal my-auto"
            aria-label="Error location"
            >{{ $parent.issues[issue_id].checkpoint }}</span
          >
        </h2>
        <span class="dot-right">
          <span class="text-muted">
            <span class="text-body">{{
              $parent.getUsername($parent.issues[issue_id].created_by_id) ||
              "ghost"
            }}</span>
            <span
              class="badge rounded-pill bg-primary ml-1"
              v-if="
                $parent.getIntegration($parent.issues[issue_id].created_by_id)
              "
              >Integration</span
            >
            opened this issue
            <timeago
              :datetime="$parent.issues[issue_id].created_at"
              :auto-update="60"
              :includeSeconds="true"
            />
          </span>
          <span class="pl-1 text-muted">
            Last modified
            <timeago
              class="text-body"
              :datetime="$parent.issues[issue_id].last_modified"
              :auto-update="60"
              :includeSeconds="true"
            />
          </span>
          <span class="pl-1 text-muted">
            <span class="text-body">{{
              $parent.issues[issue_id].occurrences
            }}</span>
            occurrences
          </span>
          <span class="pl-1 text-muted">
            Assigned to
            <span class="text-body">
              {{
                $parent.getUsername($parent.issues[issue_id].assignee_id) ||
                "ghost"
              }}
            </span>
          </span>
        </span>
      </div>

      <div class="p-4 border-bottom border-muted d-flex flex-column">
        <div class="d-flex mt-4">
          <img
            width="40"
            height="40"
            :src="$parent.getAvatar($parent.issues[issue_id].created_by_id)"
          />
          <div class="card ml-3" style="align-items: stretch; width: 100%">
            <div class="card-header text-black-50">
              <b class="text-dark">{{
                $parent.getUsername($parent.issues[issue_id].created_by_id) ||
                "ghost"
              }}</b>
              <span
                class="badge rounded-pill bg-primary ml-1"
                v-if="
                  $parent.getIntegration($parent.issues[issue_id].created_by_id)
                "
                >Integration</span
              >
              created this issue
              <timeago
                :datetime="$parent.issues[issue_id].created_at"
                :auto-update="60"
                :includeSeconds="true"
              />
            </div>
            <div class="card-body">
              {{ $parent.issues[issue_id].description }}
              <pre
                class="bg-light mt-3 mb-0 rounded-lg p-3 border border-muted"
                v-if="$parent.issues[issue_id].traceback"
                >{{ $parent.issues[issue_id].traceback }}
              </pre>
            </div>
          </div>
        </div>
        <div v-for="(comment, index) in comments" v-bind:key="index">
          <div v-if="comment.type == 0" class="d-flex mt-4">
            <img
              width="40"
              height="40"
              :src="$parent.getAvatar(comment.created_by_id)"
            />
            <div class="card ml-3" style="align-items: stretch; width: 100%">
              <div class="card-header text-black-50 comment-text">
                <b class="text-dark">{{
                  $parent.getUsername(comment.created_by_id) || "ghost"
                }}</b>
                commented
                <timeago
                  :datetime="comment.created_at"
                  :auto-update="60"
                  :includeSeconds="true"
                />
              </div>
              <div class="card-body" style="white-space: pre-wrap">
                {{ comment.content }}
              </div>
            </div>
          </div>
          <div v-else-if="comment.type == 1" class="d-flex ml-5 mt-4">
            <svg-icon
              width="30"
              height="30"
              type="mdi"
              :style="{ color: statusBackground[comment.issue_marked] }"
              :path="statusIcon[comment.issue_marked]"
              class="ml-4 mr-2"
            />
            <div
              class="align-self-middle ml-2 d-flex"
              style="align-items: stretch; width: 100%"
            >
              <div class="text-dark my-auto comment-text">
                Issue marked
                <b>{{ statusText[comment.issue_marked] }}</b>
                <timeago
                  :datetime="comment.created_at"
                  :auto-update="60"
                  :includeSeconds="true"
                />
                by
                <b>{{
                  $parent.getUsername(comment.created_by_id) || "ghost"
                }}</b>
              </div>
            </div>
          </div>
          <div v-else-if="comment.type == 2" class="d-flex ml-5 mt-4">
            <svg-icon
              width="30"
              height="30"
              type="mdi"
              :path="comment.comments_opened ? mdiLockOpenVariant : mdiLock"
              class="ml-4 mr-2 text-dark"
            />
            <div
              class="align-self-middle ml-2 d-flex"
              style="align-items: stretch; width: 100%"
            >
              <div class="text-dark my-auto comment-text">
                Comments were
                <b>{{ comment.comments_opened ? "" : "un" }}locked</b>
                <timeago
                  :datetime="comment.created_at"
                  :auto-update="60"
                  :includeSeconds="true"
                />
                by
                <b>{{
                  $parent.getUsername(comment.created_by_id) || "ghost"
                }}</b>
              </div>
            </div>
          </div>
        </div>

        <div
          class="btn btn-light border border-muted mx-auto my-4 px-5"
          v-if="!comments_end"
          @click="
            comments_page++;
            fetchComments(comments_page);
          "
        >
          Load more comments
        </div>
      </div>

      <!-- 
      <div class="d-flex ml-5 mt-4">
        <svg-icon
          width="40"
          height="40"
          type="mdi"
          :path="mdiTrayRemove"
          class="ml-4 mr-2"
        />
        <div
          class="align-self-middle ml-2 d-flex"
          style="align-items: stretch; width: 100%"
        >
          <div class="text-dark my-auto">
            Issue marked <b>invalid</b> 3 days ago by <b>ImRock</b>
          </div>
        </div>
      </div>

      <div class="d-flex mt-4">
        <img
          width="40"
          height="40"
          src="https://cdn.discordapp.com/avatars/143090142360371200/a_70444022ea3e5d73dd00d59c5578b07e.png"
        />
        <div class="card ml-3" style="align-items: stretch; width: 100%">
          <div class="card-header text-black-50">
            <b class="text-dark">ImRock</b> commented 3 days ago
          </div>
          <div class="card-body">This is a message</div>
        </div>
      </div>
      -->

      <div class="p-4">
        <div v-if="!$root.userAuthenticated">
          <div
            class="border border-muted rounded-sm py-5 text-muted text-center bg-muted card-header mb-2"
          >
            You must be logged in to comment
          </div>
          <button
            v-if="!$root.userAuthenticated"
            type="button"
            class="btn btn-success mr-2"
            :disabled="true"
          >
            Comment
          </button>
        </div>
        <div v-else-if="$parent.issues[issue_id].comments_locked">
          <div
            class="border border-muted rounded-sm py-5 text-muted text-center bg-muted card-header mb-2"
          >
            Comments are locked on this issue
          </div>
          <button
            v-if="!$root.userAuthenticated"
            type="button"
            class="btn btn-success mr-2"
            :disabled="true"
          >
            Comment
          </button>
        </div>
        <div v-else>
          <form-input
            v-model="comment"
            :type="'area'"
            :placeholder="'Leave a comment'"
            class="mb-2"
          />
          <div class="d-flex">
            <button
              v-if="$parent.elevated || !$parent.project.settings.limited"
              type="button"
              class="btn btn-success mr-2"
              :disabled="
                $parent.project.settings.archived ||
                $parent.issues[issue_id].comments_locked ||
                !validRequest()
              "
              @click="sendComment(comment)"
            >
              Comment
            </button>
            <button
              v-else
              type="button"
              class="btn btn-success mr-2"
              :disabled="true"
            >
              You cannot comment
            </button>
          </div>
        </div>
        <div
          class="pt-2"
          v-if="
            $parent.elevated ||
            $parent.issues[issue.id].created_by == $root.user.id
          "
        >
          <button class="btn btn-dark" @click="deleteIssue()">
            Delete Issue
          </button>
        </div>
        <div class="pt-2" v-if="$parent.elevated && !$parent.project.settings.archived">
          <div class="btn-group dropright mr-2">
            <button
              class="btn btn-secondary btn-sm dropdown-toggle"
              type="button"
              data-toggle="dropdown"
              aria-expanded="false"
              aria-label="Modify issue"
            >
              <svg-icon type="mdi" :height="16" :path="icons[this.marked]" />
              {{ text[this.marked] }}
            </button>
            <ul class="dropdown-menu">
              <li>
                <a class="dropdown-item" v-on:click.prevent="marked = 'none'">
                  <svg-icon type="mdi" :height="16" :path="mdiDotsHorizontal" />
                  Select Action</a
                >
              </li>
              <li>
                <hr class="dropdown-divider" />
              </li>
              <li>
                <a
                  class="dropdown-item"
                  v-on:click.prevent="marked = 'resolved'"
                >
                  <svg-icon type="mdi" :height="16" :path="mdiTray" />
                  Mark Resolved</a
                >
              </li>
              <li>
                <a class="dropdown-item" v-on:click.prevent="marked = 'active'">
                  <svg-icon type="mdi" :height="16" :path="mdiTrayFull" />
                  Mark Active</a
                >
              </li>
              <li>
                <a class="dropdown-item" v-on:click.prevent="marked = 'open'">
                  <svg-icon type="mdi" :height="16" :path="mdiTrayAlert" />
                  Mark Open</a
                >
              </li>
              <li>
                <a
                  class="dropdown-item"
                  v-on:click.prevent="marked = 'invalid'"
                >
                  <svg-icon type="mdi" :height="16" :path="mdiTrayRemove" />
                  Mark Invalid</a
                >
              </li>
              <li>
                <hr class="dropdown-divider" />
              </li>
              <li>
                <a class="dropdown-item" v-on:click.prevent="marked = 'lock'">
                  <svg-icon type="mdi" :height="16" :path="mdiLock" />
                  Lock Comments</a
                >
              </li>
              <li>
                <a class="dropdown-item" v-on:click.prevent="marked = 'unlock'">
                  <svg-icon
                    type="mdi"
                    :height="16"
                    :path="mdiLockOpenVariant"
                  />
                  Unlock Comments</a
                >
              </li>
            </ul>
          </div>
          <button
            class="btn btn-outline-secondary btn-sm"
            aria-label="Execute actions"
            @click="
              $parent.execute(marked, [$route.params.issueid]);
              fetchComments(comments_page);
            "
            :disabled="
              marked == 'none' ||
              marked == statusText[$parent.issues[issue_id].type].toLowerCase()
            "
            v-if="$parent.elevated"
          >
            <div
              v-if="$parent.executing"
              style="width: 0.6rem; height: 0.6rem; vertical-align: sub"
              class="spinner-border"
              role="status"
            >
              <span class="visually-hidden">Loading...</span>
            </div>
            <svg-icon
              v-else
              type="mdi"
              style="width: 1rem; height: 1rem; vertical-align: sub"
              :path="mdiPlay"
            />
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import axios from "axios";
import qs from "qs";
import FormInput from "@/components/FormInput.vue";
import SvgIcon from "@jamescoyle/vue-icon";
import JSONBig from "json-bigint";
import {
  mdiPlay,
  mdiChevronLeft,
  mdiAlertCircle,
  mdiTray,
  mdiTrayFull,
  mdiTrayAlert,
  mdiTrayRemove,
  mdiLock,
  mdiLockOpenVariant,
  mdiDotsHorizontal,
  mdiAccountPlus,
  mdiAccountRemove,
} from "@mdi/js";
var jsonBig = JSONBig({ storeAsString: true });

function getIssue(projectID, issueID, callback) {
  axios
    .get("/api/project/" + projectID + "/issue/" + issueID, {
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
  components: { SvgIcon, FormInput },
  name: "ViewProjectIssue",
  data() {
    return {
      issue_error: undefined,
      issue_loading: true,

      comment: "",
      comments: [],
      comments_page: 0,
      comments_end: false,

      statusIcon: {
        0: mdiTrayFull,
        1: mdiTrayAlert,
        2: mdiTrayRemove,
        3: mdiTray,
      },
      statusText: {
        0: "Active",
        1: "Open",
        2: "Invalid",
        3: "Resolved",
      },
      statusBackground: {
        0: "#0d6efd",
        1: "#dc3545",
        2: "#6c757d",
        3: "#28a745",
      },

      marked: "none",

      icons: {
        none: mdiDotsHorizontal,
        resolved: mdiTray,
        active: mdiTrayFull,
        open: mdiTrayAlert,
        invalid: mdiTrayRemove,
        lock: mdiLock,
        unlock: mdiLockOpenVariant,
      },

      text: {
        none: "Select Action",
        resolved: "Mark resolved",
        active: "Mark active",
        open: "Mark open",
        invalid: "Mark invalid",
        lock: "Lock comments",
        unlock: "Unlock comments",
      },

      mdiPlay: mdiPlay,
      mdiDotsHorizontal: mdiDotsHorizontal,
      mdiAccountPlus: mdiAccountPlus,
      mdiAccountRemove: mdiAccountRemove,
      mdiChevronLeft: mdiChevronLeft,
      mdiAlertCircle: mdiAlertCircle,
      mdiTray: mdiTray,
      mdiTrayFull: mdiTrayFull,
      mdiTrayAlert: mdiTrayAlert,
      mdiTrayRemove: mdiTrayRemove,
      mdiLock: mdiLock,
      mdiLockOpenVariant: mdiLockOpenVariant,
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
        next((vm) => {
          vm.issue_id = issueID;
          vm.setData(err, project);
        });
      });
    } else {
      next();
    }
  },
  beforeRouteUpdate(to, from, next) {
    var projectID = to.params?.id;
    var issueID = to.params?.issueid;
    this.issue_id = issueID;
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
    deleteIssue() {
      if (
        window.confirm(
          "Are you sure you want to delete this issue. This cannot be undone."
        )
      ) {
        axios
          .post(
            "/api/project/" +
              this.$route.params.id +
              "/issue/" +
              this.$route.params.issueid +
              "/delete",
            {
              transformResponse: [(data) => jsonBig.parse(data)],
            }
          )
          .then((result) => {
            var data = result.data;
            if (data.success) {
              this.$bvToast.toast(
                `Issue was deleted. Redirecting you to issues...`,
                {
                  title: "Successfully Deleted",
                  appendToast: true,
                }
              );
              setTimeout(() => {
                this.$router.push(
                  "/project/" + this.$route.params.id + "/issues"
                );
              }, 3000);
            }
          });
      }
    },
    removeDuplicateComments() {
      var _comments = Array.from(
        new Set(Object.values(this.comments).map((c) => c.id))
      ).map((id) => {
        for (const key in this.comments) {
          if (this.comments[key].id == id) {
            return this.comments[key];
          }
        }
      });
      return _comments;
    },
    fetchComments(page) {
      axios
        .get(
          "/api/project/" +
            this.$route.params.id +
            "/issue/" +
            this.$route.params.issueid +
            "/comments?page=" +
            page,
          {
            transformResponse: [(data) => jsonBig.parse(data)],
          }
        )
        .then((result) => {
          var data = result.data;
          if (data.success) {
            var userQuery = [];
            this.comments_page = data.data.page;
            data.data.comments.forEach((comment) => {
              comment.issue_marked = comment.issue_marked | 0;
              this.comments.push(comment);
              if (
                comment.created_by_id != 0 &&
                !(comment.created_by_id in this.$parent.contributors) &&
                !userQuery.includes(comment.created_by_id)
              ) {
                userQuery.push(comment.created_by_id);
              }
            });

            if (data.data.end) {
              this.comments_end = true;
            }

            this.comments = this.removeDuplicateComments();
            this.lazyLoad(userQuery);
          }
        });
    },
    lazyLoad(userQuery) {
      if (userQuery.length > 0) {
        // Fetch contributors from a passed user query
        axios
          .get("/api/project/" + this.$route.params.id + "/lazy", {
            transformResponse: [(data) => jsonBig.parse(data)],
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
                this.$set(this.$parent.contributors, user.id, user);
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
              this.issue_error = error.text || error.toString();
            }
          })
          .finally(() => {
            this.$parent.contributors_loaded = true;
          });
      } else {
        this.$parent.contributors_loaded = true;
      }
    },
    setData(err, response) {
      if (err && err != response) {
        this.issue_error = err.toString();
        this.issue_loading = false;
        return;
      } else {
        this.issue_error = undefined;
        var issue_id = this.$route.params.issueid;
        this.$set(this.$parent.issues, issue_id, response.issue);
        this.issue = this.$parent.issues[issue_id];
        this.issue_loading = false;
      }
      this.fetchComments(0);

      var userQuery = [];
      if (
        response.issue.created_by_id != 0 &&
        !(response.issue.created_by_id in this.$parent.contributors) &&
        !userQuery.includes(response.issue.created_by_id)
      ) {
        userQuery.push(response.issue.created_by_id);
      }
      if (
        response.issue.assignee_id != 0 &&
        !(response.issue.assignee_id in this.$parent.contributors) &&
        !userQuery.includes(response.issue.assignee_id)
      ) {
        userQuery.push(response.issue.assignee_id);
      }
      this.$parent.lazyLoad(userQuery);
    },
    validRequest() {
      if (this.comment.trim() == "") {
        return false;
      }
      return true;
    },
    sendComment(comment) {
      axios
        .post(
          "/api/project/" +
            this.$route.params.id +
            "/issue/" +
            this.$route.params.issueid +
            "/comments",
          qs.stringify({
            content: comment,
          }),
          {
            transformResponse: [(data) => jsonBig.parse(data)],
          }
        )
        .then((result) => {
          var data = result.data;
          if (data.success) {
            if (this.comments_end) {
              // We will only add the comment to the end
              // when they are at the end of the message list
              this.comments.push(data.data);
            }
            // this.comments.push(data.data);
          } else {
            this.issue_error = data.error;
          }
        })
        .catch((error) => {
          if (error.response?.data) {
            this.issue_error = error.response.data.error || error.response.data;
          } else {
            this.issue_error = error.text || error.toString();
          }
        });
    },
  },
};
</script>

<style scoped>
.comment-text > * {
  padding: 0 2px;
}
.circle {
  min-width: 40px;
  max-width: 40px;
  min-height: 40px;
  max-height: 40px;
  border: darkgrey dashed 1px;
  border-radius: 100%;
}

.dot-right > span:not(:last-child)::after {
  content: "•";
  padding: 0 0 0 5px;
}

/* width */
pre::-webkit-scrollbar {
  width: 14px;
  height: 24px;
}

/* Handle */
pre::-webkit-scrollbar-thumb {
  height: 6px;
  border: 8px solid rgba(0, 0, 0, 0);
  background-color: #888;
  background-clip: padding-box;
  border-radius: 20px;
}
/* Handle on hover */
pre::-webkit-scrollbar-thumb:hover {
  background: #555;
  background-clip: padding-box;
}
pre::-webkit-scrollbar-button {
  width: 0;
  height: 0;
  display: none;
}
pre::-webkit-scrollbar-corner {
  background-color: transparent;
}
</style>
