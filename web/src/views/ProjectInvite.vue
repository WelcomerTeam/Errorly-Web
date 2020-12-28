<template>
  <div class="text-center py-5" v-if="this.error">
    <error :message="this.error" />
  </div>
  <div v-else-if="this.invite">
    <div class="container-xl p-4">
      <div class="text-center my-5">
        <svg-icon
          class="mb-2 text-muted"
          type="mdi"
          :width="64"
          :height="64"
          :path="mdiMailboxOpen"
        />
        <h3>You have been invited to <b>{{ project }}</b></h3>

        <button
          type="button"
          class="btn btn-dark mt-5"
          @click="acceptInvite()"
        >
          Accept Invite
        </button>
      </div>
    </div>
  </div>
</template>

<script>
import axios from "axios";
import JSONBig from "json-bigint";
import SvgIcon from "@jamescoyle/vue-icon";
import {
  mdiMailboxOpen,
} from "@mdi/js";
import Error from "@/components/Error.vue";
var jsonBig = JSONBig({ storeAsString: true });

function getInvite(projectID, inviteID, callback) {
  axios
    .get("/api/project/" + projectID + "/invite/" + inviteID, {
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
  components: {
    Error,
    SvgIcon,
  },
  name: "ProjectInvite",
  data() {
    var data = {
      mdiMailboxOpen,

      error: undefined,
      invite: undefined,
      project: undefined,
    };
    return data;
  },
  beforeRouteEnter(to, from, next) {
    getInvite(to.params.id, to.params.invite, (err, invite) => {
      next((vm) => vm.setData(err, invite));
    });
  },
  beforeRouteUpdate(to, from, next) {
    getInvite(to.params.id, to.params.invite, (err, invite) => {
      this.setData(err, invite);
      next();
    });
  },
  methods: {
    setData(err, response) {
      if (err && err != response) {
        this.error = err.toString();
      } else {
        this.error = undefined;
        this.invite = response.valid_invite;
        this.project = response.project_name;
      }
    },
    acceptInvite() {
      axios
        .post("/api/project/" + this.$route.params.id + "/invite/" + this.$route.params.invite, {
              transformResponse: [(data) => jsonBig.parse(data)],
            })
            .then((result) => {
              var data = result.data;
              if (data.success) {
                this.$bvToast.toast(
                  `You have joined this project. Redirecting you to the project...`,
                  {
                    title: "Invite accepted!",
                    appendToast: true,
                  }
                );
                this.$root.fetchMe();
                setTimeout(() => {
                  this.$router.push("/project/" + this.$route.params.id);
                }, 3000);
              } else {
                this.$bvToast.toast(data.error, {
                  title: "Failed to accept invite",
                  appendToast: true,
                });
              }
            })
            .catch((error) => {
              if (error.response?.data) {
                this.$bvToast.toast(
                  error.response.data.error || error.response.data,
                  {
                    title: "Failed to accept invite",
                    appendToast: true,
                  }
                );
              } else {
                this.$bvToast.toast(error.text || error.toString(), {
                  title: "Failed to accept invite",
                  appendToast: true,
                });
              }
            });
    },
  },
};
</script>
