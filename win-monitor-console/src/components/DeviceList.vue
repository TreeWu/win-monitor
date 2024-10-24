<template>
  <a-menu
      mode="inline"
      :selectedKeys="[selectedDeviceId]"
      @click="handleClick"
      style="height: 100%; border-right: 0"
  >
    <a-menu-item v-for="device in devices" :key="device.hostID">
      {{ device.hostname }}
    </a-menu-item>
  </a-menu>
</template>

<script>
export default {
  props: {
    devices: {
      type: Array,
      required: true
    },
    selectedDeviceId: {
      type: String,
      default: null
    }
  },
  emits: ['device-selected'],
  methods: {
    handleClick(e) {
      const selectedDevice = this.devices.find(device => device.hostID === e.key);
      this.$emit('device-selected', selectedDevice);
    }
  }
};
</script>
