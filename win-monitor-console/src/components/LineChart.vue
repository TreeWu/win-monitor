<template>
  <v-chart :option="chartOptions" style="width: 100%; height: 250px;"/>
</template>

<script>
import {defineComponent, ref, watch} from 'vue';
import {use} from 'echarts/core';
import {CanvasRenderer} from 'echarts/renderers';
import {LineChart} from 'echarts/charts';
import {GridComponent, LegendComponent, TitleComponent, TooltipComponent} from 'echarts/components';
import VChart from 'vue-echarts';
import dayjs from 'dayjs';

use([CanvasRenderer, LineChart, GridComponent, TooltipComponent, TitleComponent, LegendComponent]);

export default defineComponent({
  components: {
    VChart
  },
  props: {
    data: {
      type: Array,
      required: true
    },
    seriesName: {
      type: String,
      required: true
    }
  },
  setup(props) {
    const chartOptions = ref({
      xAxis: {type: 'category', data: []},
      yAxis: {type: 'value'},
      series: []
    });

    watch(() => props.data, (ndata) => {
      if (ndata.length) {
        const newData = ndata.reduce((acc, cur) => {
          if (acc.length > 1) {
            if (acc[acc.length - 1].boot_time !== cur.boot_time) {
              let timeA = acc[acc.length - 1].time
              // 每隔10分钟，补充一个时间点
              while (timeA < cur.time) {
                acc.push({time: timeA, per: 0, total: 0, used: 0, free: 0, boot_time: 0})
                timeA = timeA + 60000
              }
            }
          }
          acc.push(cur)
          return acc
        }, []);
        // 每隔10分钟，补充一个时间点
        let lastTime = newData[newData.length - 1].time
        for (let i = lastTime; i < dayjs().valueOf(); i = i + 60000) {
          newData.push({time: i, per: 0, total: 0, used: 0, free: 0, boot_time: 0})
        }
        newData.forEach(v => {
          v.per = (v.per).toFixed(2);
          if (v.type === "mem" || v.type === "disk") {
            v.total = (v.total / 1024 / 1024 / 1024).toFixed(2);
            v.used = (v.used / 1024 / 1024 / 1024).toFixed(2);
            v.free = (v.free / 1024 / 1024 / 1024).toFixed(2);
          }
        })
        chartOptions.value = {
          autoresize: true,
          tooltip: {
            trigger: 'axis'
          },
          legend: {
            data: ['per', 'total', 'used', 'free',]
          },
          xAxis: {
            type: 'category',
            boundaryGap: true,
            data: newData.map(item => dayjs(item.time).format('YYYY-MM-DD HH:mm'))
          },
          yAxis: {},
          series: [{
            name: "per",
            type: 'line',
            data: newData.map(item => item.per)
          }, {
            name: "total",
            type: 'line',
            data: newData.map(item => item.total)
          }, {
            name: "used",
            type: 'line',
            data: newData.map(item => item.used)
          },
            {
              name: "free",
              type: 'line',
              data: newData.map(item => item.free)
            }], dataZoom: [
            {
              type: 'slider',
              start: 0,
              end: 100,
            },
            {
              type: 'inside',
              start: 0,
              end: 100,
            },
          ],
        };
      }
    }, {immediate: true});

    return {chartOptions};
  }
});
</script>
