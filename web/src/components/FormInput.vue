<template>
  <div class="form-check" v-if="type == 'checkbox'">
    <input
      class="form-check-input"
      type="checkbox"
      :id="id"
      :checked="value"
      v-on:change="updateValue($event.target.checked)"
      :disabled="disabled"
    />
    <label class="form-check-label" :for="id">{{ label }}</label>
  </div>
  <div v-else-if="type == 'text'">
    <label :for="id" class="col-sm-12 form-label">{{ label }}</label>
    <input
      type="text"
      class="form-control"
      :id="id"
      :value="value"
      v-on:change="updateValue($event.target.value)"
      :disabled="disabled"
      :placeholder="placeholder"
    />
  </div>
  <div v-else-if="type == 'list'">
    <label :for="id" class="col-sm-12 form-label">{{ label }}</label>
    <input
      type="text"
      class="form-control"
      :id="id"
      :value="value"
      placeholder="A,B,C"
      v-on:change="
        updateValue(Array.from(new Set($event.target.value.split(','))))
      "
      :disabled="disabled"
    />
  </div>
  <div v-else-if="type == 'number'">
    <label :for="id" class="col-sm-12 form-label">{{ label }}</label>
    <input
      type="number"
      class="form-control"
      :id="id"
      :value="value"
      v-on:change="updateValue(Number($event.target.value))"
      :disabled="disabled"
      :placeholder="placeholder"
    />
  </div>
  <div v-else-if="type == 'password'">
    <label :for="id" class="col-sm-12 form-label">{{ label }}</label>
    <div class="input-group">
      <input
        type="password"
        class="form-control"
        :id="id"
        autocomplete
        :value="value"
        v-on:change="updateValue($event.target.value)"
        :disabled="disabled"
        :placeholder="placeholder"
      />
      <button
        class="btn btn-outline-dark"
        type="button"
        v-on:click="copyFormInputPassword()"
      >
        Copy
      </button>
    </div>
  </div>
  <div v-else-if="type == 'select'">
    <label :for="id" class="col-sm-12 form-label">{{ label }}</label>
    <select
      class="form-select"
      :id="id"
      v-on:change="updateValue($event.target.value)"
      :disabled="disabled"
    >
      <option
        v-for="(item, index) in values"
        v-bind:key="index"
        selected="item == value"
      >
        {{ item }}
      </option>
    </select>
  </div>
  <span class="badge bg-warning text-dark" v-else
    >Invalid type "{{ type }}" for "{{ id }}"</span
  >
</template>

<script>
export default {
  props: ["type", "id", "label", "values", "value", "disabled", "placeholder"],
  methods: {
    copyFormInputPassword: function () {
      var elem = document.createElement("textarea");
      elem.value = this.$el.lastChild.firstChild.value;
      elem.type = "hidden";
      document.body.append(elem);

      elem.select();
      elem.setSelectionRange(0, 99999);

      document.execCommand("copy");
      elem.parentElement.removeChild(elem);
    },
    updateValue: function (value) {
      this.$emit("input", value);
    },
  },
};
</script>
