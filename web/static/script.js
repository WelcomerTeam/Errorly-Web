// Install any plugins
Vue.use(VueChartJs);

Vue.component("pie-chart", {
    extends: VueChartJs.Pie,
    mixins: [VueChartJs.mixins.reactiveProp],
    props: ['chartData', 'options'],
    mounted() {
        this.renderChart(this.chartData, this.options)
    },
})

const routes = [
    { path: "/", component: { template: "<div>a</div>" } },
    { path: "/b", component: { template: "<div>a</div>" } },
]

const router = new VueRouter({
    routes
})

// el: '#app'


const vue = new Vue({
    router
}).$mount("#app")