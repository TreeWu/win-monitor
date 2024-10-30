<template>
  <a-menu
      mode="inline"
      :selectedKeys="[selectedDeviceId]"
      @click="handleClick"
      style="height: 100%; border-right: 0"
  >
    <a-menu-item v-for="device in devices" :key="device.hostID">
      <span>{{ device.customName || device.hostname }}</span>
      <a-button type="ghost" @click.stop="handleEdit(device)">
        <template #icon>
          <EditTwoTone/>
        </template>
      </a-button>
    </a-menu-item>
  </a-menu>
  <a-modal
      v-model:open="isModalVisible"
      title="编辑设备信息"
      @ok="submitForm"
      @cancel="isModalVisible = false"
  >
    <a-form :model="formData" layout="horizontal" :labelCol="{span:8}">
      <a-form-item label="主机名" name="hostname">
        <a-input disabled v-model:value="formData.hostname" placeholder="请输入主机名"/>
      </a-form-item>
      <a-form-item label="HostId" name="hostID">
        <a-input disabled v-model:value="formData.hostID" placeholder="请输入主机名"/>
      </a-form-item>
      <a-form-item label="自定义主机名" name="customName">
        <a-input v-model:value="formData.customName" placeholder="请输入自定义名称"/>
      </a-form-item>
      <a-divider/>
      <a-form-item label="监控开关" name="monitorEnable">
        <a-switch v-model:checked="formData.config.monitorEnable"/>
      </a-form-item>
      <a-form-item required label="监控数据最大保存数量" name="maxMonitorSize">
        <a-input type="number" v-model:value="formData.config.maxMonitorSize" placeholder="请输入监控数据最大保存数量"/>
      </a-form-item>
      <a-form-item required label="采集间隔时间(s)" name="monitorCollectInterval">
        <a-input type="number" v-model:value="formData.config.monitorCollectInterval" placeholder="请输入采集间隔时间(s)"/>
      </a-form-item>
      <a-divider/>
      <a-form-item label="截图开关" name="screenshotEnable">
        <a-switch v-model:checked="formData.config.screenshotEnable"/>
      </a-form-item>
      <a-form-item required label="监控上传间隔时间(s)" name="monitorUploadInterval">
        <a-input type="number" v-model:value="formData.config.monitorUploadInterval" placeholder="请输入监控上传间隔时间"/>
      </a-form-item>
      <a-form-item required label="截图间隔时间(s)" name="screenshotIntervalTime">
        <a-input type="number" v-model:value="formData.config.screenshotIntervalTime"/>
      </a-form-item>
      <a-form-item required label="(x)张截图后强制上传" name="screenshotUploadIntervalCount">
        <a-input type="number" v-model:value="formData.config.screenshotUploadIntervalCount"/>
      </a-form-item>
      <a-form-item required label="截图上传相似度" name="screenshotUploadMinDistance">
        <a-input type="number" v-model:value="formData.config.screenshotUploadMinDistance"/>
      </a-form-item>
      <a-form-item label="是否上传原图" name="screenshotUploadOriginImage">
        <a-switch v-model:checked="formData.config.screenshotUploadOriginImage"/>
      </a-form-item>

    </a-form>
  </a-modal>
</template>


<script>
import {EditTwoTone} from "@ant-design/icons-vue";
import {ref} from 'vue';
import {DefaultApi} from "../apis/default-api.ts"
import {message} from 'ant-design-vue'; // 导入 message 组件


export default {
  components: {
    EditTwoTone
  },

  setup: function () {
    const api = new DefaultApi()
    const isModalVisible = ref(false);
    const formData = ref({});
    return {
      api,
      formData,
      isModalVisible,
      labelCol: {span: 4},
      wrapperCol: {span: 14},
    }
  },
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
  emits: ['device-selected', 'device-edit'],
  methods: {
    handleEdit(device) {
      this.formData = {...device}
      this.isModalVisible = true
    },
    async submitForm() {
      const resp = await this.api.apiConsoleHostConfPost({
        hostID: this.formData.hostID,
        customName: this.formData.customName,
        config: {
          maxMonitorSize: this.formData.config.maxMonitorSize,
          monitorCollectInterval: this.formData.config.monitorCollectInterval,
          monitorEnable: this.formData.config.monitorEnable,
          monitorUploadInterval: this.formData.config.monitorUploadInterval,
          screenshotEnable: this.formData.config.screenshotEnable,
          screenshotIntervalTime: this.formData.config.screenshotIntervalTime,
          screenshotUploadIntervalCount: this.formData.config.screenshotUploadIntervalCount,
          screenshotUploadMinDistance: this.formData.config.screenshotUploadMinDistance,
          screenshotUploadOriginImage: this.formData.config.screenshotUploadOriginImage
        }
      })
      if (resp.data.code === 200) {
        this.isModalVisible = false;
        this.$emit('device-edit');
      } else {
        message.error('编辑失败，请检查输入或重试！')
      }
    },
    handleClick(e) {
      const selectedDevice = this.devices.find(device => device.hostID === e.key);
      this.$emit('device-selected', selectedDevice);
    }
  }
};
</script>
