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

var vue;
var router;

axios.get("/api/dictionary").then(
    result => {
        routes = result.data.data.routes;

        router = new VueRouter({
            routes,
        })

        vue = new Vue({
            router,
            data() {
                return {
                    hasError: false,
                    error: "",

                    userLoading: true,
                    userAuthenticated: false,
                    userProjects: [],
                    user: {},

                    projectFilter: "",
                }
            },
            mounted() {
                this.fetchMe();
            },
            methods: {
                fetchMe() {
                    axios.get("/api/me")
                        .then(result => {
                            data = result.data;
                            if (data.success) {
                                this.userAuthenticated = data.data.authenticated;
                                this.userProjects = data.data.projects;
                                this.user = data.data.user;
                            } else {
                                this.hasError = true;
                                this.error = data.error;
                            }
                        })
                        .catch(error => { this.hasError = true; this.error = error; })
                        .finally(() => { this.userLoading = false })
                }
            },
            computed: {
                filterProjects() {
                    if (this.projectFilter == "") { return this.userProjects }
                    return this.userProjects.filter(object => {
                        return object.name.toLowerCase().includes(this.projectFilter.toLowerCase())
                    })
                }
            }
        }).$mount("#app")
    }
)
