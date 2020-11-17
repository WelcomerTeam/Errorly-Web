<template>
  <div aria-live="polite" aria-atomic="true">
    <div style="position: fixed; top: 8px; right: 8px">
      <div
        class="toast"
        :class="
          toast.class + (toast.onlyBody ? ' d-flex align-items-center' : '')
        "
        :id="toast.id"
        v-for="(toast, index) in toasts"
        v-bind:key="index"
      >
        <div class="toast-header" v-if="!toast.onlyBody">
          <img
            :src="toast.image"
            class="rounded mr-2"
            :alt="toast.alt"
            v-if="toast.image"
          />
          <strong class="mr-auto">{{ toast.title }}</strong>
          <small v-if="toast.displayTime"
            ><timeago
              :datetime="toast.since"
              :includeSeconds="true"
              :auto-update="60"
          /></small>
          <button
            type="button"
            :class="'btn-close ' + toast.closeClass"
            data-dismiss="toast"
            aria-label="Close"
            @click="hideToast(toast.id)"
          ></button>
        </div>
        <div class="toast-body">{{ toast.body }}</div>
        <button
          type="button"
          :class="'btn-close ml-auto mr-2 ' + toast.closeClass"
          data-dismiss="toast"
          aria-label="Close"
          v-if="toast.onlyBody"
          @click="hideToast(toast.id)"
        ></button>
      </div>
    </div>
  </div>
</template>

<script>
import { Toast } from "bootstrap";
export default {
  name: "ToastStack",
  data() {
    return {
      seed: Math.random().toString(36).substring(7),
      counter: 0,
      toasts: [],
      toastStore: {},
      visibleToasts: {},
    };
  },
  methods: {
    addToast(data) {
      for (var key in this.toasts) {
        var toast = this.toasts[key];
        if (this.visibleToasts[toast.id]) {
          var _toast = data;
          _toast.hidden = false;
          _toast.id = toast.id;
          _toast.since = new Date();
          this.toasts[key] = _toast;
          this.toastStore[_toast.id].show();
          return;
        }
      }
      this.createToast(data);
    },
    createToast(data) {
      this.counter++;
      var toastID = `toast-${this.seed}-${this.counter}`;

      data.since = new Date();
      data.id = toastID;
      this.toasts.push(data);

      this.$nextTick(function () {
        var timeout = 5000;
        var el = document.getElementById(toastID);
        var toast = new Toast(el, { delay: timeout });
        this.toastStore[toastID] = toast;

        toast.show();
        el.addEventListener("hide.bs.toast", () => {
          this.$set(this.visibleToasts, toastID, true);
        });
        el.addEventListener("show.bs.toast", () => {
          this.$set(this.visibleToasts, toastID, false);
        });
      });
    },
    hideToast(toastID) {
      this.toastStore[toastID].hide();
    },
  },
};
</script>
