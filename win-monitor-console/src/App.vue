<template>
  <a-layout style="height: 100vh;">
    <a-layout-sider width="200" style="background: #fff;">
      <device-list
          :devices="devices"
          :selectedDeviceId="selectedDeviceId"
          @device-edit="deviceEdit"
          @device-selected="handleDeviceSelected"
      />
    </a-layout-sider>
    <a-layout-content style="padding: 24px;">
      <a-collapse v-for="(data, index) in groupedDeviceData" :key="index" style="margin-bottom: 5px;">
        <a-collapse-panel :key="index" :header="`${data.type}` ">
          <line-chart :data="data.values" :seriesName="data.type"/>
        </a-collapse-panel>
      </a-collapse>
      <a-collapse v-if="deviceScreenshot && deviceScreenshot.cur">
        <a-collapse-panel :header="'截图,差异度:['+deviceScreenshot.distance+'],截图时间:'+ dayjs(deviceScreenshot.captureTime).format('YYYY-MM-DD HH:mm')">
          <a-image alt="最新" style="margin: 5px;height: 400px;width: 600px" :src="'data:image/gif;base64,'+deviceScreenshot.cur"/>
          <a-image alt="历史" style="margin: 5px;height: 400px;width: 600px" :src="'data:image/gif;base64,'+deviceScreenshot.pre"/>
        </a-collapse-panel>
      </a-collapse>
    </a-layout-content>
  </a-layout>
</template>

<script>
import {onMounted, ref} from 'vue';
import DeviceList from './components/DeviceList.vue';
import LineChart from './components/LineChart.vue';
import dayjs from "dayjs";
import {DefaultApi} from "@/apis/default-api.ts";

export default {
  computed: {
    dayjs() {
      return dayjs
    }
  },
  components: {
    DeviceList,
    LineChart
  },
  setup: function () {
    const devices = ref([]);
    const groupedDeviceData = ref([]);
    const deviceScreenshot = ref(null)
    const selectedDevice = ref(null);
    const selectedDeviceId = ref(null);
    const api = new DefaultApi()


    const fetchDevices = async () => {
      try {
        const response = await api.apiConsoleHostGet()
        devices.value = response.data.data
        if (devices.value.length > 0) {
          selectedDevice.value = devices.value[0];
          selectedDeviceId.value = devices.value[0].hostID
        }
      } catch (error) {
        console.error('Failed to fetch devices:', error);
      }
    };

    const fetchDeviceData = async () => {
      if (selectedDevice.value !== null) {
        try {
          const response = await api.apiConsoleHostHostIdGet(selectedDevice.value.hostID)
          const data = response.data.data.monitors;
          deviceScreenshot.value = response.data.data.screenshot;
          const groupedData = data.reduce((acc, item) => {
            let {type, name} = item;
            if (name) {
              type = name
            }
            if (!acc[type]) {
              acc[type] = [];
            }
            acc[type].push(item);
            return acc;
          }, {});

          groupedDeviceData.value = Object.keys(groupedData).map(type => ({
            type,
            values: groupedData[type].sort((x, y) => x.time > y.time ? 1 : -1)
          }));
        } catch (error) {
          console.error('Failed to fetch device data:', error);
        }
      }
      setTimeout(fetchDeviceData, 5000)
    };
    const deviceEdit = function () {
      fetchDevices();
    }

    const handleDeviceSelected = async (device) => {
      selectedDevice.value = device;
      selectedDeviceId.value = device.hostID
    };
    onMounted(() => {
      fetchDevices();
      fetchDeviceData()
    });
    return {
      deviceEdit,
      devices,
      groupedDeviceData,
      handleDeviceSelected,
      selectedDeviceId,
      deviceScreenshot
    };
  }
};
</script>

<style>
body, html, #app {
  height: 100%;
  margin: 0;
}
</style>
